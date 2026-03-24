package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostGreClient struct {
	db *sql.DB
}

func NewPostGreClient(db *sql.DB) *PostGreClient {
	return &PostGreClient{db: db}
}

func (n *PostGreClient) Ping() error {
	if n == nil || n.db == nil {
		return fmt.Errorf("postgres db is not initialized")
	}
	if err := n.db.Ping(); err != nil {
		return fmt.Errorf("postgres ping failed: %w", err)
	}
	return nil
}
