package dto

import (
	"time"

	"code/internal/domain"
)

type LinkVisitResponse struct {
	ID        int64     `json:"id" example:"5"`
	LinkID    int64     `json:"link_id" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2025-10-31T13:01:43Z"`
	IP        string    `json:"ip" example:"172.18.0.1"`
	UserAgent string    `json:"user_agent" example:"curl/8.5.0"`
	Referer   string    `json:"reffer" example:"https://example.com"`
	Status    int       `json:"status" example:"302"`
}

func FromVisit(visit domain.LinkVisit) LinkVisitResponse {
	return LinkVisitResponse{
		ID:        visit.ID,
		LinkID:    visit.LinkID,
		CreatedAt: visit.CreatedAt,
		IP:        visit.IP,
		UserAgent: visit.UserAgent,
		Referer:   visit.Referer,
		Status:    visit.Status,
	}
}
