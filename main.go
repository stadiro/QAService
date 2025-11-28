package main

import (
	"log"
	"net/http"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	apphttp "QAService/internal/http"
	"QAService/internal/storage"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=qa_service port=5432 sslmode=disable"
		log.Printf("DATABASE_DSN not set, using default: %s", dsn)
	}

	logger := log.New(os.Stdout, "[qa-service] ", log.LstdFlags|log.Lshortfile)
	logger.Println("initializing database connection")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	logger.Println("database connection established")

	store := storage.New(db)
	server := apphttp.NewServer(store, logger)

	addr := ":8080"
	logger.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
