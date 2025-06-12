package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/auth"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/storage"
)

type fakeShortener struct{}

func (f *fakeShortener) BatchDelete(_ []string) error {
	return nil
}

func (f *fakeShortener) Ping() error {
	return nil
}

func (f *fakeShortener) Gen(_ ...string) chan string {
	ch := make(chan string)
	close(ch)
	return ch
}

func (f *fakeShortener) Delete(_ chan string) chan string {
	ch := make(chan string)
	close(ch)
	return ch
}

func (f *fakeShortener) FanIn(_ ...chan string) chan string {
	ch := make(chan string)
	close(ch)
	return ch
}

func (f *fakeShortener) Save(_ context.Context, originalURL string, userID int) (storage.Data, error) {
	return storage.Data{OriginalURL: originalURL, ShortURL: "abc123", UUID: userID}, nil
}

func (f *fakeShortener) Get(_ context.Context, short string) (storage.Data, error) {
	return storage.Data{OriginalURL: "http://example.com", ShortURL: short}, nil
}

func (f *fakeShortener) GetAll(_ context.Context, userID int) ([]storage.Data, error) {
	return []storage.Data{{OriginalURL: "http://example.com", ShortURL: "abc123", UUID: userID}}, nil
}

func (f *fakeShortener) DeleteUserURLS(_ context.Context, userID int, shorts []string) error {
	return nil
}

func (f *fakeShortener) Stats() (int, int, error) {
	return 42, 10, nil
}

// setupStubServer returns a minimal *Server with in-memory LocalStorage and Echo routing.
func setupStubServer() *Server {
	s := &Server{
		echo:       echo.New(),
		URLService: &fakeShortener{},
		logger:     logrus.New().WithField("component", "example"),
		cfg: &config.Config{
			Server: config.Server{
				FlagSuffixAddr: "http://localhost:8080"}},
	}

	s.echo.Use(s.authMiddleware)
	s.echo.POST("/", s.Create)
	s.echo.POST("/api/shorten", s.CreateJSON)
	s.echo.GET("/:short", s.GetShort)
	s.echo.GET("/Ping", s.Ping)
	s.echo.POST("/api/shorten/Batch", s.Batch)
	s.echo.GET("/api/user/urls", s.UserURLs)
	s.echo.DELETE("/api/user/urls", s.DeleteUserURLs)

	return s
}

func (s *Server) createCustomJWT(userID int) (string, error) {
	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(auth.SecretKey)
}

func Example_create() {
	s := setupStubServer()
	token, _ := s.createCustomJWT(1)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://example.com"))
	req.Header.Set("Content-Type", "text/plain")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)
	fmt.Println("Status:", rec.Code)
	// Output:
	// Status: 201
}

func Example_createJSON() {
	s := setupStubServer()
	token, _ := s.createCustomJWT(1)
	body := map[string]string{"url": "http://example.com"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)
	fmt.Println("Status:", rec.Code)
	// Output:
	// Status: 201
}

//func Example_getShort() {
//	s := setupStubServer()
//	s.logger = logrus.New().WithField("component", "example")
//
//
//	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
//	rec := httptest.NewRecorder()
//
//	s.echo.ServeHTTP(rec, req)
//	fmt.Println("Status:", rec.Code)
//	fmt.Println("Redirect:", rec.Header().Get("Location"))
//	// Output:
//	// Status: 307
//	// Redirect: http://example.com
//}

func Example_ping() {
	s := setupStubServer()
	req := httptest.NewRequest(http.MethodGet, "/Ping", nil)
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)
	fmt.Println("Status:", rec.Code)
	// Output:
	// Status: 200
}

func Example_batch() {
	s := setupStubServer()
	token, _ := s.createCustomJWT(1)

	batch := []map[string]string{
		{"correlation_id": "1", "original_url": "http://example.com/1"},
		{"correlation_id": "2", "original_url": "http://example.com/2"},
	}
	b, _ := json.Marshal(batch)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/Batch", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)
	fmt.Println("Status:", rec.Code)
	// Output:
	// Status: 201
}

func Example_userURLs() {
	s := setupStubServer()
	token, _ := s.createCustomJWT(1)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)
	fmt.Println("Status:", rec.Code)
	// Output:
	// Status: 200
}

func Example_deleteUserURLs() {
	s := setupStubServer()
	token, _ := s.createCustomJWT(1)

	shorts := []string{"abc123"}
	b, _ := json.Marshal(shorts)
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: token})
	rec := httptest.NewRecorder()

	s.echo.ServeHTTP(rec, req)
	fmt.Println("Status:", rec.Code)
	// Output:
	// Status: 202
}
