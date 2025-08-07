package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "stocks.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	createTable := `CREATE TABLE IF NOT EXISTS stocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		underlying_symbol TEXT,
		option_type TEXT,
		strike_price REAL,
		expiry TEXT,
		price REAL NOT NULL,
		side TEXT,
		timestamp DATETIME NOT NULL
	);`
	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}
