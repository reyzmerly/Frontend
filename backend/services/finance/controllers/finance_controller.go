package controllers

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/kwizera/university-invoices/shared/repositories"
	"github.com/kwizera/university-invoices/shared/utils"
	"github.com/signintech/gopdf"
)

type FinanceController struct {
	invoiceRepo   *repositories.InvoiceRepository
	paymentRepo   *repositories.PaymentRepository
	studentRepo   *repositories.StudentRepository
	programmeRepo *repositories.ProgrammeRepository
	semesterRepo  *repositories.SemesterRepository
}

type rgb [3]uint8

var (
	brandMaroon = rgb{123, 24, 40}
	brandBlue   = rgb{37, 99, 235}
	brandGold   = rgb{212, 160, 23}
	textDark    = rgb{31, 41, 55}
	textGray    = rgb{107, 114, 128}
	lineGray    = rgb{229, 231, 235}
	white       = rgb{255, 255, 255}
)

const logoURL = "https://res.cloudinary.com/danzeqybu/image/upload/v1793788865/Screenshot_From_2026-11-04_12-39-17_ka7vft.png"

func NewFinanceController(db *sql.DB) *FinanceController {
	return &FinanceController{
		invoiceRepo:   repositories.NewInvoiceRepository(db),
		paymentRepo:   repositories.NewPaymentRepository(db),
		studentRepo:   repositories.NewStudentRepository(db),
		programmeRepo: repositories.NewProgrammeRepository(db),
		semesterRepo:  repositories.NewSemesterRepository(db),
	}
}

