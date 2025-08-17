package models

import "time"

// Alert represents a trading alert that can be sent to Kite 3 API
type Alert struct {
	ID               int       `json:"id"`
	Symbol           string    `json:"symbol"`
	UnderlyingSymbol string    `json:"underlying_symbol,omitempty"`
	OptionType       string    `json:"option_type,omitempty"` // "CALL", "PUT", or empty for stocks
	StrikePrice      float64   `json:"strike_price,omitempty"`
	Expiry           string    `json:"expiry,omitempty"`
	AlertType        string    `json:"alert_type"` // "PRICE_ABOVE", "PRICE_BELOW", "PERCENTAGE_CHANGE"
	TargetValue      float64   `json:"target_value"`
	Condition        string    `json:"condition"` // ">", "<", ">=", "<=", "=="
	Message          string    `json:"message"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	UserID           int       `json:"user_id"`
}

// AlertRequest represents the request structure for creating/updating alerts
type AlertRequest struct {
	Symbol           string  `json:"symbol"`
	UnderlyingSymbol string  `json:"underlying_symbol,omitempty"`
	OptionType       string  `json:"option_type,omitempty"`
	StrikePrice      float64 `json:"strike_price,omitempty"`
	Expiry           string  `json:"expiry,omitempty"`
	AlertType        string  `json:"alert_type"`
	TargetValue      float64 `json:"target_value"`
	Condition        string  `json:"condition"`
	Message          string  `json:"message"`
}

// AlertResponse represents the response structure for alert operations
type AlertResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Alert   *Alert `json:"alert,omitempty"`
}

// AlertsResponse represents the response structure for getting all alerts
type AlertsResponse struct {
	Success bool    `json:"success"`
	Alerts  []Alert `json:"alerts"`
}
