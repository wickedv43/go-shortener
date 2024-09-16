package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

var rs Response

// task post
// common post for test
func main() {
	client := &http.Client{}

	var r Request

	r.URL = "https://practicum.yandex.ru/"
	body, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r)

	req, err := http.NewRequest("POST", "http://localhost:8080/api/shorten", bytes.NewReader(body))
	if err != nil {
		err = errors.New("client post")
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	rBody, err := io.ReadAll(res.Body)
	_ = json.Unmarshal(rBody, &rs)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	fmt.Println(rs)

}
