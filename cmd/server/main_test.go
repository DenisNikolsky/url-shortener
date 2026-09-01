package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DenisNikolsky/url-shortener/internal/config"
	"github.com/DenisNikolsky/url-shortener/internal/repository"
)

type createURLTestResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
}

func TestURLAPI_CreateAndRedirect(t *testing.T) {
	cfg := config.Load("../../.env")

	e, db, err := setupServer(cfg)
	if err != nil {
		t.Fatalf("failed to setup server: %v", err)
	}
	defer db.Close()

	const originalURL = "https://google.com"

	// Создаём короткую ссылку.
	requestBody := []byte(`{"url":"https://google.com"}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/urls",
		bytes.NewReader(requestBody),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d, body: %s",
			http.StatusCreated,
			rec.Code,
			rec.Body.String(),
		)
	}

	var response createURLTestResponse

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ShortCode == "" {
		t.Fatal("expected short code to be set")
	}

	if response.OriginalURL != originalURL {
		t.Fatalf(
			"expected original URL %q, got %q",
			originalURL,
			response.OriginalURL,
		)
	}

	// Проверяем redirect.
	req = httptest.NewRequest(
		http.MethodGet,
		"/"+response.ShortCode,
		nil,
	)

	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusFound,
			rec.Code,
		)
	}

	location := rec.Header().Get("Location")

	if location != originalURL {
		t.Fatalf(
			"expected Location %q, got %q",
			originalURL,
			location,
		)
	}

	// Проверяем, что clicks увеличился.
	repo := repository.NewPostgresURLRepository(db)

	url, err := repo.GetByCode(
		context.Background(),
		response.ShortCode,
	)
	if err != nil {
		t.Fatalf("failed to get URL from database: %v", err)
	}

	if url.Clicks != 1 {
		t.Fatalf(
			"expected clicks to be 1, got %d",
			url.Clicks,
		)
	}

	// Удаляем тестовую запись.
	_, err = db.ExecContext(
		context.Background(),
		"DELETE FROM urls WHERE short_code = $1",
		response.ShortCode,
	)
	if err != nil {
		t.Fatalf("failed to cleanup test data: %v", err)
	}
}
