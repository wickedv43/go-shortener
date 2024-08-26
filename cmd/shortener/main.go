package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Starting server...")

	S.Init()
	log.Println("Initializing database...")

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, addNew)
	mux.HandleFunc(`/{id}`, getShort)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
