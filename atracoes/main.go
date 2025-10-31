package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	_ "modernc.org/sqlite"
)

type Atracao struct {
	ID string `json:"id"`
	CreateAtracaoRequest
	CreatedAt string `json:"created_at"`
}

type CreateAtracaoRequest struct {
	Nome             string `json:"nome"`
	Capacidade       int    `json:"capacidade"`
	TempoMedioEspera int    `json:"tempo_medio_espera"`
}

func main() {
	db, err := sql.Open("sqlite", "atracoes.db")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS atracoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nome TEXT NOT NULL,
		capacidade INTEGER NOT NULL,
		tempoMedioEspera INTEGER NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);
	`
	if _, err := db.Exec(createTable); err != nil {
		panic(err)
	}

	handler := &AtracaoHandler{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /atracoes", handler.GetAllAtracoes)
	mux.HandleFunc("POST /atracoes", handler.CreateAtracao)

	fmt.Println("Atracoes service running on :8081")
	http.ListenAndServe(":8081", mux)
}

type AtracaoHandler struct {
	db *sql.DB
}

func (h *AtracaoHandler) GetAllAtracoes(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, nome, capacidade, tempoMedioEspera, created_at FROM atracoes")
	if err != nil {
		http.Error(w, "failed to query attractions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var atracoes []Atracao
	for rows.Next() {
		var a Atracao
		if err := rows.Scan(&a.ID, &a.Nome, &a.Capacidade, &a.TempoMedioEspera, &a.CreatedAt); err != nil {
			http.Error(w, "failed to scan attraction", http.StatusInternalServerError)
			return
		}
		atracoes = append(atracoes, a)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(atracoes); err != nil {
		http.Error(w, "failed to encode attractions", http.StatusInternalServerError)
	}
}

func (h *AtracaoHandler) CreateAtracao(w http.ResponseWriter, r *http.Request) {
	var newAtracao CreateAtracaoRequest
	err := json.NewDecoder(r.Body).Decode(&newAtracao)
	if err != nil {
		http.Error(w, "Error decoding attraction", http.StatusBadRequest)
		return
	}

	if newAtracao.Nome == "" || newAtracao.Capacidade <= 0 || newAtracao.TempoMedioEspera <= 0 {
		http.Error(w, "nome, capacidade and tempoMedioEspera are required fields", http.StatusBadRequest)
		return
	}

	insertSQL := "INSERT INTO atracoes (nome, capacidade, tempoMedioEspera) VALUES (?, ?, ?)"
	_, err = h.db.Exec(insertSQL, newAtracao.Nome, newAtracao.Capacidade, newAtracao.TempoMedioEspera)
	if err != nil {
		slog.Error("Error inserting attraction into db", "error", err)
		http.Error(w, "Error inserting attraction into db", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(newAtracao); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
