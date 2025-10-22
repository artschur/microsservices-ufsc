package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Atracao struct {
	ID               string `json:"id"`
	Nome             string `json:"nome"`
	Capacidade       int    `json:"capacidade"`
	TempoMedioEspera int    `json:"tempo_medio_espera"`
	CreatedAt        string `json:"created_at"`
}

func main() {
	db, err := sql.Open("sqlite3", "atracoes.db")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS atracoes (
		id TEXT PRIMARY KEY,
		nome TEXT NOT NULL,
		capacidade INTEGER NOT NULL,
		tempoMedioEspera INTEGER NOT NULL,
		created_at TEXT NOT NULL
	);
	`
	if _, err := db.Exec(createTable); err != nil {
		panic(err)
	}

	handler := &AtracaoHandler{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /atracoes", handler.GetAllAtracoes)

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
