package store

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/db"
	"github.com/shadiestgoat/bankDataDB/snownode"
)

func (s *DBStore) MappingGetAll(ctx context.Context, authorID string) ([]*data.Mapping, error) {
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
		return scanMappingRow(row)
	})
}

func (s *DBStore) MappingGetByID(ctx context.Context, authorID, mappingID string) (*data.Mapping, error) {
	row := s.db.QueryRow(ctx, `
		SELECT
			id, name, priority,
			trans_text, trans_amount,
			res_name, res_category
		FROM mappings
		WHERE author_id = $1 AND id = $2 ORDER BY priority DESC`,
		authorID, mappingID,
	)

	m, err := scanMappingRow(row)
	if err != nil {
		if db.NoRows(err) {
			return nil, nil
		}
		return nil, err
	}

	return m, nil
}

func scanMappingRow(row interface { Scan(dest ...any) error }) (*data.Mapping, error) {
	mapping := &data.Mapping{}
	var rawRegex *string

	err := row.Scan(
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
}

func (s *DBStore) MappingInsert(ctx context.Context, authorID string, m *data.Mapping) (string, error) {
	id := snownode.NewID()

	_, err := s.db.Exec(
		ctx,
		`INSERT INTO mappings (
			id, author_id, name, priority,
			trans_text, trans_amount, res_name, res_category
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, authorID, m.Name, m.Priority,
		m.InpText.TextNil(), m.InpAmt, m.ResName, m.ResCategoryID,
	)
	if err != nil {
		return "", err
	}

	return id, nil
}
