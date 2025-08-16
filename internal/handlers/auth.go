package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/vinaykotian/stock-panel/internal/db"
	"github.com/vinaykotian/stock-panel/internal/email"
)

// Simple in-memory storage for demo purposes
// In production, use a proper database
var (
	// Simple token storage (in production, use Redis or database)
	tokens = make(map[string]time.Time)

	// Password reset tokens storage
	resetTokens = make(map[string]resetTokenData)

	// Email service instance
	emailService = email.NewEmailService()
)

// resetTokenData stores reset token information
type resetTokenData struct {
	Username  string
	Email     string
	ExpiresAt time.Time
}

// LoginRequest represents the login request structure
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents the registration request structure
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response structure
type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// RegisterResponse represents the registration response structure
type RegisterResponse struct {
	Message string `json:"message"`
}

// ForgotPasswordRequest represents the forgot password request structure
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordResponse represents the forgot password response structure
type ForgotPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResetPasswordRequest represents the reset password request structure
type ResetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// ResetPasswordResponse represents the reset password response structure
type ResetPasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// LoggingMiddleware logs all incoming requests with detailed information
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request details
		log.Printf("🌐 [REQUEST] %s %s - IP: %s - User-Agent: %s",
			r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

		// Log request headers (excluding sensitive ones)
		log.Printf("📋 [HEADERS] Content-Type: %s, Accept: %s, Content-Length: %s",
			r.Header.Get("Content-Type"), r.Header.Get("Accept"), r.Header.Get("Content-Length"))

		// Log query parameters if any
		if len(r.URL.Query()) > 0 {
			log.Printf("🔍 [QUERY] %s", r.URL.RawQuery)
		}

		// Log request body for POST/PUT requests (if it's JSON)
		if r.Method == "POST" || r.Method == "PUT" {
			if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				body, err := io.ReadAll(r.Body)
				if err == nil && len(body) > 0 {
					// Mask sensitive data in the log
					bodyStr := string(body)
					if strings.Contains(bodyStr, "password") {
						bodyStr = strings.ReplaceAll(bodyStr, `"password":"[^"]*"`, `"password":"***"`)
					}
					if strings.Contains(bodyStr, "token") {
						bodyStr = strings.ReplaceAll(bodyStr, `"token":"[^"]*"`, `"token":"***"`)
					}
					log.Printf("📄 [BODY] %s", bodyStr)
					// Restore the body for the handler
					r.Body = io.NopCloser(strings.NewReader(string(body)))
				}
			}
		}

		// Call the next handler
		next(w, r)

		// Log response time
		duration := time.Since(start)
		log.Printf("⏱️  [RESPONSE] %s %s - Duration: %v", r.Method, r.URL.Path, duration)
	}
}

// AuthMiddleware checks if the request has a valid authentication token
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login page, register page, forgot password, reset password, and static assets
		if strings.HasPrefix(r.URL.Path, "/login") ||
			strings.HasPrefix(r.URL.Path, "/register") ||
			strings.HasPrefix(r.URL.Path, "/forgot-password") ||
			strings.HasPrefix(r.URL.Path, "/reset-password") ||
			strings.HasPrefix(r.URL.Path, "/auth/") ||
			strings.HasPrefix(r.URL.Path, "/web/css/") ||
			strings.HasPrefix(r.URL.Path, "/web/js/") {
			next.ServeHTTP(w, r)
			return
		}

		// For API requests, check Authorization header
		if strings.HasPrefix(r.URL.Path, "/stocks") || strings.HasPrefix(r.URL.Path, "/pnl") {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Unauthorized"}`))
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if !isValidToken(token) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Invalid token"}`))
				return
			}
		}

		// For page requests, allow them to proceed (frontend will handle auth)
		next.ServeHTTP(w, r)
	}
}

