package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/storage"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
)

var i = do.New()

func init() {
	do.Provide(i, NewServer)
	do.Provide(i, config.NewConfig)
	do.Provide(i, logger.NewLogger)

	//storages
	do.Provide(i, storage.NewLocalStorage)
	do.Provide(i, storage.NewFileStorage)
	do.Provide(i, storage.NewPostgresStorage)
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
			body := "https://practicum.yandex.ru/test"
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))

			w := httptest.NewRecorder()

			srv.echo.ServeHTTP(w, request)
			res := w.Result()
			require.Equal(t, test.want.code, res.StatusCode)

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, resBody)

			err = res.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))

			os.Remove(srv.cfg.Server.FlagStoragePath)
		})

	}
}
