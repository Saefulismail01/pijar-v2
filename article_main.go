// package main

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"
// 	"os"
// 	"pijar/delivery"

// 	"github.com/joho/godotenv"
// 	_ "github.com/lib/pq"
// )

// func init() {
// 	// Load .env file
// 	if err := godotenv.Load(); err != nil {
// 		log.Printf("Warning: .env file not found or error loading: %v", err)
// 	}
// }

// func main() {
// 	// Load database configuration from environment variables
// 	dbHost := getEnv("DB_HOST", "localhost")
// 	dbPort := getEnv("DB_PORT", "5432")
// 	dbUser := getEnv("DB_USER", "postgres")
// 	dbPass := getEnv("DB_PASS", "")
// 	dbName := getEnv("DB_NAME", "")

// 	if dbName == "" {
// 		log.Fatal("DB_NAME environment variable is required")
// 	}
// 	if dbPass == "" {
// 		log.Fatal("DB_PASSWORD environment variable is required")
// 	}

// 	// Construct database connection string
// 	dbConnectionString := fmt.Sprintf(
// 		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		dbHost, dbPort, dbUser, dbPass, dbName,
// 	)

// 	// Connect to database
// 	db, err := sql.Open("postgres", dbConnectionString)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}
// 	defer db.Close()

// 	// Configure connection pool
// 	db.SetMaxOpenConns(25)
// 	db.SetMaxIdleConns(5)

// 	// Test database connection
// 	err = db.Ping()
// 	if err != nil {
// 		log.Fatalf("Failed to ping database: %v", err)
// 	}
// 	log.Println("Successfully connected to database")

// 	// Create and initialize server
// 	server, err := delivery.NewServer(db)
// 	if err != nil {
// 		log.Fatalf("Failed to create server: %v", err)
// 	}

// 	// Run server
// 	if err := server.Run(); err != nil {
// 		log.Fatalf("Server error: %v", err)
// 	}
// }

// func getEnv(key, fallback string) string {
// 	if value, exists := os.LookupEnv(key); exists && value != "" {
// 		return value
// 	}
// 	return fallback
// }

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"pijar/delivery"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading: %v", err)
	}
}

func main() {
	// Load database configuration from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASS", "")
	dbName := getEnv("DB_NAME", "")

	if dbName == "" {
		log.Fatal("DB_NAME environment variable is required")
	}
	if dbPass == "" {
		log.Fatal("DB_PASSWORD environment variable is required")
	}

	// Construct database connection string
	dbConnectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName,
	)

	// Connect to database
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Test database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Create and initialize server
	server, err := delivery.NewServer(db)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Run server
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
