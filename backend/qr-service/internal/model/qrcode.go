package model

import "time"

type QrCode struct {
	ID           string    `json:"id"`
	Label        string    `json:"label"`
	URL          string    `json:"url"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"-"`
	CreatedAtIso string    `json:"createdAtIso"`
}

func (q QrCode) NormalizeForResponse() QrCode {
	q.CreatedAtIso = q.CreatedAt.UTC().Format(time.RFC3339)
	return q
}
