package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/vinaykotian/stock-panel/internal/db"
	"github.com/vinaykotian/stock-panel/internal/handlers"
)

func main() {
	db.InitDB()
	defer db.DB.Close()

	// Start token cleanup goroutine
	handlers.StartTokenCleanup()

	// Root redirect to login
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	// Authentication routes
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pages/login.html")
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/pages/register.html")
	})

	http.HandleFunc("/auth/login", handlers.HandleLogin)
	http.HandleFunc("/auth/register", handlers.HandleRegister)
	http.HandleFunc("/auth/logout", handlers.HandleLogout)
	http.HandleFunc("/auth/verify", handlers.HandleVerifyToken)

	// Serve /web and /web/ as the main page, hide /pages in the URL
	http.HandleFunc("/web", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/", http.StatusFound)
	})

	// Serve /web/list and /web/list/ as the list page (protected)
	http.HandleFunc("/web/list", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/list/", http.StatusFound)
	}))

	http.HandleFunc("/web/list/", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	// Serve /web/dashboard and /web/dashboard/ as the dashboard page (protected)
	http.HandleFunc("/web/dashboard", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/dashboard/", http.StatusFound)
	}))

	http.HandleFunc("/web/dashboard/", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	// Serve /web/ as the main page (protected)
	http.HandleFunc("/web/", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	// API endpoint (protected)
	http.HandleFunc("/stocks", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CollectStock(w, r)
		case http.MethodGet:
			handlers.GetStocks(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	// New endpoint for daily P&L (protected)
	http.HandleFunc("/pnl", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlers.GetDailyPnL(w, r)
	}))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
