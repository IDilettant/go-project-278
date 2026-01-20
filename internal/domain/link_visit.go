package domain

import "time"

type LinkVisit struct {
	ID        int64
	LinkID    int64
	CreatedAt time.Time
	IP        string
	UserAgent string
	Referer   string
	Status    int
}
