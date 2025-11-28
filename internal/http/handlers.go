package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"QAService/internal/storage"
)

type Server struct {
	store *storage.Store
	log   *log.Logger
	mux   *http.ServeMux
}

func NewServer(store *storage.Store, logger *log.Logger) *Server {
	s := &Server{
		store: store,
		log:   logger,
		mux:   http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/questions", s.handleQuestions)
	s.mux.HandleFunc("/questions/", s.handleQuestionByID)
	s.mux.HandleFunc("/answers/", s.handleAnswerByID)
}

type createQuestionRequest struct {
	Text string `json:"text"`
}

type createAnswerRequest struct {
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}

func (s *Server) handleQuestions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listQuestions(w, r)
	case http.MethodPost:
		s.createQuestion(w, r)
	default:
		s.logRequest(r, "method not allowed for /questions")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listQuestions(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "listing questions")
	questions, err := s.store.ListQuestions(r.Context())
	if err != nil {
		s.logRequest(r, "list questions failed: %v", err)
		s.internalError(w, err)
		return
	}
	s.logRequest(r, "list questions succeeded count=%d", len(questions))
	s.respondJSON(w, http.StatusOK, questions)
}

func (s *Server) createQuestion(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r, "creating question")
	var req createQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logRequest(r, "invalid json: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		s.logRequest(r, "validation failed: empty text")
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}
	q, err := s.store.CreateQuestion(r.Context(), req.Text)
	if err != nil {
		s.logRequest(r, "create question failed: %v", err)
		s.internalError(w, err)
		return
	}
	s.logRequest(r, "question created id=%d", q.ID)
	s.respondJSON(w, http.StatusCreated, q)
}

func (s *Server) handleQuestionByID(w http.ResponseWriter, r *http.Request) {
	// Possible paths:
	// /questions/{id}
	// /questions/{id}/answers
	path := strings.TrimPrefix(r.URL.Path, "/questions/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		s.logRequest(r, "question id missing in path")
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		s.logRequest(r, "invalid question id: %q", parts[0])
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// /questions/{id}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.getQuestion(w, r, uint(id))
		case http.MethodDelete:
			s.deleteQuestion(w, r, uint(id))
		default:
			s.logRequest(r, "method %s not allowed for /questions/{id}", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	// /questions/{id}/answers
	if len(parts) == 2 && parts[1] == "answers" {
		if r.Method == http.MethodPost {
			s.createAnswer(w, r, uint(id))
			return
		}
		s.logRequest(r, "method %s not allowed for /questions/{id}/answers", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.logRequest(r, "path not found under /questions: %s", r.URL.Path)
	http.NotFound(w, r)
}

func (s *Server) getQuestion(w http.ResponseWriter, r *http.Request, id uint) {
	s.logRequest(r, "retrieving question id=%d", id)
	q, err := s.store.GetQuestionWithAnswers(r.Context(), id)
	if err != nil {
		if err == storage.ErrQuestionNotFound {
			s.logRequest(r, "question id=%d not found", id)
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		s.logRequest(r, "get question id=%d failed: %v", id, err)
		s.internalError(w, err)
		return
	}
	s.logRequest(r, "question id=%d retrieved", id)
	s.respondJSON(w, http.StatusOK, q)
}

func (s *Server) deleteQuestion(w http.ResponseWriter, r *http.Request, id uint) {
	s.logRequest(r, "deleting question id=%d", id)
	if err := s.store.DeleteQuestion(r.Context(), id); err != nil {
		s.logRequest(r, "delete question id=%d failed: %v", id, err)
		s.internalError(w, err)
		return
	}
	s.logRequest(r, "question id=%d deleted", id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createAnswer(w http.ResponseWriter, r *http.Request, questionID uint) {
	s.logRequest(r, "creating answer for question id=%d", questionID)
	var req createAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logRequest(r, "invalid json for answer: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.UserID) == "" || strings.TrimSpace(req.Text) == "" {
		s.logRequest(r, "validation failed: user_id or text empty")
		http.Error(w, "user_id and text are required", http.StatusBadRequest)
		return
	}

	a, err := s.store.CreateAnswer(r.Context(), questionID, req.UserID, req.Text)
	if err != nil {
		if err == storage.ErrQuestionNotFound {
			s.logRequest(r, "question id=%d not found while creating answer", questionID)
			http.Error(w, "question not found", http.StatusNotFound)
			return
		}
		s.logRequest(r, "create answer failed: %v", err)
		s.internalError(w, err)
		return
	}
	s.logRequest(r, "answer created id=%d for question id=%d", a.ID, questionID)
	s.respondJSON(w, http.StatusCreated, a)
}

func (s *Server) handleAnswerByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/answers/")
	id, err := strconv.ParseUint(strings.Trim(path, "/"), 10, 64)
	if err != nil {
		s.logRequest(r, "invalid answer id: %q", path)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getAnswer(w, r, uint(id))
	case http.MethodDelete:
		s.deleteAnswer(w, r, uint(id))
	default:
		s.logRequest(r, "method %s not allowed for /answers/{id}", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getAnswer(w http.ResponseWriter, r *http.Request, id uint) {
	s.logRequest(r, "retrieving answer id=%d", id)
	a, err := s.store.GetAnswer(r.Context(), id)
	if err != nil {
		if err == storage.ErrAnswerNotFound {
			s.logRequest(r, "answer id=%d not found", id)
			http.Error(w, "answer not found", http.StatusNotFound)
			return
		}
		s.logRequest(r, "get answer id=%d failed: %v", id, err)
		s.internalError(w, err)
		return
	}
	s.logRequest(r, "answer id=%d retrieved", id)
	s.respondJSON(w, http.StatusOK, a)
}

func (s *Server) deleteAnswer(w http.ResponseWriter, r *http.Request, id uint) {
	s.logRequest(r, "deleting answer id=%d", id)
	if err := s.store.DeleteAnswer(r.Context(), id); err != nil {
		s.logRequest(r, "delete answer id=%d failed: %v", id, err)
		s.internalError(w, err)
		return
	}
	s.logRequest(r, "answer id=%d deleted", id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			s.log.Printf("failed to encode json: %v", err)
		}
	}
}

func (s *Server) internalError(w http.ResponseWriter, err error) {
	s.log.Printf("internal error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func (s *Server) logRequest(r *http.Request, format string, args ...interface{}) {
	s.log.Printf("%s %s - %s", r.Method, r.URL.Path, fmt.Sprintf(format, args...))
}
