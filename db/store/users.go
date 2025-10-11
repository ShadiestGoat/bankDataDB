package store

import (
	"context"

	"github.com/shadiestgoat/bankDataDB/snownode"
)

type User struct {
	ID       string
	Name     string
	Password string
}

func (s *DBStore) GetUserByName(ctx context.Context, name string) (*User, error) {
	var (
		userID string
		realP  string
	)
	err := s.db.QueryRow(ctx, `SELECT id, password FROM users WHERE username = $1`, name).Scan(&userID, &realP)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       userID,
		Name:     name,
		Password: realP,
	}, nil
}

// Create a user in the DB
// Returns the ID & an err
// The password should be encrypted
func (s *DBStore) NewUser(ctx context.Context, username string, password []byte) (string, error) {
	id := snownode.NewID()
	_, err := s.db.Exec(ctx, `INSERT INTO users (id, username, password) VALUES ($1, $2, $3)`, id, username, password)

	return id, err
}
