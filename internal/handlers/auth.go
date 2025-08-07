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
)

// Simple in-memory storage for demo purposes
// In production, use a proper database
var (
	users = map[string]string{
		"admin": "password123", // Default credentials
		"demo":  "demo123",
	}

	// User emails storage
	userEmails = map[string]string{
		"admin": "admin@example.com",
		"demo":  "demo@example.com",
	}

	// Simple token storage (in production, use Redis or database)
	tokens = make(map[string]time.Time)
)

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

// AuthMiddleware checks if the request has a valid authentication token
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for login page, register page, and static assets
		if strings.HasPrefix(r.URL.Path, "/login") ||
			strings.HasPrefix(r.URL.Path, "/register") ||
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

	// Validate credentials
	if !isValidCredentials(loginReq.Username, loginReq.Password) {
		log.Printf("Invalid credentials for user: %s", loginReq.Username)
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
	if _, exists := users[registerReq.Username]; exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(RegisterResponse{
			Message: "Username already exists",
		})
		return
	}

	// Check if email already exists
	for _, email := range userEmails {
		if email == registerReq.Email {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(RegisterResponse{
				Message: "Email already registered",
			})
			return
		}
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

	// Store new user
	users[registerReq.Username] = registerReq.Password
	userEmails[registerReq.Username] = registerReq.Email

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

// isValidCredentials checks if the provided username and password are valid
func isValidCredentials(username, password string) bool {
	if storedPassword, exists := users[username]; exists {
		return storedPassword == password
	}
	return false
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
