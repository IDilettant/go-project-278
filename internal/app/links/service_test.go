package links

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"code/internal/domain"
)

type stubRepo struct {
	t testing.TB

	listAllFunc        func(context.Context) ([]domain.Link, error)
	listPageFunc       func(context.Context, int32, int32) ([]domain.Link, error)
	countFunc          func(context.Context) (int64, error)
	getByIDFunc        func(context.Context, int64) (domain.Link, error)
	getByShortNameFunc func(context.Context, string) (domain.Link, error)
	createFunc         func(context.Context, string, string) (domain.Link, error)
	updateFunc         func(context.Context, int64, string, string) (domain.Link, error)
	deleteFunc         func(context.Context, int64) error
}

func (s *stubRepo) ListAll(ctx context.Context) ([]domain.Link, error) {
	s.t.Helper()

	if s.listAllFunc == nil {
		s.t.Fatalf("unexpected ListAll call")
	}

	return s.listAllFunc(ctx)
}

func (s *stubRepo) ListPage(ctx context.Context, offset, limit int32) ([]domain.Link, error) {
	s.t.Helper()

	if s.listPageFunc == nil {
		s.t.Fatalf("unexpected ListPage call")
	}

	return s.listPageFunc(ctx, offset, limit)
}

func (s *stubRepo) Count(ctx context.Context) (int64, error) {
	s.t.Helper()

	if s.countFunc == nil {
		s.t.Fatalf("unexpected Count call")
	}

	return s.countFunc(ctx)
}

func (s *stubRepo) GetByID(ctx context.Context, id int64) (domain.Link, error) {
	s.t.Helper()

	if s.getByIDFunc == nil {
		s.t.Fatalf("unexpected GetByID call")
	}

	return s.getByIDFunc(ctx, id)
}

func (s *stubRepo) GetByShortName(ctx context.Context, shortName string) (domain.Link, error) {
	s.t.Helper()

	if s.getByShortNameFunc == nil {
		s.t.Fatalf("unexpected GetByShortName call")
	}

	return s.getByShortNameFunc(ctx, shortName)
}

func (s *stubRepo) Create(ctx context.Context, originalURL, shortName string) (domain.Link, error) {
	s.t.Helper()

	if s.createFunc == nil {
		s.t.Fatalf("unexpected Create call")
	}

	return s.createFunc(ctx, originalURL, shortName)
}

func (s *stubRepo) Update(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
	s.t.Helper()

	if s.updateFunc == nil {
		s.t.Fatalf("unexpected Update call")
	}

	return s.updateFunc(ctx, id, originalURL, shortName)
}

func (s *stubRepo) Delete(ctx context.Context, id int64) error {
	s.t.Helper()

	if s.deleteFunc == nil {
		s.t.Fatalf("unexpected Delete call")
	}

	return s.deleteFunc(ctx, id)
}

func TestServiceCreate_AutoShortNameRetries(t *testing.T) {
	ctx := context.Background()
	var calls int
	var lastShortName string

	repo := &stubRepo{
		t: t,
		createFunc: func(ctx context.Context, originalURL, shortName string) (domain.Link, error) {
			calls++
			lastShortName = shortName
			if calls < 3 {
				return domain.Link{}, domain.ErrShortNameConflict
			}

			return domain.Link{
				ID:          1,
				OriginalURL: originalURL,
				ShortName:   shortName,
			}, nil
		},
	}

	svc := New(repo)
	link, err := svc.Create(ctx, "https://example.com", "")
	require.NoError(t, err)
	require.Equal(t, 3, calls)
	require.NoError(t, domain.ValidateShortName(lastShortName))
	require.Equal(t, lastShortName, link.ShortName)
}

func TestServiceCreate_AutoShortNameExhausted(t *testing.T) {
	ctx := context.Background()
	var calls int

	repo := &stubRepo{
		t: t,
		createFunc: func(ctx context.Context, originalURL, shortName string) (domain.Link, error) {
			calls++
			return domain.Link{}, domain.ErrShortNameConflict
		},
	}

	svc := New(repo)
	_, err := svc.Create(ctx, "https://example.com", "")
	require.ErrorIs(t, err, domain.ErrShortNameConflict)
	require.Equal(t, autoShortNameAttempts, calls)
}

