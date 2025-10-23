package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type IngressoType string

const (
	Predetermined   IngressoType = "predeterminado"
	UnlimitedDayUse IngressoType = "day-use"
	Passport        IngressoType = "passport"
)

type Ingresso struct {
	ID                string         `json:"id"`
	UserId            string         `json:"user_id"`
	Type              IngressoType   `json:"type"`
	RemainingAccesses sql.NullInt64  `json:"remaining_accesses,omitempty"`
	ValidUntil        sql.NullString `json:"valid_until,omitempty"`
	CreatedAt         string         `json:"created_at"`
}

type SellTicketRequest struct {
	UserID string       `json:"user_id"`
	Type   IngressoType `json:"type"`
}

func main() {
	db, err := sql.Open("sqlite", "ingressos.db")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		fmt.Printf("failed to connect to db: %v", err)
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS ingressos (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		remaining_accesses INTEGER,
		valid_until DATE,
		created_at TEXT NOT NULL
	);
	`
	if _, err := db.Exec(createTable); err != nil {
		fmt.Printf("Failed to create ingressos table: %v\n", err)
		return
	}

	handler := &IngressoHandler{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ingresso/{userId}", handler.GetUserTicket)
	mux.HandleFunc("POST /ingresso/validate", handler.ValidateTicket)
	mux.HandleFunc("POST /ingresso/sell", handler.SellTicket)

	fmt.Println("Ingressos service running on :8080")
	http.ListenAndServe(":8080", mux)
}

type IngressoHandler struct {
	db *sql.DB
}

func (h *IngressoHandler) SellTicket(w http.ResponseWriter, r *http.Request) {
	var req SellTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.Type == "" {
		http.Error(w, "user_id and type are required", http.StatusBadRequest)
		return
	}

	newIngresso := Ingresso{
		ID:        uuid.NewString(),
		UserId:    req.UserID,
		Type:      req.Type,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	switch req.Type {
	case Predetermined:
		newIngresso.RemainingAccesses = sql.NullInt64{Int64: 5, Valid: true}
		// No expiration for predetermined tickets
		newIngresso.ValidUntil = sql.NullString{Valid: false}
	case UnlimitedDayUse:
		// Expires at the end of the day
		newIngresso.ValidUntil = sql.NullString{
			String: time.Now().Format("2006-01-02"),
			Valid:  true,
		}
	case Passport:
		// Passport could have a longer validity, e.g., 1 year
		newIngresso.ValidUntil = sql.NullString{
			String: time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
			Valid:  true,
		}
	default:
		http.Error(w, "invalid ticket type", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO ingressos (id, user_id, type, remaining_accesses, valid_until, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := h.db.Exec(
		query,
		newIngresso.ID,
		newIngresso.UserId,
		newIngresso.Type,
		newIngresso.RemainingAccesses,
		newIngresso.ValidUntil,
		newIngresso.CreatedAt,
	)

	if err != nil {
		log.Printf("error inserting ingresso: %v", err)
		http.Error(w, "failed to create ticket", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newIngresso)
}

func (h *IngressoHandler) ValidateTicket(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userId")
	if userId == "" {
		http.Error(w, "userId query parameter is required", http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	query := `
		SELECT id, user_id, type, remaining_accesses, valid_until
		FROM ingressos
		WHERE user_id = ? AND (valid_until IS NULL OR valid_until >= date('now'))
	`
	row := tx.QueryRow(query, userId)

	var ingresso Ingresso
	err = row.Scan(&ingresso.ID, &ingresso.UserId, &ingresso.Type, &ingresso.RemainingAccesses, &ingresso.ValidUntil)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "no valid ticket found for user", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to query ticket", http.StatusInternalServerError)
		return
	}

	if ingresso.Type == Predetermined {
		if !ingresso.RemainingAccesses.Valid || ingresso.RemainingAccesses.Int64 <= 0 {
			http.Error(w, "no remaining accesses for this ticket", http.StatusBadRequest)
			return
		}
		_, err := tx.Exec("UPDATE ingressos SET remaining_accesses = remaining_accesses - 1 WHERE id = ?", ingresso.ID)
		if err != nil {
			http.Error(w, "failed to decrement remaining accesses", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ticket validated successfully for user %s", userId)
}

func (h *IngressoHandler) GetUserTicket(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "" {
		http.Error(w, "userId path parameter is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	row := h.db.QueryRow("SELECT id, user_id, type, remaining_accesses, valid_until, created_at FROM ingressos WHERE user_id = ?;", userId)

	var ingr Ingresso
	err := row.Scan(&ingr.ID, &ingr.UserId, &ingr.Type, &ingr.RemainingAccesses, &ingr.ValidUntil, &ingr.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Printf("error scanning ingresso: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	jsonRes, err := json.Marshal(ingr)
	if err != nil {
		fmt.Println("error marshling to json")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Write(jsonRes)
}
