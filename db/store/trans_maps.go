package store

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/utils/erriter"
)

func (s *DBStore) TransMapsMapExisting(ctx context.Context, updateName bool, newVal any, authorID string, amtMatcher *float64, txtMatcher *regexp.Regexp) (*erriter.Iter[string], int, error) {
	args := pgx.NamedArgs{
		"author_id": authorID,
		"new_val":   newVal,
	}

	conditions := []string{}

	if amtMatcher != nil {
		conditions = append(conditions, "amount = @amt")
		args["amt"] = *amtMatcher
	}
	if txtMatcher != nil {
		conditions = append(conditions, "description ~ @desc")
		args["desc"] = txtMatcher.String()
	}

	col := "resolved_category"
	if updateName {
		col = "resolved_name"
	}

	rows, err := s.db.Query(
		ctx,
		fmt.Sprintf(
			`
			UPDATE
				transactions
			SET
				%s = @new_val
			WHERE
				%s IS NULL
					AND
				author_id = @author_id
					AND
				%s
			RETURNING id`,
			col, col, strings.Join(conditions, " AND "),
		),
		args,
	)
	if err != nil {
		return nil, 0, err
	}

	errIter := erriter.New(func(yield func(string) bool) error {
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			if !yield(id) {
				break
			}
		}

		rows.Close()
		return rows.Err()
	})

	errIter.Close = func() {
		rows.Close()
	}

	return errIter, int(rows.CommandTag().RowsAffected()), nil
}

func (s *DBStore) TransMapsInsert(ctx context.Context, transIDs erriter.Iter[string], mappingID string, mappedName bool) error {
	transIDs.Trans = func(v string) []any {
		return []any{v, mappingID, mappedName}
	}

	_, err := s.db.CopyFrom(
		ctx,
		pgx.Identifier{`mapped_transactions`},
		[]string{`trans_id`, `mapping_id`, `updated_name`},
		&transIDs,
	)

	return err
}

type TransMapsBatch struct {
	// trans_id, mapping_id, updated_name
	Rows [][]any
}

func (t *TransMapsBatch) Insert(transID, mappingID string, updatedName bool) {
	t.Rows = append(t.Rows, []any{transID, mappingID, updatedName})
}

func (s *DBStore) TransMapsInsertBatch(ctx context.Context, b *TransMapsBatch) error {
	_, err := s.db.CopyFrom(
		ctx,
		pgx.Identifier{`mapped_transactions`},
		[]string{`trans_id`, `mapping_id`, `updated_name`},
		pgx.CopyFromRows(b.Rows),
	)

	return err
}
