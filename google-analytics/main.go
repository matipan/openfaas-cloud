package main

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func main() {
	c := NewClient(uuid.New().String(), "UA-139317174-1", "ofc", "1")
	go c.ListenAndSend(context.TODO())

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/healthz", healthzHandler)
	r.HandleFunc("/report", c.ReportHandler)

	log.Println("Listening on port 8080")
	http.ListenAndServe(":8080", r)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
