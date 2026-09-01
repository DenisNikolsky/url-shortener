package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DenisNikolsky/url-shortener/internal/model"
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
	tests := []struct {
		name    string
		input   string
		repoErr error
		wantErr bool
	}{
		{
			name:  "valid HTTPS URL",
			input: "https://google.com",
		},
		{
			name:  "valid HTTP URL",
			input: "http://google.com",
		},
		{
			name:    "invalid URL",
			input:   "not-a-url",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			input:   "ftp://google.com",
			wantErr: true,
		},
		{
			name:    "URL without host",
			input:   "http:///google.com",
			wantErr: true,
		},
		{
			name:    "repository error",
			input:   "https://google.com",
			repoErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockURLRepository{
				createErr: tt.repoErr,
			}

			service := NewURLService(repo)

			url, err := service.Create(
				context.Background(),
				tt.input,
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				// При ошибке URL repository.Create
				// не должен вызываться.
				if tt.repoErr == nil && repo.createdURL != nil {
					t.Fatal("repository.Create should not be called")
				}

				if tt.repoErr != nil && repo.createdURL == nil {
					t.Fatal("repository.Create should be called")
				}

				if url != nil {
					t.Fatalf(
						"expected nil URL, got %+v",
						url,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if url == nil {
				t.Fatal("expected URL, got nil")
			}

			if url.OriginalURL != tt.input {
				t.Errorf(
					"expected original URL %s, got %s",
					tt.input,
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

			if repo.createdURL != url {
				t.Error("repository received different URL")
			}
		})
	}
}

func TestURLService_GetByCode(t *testing.T) {
	tests := []struct {
		name               string
		code               string
		url                *model.URL
		getByCodeErr       error
		incrementClicksErr error
		wantErr            bool
		wantServiceErr     error
	}{
		{
			name: "success",
			code: "abc123",
			url: &model.URL{
				ID:          1,
				ShortCode:   "abc123",
				OriginalURL: "https://google.com",
				Clicks:      0,
			},
		},
		{
			name:           "URL not found",
			code:           "abc123",
			getByCodeErr:   sql.ErrNoRows,
			wantErr:        true,
			wantServiceErr: ErrURLNotFound,
		},
		{
			name:         "repository error",
			code:         "abc123",
			getByCodeErr: errors.New("database error"),
			wantErr:      true,
		},
		{
			name: "increment clicks error",
			code: "abc123",
			url: &model.URL{
				ID:          1,
				ShortCode:   "abc123",
				OriginalURL: "https://google.com",
				Clicks:      0,
			},
			incrementClicksErr: errors.New("database error"),
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockURLRepository{
				url:                tt.url,
				getByCodeErr:       tt.getByCodeErr,
				incrementClicksErr: tt.incrementClicksErr,
			}

			service := NewURLService(repo)

			url, err := service.GetByCode(
				context.Background(),
				tt.code,
			)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if tt.wantServiceErr != nil &&
					!errors.Is(err, tt.wantServiceErr) {
					t.Fatalf(
						"expected error %v, got %v",
						tt.wantServiceErr,
						err,
					)
				}

				if url != nil {
					t.Fatalf(
						"expected nil URL, got %+v",
						url,
					)
				}

				// Если GetByCode завершился ошибкой,
				// IncrementClicks вызываться не должен.
				if tt.getByCodeErr != nil &&
					repo.incrementedCode != "" {
					t.Fatal(
						"IncrementClicks should not be called when GetByCode fails",
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if url == nil {
				t.Fatal("expected URL, got nil")
			}

			if url.ShortCode != tt.code {
				t.Errorf(
					"expected short code %s, got %s",
					tt.code,
					url.ShortCode,
				)
			}

			if url.OriginalURL != "https://google.com" {
				t.Errorf(
					"expected original URL https://google.com, got %s",
					url.OriginalURL,
				)
			}

			if repo.incrementedCode != tt.code {
				t.Errorf(
					"expected clicks increment for %s, got %s",
					tt.code,
					repo.incrementedCode,
				)
			}

			if url.Clicks != 1 {
				t.Errorf(
					"expected clicks 1, got %d",
					url.Clicks,
				)
			}
		})
	}
}
