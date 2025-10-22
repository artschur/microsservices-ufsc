package main

import (
	"encoding/json"
	"errors"
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
	atracoes, err := getAttractions()
	if err != nil {
		panic(err)
	}
	for _, atracao := range atracoes {
		println("Atracao:", atracao.Id, "Capacidade:", atracao.Capacidade)
	}
	queueMap := initAtraction(atracoes)

	mux := http.NewServeMux()
	queueHandler := &QueueHandler{
		db: queueMap,
	}
	go decrementAllQueuesWorker(queueMap)
	mux.HandleFunc("POST /fila/increment", queueHandler.HandleIncrement)
	mux.HandleFunc("GET /fila", queueHandler.GetAllQueues)

	println("Fila service running on :8081")
	http.ListenAndServe(":8081", mux)
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
	json.NewEncoder(w).Encode(h.db)
	w.WriteHeader(http.StatusOK)
}
func initAtraction(atracoes []atracao) map[string]atracao {
	var queueMap = make(map[string]atracao, len(atracoes))
	for _, attr := range atracoes {
		queueMap[attr.Id] = atracao{
			Id:           attr.Id,
			Nome:         attr.Nome,
			Capacidade:   attr.Capacidade,
			CurrentQueue: 0,
		}
	}
	return queueMap
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
	resp, err := http.Get("http://localhost:8080/atracoes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
