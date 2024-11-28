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

	bidy := `[{
        "correlation_id": "<33>",
        "original_url": "<https://p2racticum.yandex.ru/learn>"
    },{
        "correlation_id": "<икатор>",
        "original_url": "<https://pra3cticum.yandex.ru/learn/go-advanced/courses/6e4a1d46-9b38-4936-93c4-62f9ec2db45a/sprints/366447/topics/f04453a3-f8c6-4b19-bb87-454f61520c4e/lessons/c5404109-dc98-4636-ae51-3c3d284b129f/>"
    }]`

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