// HandleLogin processes login requests
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleLogin called with method: %s", r.Method)

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Request body received: %s", string(body))

	// Parse login request
	var loginReq LoginRequest
	if err := json.Unmarshal(body, &loginReq); err != nil {
		log.Printf("Failed to parse login request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Parsed login request - Username: %s, Password: %s", loginReq.Username, "***")

	// Validate credentials using database
	_, storedPassword, err := db.GetUserByUsername(loginReq.Username)
	if err != nil {
		log.Printf("User not found: %s", loginReq.Username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Message: "Invalid username or password",
		})
		return
	}

	if storedPassword != loginReq.Password {
		log.Printf("Invalid password for user: %s", loginReq.Username)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Message: "Invalid username or password",
		})
		return
	}

	log.Printf("Credentials validated successfully for user: %s", loginReq.Username)

	// Generate token
	token, err := generateToken()
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store token with expiration
	tokens[token] = time.Now().Add(24 * time.Hour) // 24 hour expiration

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := LoginResponse{
		Token:   token,
		Message: "Login successful",
	}
	log.Printf("Sending successful login response for user: %s", loginReq.Username)
	json.NewEncoder(w).Encode(response)
}

// HandleLogout processes logout requests
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		delete(tokens, token)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout successful",
	})
}

// HandleRegister processes registration requests
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse registration request
	var registerReq RegisterRequest
	if err := json.Unmarshal(body, &registerReq); err != nil {
		log.Printf("Failed to parse registration request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate input
	if registerReq.Username == "" || registerReq.Email == "" || registerReq.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{
			Message: "All fields are required",
		})
		return
	}

	// Check if username already exists
	_, _, err = db.GetUserByUsername(registerReq.Username)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(RegisterResponse{
			Message: "Username already exists",
		})
		return
	}

	// Check if email already exists
	_, err = db.GetUserByEmail(registerReq.Email)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(RegisterResponse{
			Message: "Email already registered",
		})
		return
	}

	// Validate password strength (basic validation)
	if len(registerReq.Password) < 8 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{
			Message: "Password must be at least 8 characters long",
		})
		return
	}

	// Store new user in database
	err = db.CreateUser(registerReq.Username, registerReq.Email, registerReq.Password)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("New user registered: %s", registerReq.Username)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		Message: "Account created successfully",
	})
}

// HandleVerifyToken verifies if a token is valid
func HandleVerifyToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if !isValidToken(token) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Token is valid",
	})
}

// isValidToken checks if the provided token is valid and not expired
func isValidToken(token string) bool {
	if expiration, exists := tokens[token]; exists {
		if time.Now().Before(expiration) {
			return true
		}
		// Token expired, remove it
		delete(tokens, token)
	}
	return false
}

// generateToken creates a new random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CleanupExpiredTokens removes expired tokens from memory
func CleanupExpiredTokens() {
	for token, expiration := range tokens {
		if time.Now().After(expiration) {
			delete(tokens, token)
		}
	}
	// Also clean up expired reset tokens
	CleanupExpiredResetTokens()
}

// StartTokenCleanup starts a goroutine to periodically clean up expired tokens
func StartTokenCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Clean up every hour
		defer ticker.Stop()

		for range ticker.C {
			CleanupExpiredTokens()
		}
	}()
}

