package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/adapters/httpapi/dto"
	"code/internal/app/links"
	"code/internal/domain"
)

func (h *Handler) ListLinkVisits(c *gin.Context) {
	rng, _, hasRange, err := parseRangeFromRequest(c)
	if err != nil {
		writeInvalidRange(c)

		return
	}

	rawSort, ok := parseReactAdminSort(c.Query("sort"))
	if !ok {
		h.fail(c, links.ErrInvalidSort)

		return
	}

	sort, err := links.NormalizeAndValidateSort(rawSort, links.DefaultLinkVisitsSort, links.AllowedLinkVisitsSortFields())
	if err != nil {
		h.fail(c, err)

		return
	}

	query := links.LinkVisitsQuery{Sort: sort}
	if hasRange {
		query.Range = &rng
	}

	items, total, err := h.fetchVisits(c, query)
	if err != nil {
		h.fail(c, err)

		return
	}

	if hasRange {
		if len(items) == 0 {
			c.Header("Content-Range", fmt.Sprintf("link_visits */%d", total))
		} else {
			end := rng.Start + len(items) - 1
			c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", rng.Start, end, total))
		}
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) fetchVisits(
	c *gin.Context,
	query links.LinkVisitsQuery,
) ([]dto.LinkVisitResponse, int64, error) {
	visits, total, err := h.svc.ListLinkVisits(c.Request.Context(), query)
	if err != nil {
		return nil, 0, err
	}

	return toVisitResponses(visits), total, nil
}

func toVisitResponses(visits []domain.LinkVisit) []dto.LinkVisitResponse {
	items := make([]dto.LinkVisitResponse, 0, len(visits))
	for _, visit := range visits {
		items = append(items, dto.FromVisit(visit))
	}

	return items
}
