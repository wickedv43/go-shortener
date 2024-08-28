package main

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_addNew(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive test #1",
			want: want{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			S.Init()

			body := "https://practicum.yandex.ru/"
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))

			w := httptest.NewRecorder()
			addNew(w, request)

			res := w.Result()

			require.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			require.NotEmpty(t, resBody)
			require.Contains(t, string(resBody), "http://localhost:8080/")
			require.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func Test_getShort(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "positive test #1",
			want: want{
				code: http.StatusTemporaryRedirect,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			S.Init()

			body := "https://practicum.yandex.ru/"
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))

			w := httptest.NewRecorder()
			addNew(w, request)

			res := w.Result()

			require.Equal(t, http.StatusCreated, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, resBody)

			resBodyStr := string(resBody)
			short, ok := strings.CutPrefix(resBodyStr, "http://localhost:8080/")
			require.True(t, ok)

			req := httptest.NewRequest(http.MethodGet, "/"+short, nil)
			w = httptest.NewRecorder()
			getShort(w, req)
			res = w.Result()
			require.Equal(t, test.want.code, res.StatusCode)
			require.Equal(t, body, res.Header.Get("Location"))
		})
	}
}
