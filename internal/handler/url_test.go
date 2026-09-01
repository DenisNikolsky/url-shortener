package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DenisNikolsky/url-shortener/internal/model"
	"github.com/DenisNikolsky/url-shortener/internal/service"
	"github.com/labstack/echo/v4"
)

type mockURLService struct {
	createURL    *model.URL
	createErr    error
	getByCodeURL *model.URL
	getByCodeErr error
}

func (m *mockURLService) Create(
	ctx context.Context,
	originalURL string,
) (*model.URL, error) {
	return m.createURL, m.createErr
}

func (m *mockURLService) GetByCode(
	ctx context.Context,
	code string,
) (*model.URL, error) {
	return m.getByCodeURL, m.getByCodeErr
}

func TestURLHandler_Create(t *testing.T) {
	e := echo.New()

	service := &mockURLService{
		createURL: &model.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://google.com",
		},
	}

	handler := NewURLHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/urls",
		strings.NewReader(`{"url":"https://google.com"}`),
	)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.Create(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	var response createURLResponse

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ShortCode != "abc123" {
		t.Errorf(
			"expected short code abc123, got %s",
			response.ShortCode,
		)
	}

	if response.OriginalURL != "https://google.com" {
		t.Errorf(
			"expected original URL https://google.com, got %s",
			response.OriginalURL,
		)
	}
}

func TestURLHandler_Create_InvalidRequest(t *testing.T) {
	e := echo.New()

	service := &mockURLService{}

	handler := NewURLHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/urls",
		strings.NewReader(`{"url":`),
	)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.Create(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestURLHandler_Create_ServiceError(t *testing.T) {
	e := echo.New()

	service := &mockURLService{
		createErr: service.ErrInvalidURL,
	}

	handler := NewURLHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/urls",
		strings.NewReader(`{"url":"not-a-url"}`),
	)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.Create(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestURLHandler_Redirect(t *testing.T) {
	e := echo.New()

	urlService := &mockURLService{
		getByCodeURL: &model.URL{
			ShortCode:   "abc123",
			OriginalURL: "https://google.com",
		},
	}

	handler := NewURLHandler(urlService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc123",
		nil,
	)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.SetPath("/:code")
	c.SetParamNames("code")
	c.SetParamValues("abc123")

	err := handler.Redirect(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusFound,
			rec.Code,
		)
	}

	location := rec.Header().Get("Location")

	if location != "https://google.com" {
		t.Errorf(
			"expected Location https://google.com, got %s",
			location,
		)
	}
}

func TestURLHandler_Redirect_NotFound(t *testing.T) {
	e := echo.New()

	urlService := &mockURLService{
		getByCodeErr: service.ErrURLNotFound,
	}

	handler := NewURLHandler(urlService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc123",
		nil,
	)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.SetPath("/:code")
	c.SetParamNames("code")
	c.SetParamValues("abc123")

	err := handler.Redirect(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestURLHandler_Redirect_ServiceError(t *testing.T) {
	e := echo.New()

	urlService := &mockURLService{
		getByCodeErr: errors.New("database error"),
	}

	handler := NewURLHandler(urlService)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc123",
		nil,
	)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	c.SetPath("/:code")
	c.SetParamNames("code")
	c.SetParamValues("abc123")

	err := handler.Redirect(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestURLHandler_Create_ServiceInternalError(t *testing.T) {
	e := echo.New()

	urlService := &mockURLService{
		createErr: errors.New("database error"),
	}

	handler := NewURLHandler(urlService)

	req := httptest.NewRequest(
		http.MethodPost,
		"/urls",
		strings.NewReader(`{"url":"https://google.com"}`),
	)

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := handler.Create(c)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}
