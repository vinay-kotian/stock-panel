package main

import (
	"log"
	"net/http"

	"github.com/vinaykotian/stock-panel/internal/db"
	"github.com/vinaykotian/stock-panel/internal/handlers"
)

func main() {
	db.InitDB()
	defer db.DB.Close()

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
