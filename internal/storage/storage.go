package storage

import "context"

type Data struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	DeletedFlag bool   `json:"is_deleted"`
}

type DataKeeper interface {
	Save(c context.Context, d Data) error
	Get(c context.Context, s string) (Data, error)
	GetAll(c context.Context, userID int) ([]Data, error)
	BatchDelete(short []string) error

	HealthCheck() error

	Close() error
}
