package main

import (
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
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
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		Get(`http://localhost:8080/api/user/urls`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(resp.Body()), resp.StatusCode())

}
