package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type atracao struct {
	Id                   string `json:"id"`
	Nome                 string `json:"nome"`
	TempoEsperaPorPessoa int    `json:"tempoEsperaPorPessoa"`
	Capacidade           int    `json:"capacidade"`
}

func main() {
	atracoes, err := getAttractions()
	if err != nil {
		panic(err)
	}
	for _, atracao := range atracoes {
		println("Atracao:", atracao.Id, "Tempo de espera por pessoa:", atracao.TempoEsperaPorPessoa)
	}
	queueMap := initAtraction(atracoes)

	mux := http.NewServeMux()
	queueHandler := &QueueHandler{
		db: queueMap,
	}
	mux.HandleFunc("/increment", queueHandler.HandleIncrement)
}

type QueueHandler struct {
	db map[string]int
}

func (h *QueueHandler) HandleIncrement(w http.ResponseWriter, r *http.Request) {
	attractionId := r.URL.Query().Get("id")
	if attractionId == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	h.db[attractionId]++
	w.WriteHeader(http.StatusOK)

}

func initAtraction(atracoes []atracao) map[string]int {
	var queueMap = make(map[string]int, len(atracoes))
	for _, atracao := range atracoes {
		queueMap[atracao.Id] = 0
	}
	return queueMap
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