func TestServiceCreate_ExplicitShortNameValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid short_name", func(t *testing.T) {
		repo := &stubRepo{
			t: t,
		}

		svc := New(repo)
		_, err := svc.Create(ctx, "https://example.com", "ab_cd")
		require.ErrorIs(t, err, domain.ErrInvalidShortName)
	})

	t.Run("conflict does not retry", func(t *testing.T) {
		var calls int
		repo := &stubRepo{
			t: t,
			createFunc: func(ctx context.Context, originalURL, shortName string) (domain.Link, error) {
				calls++
				return domain.Link{}, domain.ErrShortNameConflict
			},
		}

		svc := New(repo)
		_, err := svc.Create(ctx, "https://example.com", "abcd")
		require.ErrorIs(t, err, domain.ErrShortNameConflict)
		require.Equal(t, 1, calls)
	})
}

func TestServiceUpdate_AutoShortNameRetries(t *testing.T) {
	ctx := context.Background()
	var calls int
	var lastShortName string

	repo := &stubRepo{
		t: t,
		getByIDFunc: func(ctx context.Context, id int64) (domain.Link, error) {
			return domain.Link{ID: id, ShortName: "abcd"}, nil
		},
		updateFunc: func(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
			calls++
			lastShortName = shortName
			if calls < 3 {
				return domain.Link{}, domain.ErrShortNameConflict
			}

			return domain.Link{ID: id, OriginalURL: originalURL, ShortName: shortName}, nil
		},
	}

	svc := New(repo)
	link, err := svc.Update(ctx, 1, "https://example.com/new", "")
	require.NoError(t, err)
	require.Equal(t, 3, calls)
	require.NoError(t, domain.ValidateShortName(lastShortName))
	require.NotEqual(t, "abcd", lastShortName)
	require.Equal(t, lastShortName, link.ShortName)
}

func TestServiceUpdate_ExplicitShortName(t *testing.T) {
	ctx := context.Background()

	repo := &stubRepo{
		t: t,
		updateFunc: func(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
			require.Equal(t, "zzzz", shortName)
			return domain.Link{ID: id, OriginalURL: originalURL, ShortName: shortName}, nil
		},
	}

	svc := New(repo)
	link, err := svc.Update(ctx, 1, "https://example.com/new", "zzzz")
	require.NoError(t, err)
	require.Equal(t, "zzzz", link.ShortName)
}

func TestServiceUpdate_InvalidShortName(t *testing.T) {
	ctx := context.Background()

	repo := &stubRepo{
		t: t,
		updateFunc: func(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
			t.Fatalf("Update should not be called")
			return domain.Link{}, nil
		},
	}

	svc := New(repo)
	_, err := svc.Update(ctx, 1, "https://example.com/new", "ab_cd")
	require.ErrorIs(t, err, domain.ErrInvalidShortName)
}

func TestServiceUpdate_NotFound(t *testing.T) {
	ctx := context.Background()

	repo := &stubRepo{
		t: t,
		getByIDFunc: func(ctx context.Context, id int64) (domain.Link, error) {
			return domain.Link{}, domain.ErrNotFound
		},
		updateFunc: func(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
			t.Fatalf("Update should not be called")
			return domain.Link{}, nil
		},
	}

	svc := New(repo)
	_, err := svc.Update(ctx, 1, "https://example.com/new", "")
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestServiceUpdate_InvalidOriginalURL(t *testing.T) {
	ctx := context.Background()

	repo := &stubRepo{
		t: t,
		getByIDFunc: func(ctx context.Context, id int64) (domain.Link, error) {
			t.Fatalf("GetByID should not be called")
			return domain.Link{}, nil
		},
		updateFunc: func(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
			t.Fatalf("Update should not be called")
			return domain.Link{}, nil
		},
	}

	svc := New(repo)
	_, err := svc.Update(ctx, 1, "not-a-url", "")
	require.ErrorIs(t, err, domain.ErrInvalidURL)
}
