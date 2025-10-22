package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type IngressoType string

const (
	Predetermined   IngressoType = "predeterminado"
	UnlimitedDayUse IngressoType = "day-use"
	Passport        IngressoType = "passport"
)

type ingresso struct {
	ID     string
	UserId string
	Type   IngressoType
}

func main() {
	db, err := sql.Open("sqlite", "usuarios")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		fmt.Printf("failed to connect to db: %v", err)
	}
	createTable := `
	CREATE TABLE IF NOT EXISTS ingressos (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		type TEXT NOT NULL,
		remaining_accesses INTEGER,
		valid_until TEXT,
		created_at TEXT NOT NULL
	);
	`
	if _, err := db.Exec(createTable); err != nil {
		fmt.Printf("Failed to create ingressos table: %v\n", err)
		return
	}

	handler := &IngressoHandler{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ingresso/{userId}", handler.CheckIfUserHasTicket)
}

type IngressoHandler struct {
	db *sql.DB
}

func (u *IngressoHandler) CheckIfUserHasTicket(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query()
	userq := url.Get("user")
	if userq == "" {
		fmt.Println("user query cant be empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp, err := u.db.Query("SELECT id, user_id, type FROM ingressos WHERE user_id = ?;", userq)
	if err != nil {
		fmt.Printf("failed to get ingresso: %v", err)
	}
	defer resp.Close()

	ingressos := []ingresso{}
	for resp.Next() {
		var ingr ingresso
		err := resp.Scan(&ingr.ID, &ingr.UserId, &ingr.Type)
		if err != nil {
			fmt.Printf("error scanning ingresso: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error scanning ingresso"))
			return

		}
		ingressos = append(ingressos, ingr)
	}
	jsonRes, err := json.Marshal(ingressos)
	if err != nil {
		fmt.Println("error marshling to json")
		return
	}

	w.Write(jsonRes)
}
