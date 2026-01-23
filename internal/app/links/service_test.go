package links

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"code/internal/domain"
)

type stubRepo struct {
	t testing.TB

	listAllFunc        func(context.Context, Sort) ([]domain.Link, error)
	listPageFunc       func(context.Context, int32, int32, Sort) ([]domain.Link, error)
	countFunc          func(context.Context) (int64, error)
	getByIDFunc        func(context.Context, int64) (domain.Link, error)
	getByShortNameFunc func(context.Context, string) (domain.Link, error)
	createFunc         func(context.Context, string, string) (domain.Link, error)
	updateFunc         func(context.Context, int64, string, string) (domain.Link, error)
	deleteFunc         func(context.Context, int64) error
}

type stubVisitsRepo struct {
	t testing.TB

	createFunc   func(context.Context, domain.LinkVisit) (int64, error)
	listAllFunc  func(context.Context, Sort) ([]domain.LinkVisit, error)
	listPageFunc func(context.Context, int32, int32, Sort) ([]domain.LinkVisit, error)
	countFunc    func(context.Context) (int64, error)
}

func (s *stubVisitsRepo) Create(ctx context.Context, visit domain.LinkVisit) (int64, error) {
	s.t.Helper()

	if s.createFunc == nil {
		s.t.Fatalf("unexpected Create call")
	}

	return s.createFunc(ctx, visit)
}

func (s *stubVisitsRepo) ListAll(ctx context.Context, sort Sort) ([]domain.LinkVisit, error) {
	s.t.Helper()

	if s.listAllFunc == nil {
		s.t.Fatalf("unexpected ListAll call")
	}

	return s.listAllFunc(ctx, sort)
}

func (s *stubVisitsRepo) ListPage(ctx context.Context, offset, limit int32, sort Sort) ([]domain.LinkVisit, error) {
	s.t.Helper()

	if s.listPageFunc == nil {
		s.t.Fatalf("unexpected ListPage call")
	}

	return s.listPageFunc(ctx, offset, limit, sort)
}

func (s *stubVisitsRepo) Count(ctx context.Context) (int64, error) {
	s.t.Helper()

	if s.countFunc == nil {
		s.t.Fatalf("unexpected Count call")
	}

	return s.countFunc(ctx)
}

func (s *stubRepo) ListAll(ctx context.Context, sort Sort) ([]domain.Link, error) {
	s.t.Helper()

	if s.listAllFunc == nil {
		s.t.Fatalf("unexpected ListAll call")
	}

	return s.listAllFunc(ctx, sort)
}

func (s *stubRepo) ListPage(ctx context.Context, offset, limit int32, sort Sort) ([]domain.Link, error) {
	s.t.Helper()

	if s.listPageFunc == nil {
		s.t.Fatalf("unexpected ListPage call")
	}

	return s.listPageFunc(ctx, offset, limit, sort)
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

	svc := New(repo, nil, nil)
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

	svc := New(repo, nil, nil)
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

		svc := New(repo, nil, nil)
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

		svc := New(repo, nil, nil)
		_, err := svc.Create(ctx, "https://example.com", "abcd")
		require.ErrorIs(t, err, domain.ErrShortNameConflict)
		require.Equal(t, 1, calls)
	})
}

func TestServiceUpdate_EmptyShortNameGenerates(t *testing.T) {
	ctx := context.Background()
	var gotShortName string

	repo := &stubRepo{
		t: t,
		updateFunc: func(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
			gotShortName = shortName
			return domain.Link{ID: id, OriginalURL: originalURL, ShortName: shortName}, nil
		},
	}

	svc := New(repo, nil, nil)
	link, err := svc.Update(ctx, 1, "https://example.com/new", "")
	require.NoError(t, err)
	require.NotEmpty(t, gotShortName)
	require.NoError(t, domain.ValidateShortName(gotShortName))
	require.Equal(t, gotShortName, link.ShortName)
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

	svc := New(repo, nil, nil)
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

	svc := New(repo, nil, nil)
	_, err := svc.Update(ctx, 1, "https://example.com/new", "ab_cd")
	require.ErrorIs(t, err, domain.ErrInvalidShortName)
}

func TestServiceUpdate_ConflictReturnsError(t *testing.T) {
	ctx := context.Background()

	repo := &stubRepo{
		t: t,
		updateFunc: func(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
			return domain.Link{}, domain.ErrShortNameConflict
		},
	}

	svc := New(repo, nil, nil)
	_, err := svc.Update(ctx, 1, "https://example.com/new", "conflict")
	require.ErrorIs(t, err, domain.ErrShortNameConflict)
}

func TestServiceUpdate_NotFound(t *testing.T) {
	ctx := context.Background()

	repo := &stubRepo{
		t: t,
		updateFunc: func(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
			return domain.Link{}, domain.ErrNotFound
		},
	}

	svc := New(repo, nil, nil)
	_, err := svc.Update(ctx, 1, "https://example.com/new", "abcd")
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestServiceUpdate_InvalidOriginalURL(t *testing.T) {
	ctx := context.Background()

	repo := &stubRepo{
		t: t,
		updateFunc: func(ctx context.Context, id int64, originalURL, shortName string) (domain.Link, error) {
			t.Fatalf("Update should not be called")
			return domain.Link{}, nil
		},
	}

	svc := New(repo, nil, nil)
	_, err := svc.Update(ctx, 1, "not-a-url", "abcd")
	require.ErrorIs(t, err, domain.ErrInvalidURL)
}

func TestServiceRedirect_VisitCreateFailureDoesNotFail(t *testing.T) {
	ctx := context.Background()
	link := domain.Link{
		ID:          1,
		OriginalURL: "https://example.com",
		ShortName:   "code",
	}
	var createCalls int

	repo := &stubRepo{
		t: t,
		getByShortNameFunc: func(ctx context.Context, shortName string) (domain.Link, error) {
			require.Equal(t, "code", shortName)
			return link, nil
		},
	}

	visitsRepo := &stubVisitsRepo{
		t: t,
		createFunc: func(ctx context.Context, visit domain.LinkVisit) (int64, error) {
			createCalls++
			return 0, errors.New("write failed")
		},
	}

	svc := New(repo, visitsRepo, nil)
	url, status, err := svc.Redirect(ctx, "code", VisitMeta{})
	require.NoError(t, err)
	require.Equal(t, link.OriginalURL, url)
	require.Equal(t, redirectStatusFound, status)
	require.Equal(t, 1, createCalls)
}
