package links

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAndValidateSort_Default(t *testing.T) {
	def := Sort{Field: SortFieldID, Order: SortDesc}
	got, err := NormalizeAndValidateSort(Sort{}, def, AllowedLinksSortFields())
	require.NoError(t, err)
	require.Equal(t, def, got)
}

func TestNormalizeAndValidateSort_InvalidOrder(t *testing.T) {
	_, err := NormalizeAndValidateSort(
		Sort{Field: SortFieldID, Order: SortOrder("DOWN")},
		DefaultLinksSort,
		AllowedLinksSortFields(),
	)
	require.True(t, errors.Is(err, ErrInvalidSort))
}

func TestNormalizeAndValidateSort_UnknownField(t *testing.T) {
	_, err := NormalizeAndValidateSort(
		Sort{Field: SortField("nope"), Order: SortAsc},
		DefaultLinksSort,
		AllowedLinksSortFields(),
	)
	require.True(t, errors.Is(err, ErrInvalidSort))
}

func TestNormalizeAndValidateSort_TrimsAndNormalizes(t *testing.T) {
	raw := Sort{Field: SortField("  ID "), Order: SortOrder(" asc ")}
	got, err := NormalizeAndValidateSort(raw, DefaultLinksSort, AllowedLinksSortFields())
	require.NoError(t, err)
	require.Equal(t, Sort{Field: SortFieldID, Order: SortAsc}, got)
}

func TestNormalizeAndValidateSort_NilAllowed(t *testing.T) {
	_, err := NormalizeAndValidateSort(
		Sort{Field: SortFieldID, Order: SortAsc},
		DefaultLinksSort,
		nil,
	)
	require.True(t, errors.Is(err, ErrInvalidSort))
}
