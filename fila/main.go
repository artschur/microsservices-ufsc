package main

import (
	"encoding/json"
	"errors"
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
	// Start with an empty map and load attractions in the background with retries
	queueMap := make(map[string]atracao)

	mux := http.NewServeMux()
	queueHandler := &QueueHandler{
		db: queueMap,
	}

	// Load attractions with retry and then start the worker
	go func() {
		for {
			atracoes, err := getAttractions()
			if err != nil {
				log.Printf("waiting for atracoes service: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}
			for _, a := range atracoes {
				queueMap[a.Id] = atracao{
					Id:           a.Id,
					Nome:         a.Nome,
					Capacidade:   a.Capacidade,
					CurrentQueue: 0,
				}
			}
			go decrementAllQueuesWorker(queueMap)
			return
		}
	}()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /fila/increment", queueHandler.HandleIncrement)
	mux.HandleFunc("GET /fila", queueHandler.GetAllQueues)

	println("Fila service running on :8083")
	http.ListenAndServe(":8083", mux)
}

type QueueHandler struct {
	db map[string]atracao
}

func (h *QueueHandler) HandleIncrement(w http.ResponseWriter, r *http.Request) {
	attractionId := r.URL.Query().Get("id")
	if attractionId == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	currAttr, exists := h.db[attractionId]
	if !exists {
		http.Error(w, "attraction not found", http.StatusNotFound)
		return
	}

	h.db[attractionId] = atracao{
		Id:           currAttr.Id,
		Nome:         currAttr.Nome,
		Capacidade:   currAttr.Capacidade,
		CurrentQueue: currAttr.CurrentQueue + 1,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(h.db[attractionId])
}

func (h *QueueHandler) GetAllQueues(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.db)
}

func decrementAllQueuesWorker(queueMap map[string]atracao) {
	for {
		for id, attr := range queueMap {
			if attr.CurrentQueue > 0 {
				queueMap[id] = atracao{
					Id:           attr.Id,
					Nome:         attr.Nome,
					Capacidade:   attr.Capacidade,
					CurrentQueue: attr.CurrentQueue - 1,
				}
			}
		}
		time.Sleep(15 * time.Second)
	}
}

func getAttractions() ([]atracao, error) {
	// Call atracoes service directly instead of via gateway
	resp, err := http.Get("http://atracoes:8081/atracoes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("atracoes returned status %d", resp.StatusCode)
	}

	var atracoes []atracao
	err = json.NewDecoder(resp.Body).Decode(&atracoes)
	if err != nil {
		return nil, err
	}
	if len(atracoes) == 0 {
		return nil, errors.New("no attractions found")
	}
	return atracoes, nil
}
