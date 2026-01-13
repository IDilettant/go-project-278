package links

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"code/internal/domain"
)

const (
	autoShortNameAttempts = 5

	shortNameRandBytes = 6
	shortNameLen       = 8

	createErrWrapFmt = "links create: %w"
)

var shortNameAlphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type Service struct {
	repo Repo
}

func New(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListAll(ctx context.Context) ([]domain.Link, error) {
	items, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("links list all: %w", err)
	}

	return items, nil
}

func (s *Service) ListPage(
	ctx context.Context,
	offset, limit int32,
	needTotal bool,
) ([]domain.Link, int64, error) {
	items, err := s.repo.ListPage(ctx, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("links list page: %w", err)
	}

	if !needTotal {
		return items, -1, nil
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
	if err := domain.ValidateShortName(shortName); err != nil {
		return domain.Link{}, err
	}

	link, err := s.repo.GetByShortName(ctx, shortName)
	if err != nil {
		return domain.Link{}, fmt.Errorf("links get by short name: %w", err)
	}

	return link, nil
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

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Link{}, fmt.Errorf("links get by id: %w", err)
	}

	if shortName == "" {
		shortName = existing.ShortName
	} else if shortName != existing.ShortName {
		return domain.Link{}, domain.ErrShortNameImmutable
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

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("links delete: %w", err)
	}

	return nil
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
		if err == domain.ErrShortNameConflict {
			continue
		}

		if err != nil {
			return domain.Link{}, fmt.Errorf(createErrWrapFmt, err)
		}

		return link, nil
	}

	return domain.Link{}, fmt.Errorf(createErrWrapFmt, domain.ErrShortNameConflict)
}

func generateShortName() (string, error) {
	out := make([]rune, shortNameLen)
	max := big.NewInt(int64(len(shortNameAlphabet)))

	for i := range shortNameLen {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("rand int: %w", err)
		}

		out[i] = shortNameAlphabet[n.Int64()]
	}

	return string(out), nil
}
