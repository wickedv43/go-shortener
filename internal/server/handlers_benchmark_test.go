package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/labstack/echo/v4"
	"github.com/wickedv43/go-shortener/internal/mocks"
	"github.com/wickedv43/go-shortener/internal/storage"
	"go.uber.org/mock/gomock"
)

func BenchmarkGetShort(b *testing.B) {
	srv := setupTestServer(b, func(mock *mocks.MockDataKeeper) {
		mock.EXPECT().
			Get(gomock.Any(), "abc123").
			Return(storage.Data{
				OriginalURL: "https://example.com",
			}, nil).
			AnyTimes()
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		rec := httptest.NewRecorder()
		ctx := srv.echo.NewContext(req, rec)
		ctx.SetParamNames("short")
		ctx.SetParamValues("abc123")

		_ = srv.getShort(ctx)
	}
}
