package main

import (
	"bytes"
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
)

// task post
// common post for test
func main() {
	client := &http.Client{}

	body := "https://practicum.yandex.ru/"

	req, err := http.NewRequest("POST", "http://localhost:8080/", bytes.NewReader([]byte(body)))
	if err != nil {
		err = errors.New("client post")
		logrus.Error(err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		logrus.Error(err)
	}
	defer resp.Body.Close()

}
