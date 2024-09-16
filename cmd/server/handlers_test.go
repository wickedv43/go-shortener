package server

import (
	"bytes"
	"encoding/json"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/cmd/config"
	"github.com/wickedv43/go-shortener/cmd/logger"
	"github.com/wickedv43/go-shortener/cmd/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var i = do.New()

func init() {
	do.Provide(i, NewServer)
	do.Provide(i, config.NewConfig)
	do.Provide(i, storage.NewStorage)
	do.Provide(i, logger.NewLogger)
}

// Test for "/"
func Test_addNew(t *testing.T) {
	var srv = do.MustInvoke[*Server](i)

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

			body := "https://practicum.yandex.ru/"
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))

			w := httptest.NewRecorder()

			srv.engine.ServeHTTP(w, request)
			res := w.Result()
			require.Equal(t, test.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, resBody)

			err = res.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

		})

	}
}

// Test for "/:short"
func Test_getShort(t *testing.T) {
	var srv = do.MustInvoke[*Server](i)

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

			url := "https://practicum.yandex.ru/"
			short := Shorting()
			srv.storage.Put(url, short)

			req := httptest.NewRequest(http.MethodGet, "/"+short, nil)

			w := httptest.NewRecorder()

			srv.engine.ServeHTTP(w, req)

			res := w.Result()

			err := res.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.code, res.StatusCode)
			require.Equal(t, url, res.Header.Get("Location"))
		})
	}
}

func Test_addNewJSON(t *testing.T) {
	var srv = do.MustInvoke[*Server](i)

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
				contentType: "application/json; charset=utf-8",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var r Expand
			r.URL = "https://practicum.yandex.ru/"

			body, err := json.Marshal(r)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))

			w := httptest.NewRecorder()

			srv.engine.ServeHTTP(w, req)

			res := w.Result()

			err = res.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.code, res.StatusCode)
			require.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}
