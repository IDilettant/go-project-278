package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"code/internal/adapters/http/dto"
)

func (h *Handler) ListLinkVisits(c *gin.Context) {
	rng, _, hasRange, err := parseListRange(rangeValue(c))
	if err != nil {
		writeInvalidRange(c)
		return
	}

	var (
		items []dto.LinkVisitResponse
		total int64
	)

	if hasRange {
		visits, count, err := h.svc.ListVisits(c.Request.Context(), rng)
		if err != nil {
			h.fail(c, err)
			return
		}

		total = count
		items = make([]dto.LinkVisitResponse, 0, len(visits))
		for _, visit := range visits {
			items = append(items, dto.FromVisit(visit))
		}
	} else {
		visits, err := h.svc.ListVisitsAll(c.Request.Context())
		if err != nil {
			h.fail(c, err)
			return
		}

		items = make([]dto.LinkVisitResponse, 0, len(visits))
		for _, visit := range visits {
			items = append(items, dto.FromVisit(visit))
		}
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
