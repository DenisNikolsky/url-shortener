package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DenisNikolsky/url-shortener/internal/model"
)

type PostgresURLRepository struct {
	db *sql.DB
}

func NewPostgresURLRepository(db *sql.DB) *PostgresURLRepository {
	return &PostgresURLRepository{
		db: db,
	}
}

func (r *PostgresURLRepository) Create(
	ctx context.Context,
	url *model.URL,
) error {
	query := `
		INSERT INTO urls (
			short_code,
			original_url
		)
		VALUES ($1, $2)
		RETURNING
			id,
			created_at,
			clicks
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		url.ShortCode,
		url.OriginalURL,
	).Scan(
		&url.ID,
		&url.CreatedAt,
		&url.Clicks,
	)

	if err != nil {
		return fmt.Errorf("create url: %w", err)
	}

	return nil
}

func (r *PostgresURLRepository) GetByCode(
	ctx context.Context,
	code string,
) (*model.URL, error) {
	query := `
		SELECT
			id,
			short_code,
			original_url,
			created_at,
			clicks
		FROM urls
		WHERE short_code = $1
	`

	var url model.URL

	err := r.db.QueryRowContext(
		ctx,
		query,
		code,
	).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.Clicks,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, fmt.Errorf("get url by code: %w", err)
	}

	return &url, nil
}

func (r *PostgresURLRepository) IncrementClicks(
	ctx context.Context,
	code string,
) error {
	query := `
		UPDATE urls
		SET clicks = clicks + 1
		WHERE short_code = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		code,
	)
	if err != nil {
		return fmt.Errorf("increment clicks: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
