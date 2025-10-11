package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/db"
	"github.com/shadiestgoat/bankDataDB/snownode"
)

func (s *DBStore) InsertCheckpoint(batch *pgx.Batch, date time.Time, amt float64) {
	batch.Queue(`INSERT INTO checkpoints (created_at, amount) VALUES ($1, $2) ON CONFLICT DO UPDATE SET amount = $2`, date, amt)
}

type TransactionBatch struct {
	Rows [][]any
}

func (t *TransactionBatch) Insert(authedAt, settledAt time.Time, authorID string, desc string, amt float64, resolvedName *string, resolvedCatID *string) {
	t.Rows = append(t.Rows, []any{
		snownode.NewID(),
		authorID,
		authedAt,
		settledAt,
		desc,
		amt,
		resolvedName,
		resolvedCatID,
	})
}

func (s *DBStore) InsertTransactions(ctx context.Context, b *TransactionBatch) (int64, error) {
	return s.db.CopyFrom(ctx, pgx.Identifier{`transactions`}, []string{
		`id`,
		`author_id`,
		`authed_at`,
		`settled_at`,
		`description`,
		`amount`,
		`resolved_name`,
		`resolved_category`,
	}, pgx.CopyFromRows(b.Rows))
}

func (s *DBStore) GetTransactions(ctx context.Context, authorID string, amount, offset int, orderColumn string, asc bool) ([]*data.Transactions, error) {
	rows, err := s.db.Query(
		ctx,
		fmt.Sprintf(`
			SELECT
				settled_at, authed_at, description, amount, resolved_name, resolved_category
			FROM transactions
			WHERE author_id = $1
			ORDER BY %s %s
			LIMIT $3
			OFFSET $4
		`, orderColumn, db.AscKey(asc)),
		authorID, orderColumn, amount, offset,
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (*data.Transactions, error) {
		t := &data.Transactions{}
		err := row.Scan(&t.SettledAt, &t.AuthedAt, &t.Desc, &t.Amount, &t.ResolvedName, &t.ResolvedCategoryID)
		return t, err
	})
}
