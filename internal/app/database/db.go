package database

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func ConnectDatabase() *sql.DB {
	godotenv.Load()

	username := os.Getenv("POSTGRES_USERNAME")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DATABASE")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}

	return db
}
