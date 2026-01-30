package links

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"code/internal/domain"
)

const (
	autoShortNameAttempts = 5
	shortNameLen          = 8

	createErrWrapFmt = "links create: %w"

	shortNameAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	redirectStatusFound = 302
)

var errVisitsRepoNil = errors.New("link visits repo is nil")

type Service struct {
	repo       Repo
	visitsRepo VisitsRepo
	log        Logger
}

func New(repo Repo, visitsRepo VisitsRepo, log Logger) *Service {
	if log == nil {
		log = NopLogger{}
	}

	return &Service{repo: repo, visitsRepo: visitsRepo, log: log}
}

var _ UseCase = (*Service)(nil)

func (s *Service) ListLinks(ctx context.Context, query LinksQuery) ([]domain.Link, int64, error) {
	if query.Range == nil {
		items, err := s.repo.ListAll(ctx, query.Sort)
		if err != nil {
			return nil, 0, fmt.Errorf("links list all: %w", err)
		}

		return items, -1, nil
	}

	items, err := s.repo.ListPage(ctx, int32(query.Range.Start), int32(query.Range.Count), query.Sort)
	if err != nil {
		return nil, 0, fmt.Errorf("links list page: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("links count: %w", err)
	}

	return items, total, nil
}

func (s *Service) Get(ctx context.Context, id int64) (domain.Link, error) {
	link, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Link{}, fmt.Errorf("links get by id: %w", err)
	}

	return link, nil
}

func (s *Service) GetByShortName(ctx context.Context, shortName string) (domain.Link, error) {
	err := domain.ValidateShortName(shortName)
	if err != nil {
		return domain.Link{}, err
	}

	link, err := s.repo.GetByShortName(ctx, shortName)
	if err != nil {
		return domain.Link{}, fmt.Errorf("links get by short name: %w", err)
	}

	return link, nil
}

func (s *Service) Redirect(ctx context.Context, shortName string, meta VisitMeta) (string, int, error) {
	link, err := s.GetByShortName(ctx, shortName)
	if err != nil {
		return "", 0, err
	}

	status := redirectStatusFound

	if s.visitsRepo != nil {
		visit := domain.LinkVisit{
			LinkID:    link.ID,
			CreatedAt: time.Now().UTC(),
			IP:        meta.IP,
			UserAgent: meta.UserAgent,
			Referer:   meta.Referer,
			Status:    status,
		}

		if _, err := s.visitsRepo.Create(ctx, visit); err != nil {
			s.log.With(
				"code", shortName,
				"link_id", link.ID,
			).Warn("link visit create failed", "err", err)
		}
	}

	return link.OriginalURL, status, nil
}

func (s *Service) Create(ctx context.Context, originalURL, shortName string) (domain.Link, error) {
	originalURL = strings.TrimSpace(originalURL)
	shortName = strings.TrimSpace(shortName)

	if err := domain.ValidateOriginalURL(originalURL); err != nil {
		return domain.Link{}, err
	}

	if shortName == "" {
		return s.createWithGeneratedShortName(ctx, originalURL)
	}

	if err := domain.ValidateShortName(shortName); err != nil {
		return domain.Link{}, err
	}

	link, err := s.repo.Create(ctx, originalURL, shortName)
	if err != nil {
		return domain.Link{}, fmt.Errorf(createErrWrapFmt, err)
	}

	return link, nil
}

func (s *Service) Update(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
	originalURL = strings.TrimSpace(originalURL)
	shortName = strings.TrimSpace(shortName)

	if err := domain.ValidateOriginalURL(originalURL); err != nil {
		return domain.Link{}, err
	}

	if shortName == "" {
		return s.updateWithGeneratedShortName(ctx, id, originalURL)
	}

	if err := domain.ValidateShortName(shortName); err != nil {
		return domain.Link{}, err
	}

	link, err := s.repo.Update(ctx, id, originalURL, shortName)
	if err != nil {
		return domain.Link{}, fmt.Errorf("links update: %w", err)
	}

	return link, nil
}

func (s *Service) updateWithGeneratedShortName(
	ctx context.Context,
	id int64,
	originalURL string,
) (domain.Link, error) {
	for range autoShortNameAttempts {
		gen, err := generateShortName()
		if err != nil {
			return domain.Link{}, fmt.Errorf("links generate short name: %w", err)
		}

		link, err := s.repo.Update(ctx, id, originalURL, gen)
		if errors.Is(err, domain.ErrShortNameConflict) {
			continue
		}

		if err != nil {
			return domain.Link{}, fmt.Errorf("links update: %w", err)
		}

		return link, nil
	}

	return domain.Link{}, domain.ErrShortNameConflict
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("links delete: %w", err)
	}

	return nil
}

func (s *Service) ListLinkVisits(ctx context.Context, query LinkVisitsQuery) ([]domain.LinkVisit, int64, error) {
	if s.visitsRepo == nil {
		return nil, 0, errVisitsRepoNil
	}

	if query.Range == nil {
		items, err := s.visitsRepo.ListAll(ctx, query.Sort)
		if err != nil {
			return nil, 0, fmt.Errorf("link visits list all: %w", err)
		}

		return items, -1, nil
	}

	items, err := s.visitsRepo.ListPage(ctx, int32(query.Range.Start), int32(query.Range.Count), query.Sort)
	if err != nil {
		return nil, 0, fmt.Errorf("link visits list page: %w", err)
	}

	total, err := s.visitsRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("link visits count: %w", err)
	}

	return items, total, nil
}

func (s *Service) createWithGeneratedShortName(
	ctx context.Context,
	originalURL string,
) (domain.Link, error) {
	for range autoShortNameAttempts {
		gen, err := generateShortName()
		if err != nil {
			return domain.Link{}, fmt.Errorf("links generate short name: %w", err)
		}

		link, err := s.repo.Create(ctx, originalURL, gen)
		if errors.Is(err, domain.ErrShortNameConflict) {
			continue
		}

		if err != nil {
			return domain.Link{}, fmt.Errorf(createErrWrapFmt, err)
		}

		return link, nil
	}

	return domain.Link{}, domain.ErrShortNameConflict
}

func generateShortName() (string, error) {
	alphaLen := len(shortNameAlphabet)
	cutoff := (256 / alphaLen) * alphaLen

	out := make([]byte, shortNameLen)
	filled := 0

	var buf [32]byte
	for filled < shortNameLen {
		_, err := rand.Read(buf[:])
		if err != nil {
			return "", fmt.Errorf("rand read: %w", err)
		}

		for _, b := range buf {
			if filled >= shortNameLen {
				break
			}

			if int(b) >= cutoff {
				continue
			}

			out[filled] = shortNameAlphabet[int(b)%alphaLen]
			filled++
		}
	}

	return string(out), nil
}
