package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/vinaykotian/stock-panel/internal/db"
	"github.com/vinaykotian/stock-panel/internal/models"
)

func CollectStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Log the raw request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Incoming POST /stocks request: %s", string(body))
	// Decode the body into Stock
	var s models.Stock
	if err := json.Unmarshal(body, &s); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.Timestamp = time.Now()
	_, err = db.DB.Exec(
		"INSERT INTO stocks (symbol, underlying_symbol, option_type, strike_price, expiry, price, side, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		s.Symbol, s.UnderlyingSymbol, s.OptionType, s.StrikePrice, s.Expiry, s.Price, s.Side, s.Timestamp,
	)
	if err != nil {
		log.Printf("Failed to insert stock: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetStocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.DB.Query("SELECT symbol, underlying_symbol, option_type, strike_price, expiry, price, side, timestamp FROM stocks")
	if err != nil {
		log.Printf("Failed to query stocks: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var stocks []models.Stock
	for rows.Next() {
		var s models.Stock
		var ts string
		if err := rows.Scan(&s.Symbol, &s.UnderlyingSymbol, &s.OptionType, &s.StrikePrice, &s.Expiry, &s.Price, &s.Side, &ts); err != nil {
			log.Printf("Failed to scan stock: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		t, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			t, _ = time.Parse(time.RFC3339, ts)
		}
		s.Timestamp = t
		stocks = append(stocks, s)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stocks)
}

// Add this function to serve static files from the web directory
func ServeWeb(staticDir string) http.Handler {
	return http.StripPrefix("/web/", http.FileServer(http.Dir(staticDir)))
}

// Redirect /web/ to /web/pages/index.html for convenience
func WebRootRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/web/pages/index.html", http.StatusFound)
}

// GetDailyPnL returns consolidated profit and loss for each day
func GetDailyPnL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	rows, err := db.DB.Query("SELECT price, side, timestamp FROM stocks")
	if err != nil {
		log.Printf("Failed to query stocks: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type dailyPnL struct {
		Date string  `json:"date"`
		PnL  float64 `json:"pnl"`
	}

	pnlMap := make(map[string]float64)
	for rows.Next() {
		var price float64
		var side string
		var ts string
		if err := rows.Scan(&price, &side, &ts); err != nil {
			log.Printf("Failed to scan stock: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		t, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			t, _ = time.Parse(time.RFC3339, ts)
		}
		date := t.Format("2006-01-02")
		if side == "SELL" {
			pnlMap[date] += price
		} else if side == "BUY" {
			pnlMap[date] -= price
		}
	}

	var result []dailyPnL
	for date, pnl := range pnlMap {
		result = append(result, dailyPnL{Date: date, PnL: pnl})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
