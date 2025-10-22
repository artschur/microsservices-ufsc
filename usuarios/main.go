package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type user struct {
	ID       string
	name     string
	email    string
	telefone string
}

func main() {
	db, err := sql.Open("sqlite", "./usuarios.db")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		fmt.Printf("failed to connect to db: %v", err)
	}
	_, err = db.Exec("CREATE IF NOT EXISTS usuarios ;")
	if err != nil {
		fmt.Println("Failed to create usuarios database")
	}

	userHandler := UserHandler{
		db: db,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /user", userHandler.GetUser)
	mux.HandleFunc("POST /user", userHandler.CreateUser)

}

type UserHandler struct {
	db *sql.DB
}

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser user
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		fmt.Printf("error decoding user: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if newUser.name == "" || newUser.email == "" || newUser.telefone == "" {
		http.Error(w, "name, email and telefone are required fields", http.StatusBadRequest)
		return
	}

	res, err = u.db.Exec("INSERT INTO users (id, name, email) VALUES (?, ?, ?) RETURNING id;", newUser.ID, newUser.name, newUser.email)
	if err != nil {
		fmt.Printf("error inserting user into db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.
}

func (u *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query()
	userId := url.Get("id")
	if userId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := u.db.Query("SELECT id, name, email FROM users WHERE id = ?;", userId)
	if err != nil {
		fmt.Printf("failed to get users: %v", err)
	}
	defer resp.Close()

	var user user
	for resp.Next() {
		err := resp.Scan(&user.ID, &user.name, &user.email)
		if err != nil {
			fmt.Printf("error scanning user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	jsonRes, err := json.Marshal(user)
	if err != nil {
		fmt.Println("error marshling to json")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonRes)

}
