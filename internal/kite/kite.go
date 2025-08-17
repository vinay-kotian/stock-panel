package kite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// KiteService handles communication with Kite 3 API
type KiteService struct {
	APIKey     string
	APISecret  string
	BaseURL    string
	HTTPClient *http.Client
}

// AlertPayload represents the structure for sending alerts to Kite
type AlertPayload struct {
	Symbol           string  `json:"symbol"`
	UnderlyingSymbol string  `json:"underlying_symbol,omitempty"`
	OptionType       string  `json:"option_type,omitempty"`
	StrikePrice      float64 `json:"strike_price,omitempty"`
	Expiry           string  `json:"expiry,omitempty"`
	AlertType        string  `json:"alert_type"`
	TargetValue      float64 `json:"target_value"`
	Condition        string  `json:"condition"`
	Message          string  `json:"message"`
	Timestamp        string  `json:"timestamp"`
	UserID           int     `json:"user_id"`
}

// KiteResponse represents the response from Kite API
type KiteResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewKiteService creates a new Kite service instance
func NewKiteService(apiKey, apiSecret, baseURL string) *KiteService {
	return &KiteService{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendAlert sends an alert to Kite 3 API
func (k *KiteService) SendAlert(alert AlertPayload) error {
	// Create the request payload
	payload := map[string]interface{}{
		"symbol":            alert.Symbol,
		"underlying_symbol": alert.UnderlyingSymbol,
		"option_type":       alert.OptionType,
		"strike_price":      alert.StrikePrice,
		"expiry":            alert.Expiry,
		"alert_type":        alert.AlertType,
		"target_value":      alert.TargetValue,
		"condition":         alert.Condition,
		"message":           alert.Message,
		"timestamp":         alert.Timestamp,
		"user_id":           alert.UserID,
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal alert payload: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", k.BaseURL+"/alerts", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", k.APIKey)
	req.Header.Set("X-API-Secret", k.APISecret)

	// Send request
	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Kite API: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Kite API returned status %d", resp.StatusCode)
	}

	// Parse response
	var kiteResp KiteResponse
	if err := json.NewDecoder(resp.Body).Decode(&kiteResp); err != nil {
		return fmt.Errorf("failed to decode Kite API response: %v", err)
	}

	// Check if the response indicates success
	if kiteResp.Status != "success" {
		return fmt.Errorf("Kite API error: %s", kiteResp.Message)
	}

	log.Printf("✅ Alert sent to Kite API successfully: %s", alert.Symbol)
	return nil
}

// SendBulkAlerts sends multiple alerts to Kite 3 API
func (k *KiteService) SendBulkAlerts(alerts []AlertPayload) error {
	for _, alert := range alerts {
		if err := k.SendAlert(alert); err != nil {
			log.Printf("❌ Failed to send alert for %s: %v", alert.Symbol, err)
			// Continue with other alerts even if one fails
			continue
		}
	}
	return nil
}

// TestConnection tests the connection to Kite 3 API
func (k *KiteService) TestConnection() error {
	req, err := http.NewRequest("GET", k.BaseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}

	req.Header.Set("X-API-Key", k.APIKey)
	req.Header.Set("X-API-Secret", k.APISecret)

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Kite API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Kite API health check failed with status %d", resp.StatusCode)
	}

	log.Printf("✅ Kite API connection test successful")
	return nil
}

// GetAlertStatus retrieves the status of alerts from Kite 3 API
func (k *KiteService) GetAlertStatus(userID int) ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/alerts/status?user_id=%d", k.BaseURL, userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create status request: %v", err)
	}

	req.Header.Set("X-API-Key", k.APIKey)
	req.Header.Set("X-API-Secret", k.APISecret)

	resp, err := k.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert status from Kite API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Kite API status request failed with status %d", resp.StatusCode)
	}

	var response struct {
		Status string                   `json:"status"`
		Data   []map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %v", err)
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("Kite API returned error status")
	}

	return response.Data, nil
}
