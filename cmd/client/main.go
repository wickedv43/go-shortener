package main

import (
	"bytes"
	"compress/gzip"
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

	bidy := `[ 
{
        "correlation_id": "<11>",
        "original_url": "fff"
    },
{
        "correlation_id": "<12>",
        "original_url": "ff3"
    }
]`

	_, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r)

	req, err := http.NewRequest("POST", "http://localhost:8080/api/shorten/batch", bytes.NewReader([]byte(bidy)))
	if err != nil {
		err = errors.New("client post")
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Accept-Encoding", "gzip")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	//gzip reader
	bodyGZIP, err := gzip.NewReader(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	rBody, err := io.ReadAll(bodyGZIP)
	_ = json.Unmarshal(rBody, &rs)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	fmt.Println(rs)
	fmt.Println(string(rBody))
	fmt.Println(res.StatusCode, res.Header.Get("Content-Encoding"), res.Header.Get("Location"))

}
