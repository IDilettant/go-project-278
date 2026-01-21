package dto

import "code/internal/domain"

type LinkResponse struct {
	ID          int64  `json:"id" example:"1"`
	OriginalURL string `json:"original_url" example:"https://example.com"`
	ShortName   string `json:"short_name" example:"abc123"`
	ShortURL    string `json:"short_url" example:"https://example.com/r/abc123"`
}

func FromDomain(link domain.Link, baseURL string) LinkResponse {
	return LinkResponse{
		ID:          link.ID,
		OriginalURL: link.OriginalURL,
		ShortName:   link.ShortName,
		ShortURL:    baseURL + "/r/" + link.ShortName,
	}
}
