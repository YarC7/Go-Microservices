package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

// GetDB returns the database connection
func GetDB() *sql.DB {
	if db != nil {
		return db
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5436") // Different port for payment service
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "canh177")
	dbname := getEnv("DB_NAME", "payment_db")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	log.Println("Successfully connected to payment database")
	return db
}

// InitSchema creates the necessary tables
func InitSchema(database *sql.DB) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS payments (
		id SERIAL PRIMARY KEY,
		order_id INTEGER NOT NULL,
		customer_id INTEGER NOT NULL,
		amount DECIMAL(10, 2) NOT NULL,
		currency VARCHAR(3) NOT NULL DEFAULT 'USD',
		status VARCHAR(50) NOT NULL,
		stripe_payment_id VARCHAR(255),
		stripe_client_secret VARCHAR(255),
		payment_method VARCHAR(50),
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
	CREATE INDEX IF NOT EXISTS idx_payments_customer_id ON payments(customer_id);
	CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
	CREATE INDEX IF NOT EXISTS idx_payments_stripe_payment_id ON payments(stripe_payment_id);
	`

	_, err := database.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Failed to create payments table: ", err)
	}

	log.Println("Payment database schema initialized successfully")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}