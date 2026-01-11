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
)

var shortNameAlphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type Service struct {
	repo Repo
}

func New(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]domain.Link, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id int64) (domain.Link, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByShortName(ctx context.Context, shortName string) (domain.Link, error) {
	if err := domain.ValidateShortName(shortName); err != nil {
		return domain.Link{}, err
	}

	return s.repo.GetByShortName(ctx, shortName)
}

func (s *Service) Create(ctx context.Context, originalURL, shortName string) (domain.Link, error) {
	originalURL = strings.TrimSpace(originalURL)
	shortName = strings.TrimSpace(shortName)

	if err := domain.ValidateOriginalURL(originalURL); err != nil {
		return domain.Link{}, err
	}

	if shortName == "" {
		for range autoShortNameAttempts {
			gen, err := generateShortName()
			if err != nil {
				return domain.Link{}, err
			}

			link, err := s.repo.Create(ctx, originalURL, gen)
			if err == domain.ErrShortNameConflict {
				continue
			}

			if err != nil {
				return domain.Link{}, err
			}

			return link, nil
		}

		return domain.Link{}, domain.ErrShortNameConflict
	}

	if err := domain.ValidateShortName(shortName); err != nil {
		return domain.Link{}, err
	}

	return s.repo.Create(ctx, originalURL, shortName)
}

func (s *Service) Update(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
	originalURL = strings.TrimSpace(originalURL)
	shortName = strings.TrimSpace(shortName)

	if err := domain.ValidateOriginalURL(originalURL); err != nil {
		return domain.Link{}, err
	}

	if err := domain.ValidateShortName(shortName); err != nil {
		return domain.Link{}, err
	}

	return s.repo.Update(ctx, id, originalURL, shortName)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
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
