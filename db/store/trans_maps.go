package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/data"
)

func (s *DBStore) TransMapsMapExisting(ctx context.Context, updateName bool, authorID string, m *data.Mapping) (int, error) {
	// TODO: In the future, maybe we can limit this to X rows per chunk or smt?
	args := pgx.NamedArgs{
		"author_id": authorID,
		"priority": m.Priority,
		"match_name": updateName,
		"mapping_id": m.ID,
	}
	if updateName {
		args["new_val"] = *m.ResName
	} else {
		args["new_val"] = *m.ResCategoryID
	}

	conditions := []string{}

	if m.InpAmt != nil {
		conditions = append(conditions, "amount = @amt")
		args["amt"] = *m.InpAmt
	}
	if m.InpText != nil {
		conditions = append(conditions, "description ~ @desc")
		args["desc"] = *m.InpText.TextNil()
	}

	col := "resolved_category"
	if updateName {
		col = "resolved_name"
	}

	res, err := s.db.Exec(
		ctx,
		fmt.Sprintf(
			`
			WITH eligible AS (
				SELECT t.id, mapping_id
				FROM transactions AS t
				LEFT JOIN mapped_transactions ON t.id = trans_id AND updated_name = @match_name
				LEFT JOIN mappings AS m ON m.id = mapping_id
				WHERE
					t.author_id = @author_id
						AND
					-- is mapped or not manually overridden
					(priority IS NOT NULL OR %s IS NULL)
						AND
					-- is not mapped or has lower priority
					(priority IS NULL OR priority < @priority)
			), deleted AS (
				DELETE FROM mapped_transactions mp
				USING eligible e
				WHERE mp.mapping_id = e.mapping_id AND updated_name = @match_name
			), updated AS (
				UPDATE transactions
				SET %s = @new_val
				FROM eligible
				WHERE transactions.id = eligible.id AND %s
				RETURNING transactions.id
			) INSERT INTO mapped_transactions (trans_id, mapping_id, updated_name)
				SELECT id AS trans_id, @mapping_id AS mapping_id, @match_name AS updated_name FROM updated
			`, col, col, strings.Join(conditions, " AND "),
		),
		args,
	)
	if err != nil {
		return 0, err
	}

	return int(res.RowsAffected()), nil
}

func (s *DBStore) TransMapsInsert(ctx context.Context, transIDs []string, mappingID string, mappedName bool) error {
	i := 0
	_, err := s.db.CopyFrom(
		ctx,
		pgx.Identifier{`mapped_transactions`},
		[]string{`trans_id`, `mapping_id`, `updated_name`},
		pgx.CopyFromFunc(func() (row []any, err error) {
			if i >= len(transIDs) {
				return nil, nil
			}

			defer func() {
				i++
			}()

			return []any{transIDs[i], mappingID, mappedName}, nil
		}),
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
