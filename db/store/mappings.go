package store

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/data"
)

func (s *DBStore) updateBasedOnMapping(ctx context.Context, col string, newVal any, authorID string, amtMatcher *float64, txtMatcher *regexp.Regexp) (int, error) {
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

	res, err := s.db.Exec(
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
				`,
			col, col,
		)+strings.Join(conditions, " AND "),
		args,
	)
	if err != nil {
		return 0, err
	}

	return int(res.RowsAffected()), nil
}

func (s *DBStore) UpdateTransCatsUsingMapping(ctx context.Context, newCategoryID string, authorID string, amtMatcher *float64, txtMatcher *regexp.Regexp) (int, error) {
	return s.updateBasedOnMapping(ctx, `resolved_category`, newCategoryID, authorID, amtMatcher, txtMatcher)
}

func (s *DBStore) UpdateTransNamesUsingMapping(ctx context.Context, newName string, authorID string, amtMatcher *float64, txtMatcher *regexp.Regexp) (int, error) {
	return s.updateBasedOnMapping(ctx, `resolved_name`, newName, authorID, amtMatcher, txtMatcher)
}

func (s *DBStore) GetMappingsForAuthor(ctx context.Context, authorID string) ([]*data.Mapping, error) {
	rows, err := s.db.Query(
		ctx,
		`
		SELECT
			id, name, priority,
			trans_text, trans_amount,
			res_name, res_category
		FROM mappings
		WHERE author_id = $1 ORDER BY priority DESC`,
		authorID,
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (*data.Mapping, error) {
		mapping := &data.Mapping{}
		var rawRegex *string

		err = row.Scan(
			&mapping.ID, &mapping.Name, &mapping.Priority,
			&rawRegex, &mapping.InpAmt,
			&mapping.ResName, &mapping.ResCategoryID,
		)
		if err != nil {
			return nil, err
		}

		if rawRegex != nil {
			reg, err := regexp.CompilePOSIX(*rawRegex)
			if err != nil {
				// TODO: Log smt here idk
			} else {
				mapping.InpText = (*data.MarshallableRegexp)(reg)
			}
		}

		return mapping, nil
	})
}
