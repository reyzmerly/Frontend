package hdev

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Service struct {
	apiID       string
	apiKey      string
	callbackURL string
	timeoutMS   int
	httpClient  *http.Client
}

type InitiatePaymentRequest struct {
	Ref    string `json:"ref"`
	Tel    string `json:"tel"`
	TxRef  string `json:"tx_ref"`
	Amount int    `json:"amount"`
	Link   string `json:"link"`
}

type InitiatePaymentResponse struct {
	Status  string `json:"status"`
	TxRef   string `json:"tx_ref"`
	Message string `json:"message,omitempty"`
}

type StatusResponse struct {
	Status  string `json:"status"`
	TxRef   string `json:"tx_ref"`
	Message string `json:"message,omitempty"`
}

func NewService(apiID, apiKey, callbackURL string, timeoutMS int) *Service {
	return &Service{
		apiID:       apiID,
		apiKey:      apiKey,
		callbackURL: callbackURL,
		timeoutMS:   timeoutMS,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutMS) * time.Millisecond,
		},
	}
}

func (s *Service) InitiatePayment(req InitiatePaymentRequest) (*InitiatePaymentResponse, error) {
	// Build form data
	formData := url.Values{}
	formData.Set("ref", req.Ref)
	formData.Set("tel", req.Tel)
	formData.Set("tx_ref", req.TxRef)
	formData.Set("amount", fmt.Sprintf("%d", req.Amount))
	formData.Set("link", req.Link)

	// URL with API ID and Key in path
	url := fmt.Sprintf("https://payment.hdevtech.cloud/api_pay/api/%s/%s", s.apiID, s.apiKey)
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create HDEV payment request")
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HDEV payment request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("HDEV API returned status %d", resp.StatusCode)
	}

	var response InitiatePaymentResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (s *Service) CheckTransactionStatus(txRef string) (*StatusResponse, error) {
	// Build form data
	formData := url.Values{}
	formData.Set("ref", "read")
	formData.Set("tx_ref", txRef)

	// URL with API ID and Key in path
	url := fmt.Sprintf("https://payment.hdevtech.cloud/api_pay/api/%s/%s", s.apiID, s.apiKey)
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create HDEV status request")
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HDEV status request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("HDEV API returned status %d", resp.StatusCode)
	}

	var response StatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (s *Service) CallbackURL() string {
	return s.callbackURL
}
