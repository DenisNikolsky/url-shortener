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
	defer db.Close()

	defer func() {
		_, _ = db.Exec(
			"DELETE FROM urls WHERE short_code = $1",
			"test123",
		)
	}()

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
