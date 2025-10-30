package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Function to connect to the PostgreSQL Database, connects to URL using pgx & returns the connection
// if successful
func Connect(dbURL string) *sql.DB {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to open DB connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	fmt.Println("Successful database connection")

	return db
}