func (c *FinanceController) GetDashboard(w http.ResponseWriter, r *http.Request) {
	invoices, _, _ := c.invoiceRepo.List(0, "", "", "", "", "", "", 1, 1000)
	payments, _, _ := c.paymentRepo.List(0, "", 1, 1000)

	if len(invoices) == 0 && len(payments) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	totalInvoiced := int64(0)
	totalCollected := int64(0)
	overdueAmount := int64(0)
	pendingPayments := 0
	failedPayments := 0

	for _, inv := range invoices {
		totalInvoiced += inv["total_amount"].(int64)
		if inv["status"].(string) == "Overdue" {
			overdueAmount += inv["balance"].(int64)
		}
	}

	for _, pay := range payments {
		if pay["status"].(string) == "Paid" {
			totalCollected += pay["amount"].(int64)
		} else if pay["status"].(string) == "Pending" {
			pendingPayments++
		} else if pay["status"].(string) == "Failed" {
			failedPayments++
		}
	}

	outstandingBalance := totalInvoiced - totalCollected

	dashboard := map[string]interface{}{
		"total_invoiced":      totalInvoiced,
		"total_collected":     totalCollected,
		"outstanding_balance": outstandingBalance,
		"overdue_amount":      overdueAmount,
		"pending_payments":    pendingPayments,
		"failed_payments":     failedPayments,
		"total_students":      0,
		"active_invoices":     len(invoices),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

func (c *FinanceController) GetReportSummary(w http.ResponseWriter, r *http.Request) {
	invoices, _, _ := c.invoiceRepo.List(0, "", "", "", "", "", "", 1, 10000)
	payments, _, _ := c.paymentRepo.List(0, "", 1, 10000)

	if len(invoices) == 0 && len(payments) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	totalInvoiced := int64(0)
	totalCollected := int64(0)

	for _, inv := range invoices {
		totalInvoiced += inv["total_amount"].(int64)
	}

	for _, pay := range payments {
		if pay["status"].(string) == "Paid" {
			totalCollected += pay["amount"].(int64)
		}
	}

	outstandingBalance := totalInvoiced - totalCollected
	collectionRate := float64(0)
	if totalInvoiced > 0 {
		collectionRate = float64(totalCollected) / float64(totalInvoiced) * 100
	}

	summary := map[string]interface{}{
		"total_invoiced":      totalInvoiced,
		"total_collected":     totalCollected,
		"outstanding_balance": outstandingBalance,
		"collection_rate":     collectionRate,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (c *FinanceController) GetAccountsReport(w http.ResponseWriter, r *http.Request) {
	invoices, _, _ := c.invoiceRepo.List(0, "", "", "", "", "", "", 1, 10000)

	if len(invoices) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	accounts := make(map[uint]map[string]interface{})
	for _, inv := range invoices {
		student, ok := inv["student"].(map[string]interface{})
		if !ok {
			continue
		}
		studentID, ok := student["id"].(uint)
		if !ok {
			continue
		}

		if _, exists := accounts[studentID]; !exists {
			accounts[studentID] = map[string]interface{}{
				"student_id":          studentID,
				"student_number":      student["student_number"],
				"student_name":        student["full_name"],
				"programme":           "",
				"total_invoiced":      int64(0),
				"total_paid":          int64(0),
				"outstanding_balance": int64(0),
				"invoice_count":       0,
			}
		}

		totalAmount := int64(0)
		if ta, ok := inv["total_amount"].(int64); ok {
			totalAmount = ta
		}
		paidAmount := int64(0)
		if pa, ok := inv["paid_amount"].(int64); ok {
			paidAmount = pa
		}
		balance := totalAmount - paidAmount

		accounts[studentID]["total_invoiced"] = accounts[studentID]["total_invoiced"].(int64) + totalAmount
		accounts[studentID]["total_paid"] = accounts[studentID]["total_paid"].(int64) + paidAmount
		accounts[studentID]["outstanding_balance"] = accounts[studentID]["outstanding_balance"].(int64) + balance
		accounts[studentID]["invoice_count"] = accounts[studentID]["invoice_count"].(int) + 1
	}

	var result []map[string]interface{}
	for _, acc := range accounts {
		result = append(result, acc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (c *FinanceController) SearchStudents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}

	students, err := c.studentRepo.SearchStudents(query, limit)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to fetch students"}`, http.StatusInternalServerError)
		return
	}

	if len(students) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

func (c *FinanceController) ListProgrammes(w http.ResponseWriter, r *http.Request) {
	programmes, err := c.programmeRepo.ListAll()
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to fetch programmes"}`, http.StatusInternalServerError)
		return
	}

	if len(programmes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programmes)
}

func (c *FinanceController) ListSemesters(w http.ResponseWriter, r *http.Request) {
	semesters, err := c.semesterRepo.ListAll()
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to fetch semesters"}`, http.StatusInternalServerError)
		return
	}

	if len(semesters) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(semesters)
}

func (c *FinanceController) ExportReport(w http.ResponseWriter, r *http.Request) {
	reportType := r.URL.Query().Get("type")
	format := r.URL.Query().Get("format")
	if reportType == "" {
		reportType = "accounts"
	}
	if format == "" {
		format = "csv"
	}

	// Get accounts data
	invoices, _, _ := c.invoiceRepo.List(0, "", "", "", "", "", "", 1, 10000)

	accounts := make(map[uint]map[string]interface{})
	for _, inv := range invoices {
		student, ok := inv["student"].(map[string]interface{})
		if !ok {
			continue
		}
		studentID, ok := student["id"].(uint)
		if !ok {
			continue
		}

		if _, exists := accounts[studentID]; !exists {
			accounts[studentID] = map[string]interface{}{
				"student_id":          studentID,
				"student_number":      student["student_number"],
				"student_name":        student["full_name"],
				"total_invoiced":      int64(0),
				"total_paid":          int64(0),
				"outstanding_balance": int64(0),
				"invoice_count":       0,
			}
		}

		totalAmount := int64(0)
		if ta, ok := inv["total_amount"].(int64); ok {
			totalAmount = ta
		}
		paidAmount := int64(0)
		if pa, ok := inv["paid_amount"].(int64); ok {
			paidAmount = pa
		}
		balance := totalAmount - paidAmount

		accounts[studentID]["total_invoiced"] = accounts[studentID]["total_invoiced"].(int64) + totalAmount
		accounts[studentID]["total_paid"] = accounts[studentID]["total_paid"].(int64) + paidAmount
		accounts[studentID]["outstanding_balance"] = accounts[studentID]["outstanding_balance"].(int64) + balance
		accounts[studentID]["invoice_count"] = accounts[studentID]["invoice_count"].(int) + 1
	}

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename="+reportType+"_report.csv")

		writer := csv.NewWriter(w)
		defer writer.Flush()

		// Write CSV header
		writer.Write([]string{"Student ID", "Student Number", "Student Name", "Total Invoiced", "Total Paid", "Outstanding Balance", "Invoice Count"})

		// Write CSV rows
		for _, acc := range accounts {
			record := []string{
				strconv.FormatUint(uint64(acc["student_id"].(uint)), 10),
				acc["student_number"].(string),
				acc["student_name"].(string),
				strconv.FormatInt(acc["total_invoiced"].(int64), 10),
				strconv.FormatInt(acc["total_paid"].(int64), 10),
				strconv.FormatInt(acc["outstanding_balance"].(int64), 10),
				strconv.Itoa(acc["invoice_count"].(int)),
			}
			writer.Write(record)
		}
	} else if format == "pdf" {
		c.generateAccountsPDF(w, accounts, reportType)
	} else {
		http.Error(w, `{"code":"BAD_REQUEST","message":"Invalid format. Use 'csv' or 'pdf'"}`, http.StatusBadRequest)
		return
	}
}

func (c *FinanceController) generateAccountsPDF(w http.ResponseWriter, accounts map[uint]map[string]interface{}, reportType string) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4Landscape})

	if err := utils.AddSystemPDFFonts(&pdf); err != nil {
		log.Printf("finance report PDF: %v", err)
		http.Error(w, `{"code":"INTERNAL_SERVER_ERROR","message":"Unable to generate finance report PDF"}`, http.StatusInternalServerError)
		return
	}

	const pageW, margin = 841.89, 40.0 // A4 Landscape width
	contentW := pageW - 2*margin
	const headerHeight = 120.0
	const footerHeight = 50.0
	const rowH = 30.0
	const maxRowsPerPage = 18

	fill := func(col rgb, x, y, w, h float64) {
		pdf.SetFillColor(col[0], col[1], col[2])
		pdf.RectFromUpperLeftWithStyle(x, y, w, h, "F")
	}
	text := func(font string, size float64, col rgb, x, y float64, s string) {
		pdf.SetFont(font, "", size)
		pdf.SetTextColor(col[0], col[1], col[2])
		pdf.SetXY(x, y)
		pdf.Cell(nil, s)
	}
	textAligned := func(font string, size float64, col rgb, x, y, w float64, align int, s string) {
		pdf.SetFont(font, "", size)
		pdf.SetTextColor(col[0], col[1], col[2])
		pdf.SetXY(x, y)
		pdf.CellWithOption(&gopdf.Rect{W: w, H: size + 4}, s, gopdf.CellOption{Align: align})
	}
	hline := func(col rgb, width, y float64) {
		pdf.SetLineWidth(width)
		pdf.SetStrokeColor(col[0], col[1], col[2])
		pdf.Line(margin, y, pageW-margin, y)
	}

	// Convert accounts map to slice for pagination
	var accountList []map[string]interface{}
	for _, acc := range accounts {
		accountList = append(accountList, acc)
	}

	// Calculate number of pages needed
	totalPages := (len(accountList) + maxRowsPerPage - 1) / maxRowsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	for pageNum := 0; pageNum < totalPages; pageNum++ {
		pdf.AddPage()

		// Header
		fill(brandMaroon, 0, 0, pageW, 12)
		fill(brandGold, 0, 12, pageW, 3)

		// Add logo
		logoData, logoImg := fetchLogo()
		textX := margin
		if logoImg != nil {
			b := logoImg.Bounds()
			h := 50.0
			wd := h * float64(b.Dx()) / float64(b.Dy())
			if wd > 160 {
				wd = 160
				h = wd * float64(b.Dy()) / float64(b.Dx())
			}
			if holder, err := gopdf.ImageHolderByBytes(logoData); err == nil {
				if err := pdf.ImageByHolder(holder, margin, 35, &gopdf.Rect{W: wd, H: h}); err == nil {
					textX = margin + wd + 15
				}
			}
		}

		text("bold", 18, brandMaroon, textX, 42, "UNIVERSITY OF KIGALI")
		text("regular", 10, brandBlue, textX, 68, "Accounts Report")
		textAligned("regular", 9, textGray, pageW-margin-200, 68, 200, gopdf.Right, fmt.Sprintf("Page %d of %d", pageNum+1, totalPages))

		hline(lineGray, 1, 105)

		// Table header
		text("bold", 12, brandMaroon, margin, 125, "Student Accounts")
		fill(brandGold, margin, 143, 40, 2)

		// Column headers
		y := 158.0
		fill(brandBlue, margin, y, contentW, rowH)
		text("bold", 11, white, margin+5, y+8, "Student No")
		text("bold", 11, white, margin+120, y+8, "Student Name")
		textAligned("bold", 11, white, margin+300, y+8, 100, gopdf.Right, "Invoiced")
		textAligned("bold", 11, white, margin+420, y+8, 100, gopdf.Right, "Paid")
		textAligned("bold", 11, white, margin+540, y+8, 100, gopdf.Right, "Balance")
		textAligned("bold", 11, white, margin+660, y+8, 80, gopdf.Right, "Invoices")

		y += rowH

		// Data rows for this page
		startIdx := pageNum * maxRowsPerPage
		endIdx := startIdx + maxRowsPerPage
		if endIdx > len(accountList) {
			endIdx = len(accountList)
		}

		for i := startIdx; i < endIdx; i++ {
			acc := accountList[i]

			// Alternate row colors
			if i%2 == 0 {
				fill(mix(brandBlue, 0.95), margin, y, contentW, rowH)
			}

			text("regular", 10, textDark, margin+5, y+8, acc["student_number"].(string))
			text("regular", 10, textDark, margin+120, y+8, acc["student_name"].(string))
			textAligned("regular", 10, textDark, margin+300, y+8, 100, gopdf.Right, formatRWF(acc["total_invoiced"].(int64)))
			textAligned("regular", 10, textDark, margin+420, y+8, 100, gopdf.Right, formatRWF(acc["total_paid"].(int64)))
			textAligned("regular", 10, textDark, margin+540, y+8, 100, gopdf.Right, formatRWF(acc["outstanding_balance"].(int64)))
			textAligned("regular", 10, textDark, margin+660, y+8, 80, gopdf.Right, strconv.Itoa(acc["invoice_count"].(int)))

			y += rowH
		}

		// Footer
		hline(brandMaroon, 2, 800)
		text("regular", 9, textGray, margin, 815, "Generated on "+fmt.Sprintf("%s", reportType))
		fill(brandGold, 0, 835, pageW, 3)
		fill(brandMaroon, 0, 838, pageW, 15)
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+reportType+"_report.pdf")
	pdf.Write(w)
}

func mix(c rgb, t float64) rgb {
	return rgb{
		uint8(float64(c[0]) + (255-float64(c[0]))*t),
		uint8(float64(c[1]) + (255-float64(c[1]))*t),
		uint8(float64(c[2]) + (255-float64(c[2]))*t),
	}
}

func formatRWF(n int64) string {
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s + " RWF"
}

func fetchLogo() ([]byte, image.Image) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(logoURL)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil
	}
	return data, img
}
