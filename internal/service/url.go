package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	neturl "net/url"

	"github.com/DenisNikolsky/url-shortener/internal/model"
)

const shortCodeLength = 6

var (
	ErrInvalidURL  = errors.New("invalid URL")
	ErrURLNotFound = errors.New("URL not found")
)

type URLRepository interface {
	Create(ctx context.Context, url *model.URL) error
	GetByCode(ctx context.Context, code string) (*model.URL, error)
	IncrementClicks(ctx context.Context, code string) error
}

type URLService interface {
	Create(ctx context.Context, originalURL string) (*model.URL, error)
	GetByCode(ctx context.Context, code string) (*model.URL, error)
}

type urlService struct {
	repo URLRepository
}

func NewURLService(repo URLRepository) URLService {
	return &urlService{
		repo: repo,
	}
}

func (s *urlService) Create(
	ctx context.Context,
	originalURL string,
) (*model.URL, error) {
	parsedURL, err := neturl.ParseRequestURI(originalURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("%w: URL must use http or https", ErrInvalidURL)
	}

	if parsedURL.Host == "" {
		return nil, fmt.Errorf("%w: URL must have a host", ErrInvalidURL)
	}

	shortCode, err := generateShortCode(shortCodeLength)
	if err != nil {
		return nil, fmt.Errorf(
			"generate short code: %w",
			err,
		)
	}

	urlModel := &model.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
	}

	if err := s.repo.Create(ctx, urlModel); err != nil {
		return nil, fmt.Errorf(
			"create URL: %w",
			err,
		)
	}

	return urlModel, nil
}

func (s *urlService) GetByCode(
	ctx context.Context,
	code string,
) (*model.URL, error) {
	url, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrURLNotFound
		}

		return nil, fmt.Errorf("get URL by code: %w", err)
	}

	if err := s.repo.IncrementClicks(ctx, code); err != nil {
		return nil, fmt.Errorf(
			"increment clicks: %w",
			err,
		)
	}

	url.Clicks++

	return url, nil
}

func generateShortCode(length int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	code := make([]byte, length)

	for i := range code {
		n, err := rand.Int(
			rand.Reader,
			big.NewInt(int64(len(alphabet))),
		)
		if err != nil {
			return "", fmt.Errorf(
				"generate random number: %w",
				err,
			)
		}

		code[i] = alphabet[n.Int64()]
	}

	return string(code), nil
}
