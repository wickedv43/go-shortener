package main

import (
	"fmt"

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

	r := resty.New()

	resp, err := r.R().
		SetHeader("Content-Type", "text/plain").
		SetBody("https://www.google.com").
		Post("http://localhost:8080/")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("--------------------")
	fmt.Println(string(resp.Body()))
	fmt.Println("--------------------")
	fmt.Println(resp.StatusCode(), resp.Header().Get("Content-Encoding"), resp.Header().Get("Location"))
	fmt.Println(resp.Cookies())

}
