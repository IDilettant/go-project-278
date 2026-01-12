package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"code/internal/app/links"
	"code/internal/domain"
	"code/internal/repository/postgres/sqlcgen"
)

// PostgreSQL SQLSTATE error codes.
// See: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	sqlStateUniqueViolation = "23505"
)

type Repo struct {
	q *sqlcgen.Queries
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{q: sqlcgen.New(db)}
}

var _ links.Repo = (*Repo)(nil)

func (r *Repo) ListAll(ctx context.Context) ([]domain.Link, error) {
	rows, err := r.q.ListLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: list all links: %w", err)
	}

	out := make([]domain.Link, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRow(row))
	}

	return out, nil
}

func (r *Repo) ListPage(ctx context.Context, offset, limit int32) ([]domain.Link, error) {
	rows, err := r.q.ListLinksPage(ctx, sqlcgen.ListLinksPageParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: list links page: %w", err)
	}

	out := make([]domain.Link, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRow(row))
	}

	return out, nil
}

func (r *Repo) Count(ctx context.Context) (int64, error) {
	total, err := r.q.CountLinks(ctx)
	if err != nil {
		return 0, fmt.Errorf("postgres: count links: %w", err)
	}

	return total, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (domain.Link, error) {
	row, err := r.q.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Link{}, domain.ErrNotFound
		}

		return domain.Link{}, fmt.Errorf("postgres: get link by id: %w", err)
	}

	return mapRow(row), nil
}

func (r *Repo) GetByShortName(ctx context.Context, shortName string) (domain.Link, error) {
	row, err := r.q.GetLinkByShortName(ctx, shortName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Link{}, domain.ErrNotFound
		}

		return domain.Link{}, fmt.Errorf("postgres: get link by short name: %w", err)
	}

	return mapRow(row), nil
}

func (r *Repo) Create(ctx context.Context, originalURL, shortName string) (domain.Link, error) {
	row, err := r.q.CreateLink(ctx, sqlcgen.CreateLinkParams{
		OriginalUrl: originalURL,
		ShortName:   shortName,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Link{}, domain.ErrShortNameConflict
		}

		return domain.Link{}, fmt.Errorf("postgres: create link: %w", err)
	}

	return mapRow(row), nil
}

func (r *Repo) Update(
	ctx context.Context,
	id int64,
	originalURL, shortName string,
) (domain.Link, error) {
	row, err := r.q.UpdateLink(ctx, sqlcgen.UpdateLinkParams{
		ID:          id,
		OriginalUrl: originalURL,
		ShortName:   shortName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Link{}, domain.ErrNotFound
		}

		if isUniqueViolation(err) {
			return domain.Link{}, domain.ErrShortNameConflict
		}

		return domain.Link{}, fmt.Errorf("postgres: update link: %w", err)
	}

	return mapRow(row), nil
}

func (r *Repo) Delete(ctx context.Context, id int64) error {
	n, err := r.q.DeleteLink(ctx, id)
	if err != nil {
		return fmt.Errorf("postgres: delete link: %w", err)
	}

	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == sqlStateUniqueViolation
	}

	return false
}

func mapRow(row sqlcgen.Link) domain.Link {
	return domain.Link{
		ID:          row.ID,
		OriginalURL: row.OriginalUrl,
		ShortName:   row.ShortName,
		CreatedAt:   row.CreatedAt,
	}
}
