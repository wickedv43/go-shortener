package main

import (
	"io"
	"log"
	"net/http"
	"strings"
)

func addNew(w http.ResponseWriter, r *http.Request) {
	var url []byte
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	var short string
	ok, short := S.InStorage(string(url))
	if !ok {
		short = Shorting()
	}
	S.Save(string(url), short)
	log.Println(string(url), short)

	w.WriteHeader(http.StatusCreated)

	w.Header().Set("content-type", "text/plain; charset=UTF-8")

	resURL := "http://localhost:8080/" + short
	w.Write([]byte(resURL))
}

func getShort(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	url := r.URL.String()
	log.Println(url)
	short := strings.TrimPrefix(url, "/")
	log.Println(short)

	respUrl, ok := S.Get(short)
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	log.Println("GET:", short, "RETURN:", respUrl)

	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Header().Set("content-type", "text/plain; charset=UTF-8")
	w.Write([]byte(respUrl))
}

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
