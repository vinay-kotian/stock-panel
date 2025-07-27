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

	// Serve /web and /web/ as the main page, hide /pages in the URL
	http.HandleFunc("/web", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/", http.StatusFound)
	})

	// Serve /web/list and /web/list/ as the list page
	http.HandleFunc("/web/list", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/list/", http.StatusFound)
	})

	http.HandleFunc("/web/list/", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// Serve /web/ as the main page
	http.HandleFunc("/web/", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// API endpoint
	http.HandleFunc("/stocks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handlers.CollectStock(w, r)
		case http.MethodGet:
			handlers.GetStocks(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
