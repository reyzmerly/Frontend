package mopay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Service struct {
	apiURL      string
	apiToken    string
	walletPhone string
	callbackURL string
	timeoutMS   int
	httpClient  *http.Client
}

type InitiatePaymentRequest struct {
	TransactionID string     `json:"transactionId"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Phone         int64      `json:"phone"`
	PaymentMode   string     `json:"payment_mode"`
	Message       string     `json:"message"`
	CallbackURL   string     `json:"callbackUrl"`
	Transfers     []Transfer `json:"transfers,omitempty"`
}

type Transfer struct {
	Phone   int64   `json:"phone"`
	Amount  float64 `json:"amount"`
	Message string  `json:"message"`
}

type InitiatePaymentResponse struct {
	TransactionID string `json:"transactionId"`
	Status        int    `json:"status"`
	Message       string `json:"message"`
}

type StatusResponse struct {
	TransactionID string  `json:"transaction_id"`
	Status        int     `json:"status"`
	StatusText    string  `json:"status_text"`
	Amount        float64 `json:"amount"`
}

type CustomerInfoResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

func NewService(apiURL, apiToken, walletPhone, callbackURL string, timeoutMS int) *Service {
	return &Service{
		apiURL:      apiURL,
		apiToken:    apiToken,
		walletPhone: walletPhone,
		callbackURL: callbackURL,
		timeoutMS:   timeoutMS,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutMS) * time.Millisecond,
		},
	}
}

func (s *Service) GenerateTransactionID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("INV-%d-%d", timestamp, time.Now().Unix())
}

func (s *Service) InitiatePayment(req InitiatePaymentRequest) (*InitiatePaymentResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/initiate-payment", s.apiURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	authHeader := "Bearer " + s.apiToken
	httpReq.Header.Set("Authorization", authHeader)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate payment: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("MoPay API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response InitiatePaymentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (s *Service) CheckTransactionStatus(transactionID string) (*StatusResponse, error) {
	url := fmt.Sprintf("%s/check-status/%s", s.apiURL, transactionID)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiToken)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to check status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("MoPay API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response StatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (s *Service) GetCustomerInfo(phone string) (*CustomerInfoResponse, error) {
	url := fmt.Sprintf("%s/customer-info?phone=%s", s.apiURL, phone)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.apiToken)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("MoPay API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response CustomerInfoResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (s *Service) IsSuccessful(status *StatusResponse) bool {
	return status != nil && status.Status == 200
}

func (s *Service) IsFailed(status *StatusResponse) bool {
	return status != nil && status.Status == 400
}

func (s *Service) WalletPhone() string {
	return s.walletPhone
}

func (s *Service) CallbackURL() string {
	return s.callbackURL
}
