package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
)

func main() {
	client := &http.Client{}

	body := "https://practicum.yandex.ru/"

	req, err := http.NewRequest("POST", "http://localhost:8080/", bytes.NewReader([]byte(body)))
	if err != nil {
		err = errors.New("client post")
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
}
