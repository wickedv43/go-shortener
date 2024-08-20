package main

import (
	"log"
	"net/http"
)

func addNew(w http.ResponseWriter, r *http.Request) {
	var url []byte
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	_, err := r.Body.Read(url)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	if S.InStorage(string(url)) {
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
	}

	short := Shorting()
	S.Save(string(url), short)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("content-type", "text/plain; charset=UTF-8")

	resURL := "http://localhost:8080/" + short
	w.Write([]byte(resURL))
}

func getShort(w http.ResponseWriter, r *http.Request) {

}

func main() {
	log.Println("Starting server...")

	S.Init()
	log.Println("Initializing database...")

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, addNew)
	//mux.HandleFunc(`/{id}`, getShort)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
