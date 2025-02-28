package main

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

// task post
// common post for test
func main() {

	r := resty.New()

	cookie := &http.Cookie{
		Name:  "auth_token",
		Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzk3NDM2MTIsInVzZXJfaWQiOjMzMjZ9.WGTa3JT4kcb_S047NY9PEKLsdJMnH10z1ZiVQAizV7w",
	}

	resp, err := r.R().SetCookie(cookie).Get("http://localhost:8080/api/user/urls")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("--------------------")
	fmt.Println(string(resp.Body()))
	fmt.Println("--------------------")
	fmt.Println(resp.StatusCode(), resp.Header().Get("Content-Encoding"), resp.Header().Get("Location"))
	fmt.Println(resp.Cookies())

}
