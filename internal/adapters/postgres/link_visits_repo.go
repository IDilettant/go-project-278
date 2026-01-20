package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"code/internal/adapters/postgres/sqlcgen"
	"code/internal/app/links"
	"code/internal/domain"
)

type LinkVisitsRepo struct {
	q *sqlcgen.Queries
}

func NewLinkVisitsRepo(db *sql.DB) *LinkVisitsRepo {
	return &LinkVisitsRepo{q: sqlcgen.New(db)}
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

func (r *LinkVisitsRepo) ListAll(ctx context.Context) ([]domain.LinkVisit, error) {
	rows, err := r.q.ListLinkVisits(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: list link visits: %w", err)
	}

	out := make([]domain.LinkVisit, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapVisitRow(row))
	}

	return out, nil
}

func (r *LinkVisitsRepo) ListPage(ctx context.Context, offset, limit int32) ([]domain.LinkVisit, error) {
	rows, err := r.q.ListLinkVisitsPage(ctx, sqlcgen.ListLinkVisitsPageParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: list link visits page: %w", err)
	}

	out := make([]domain.LinkVisit, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapVisitRow(row))
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

func mapVisitRow(row sqlcgen.LinkVisit) domain.LinkVisit {
	return domain.LinkVisit{
		ID:        row.ID,
		LinkID:    row.LinkID,
		CreatedAt: row.CreatedAt,
		IP:        row.Ip,
		UserAgent: row.UserAgent,
		Referer:   row.Referer,
		Status:    int(row.Status),
	}
}
