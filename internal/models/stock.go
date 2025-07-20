package models

import "time"

// Stock represents options information and trading signals
// OptionType: "CALL" or "PUT"
// Side: "BUY" or "SELL"
type Stock struct {
	Symbol           string    `json:"symbol"`
	UnderlyingSymbol string    `json:"underlying_symbol"`
	OptionType       string    `json:"option_type"`
	StrikePrice      float64   `json:"strike_price"`
	Expiry           string    `json:"expiry"`
	Price            float64   `json:"price"`
	Side             string    `json:"side"`
	Timestamp        time.Time `json:"timestamp"`
}
