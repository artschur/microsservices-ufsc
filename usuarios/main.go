package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Telefone string `json:"telefone"`
}

func main() {
	db, err := sql.Open("sqlite", "./usuarios.db")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS usuarios (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		telefone TEXT NOT NULL
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatalf("Failed to create usuarios table: %v", err)
	}

	userHandler := &UserHandler{
		db: db,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /user", userHandler.GetUser)
	mux.HandleFunc("POST /user", userHandler.CreateUser)

	fmt.Println("Usuarios service running on :8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

type UserHandler struct {
	db *sql.DB
}

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Error decoding user", http.StatusBadRequest)
		return
	}

	if newUser.ID == "" || newUser.Name == "" || newUser.Email == "" || newUser.Telefone == "" {
		http.Error(w, "id, name, email and telefone are required fields", http.StatusBadRequest)
		return
	}

	// Correct INSERT statement for the "usuarios" table
	insertSQL := "INSERT INTO usuarios (id, name, email, telefone) VALUES (?, ?, ?, ?)"
	_, err = u.db.Exec(insertSQL, newUser.ID, newUser.Name, newUser.Email, newUser.Telefone)
	if err != nil {
		http.Error(w, "Error inserting user into db", http.StatusInternalServerError)
		return
	}

	// Return the created user object as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newUser); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (u *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("id")
	if userId == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Correct SELECT statement for the "usuarios" table
	querySQL := "SELECT id, name, email, telefone FROM usuarios WHERE id = ?;"
	row := u.db.QueryRow(querySQL, userId)

	var user User
	// Scan all fields from the row
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Telefone)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	// Return the found user as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
