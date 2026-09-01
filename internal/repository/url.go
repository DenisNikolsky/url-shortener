package repository

import (
	"context"

	"github.com/DenisNikolsky/url-shortener/internal/model"
)

type URLRepository interface {
	Create(ctx context.Context, url *model.URL) error
	GetByCode(ctx context.Context, code string) (*model.URL, error)
	IncrementClicks(ctx context.Context, code string) error
}
