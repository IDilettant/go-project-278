package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"code/internal/adapters/postgres/sqlcgen"
	"code/internal/app/links"
	"code/internal/domain"
)

type LinkVisitsRepo struct {
	db *sql.DB
	q  *sqlcgen.Queries
}

func NewLinkVisitsRepo(db *sql.DB) *LinkVisitsRepo {
	return &LinkVisitsRepo{db: db, q: sqlcgen.New(db)}
}

var _ links.VisitsRepo = (*LinkVisitsRepo)(nil)

func (r *LinkVisitsRepo) Create(ctx context.Context, visit domain.LinkVisit) (int64, error) {
	id, err := r.q.CreateLinkVisit(ctx, sqlcgen.CreateLinkVisitParams{
		LinkID:    visit.LinkID,
		CreatedAt: visit.CreatedAt,
		Ip:        visit.IP,
		UserAgent: visit.UserAgent,
		Referer:   visit.Referer,
		Status:    int32(visit.Status),
	})
	if err != nil {
		return 0, fmt.Errorf("postgres: create link visit: %w", err)
	}

	return id, nil
}

func (r *LinkVisitsRepo) ListAll(ctx context.Context, sort links.Sort) ([]domain.LinkVisit, error) {
	orderBy, err := orderByLinkVisits(sort)
	if err != nil {
		return nil, err
	}

	return r.listLinkVisits(ctx, orderBy, nil, nil, "list link visits")
}

func (r *LinkVisitsRepo) ListPage(ctx context.Context, offset, limit int32, sort links.Sort) ([]domain.LinkVisit, error) {
	orderBy, err := orderByLinkVisits(sort)
	if err != nil {
		return nil, err
	}

	return r.listLinkVisits(ctx, orderBy, &limit, &offset, "list link visits page")
}

func (r *LinkVisitsRepo) listLinkVisits(
	ctx context.Context,
	orderBy string,
	limit, offset *int32,
	op string,
) ([]domain.LinkVisit, error) {
	builder := sq.Select(sqlVisitsSelectCols...).
		From(sqlTableLinkVisits + " " + sqlAliasVisits).
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
		return nil, fmt.Errorf("postgres: %s: %w", op, err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var out []domain.LinkVisit
	for rows.Next() {
		var item domain.LinkVisit
		var status int32
		if err := rows.Scan(
			&item.ID,
			&item.LinkID,
			&item.CreatedAt,
			&item.IP,
			&item.UserAgent,
			&item.Referer,
			&status,
		); err != nil {
			return nil, fmt.Errorf("postgres: %s: %w", op, err)
		}
		
		item.Status = int(status)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: %s: %w", op, err)
	}

	return out, nil
}

func (r *LinkVisitsRepo) Count(ctx context.Context) (int64, error) {
	total, err := r.q.CountLinkVisits(ctx)
	if err != nil {
		return 0, fmt.Errorf("postgres: count link visits: %w", err)
	}

	return total, nil
}
