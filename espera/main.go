package main

import (
	"encoding/json"
	"fmt"
	"log"
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

	t := time.NewTicker(30 * time.Second)
	go func() {
		for {
			<-t.C
			err := handler.calculateWaitingTime()
			if err != nil {
				log.Println("error calculating waiting time:", err)
				continue
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /espera/{attractionId}", handler.GetWaitingTimeForAttraction)

	http.ListenAndServe(":8084", mux)
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

func (h *WaitingTimeHandler) calculateWaitingTime() error {
	// Call fila service directly instead of via gateway
	resp, err := http.Get("http://fila:8083/fila")
	if err != nil {
		return fmt.Errorf("failed to fetch attraction queues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fila returned status %d", resp.StatusCode)
	}

	var attractionQueues map[string]atracao
	if err := json.NewDecoder(resp.Body).Decode(&attractionQueues); err != nil {
		return fmt.Errorf("failed to decode attraction queues: %w", err)
	}
	for _, attraction := range attractionQueues {
		h.db[attraction.Id] = attraction.CurrentQueue * 15
	}
	return nil
}
