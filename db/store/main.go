package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/db"
)

//go:generate go tool github.com/sqlc-dev/sqlc/cmd/sqlc generate
//go:generate ./../../gen/generate-post.sh
//go:generate go tool github.com/vburenin/ifacemaker -f "*.go" -s DBStore -i Store -p store -c "DONT EDIT: Auto generated" -o "store.go"
//go:generate go tool github.com/vektra/mockery/v3

func NewStore(db db.DBQuerier) *DBStore { return &DBStore{db} }

type DBStore struct{ db db.DBQuerier }

func (s *DBStore) SendBatch(ctx context.Context, b *pgx.Batch) error {
	return s.db.SendBatch(ctx, b).Close()
}
