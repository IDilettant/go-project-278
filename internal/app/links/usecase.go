package links

import (
	"context"

	"code/internal/domain"
)

// UseCase is an input port for the links application.
type UseCase interface {
	ListAll(ctx context.Context) ([]domain.Link, error)
	ListPage(ctx context.Context, offset, limit int32, needTotal bool) ([]domain.Link, int64, error)
	Get(ctx context.Context, id int64) (domain.Link, error)
	GetByShortName(ctx context.Context, shortName string) (domain.Link, error)
	Create(ctx context.Context, originalURL, shortName string) (domain.Link, error)
	Update(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error)
	Delete(ctx context.Context, id int64) error
}
