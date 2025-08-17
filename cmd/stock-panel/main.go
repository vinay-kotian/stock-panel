package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
	"github.com/vinaykotian/stock-panel/internal/db"
	"github.com/vinaykotian/stock-panel/internal/handlers"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  No .env file found or error loading it: %v", err)
		log.Printf("💡 Make sure to set SMTP_USER and SMTP_PASS environment variables for email functionality")
	} else {
		log.Printf("✅ .env file loaded successfully")
	}

	db.InitDB()
	defer db.DB.Close()

	// Start token cleanup goroutine
	handlers.StartTokenCleanup()

	// Root redirect to login
	http.HandleFunc("/", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	}))

	// Authentication routes
	http.HandleFunc("/login", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pages/login.html")
	}))

	http.HandleFunc("/test-login", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "test_login.html")
	}))

	http.HandleFunc("/debug-login", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "debug_login.html")
	}))

	http.HandleFunc("/simple-test", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "simple_test.html")
	}))

	http.HandleFunc("/test-email", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "test-email.html")
	}))

	http.HandleFunc("/test-alerts", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "test-alerts.html")
	}))

	http.HandleFunc("/register", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pages/register.html")
	}))

	http.HandleFunc("/forgot-password", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pages/forgot-password.html")
	}))

	http.HandleFunc("/reset-password", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pages/reset-password.html")
	}))

	http.HandleFunc("/auth/login", handlers.LoggingMiddleware(handlers.HandleLogin))
	http.HandleFunc("/auth/register", handlers.LoggingMiddleware(handlers.HandleRegister))
	http.HandleFunc("/auth/logout", handlers.LoggingMiddleware(handlers.HandleLogout))
	http.HandleFunc("/auth/verify", handlers.LoggingMiddleware(handlers.HandleVerifyToken))
	http.HandleFunc("/auth/forgot-password", handlers.LoggingMiddleware(handlers.HandleForgotPassword))
	http.HandleFunc("/auth/reset-password", handlers.LoggingMiddleware(handlers.HandleResetPassword))
	http.HandleFunc("/auth/test-email", handlers.LoggingMiddleware(handlers.HandleTestEmail))

	// Serve /web and /web/ as the main page, hide /pages in the URL
	http.HandleFunc("/web", handlers.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/", http.StatusFound)
	}))

	// Serve /web/list and /web/list/ as the list page (protected)
	http.HandleFunc("/web/list", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/list/", http.StatusFound)
	})))

	http.HandleFunc("/web/list/", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/web/list/" || r.URL.Path == "/web/list" {
			http.ServeFile(w, r, "web/pages/list.html")
			return
		}
		// Serve other static files (js, css, etc.)
		staticPath := strings.TrimPrefix(r.URL.Path, "/web/")
		if strings.HasPrefix(staticPath, "js/") || strings.HasPrefix(staticPath, "css/") {
			http.ServeFile(w, r, "web/"+staticPath)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	})))

	// Serve /web/dashboard and /web/dashboard/ as the dashboard page (protected)
	http.HandleFunc("/web/dashboard", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/dashboard/", http.StatusFound)
	})))

	http.HandleFunc("/web/dashboard/", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/web/dashboard/" || r.URL.Path == "/web/dashboard" {
			http.ServeFile(w, r, "web/pages/dashboard.html")
			return
		}
		// Serve other static files (js, css, etc.)
		staticPath := strings.TrimPrefix(r.URL.Path, "/web/")
		if strings.HasPrefix(staticPath, "js/") || strings.HasPrefix(staticPath, "css/") {
			http.ServeFile(w, r, "web/"+staticPath)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	})))

	// Serve /web/ as the main page (protected)
	http.HandleFunc("/web/", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/web/" || r.URL.Path == "/web" {
			http.ServeFile(w, r, "web/pages/index.html")
			return
		}
		staticPath := strings.TrimPrefix(r.URL.Path, "/web/")
		if strings.HasPrefix(staticPath, "js/") || strings.HasPrefix(staticPath, "css/") {
			http.ServeFile(w, r, "web/"+staticPath)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	})))

	// API endpoint (protected)
	http.HandleFunc("/stocks", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CollectStock(w, r)
		case http.MethodGet:
			handlers.GetStocks(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	// New endpoint for daily P&L (protected)
	http.HandleFunc("/pnl", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.GetDailyPnL(w, r)
	})))

	// Alerts endpoints (protected)
	http.HandleFunc("/alerts", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CreateAlert(w, r)
		case http.MethodGet:
			handlers.GetAlerts(w, r)
		case http.MethodPut:
			handlers.UpdateAlert(w, r)
		case http.MethodDelete:
			handlers.DeleteAlert(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	// Alert toggle endpoint (protected)
	http.HandleFunc("/alerts/toggle", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.ToggleAlert(w, r)
	})))

	// Test Kite API endpoint (protected)
	http.HandleFunc("/alerts/test-kite", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.TestKiteAPI(w, r)
	})))

	// Serve /web/alerts and /web/alerts/ as the alerts page (protected)
	http.HandleFunc("/web/alerts", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/alerts/", http.StatusFound)
	})))

	http.HandleFunc("/web/alerts/", handlers.LoggingMiddleware(handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/web/alerts/" || r.URL.Path == "/web/alerts" {
			http.ServeFile(w, r, "web/pages/alerts.html")
			return
		}
		// Serve other static files (js, css, etc.)
		staticPath := strings.TrimPrefix(r.URL.Path, "/web/")
		if strings.HasPrefix(staticPath, "js/") || strings.HasPrefix(staticPath, "css/") {
			http.ServeFile(w, r, "web/"+staticPath)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	})))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
