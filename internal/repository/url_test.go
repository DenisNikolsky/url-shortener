package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DenisNikolsky/url-shortener/internal/config"
	"github.com/DenisNikolsky/url-shortener/internal/model"
)

func TestPostgresURLRepository_Create(t *testing.T) {
	cfg := config.Load("../../.env")

	db, err := NewPostgres(cfg)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM urls WHERE short_code = $1",
			"test123",
		)
		if err != nil {
			t.Errorf("failed to cleanup test data: %v", err)
		}

		db.Close()
	})

	repo := NewPostgresURLRepository(db)

	ctx := context.Background()

	url := &model.URL{
		ShortCode:   "test123",
		OriginalURL: "https://example.com",
	}

	err = repo.Create(ctx, url)
	if err != nil {
		t.Fatal(err)
	}

	if url.ID == 0 {
		t.Error("expected ID to be set")
	}

	if url.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if url.Clicks != 0 {
		t.Errorf("expected clicks to be 0, got %d", url.Clicks)
	}

	savedURL, err := repo.GetByCode(ctx, "test123")
	if err != nil {
		t.Fatal(err)
	}

	if savedURL.ID != url.ID {
		t.Errorf(
			"expected ID %d, got %d",
			url.ID,
			savedURL.ID,
		)
	}

	if savedURL.ShortCode != url.ShortCode {
		t.Errorf(
			"expected short code %q, got %q",
			url.ShortCode,
			savedURL.ShortCode,
		)
	}

	if savedURL.OriginalURL != url.OriginalURL {
		t.Errorf(
			"expected original URL %q, got %q",
			url.OriginalURL,
			savedURL.OriginalURL,
		)
	}
}

func TestPostgresURLRepository_GetByCode_NotFound(t *testing.T) {
	cfg := config.Load("../../.env")

	db, err := NewPostgres(cfg)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	defer db.Close()

	repo := NewPostgresURLRepository(db)

	_, err = repo.GetByCode(
		context.Background(),
		"does-not-exist",
	)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf(
			"expected sql.ErrNoRows, got %v",
			err,
		)
	}
}

func TestPostgresURLRepository_IncrementClicks(t *testing.T) {
	cfg := config.Load("../../.env")

	db, err := NewPostgres(cfg)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	defer db.Close()

	repo := NewPostgresURLRepository(db)

	url := &model.URL{
		ShortCode:   "click123",
		OriginalURL: "https://google.com",
	}

	err = repo.Create(context.Background(), url)
	if err != nil {
		t.Fatalf("create url: %v", err)
	}

	defer func() {
		_, _ = db.Exec(
			"DELETE FROM urls WHERE short_code = $1",
			"click123",
		)
	}()

	err = repo.IncrementClicks(
		context.Background(),
		"click123",
	)
	if err != nil {
		t.Fatalf("increment clicks: %v", err)
	}

	foundURL, err := repo.GetByCode(
		context.Background(),
		"click123",
	)
	if err != nil {
		t.Fatalf("get url after increment: %v", err)
	}

	if foundURL.Clicks != 1 {
		t.Errorf(
			"expected clicks 1, got %d",
			foundURL.Clicks,
		)
	}
}

func TestPostgresURLRepository_GetByCode(t *testing.T) {
	cfg := config.Load("../../.env")

	db, err := NewPostgres(cfg)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM urls WHERE short_code = $1",
			"get123",
		)
		if err != nil {
			t.Errorf("failed to cleanup test data: %v", err)
		}

		db.Close()
	})

	repo := NewPostgresURLRepository(db)

	ctx := context.Background()

	url := &model.URL{
		ShortCode:   "get123",
		OriginalURL: "https://example.com/some/long/url",
	}

	if err := repo.Create(ctx, url); err != nil {
		t.Fatalf("create url: %v", err)
	}

	foundURL, err := repo.GetByCode(ctx, "get123")
	if err != nil {
		t.Fatalf("get url by code: %v", err)
	}

	if foundURL.ID != url.ID {
		t.Errorf(
			"expected ID %d, got %d",
			url.ID,
			foundURL.ID,
		)
	}

	if foundURL.ShortCode != "get123" {
		t.Errorf(
			"expected short code %q, got %q",
			"get123",
			foundURL.ShortCode,
		)
	}

	if foundURL.OriginalURL != "https://example.com/some/long/url" {
		t.Errorf(
			"expected original URL %q, got %q",
			"https://example.com/some/long/url",
			foundURL.OriginalURL,
		)
	}

	if foundURL.Clicks != 0 {
		t.Errorf(
			"expected clicks 0, got %d",
			foundURL.Clicks,
		)
	}

	if foundURL.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestPostgresURLRepository_IncrementClicks_NotFound(t *testing.T) {
	cfg := config.Load("../../.env")

	db, err := NewPostgres(cfg)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	repo := NewPostgresURLRepository(db)

	err = repo.IncrementClicks(
		context.Background(),
		"does-not-exist",
	)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf(
			"expected sql.ErrNoRows, got %v",
			err,
		)
	}
}

func TestPostgresURLRepository_Create_DuplicateShortCode(t *testing.T) {
	cfg := config.Load("../../.env")

	db, err := NewPostgres(cfg)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}

	t.Cleanup(func() {
		_, err := db.Exec(
			"DELETE FROM urls WHERE short_code = $1",
			"duplicate",
		)
		if err != nil {
			t.Errorf("failed to cleanup test data: %v", err)
		}

		db.Close()
	})

	repo := NewPostgresURLRepository(db)

	ctx := context.Background()

	firstURL := &model.URL{
		ShortCode:   "duplicate",
		OriginalURL: "https://example.com/first",
	}

	if err := repo.Create(ctx, firstURL); err != nil {
		t.Fatalf("create first URL: %v", err)
	}

	secondURL := &model.URL{
		ShortCode:   "duplicate",
		OriginalURL: "https://example.com/second",
	}

	err = repo.Create(ctx, secondURL)
	if err == nil {
		t.Fatal("expected error when creating duplicate short code")
	}
}
