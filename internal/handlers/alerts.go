package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vinaykotian/stock-panel/internal/db"
	"github.com/vinaykotian/stock-panel/internal/kite"
	"github.com/vinaykotian/stock-panel/internal/models"
)

// CreateAlert handles creating a new alert
func CreateAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID := r.Context().Value("userID").(int)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var alertReq models.AlertRequest
	if err := json.Unmarshal(body, &alertReq); err != nil {
		log.Printf("Failed to unmarshal alert request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate required fields
	if alertReq.Symbol == "" || alertReq.AlertType == "" || alertReq.Condition == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Symbol, alert type, and condition are required",
		})
		return
	}

	// Insert alert into database
	result, err := db.DB.Exec(
		"INSERT INTO alerts (symbol, underlying_symbol, option_type, strike_price, expiry, alert_type, target_value, condition, message, is_active, created_at, updated_at, user_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		alertReq.Symbol, alertReq.UnderlyingSymbol, alertReq.OptionType, alertReq.StrikePrice, alertReq.Expiry, alertReq.AlertType, alertReq.TargetValue, alertReq.Condition, alertReq.Message, true, time.Now(), time.Now(), userID,
	)
	if err != nil {
		log.Printf("Failed to insert alert: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	alertID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to get last insert ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create response
	alert := &models.Alert{
		ID:               int(alertID),
		Symbol:           alertReq.Symbol,
		UnderlyingSymbol: alertReq.UnderlyingSymbol,
		OptionType:       alertReq.OptionType,
		StrikePrice:      alertReq.StrikePrice,
		Expiry:           alertReq.Expiry,
		AlertType:        alertReq.AlertType,
		TargetValue:      alertReq.TargetValue,
		Condition:        alertReq.Condition,
		Message:          alertReq.Message,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		UserID:           userID,
	}

	// Send alert to Kite 3 API if configured
	kiteAPIKey := os.Getenv("KITE_API_KEY")
	kiteAPISecret := os.Getenv("KITE_API_SECRET")
	kiteBaseURL := os.Getenv("KITE_BASE_URL")

	if kiteAPIKey != "" && kiteAPISecret != "" && kiteBaseURL != "" {
		kiteService := kite.NewKiteService(kiteAPIKey, kiteAPISecret, kiteBaseURL)

		kiteAlert := kite.AlertPayload{
			Symbol:           alert.Symbol,
			UnderlyingSymbol: alert.UnderlyingSymbol,
			OptionType:       alert.OptionType,
			StrikePrice:      alert.StrikePrice,
			Expiry:           alert.Expiry,
			AlertType:        alert.AlertType,
			TargetValue:      alert.TargetValue,
			Condition:        alert.Condition,
			Message:          alert.Message,
			Timestamp:        alert.CreatedAt.Format(time.RFC3339),
			UserID:           alert.UserID,
		}

		if err := kiteService.SendAlert(kiteAlert); err != nil {
			log.Printf("Warning: Failed to send alert to Kite API: %v", err)
			// Continue with the response even if Kite API fails
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.AlertResponse{
		Success: true,
		Message: "Alert created successfully",
		Alert:   alert,
	})
}

// GetAlerts handles getting all alerts for a user
func GetAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID := r.Context().Value("userID").(int)

	rows, err := db.DB.Query("SELECT id, symbol, underlying_symbol, option_type, strike_price, expiry, alert_type, target_value, condition, message, is_active, created_at, updated_at, user_id FROM alerts WHERE user_id = ? ORDER BY created_at DESC", userID)
	if err != nil {
		log.Printf("Failed to query alerts: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var alert models.Alert
		var createdAt, updatedAt string
		if err := rows.Scan(&alert.ID, &alert.Symbol, &alert.UnderlyingSymbol, &alert.OptionType, &alert.StrikePrice, &alert.Expiry, &alert.AlertType, &alert.TargetValue, &alert.Condition, &alert.Message, &alert.IsActive, &createdAt, &updatedAt, &alert.UserID); err != nil {
			log.Printf("Failed to scan alert: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Parse timestamps
		if t, err := time.Parse(time.RFC3339Nano, createdAt); err == nil {
			alert.CreatedAt = t
		}
		if t, err := time.Parse(time.RFC3339Nano, updatedAt); err == nil {
			alert.UpdatedAt = t
		}

		alerts = append(alerts, alert)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AlertsResponse{
		Success: true,
		Alerts:  alerts,
	})
}

// UpdateAlert handles updating an existing alert
func UpdateAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID := r.Context().Value("userID").(int)

	// Get alert ID from URL
	alertIDStr := r.URL.Query().Get("id")
	if alertIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Alert ID is required",
		})
		return
	}

	alertID, err := strconv.Atoi(alertIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Invalid alert ID",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var alertReq models.AlertRequest
	if err := json.Unmarshal(body, &alertReq); err != nil {
		log.Printf("Failed to unmarshal alert request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Update alert in database
	result, err := db.DB.Exec(
		"UPDATE alerts SET symbol = ?, underlying_symbol = ?, option_type = ?, strike_price = ?, expiry = ?, alert_type = ?, target_value = ?, condition = ?, message = ?, updated_at = ? WHERE id = ? AND user_id = ?",
		alertReq.Symbol, alertReq.UnderlyingSymbol, alertReq.OptionType, alertReq.StrikePrice, alertReq.Expiry, alertReq.AlertType, alertReq.TargetValue, alertReq.Condition, alertReq.Message, time.Now(), alertID, userID,
	)
	if err != nil {
		log.Printf("Failed to update alert: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Alert not found or you don't have permission to update it",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AlertResponse{
		Success: true,
		Message: "Alert updated successfully",
	})
}

// DeleteAlert handles deleting an alert
func DeleteAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID := r.Context().Value("userID").(int)

	// Get alert ID from URL
	alertIDStr := r.URL.Query().Get("id")
	if alertIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Alert ID is required",
		})
		return
	}

	alertID, err := strconv.Atoi(alertIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Invalid alert ID",
		})
		return
	}

	// Delete alert from database
	result, err := db.DB.Exec("DELETE FROM alerts WHERE id = ? AND user_id = ?", alertID, userID)
	if err != nil {
		log.Printf("Failed to delete alert: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Alert not found or you don't have permission to delete it",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AlertResponse{
		Success: true,
		Message: "Alert deleted successfully",
	})
}

