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

var rs Response

// task post
// common post for test
func main() {

	r := resty.New()

	cookie := &http.Cookie{
		Name:  "auth_token",
		Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzk3NDM2MTIsInVzZXJfaWQiOjMzMjZ9.WGTa3JT4kcb_S047NY9PEKLsdJMnH10z1ZiVQAizV7w",
	}

	//resp, err := r.R().SetCookie(cookie).
	//	SetHeader("Content-Type", "application/json").
	//	SetBody(`[{"correlation_id":"eff56730-ff23-4421-9692-f383d2d10c03","original_url":"http://h86daojnfcstn.biz/vrb8yp7ic91ij/soif137sky38"},{"correlation_id":"0fd7fc27-beac-41b9-a2c6-c281c06b7c02","original_url":"http://pqs1sfjqk6wfye.net/ngeh7f35unj/euwkn8jd"}]`).
	//	Post("http://localhost:8080/api/shorten/batch")
	//if err != nil {
	//	fmt.Println(err)
	//}

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
