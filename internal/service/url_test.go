package service

import (
	"context"
	"github.com/DenisNikolsky/url-shortener/internal/model"
	"testing"
)

type mockURLRepository struct {
	createdURL         *model.URL
	createErr          error
	url                *model.URL
	getByCodeErr       error
	incrementedCode    string
	incrementClicksErr error
}

func (m *mockURLRepository) Create(
	ctx context.Context,
	url *model.URL,
) error {
	m.createdURL = url

	return m.createErr
}

func (m *mockURLRepository) GetByCode(
	ctx context.Context,
	code string,
) (*model.URL, error) {
	return m.url, m.getByCodeErr
}

func (m *mockURLRepository) IncrementClicks(
	ctx context.Context,
	code string,
) error {
	m.incrementedCode = code

	return m.incrementClicksErr
}

func TestURLService_Create(t *testing.T) {
	repo := &mockURLRepository{}

	service := NewURLService(repo)

	url, err := service.Create(
		context.Background(),
		"https://google.com",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if url == nil {
		t.Fatal("expected URL, got nil")
	}

	if url.OriginalURL != "https://google.com" {
		t.Errorf(
			"expected original URL https://google.com, got %s",
			url.OriginalURL,
		)
	}

	if len(url.ShortCode) != shortCodeLength {
		t.Errorf(
			"expected short code length %d, got %d",
			shortCodeLength,
			len(url.ShortCode),
		)

	}

	if repo.createdURL == nil {
		t.Fatal("repository.Create was not called")
	}
}

func TestURLService_Create_InvalidURL(t *testing.T) {
	repo := &mockURLRepository{}

	service := NewURLService(repo)

	_, err := service.Create(
		context.Background(),
		"not-a-url",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if repo.createdURL != nil {
		t.Fatal("repository.Create should not be called")
	}
}

func TestURLService_GetByCode(t *testing.T) {
	repo := &mockURLRepository{
		url: &model.URL{
			ID:          1,
			ShortCode:   "abc123",
			OriginalURL: "https://google.com",
			Clicks:      0,
		},
	}

	service := NewURLService(repo)

	url, err := service.GetByCode(
		context.Background(),
		"abc123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if url == nil {
		t.Fatal("expected URL, got nil")
	}

	if url.ShortCode != "abc123" {
		t.Errorf(
			"expected short code abc123, got %s",
			url.ShortCode,
		)
	}

	if url.OriginalURL != "https://google.com" {
		t.Errorf(
			"expected https://google.com, got %s",
			url.OriginalURL,
		)
	}

	if repo.incrementedCode != "abc123" {
		t.Errorf(
			"expected clicks increment for abc123, got %s",
			repo.incrementedCode,
		)
	}

	if url.Clicks != 1 {
		t.Errorf(
			"expected clicks 1, got %d",
			url.Clicks,
		)
	}
}
