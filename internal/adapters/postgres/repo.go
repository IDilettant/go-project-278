package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"

	"code/internal/adapters/postgres/sqlcgen"
	"code/internal/app/links"
	"code/internal/domain"
)

// PostgreSQL SQLSTATE error codes.
// See: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	sqlStateUniqueViolation = "23505"
)

type Repo struct {
	db *sql.DB
	q  *sqlcgen.Queries
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db, q: sqlcgen.New(db)}
}

var _ links.Repo = (*Repo)(nil)

func (r *Repo) ListAll(ctx context.Context, sort links.Sort) ([]domain.Link, error) {
	orderBy, err := orderByLinks(sort)
	if err != nil {
		return nil, err
	}

	return r.listLinks(ctx, orderBy, nil, nil, "list all links")
}

func (r *Repo) ListPage(ctx context.Context, offset, limit int32, sort links.Sort) ([]domain.Link, error) {
	orderBy, err := orderByLinks(sort)
	if err != nil {
		return nil, err
	}

	return r.listLinks(ctx, orderBy, &limit, &offset, "list links page")
}

func (r *Repo) listLinks(ctx context.Context, orderBy string, limit, offset *int32, op string) ([]domain.Link, error) {
	builder := sq.Select(sqlLinksSelectCols...).
		From(sqlTableLinks + " " + sqlAliasLinks).
		OrderBy(orderBy).
		PlaceholderFormat(sq.Dollar)

	if limit != nil {
		builder = builder.Limit(uint64(*limit))
	}

	if offset != nil {
		builder = builder.Offset(uint64(*offset))
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("postgres: build %s: %w", op, err)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf(errOpFmt, op, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var out []domain.Link
	for rows.Next() {
		var item domain.Link
		if err := rows.Scan(&item.ID, &item.OriginalURL, &item.ShortName, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf(errOpFmt, op, err)
		}

		out = append(out, item)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(errOpFmt, op, err)
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
