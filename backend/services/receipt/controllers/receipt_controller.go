package controllers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/signintech/gopdf"

	"github.com/kwizera/university-invoices/shared/repositories"
	"github.com/kwizera/university-invoices/shared/utils"
)

const logoURL = "https://res.cloudinary.com/danzeqybu/image/upload/v1793788865/Screenshot_From_2026-11-04_12-39-17_ka7vft.png"

type rgb [3]uint8

// Colors taken from the University of Kigali logo
var (
	brandMaroon = rgb{123, 24, 40}  // main color
	brandBlue   = rgb{37, 99, 235}  // secondary
	brandGold   = rgb{212, 160, 23} // accent

	textDark = rgb{31, 41, 55}
	textGray = rgb{107, 114, 128}
	lineGray = rgb{229, 231, 235}
	white    = rgb{255, 255, 255}
)

type ReceiptController struct {
	receiptRepo *repositories.ReceiptRepository
}

func NewReceiptController(db *sql.DB) *ReceiptController {
	return &ReceiptController{receiptRepo: repositories.NewReceiptRepository(db)}
}

func (c *ReceiptController) GetReceipt(w http.ResponseWriter, r *http.Request) {
	receiptID, _ := strconv.ParseUint(chi.URLParam(r, "receiptId"), 10, 32)
	receipt, err := c.receiptRepo.GetByID(uint(receiptID))
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Receipt not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipt)
}

func (c *ReceiptController) GetReceiptByPaymentID(w http.ResponseWriter, r *http.Request) {
	paymentID, _ := strconv.ParseUint(chi.URLParam(r, "paymentId"), 10, 32)

	receipt, err := c.receiptRepo.GetByPaymentID(uint(paymentID))
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Receipt not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipt)
}

// ---------- helpers ----------

func str(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func num(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case uint:
		return int64(n)
	case uint64:
		return int64(n)
	case float64:
		return int64(n)
	}
	return 0
}

func formatRWF(n int64) string {
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s + " RWF"
}

// mix blends a color with white. t=0 is the original color, t=1 is white.
func mix(c rgb, t float64) rgb {
	return rgb{
		uint8(float64(c[0]) + (255-float64(c[0]))*t),
		uint8(float64(c[1]) + (255-float64(c[1]))*t),
		uint8(float64(c[2]) + (255-float64(c[2]))*t),
	}
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

func formatIssued(v interface{}) string {
	s := str(v)
	if s == "" {
		return time.Now().Format("2006-01-02 15:04:05")
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Format("2006-01-02 15:04:05")
	}
	return strings.Split(strings.Replace(s, "T", " ", 1), ".")[0]
}

// ---------- PDF ----------

func (c *ReceiptController) GetReceiptPDF(w http.ResponseWriter, r *http.Request) {
	receiptID, _ := strconv.ParseUint(chi.URLParam(r, "receiptId"), 10, 32)
	receipt, err := c.receiptRepo.GetByID(uint(receiptID))
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Receipt not found"}`, http.StatusNotFound)
		return
	}

	accent := brandMaroon
	tint := mix(accent, 0.93)
	blueTint := mix(brandBlue, 0.92)

	logoData, logoImg := fetchLogo()

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	if err := utils.AddSystemPDFFonts(&pdf); err != nil {
		log.Printf("receipt PDF: %v", err)
		http.Error(w, `{"code":"INTERNAL_SERVER_ERROR","message":"Unable to generate receipt PDF"}`, http.StatusInternalServerError)
		return
	}

	const pageW, margin = 595.28, 40.0
	contentW := pageW - 2*margin

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

	// ----- Top bars -----
	fill(brandMaroon, 0, 0, pageW, 12)
	fill(brandGold, 0, 12, pageW, 3)

	// ----- Header: logo + university name -----
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
	text("regular", 10, brandBlue, textX, 68, "Invoice Management System")

	textAligned("bold", 24, accent, pageW-margin-200, 38, 200, gopdf.Right, "RECEIPT")
	textAligned("regular", 10, textGray, pageW-margin-200, 68, 200, gopdf.Right, "No. "+str(receipt["receipt_number"]))

	hline(lineGray, 1, 105)

	// ----- Section title -----
	text("bold", 12, accent, margin, 125, "PAYMENT DETAILS")
	fill(brandGold, margin, 143, 40, 2)

	issued := formatIssued(receipt["issued_at"])

	// ----- Detail rows (zebra striped) -----
	rows := [][2]string{
		{"Receipt Number", str(receipt["receipt_number"])},
		{"Invoice Number", str(receipt["invoice_number"])},
		{"Student Name", str(receipt["student_name"])},
		{"Date Issued", issued},
	}
	y := 158.0
	const rowH = 36.0
	for i, row := range rows {
		if i%2 == 0 {
			fill(tint, margin, y, contentW, rowH)
		}
		text("regular", 11, textGray, margin+15, y+11, row[0])
		textAligned("bold", 11, textDark, margin+contentW/2, y+11, contentW/2-15, gopdf.Right, row[1])
		y += rowH
	}
	hline(lineGray, 1, y)

	// ----- Highlight boxes -----
	y += 30
	boxW := (contentW - 20) / 2
	const boxH = 85.0

	// Amount paid (maroon)
	fill(brandMaroon, margin, y, boxW, boxH)
	fill(brandGold, margin, y+boxH-3, boxW, 3)
	textAligned("regular", 10, white, margin, y+18, boxW, gopdf.Center, "AMOUNT PAID")
	textAligned("bold", 20, white, margin, y+40, boxW, gopdf.Center, formatRWF(num(receipt["amount_paid"])))

	// Remaining balance (blue)
	bx := margin + boxW + 20
	fill(blueTint, bx, y, boxW, boxH)
	fill(brandBlue, bx, y, 4, boxH)
	textAligned("regular", 10, textGray, bx, y+18, boxW, gopdf.Center, "REMAINING BALANCE")
	textAligned("bold", 20, brandBlue, bx, y+40, boxW, gopdf.Center, formatRWF(num(receipt["new_balance"])))

	// ----- Footer -----
	hline(brandMaroon, 2, 740)
	text("bold", 11, accent, margin, 755, "Thank you for your payment!")
	text("regular", 9, textGray, margin, 775, "Generated on "+issued)
	textAligned("regular", 9, textGray, pageW-margin-250, 775, 250, gopdf.Right, "This is a computer-generated receipt.")
	fill(brandGold, 0, 824, pageW, 3)
	fill(brandMaroon, 0, 827, pageW, 15)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=receipt_"+str(receipt["receipt_number"])+".pdf")
	pdf.Write(w)
}
