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

	w.Header().Set("Content-Type", "text/plain")

	resURL := "http://localhost:8080/" + short
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resURL))

}

func getShort(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	url := r.URL.String()
	short := strings.TrimPrefix(url, "/")

	respURL, ok := S.Get(short)
	if !ok {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	log.Println("GET:", short, "RETURN:", respURL)
	w.Header().Set("Location", respURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
