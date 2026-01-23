package links

import (
	"context"

	"code/internal/domain"
)

// UseCase is an input port for the links application.
type UseCase interface {
	ListLinks(ctx context.Context, query LinksQuery) ([]domain.Link, int64, error)
	Get(ctx context.Context, id int64) (domain.Link, error)
	GetByShortName(ctx context.Context, shortName string) (domain.Link, error)
	Redirect(ctx context.Context, shortName string, meta VisitMeta) (string, int, error)
	Create(ctx context.Context, originalURL, shortName string) (domain.Link, error)
	Update(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error)
	Delete(ctx context.Context, id int64) error
	ListLinkVisits(ctx context.Context, query LinkVisitsQuery) ([]domain.LinkVisit, int64, error)
}
