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

	// Create stocks table
	createStocksTable := `CREATE TABLE IF NOT EXISTS stocks (
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
	_, err = DB.Exec(createStocksTable)
	if err != nil {
		log.Fatalf("Failed to create stocks table: %v", err)
	}

	// Create users table
	createUsersTable := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = DB.Exec(createUsersTable)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	// Create alerts table
	createAlertsTable := `CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		underlying_symbol TEXT,
		option_type TEXT,
		strike_price REAL,
		expiry TEXT,
		alert_type TEXT NOT NULL,
		target_value REAL NOT NULL,
		condition TEXT NOT NULL,
		message TEXT,
		is_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		user_id INTEGER NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (id)
	);`
	_, err = DB.Exec(createAlertsTable)
	if err != nil {
		log.Fatalf("Failed to create alerts table: %v", err)
	}

	// Insert default users if they don't exist
	insertDefaultUsers := `INSERT OR IGNORE INTO users (username, email, password_hash) VALUES 
		('admin', 'admin@example.com', 'password123'),
		('demo', 'demo@example.com', 'demo123'),
		('vinay', 'vdkotian1@gmail.com', 'password123');`
	_, err = DB.Exec(insertDefaultUsers)
	if err != nil {
		log.Printf("Warning: Failed to insert default users: %v", err)
	}

	log.Printf("✅ Database initialized successfully")
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (string, error) {
	var username string
	err := DB.QueryRow("SELECT username FROM users WHERE email = ?", email).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (string, string, error) {
	var email, passwordHash string
	err := DB.QueryRow("SELECT email, password_hash FROM users WHERE username = ?", username).Scan(&email, &passwordHash)
	if err != nil {
		return "", "", err
	}
	return email, passwordHash, nil
}

// GetAllUsers retrieves all users (username, email pairs)
func GetAllUsers() (map[string]string, error) {
	rows, err := DB.Query("SELECT username, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make(map[string]string)
	for rows.Next() {
		var username, email string
		if err := rows.Scan(&username, &email); err != nil {
			return nil, err
		}
		users[username] = email
	}

	return users, nil
}

// CreateUser creates a new user
func CreateUser(username, email, passwordHash string) error {
	_, err := DB.Exec("INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		username, email, passwordHash)
	return err
}

// UpdateUserPassword updates a user's password
func UpdateUserPassword(username, newPasswordHash string) error {
	_, err := DB.Exec("UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE username = ?",
		newPasswordHash, username)
	return err
}