// HandleForgotPassword processes forgot password requests
func HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse forgot password request
	var forgotReq ForgotPasswordRequest
	if err := json.Unmarshal(body, &forgotReq); err != nil {
		log.Printf("Failed to parse forgot password request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate email
	if forgotReq.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ForgotPasswordResponse{
			Success: false,
			Message: "Email is required",
		})
		return
	}

	// Find user by email using database
	username, err := db.GetUserByEmail(forgotReq.Email)

	log.Printf("🔍 [FORGOT_PASSWORD] Email: %s, Found user: %s", forgotReq.Email, username)

	if err != nil {
		// Don't reveal if email exists or not for security
		log.Printf("⚠️  [FORGOT_PASSWORD] Email not found in system: %s", forgotReq.Email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ForgotPasswordResponse{
			Success: true,
			Message: "If the email exists in our system, a reset link has been sent.",
		})
		return
	}

	log.Printf("✅ [FORGOT_PASSWORD] User found: %s for email: %s", username, forgotReq.Email)

	// Generate reset token
	resetToken, err := generateToken()
	if err != nil {
		log.Printf("Failed to generate reset token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store reset token with expiration (1 hour)
	resetTokens[resetToken] = resetTokenData{
		Username:  username,
		Email:     forgotReq.Email,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Generate reset link
	resetLink := "http://localhost:8080/reset-password?token=" + resetToken

	// Send email if email service is configured
	if emailService.IsEmailConfigured() {
		log.Printf("📧 Email service is configured, attempting to send reset email...")
		if err := emailService.SendPasswordResetEmail(forgotReq.Email, resetLink); err != nil {
			log.Printf("❌ Failed to send password reset email to %s: %v", forgotReq.Email, err)
			// Still return success to avoid revealing if email exists
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ForgotPasswordResponse{
				Success: true,
				Message: "If the email exists in our system, a reset link has been sent.",
			})
			return
		}
		log.Printf("✅ Password reset email sent successfully to %s", forgotReq.Email)
	} else {
		// Fallback to console logging if email is not configured
		log.Printf("⚠️  Email service not configured - displaying reset link in console")
		log.Printf("🔗 Password reset link for %s: %s", forgotReq.Email, resetLink)
		log.Printf("💡 To enable email sending, set SMTP_USER and SMTP_PASS environment variables")
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ForgotPasswordResponse{
		Success: true,
		Message: "If the email exists in our system, a reset link has been sent.",
	})
}

// HandleResetPassword processes password reset requests
func HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse reset password request
	var resetReq ResetPasswordRequest
	if err := json.Unmarshal(body, &resetReq); err != nil {
		log.Printf("Failed to parse reset password request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate input
	if resetReq.Token == "" || resetReq.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResetPasswordResponse{
			Success: false,
			Message: "Token and password are required",
		})
		return
	}

	// Validate password strength
	if len(resetReq.Password) < 8 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResetPasswordResponse{
			Success: false,
			Message: "Password must be at least 8 characters long",
		})
		return
	}

	// Check if reset token exists and is valid
	resetData, exists := resetTokens[resetReq.Token]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResetPasswordResponse{
			Success: false,
			Message: "Invalid or expired reset token",
		})
		return
	}

	// Check if token is expired
	if time.Now().After(resetData.ExpiresAt) {
		delete(resetTokens, resetReq.Token)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResetPasswordResponse{
			Success: false,
			Message: "Reset token has expired",
		})
		return
	}

	// Update user password in database
	err = db.UpdateUserPassword(resetData.Username, resetReq.Password)
	if err != nil {
		log.Printf("Failed to update password for user %s: %v", resetData.Username, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Remove the used reset token
	delete(resetTokens, resetReq.Token)

	log.Printf("Password reset successfully for user: %s", resetData.Username)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResetPasswordResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}

// CleanupExpiredResetTokens removes expired reset tokens from memory
func CleanupExpiredResetTokens() {
	for token, resetData := range resetTokens {
		if time.Now().After(resetData.ExpiresAt) {
			delete(resetTokens, token)
		}
	}
}

// HandleTestEmail tests the email configuration
func HandleTestEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Parse test email request
	var testReq struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &testReq); err != nil {
		log.Printf("Failed to parse test email request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate email
	if testReq.Email == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Email is required",
		})
		return
	}

	// Check if email service is configured
	if !emailService.IsEmailConfigured() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Email service is not configured. Set SMTP_USER and SMTP_PASS environment variables.",
		})
		return
	}

	// Send test email
	testSubject := "Test Email - Stock Panel"
	testBody := `
		<html>
		<body>
			<h2>Test Email</h2>
			<p>This is a test email to verify your email configuration is working correctly.</p>
			<p>If you received this email, your email service is properly configured!</p>
			<br>
			<p>Best regards,<br>Stock Panel Team</p>
		</body>
		</html>
	`

	if err := emailService.SendEmail(testReq.Email, testSubject, testBody); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to send test email: " + err.Error(),
		})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Test email sent successfully!",
	})
}
