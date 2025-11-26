package http

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"QAService/internal/models"
	"QAService/internal/storage"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=qa_service_test port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open postgres: %v", err)
	}
	if err := db.AutoMigrate(&models.Question{}, &models.Answer{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	store := storage.New(db)
	logger := log.New(testingWriter{t: t}, "[test] ", log.LstdFlags)
	return NewServer(store, logger)
}

type testingWriter struct {
	t *testing.T
}

func (tw testingWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
}

func TestCreateAndGetQuestion(t *testing.T) {
	server := setupTestServer(t)

	body, _ := json.Marshal(map[string]string{"text": "Test question"})
	req := httptest.NewRequest(http.MethodPost, "/questions", bytes.NewReader(body))
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var created models.Question
	if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/questions/"+strconv.Itoa(int(created.ID)), nil)
	getW := httptest.NewRecorder()
	server.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getW.Code)
	}
}
