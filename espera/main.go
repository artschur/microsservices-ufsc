package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type atracao struct {
	Id           string `json:"id"`
	Nome         string `json:"nome"`
	Capacidade   int    `json:"capacidade"`
	CurrentQueue int    `json:"currentQueue"`
}

func main() {
	// id da atracao e tempo de espera em minutos
	waitingTimeMap := make(map[string]int)
	handler := &WaitingTimeHandler{
		db: waitingTimeMap,
	}
	handler.calculateWaitingTime()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /espera/{attractionId}", handler.GetWaitingTimeForAttraction)

	http.ListenAndServe(":8082", mux)
}

type WaitingTimeHandler struct {
	db map[string]int
}

func (h *WaitingTimeHandler) GetWaitingTimeForAttraction(w http.ResponseWriter, r *http.Request) {
	attractionId := r.PathValue("attractionId")
	waitingTime, exists := h.db[attractionId]
	if !exists {
		http.Error(w, "attraction not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"waiting_time": waitingTime})
}

func (h *WaitingTimeHandler) calculateWaitingTime() {
	resp, err := http.Get("http://gateway:8000/fila")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var attractionQueues map[string]atracao
	if err := json.NewDecoder(resp.Body).Decode(&attractionQueues); err != nil {
		panic(err)
	}
	for _, attraction := range attractionQueues {
		h.db[attraction.Id] = attraction.CurrentQueue * 15
	}

	time.Sleep(30 * time.Second)
}
