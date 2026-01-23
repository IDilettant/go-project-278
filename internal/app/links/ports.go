package links

import (
	"context"

	"code/internal/domain"
)

type Repo interface {
	ListAll(ctx context.Context, sort Sort) ([]domain.Link, error)
	ListPage(ctx context.Context, offset, limit int32, sort Sort) ([]domain.Link, error)
	Count(ctx context.Context) (int64, error)
	GetByID(ctx context.Context, id int64) (domain.Link, error)
	GetByShortName(ctx context.Context, shortName string) (domain.Link, error)
	Create(ctx context.Context, originalURL, shortName string) (domain.Link, error)
	Update(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error)
	Delete(ctx context.Context, id int64) error
}

type VisitsRepo interface {
	Create(ctx context.Context, visit domain.LinkVisit) (int64, error)
	ListAll(ctx context.Context, sort Sort) ([]domain.LinkVisit, error)
	ListPage(ctx context.Context, offset, limit int32, sort Sort) ([]domain.LinkVisit, error)
	Count(ctx context.Context) (int64, error)
}
