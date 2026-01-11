package links

import (
	"context"

	"code/internal/domain"
)

type Repo interface {
	List(ctx context.Context) ([]domain.Link, error)
	GetByID(ctx context.Context, id int64) (domain.Link, error)
	GetByShortName(ctx context.Context, shortName string) (domain.Link, error)
	Create(ctx context.Context, originalURL, shortName string) (domain.Link, error)
	Update(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error)
	Delete(ctx context.Context, id int64) error
}
