package service

import (
	"context"

	"github.com/DenisNikolsky/url-shortener/internal/model"
)

type URLService interface {
	Create(ctx context.Context, originalURL string) (*model.URL, error)
	GetByCode(ctx context.Context, code string) (*model.URL, error)
}
