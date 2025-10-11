package db

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

func NoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func AscKey(asc bool) string {
	if asc {
		return "ASC"
	}

	return "DESC"
}