// ToggleAlert handles toggling alert active status
func ToggleAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID := r.Context().Value("userID").(int)

	// Get alert ID from URL
	alertIDStr := r.URL.Query().Get("id")
	if alertIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Alert ID is required",
		})
		return
	}

	alertID, err := strconv.Atoi(alertIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Invalid alert ID",
		})
		return
	}

	// Toggle alert status in database
	result, err := db.DB.Exec(
		"UPDATE alerts SET is_active = NOT is_active, updated_at = ? WHERE id = ? AND user_id = ?",
		time.Now(), alertID, userID,
	)
	if err != nil {
		log.Printf("Failed to toggle alert: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get rows affected: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(models.AlertResponse{
			Success: false,
			Message: "Alert not found or you don't have permission to toggle it",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AlertResponse{
		Success: true,
		Message: "Alert status toggled successfully",
	})
}

// TestKiteAPI handles testing the Kite 3 API connection
func TestKiteAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	kiteAPIKey := os.Getenv("KITE_API_KEY")
	kiteAPISecret := os.Getenv("KITE_API_SECRET")
	kiteBaseURL := os.Getenv("KITE_BASE_URL")

	if kiteAPIKey == "" || kiteAPISecret == "" || kiteBaseURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Kite API credentials not configured. Please set KITE_API_KEY, KITE_API_SECRET, and KITE_BASE_URL environment variables.",
		})
		return
	}

	kiteService := kite.NewKiteService(kiteAPIKey, kiteAPISecret, kiteBaseURL)

	if err := kiteService.TestConnection(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to connect to Kite API: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Successfully connected to Kite API",
	})
}
