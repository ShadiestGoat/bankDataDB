package internal

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/db"
	"github.com/shadiestgoat/bankDataDB/snownode"
)

type ResourceValue interface {
	DBCols() []string
	DBColPointers() []any
	Clone() ResourceValue
}

type ResMappings struct {
	data.Mapping
	tmpRegexp *string `json:"-"`
}

func (r *ResMappings) DBCols() []string {
	return []string{`name`, `trans_text`, `trans_amount`, `res_name`, `res_category`, `priority`}
}
func (r *ResMappings) DBColPointers() []any {
	return []any{&r.Name, &r.tmpRegexp, &r.InpAmt, &r.ResName, &r.ResCategoryID, &r.Priority}
}
func (r *ResMappings) Clone() ResourceValue {return &ResMappings{}}
func (r *ResMappings) PostScan() {
	if r.tmpRegexp != nil {
		tmp, err := regexp.CompilePOSIX(*r.tmpRegexp)
		if err != nil {
			r.InpText = (*data.MarshallableRegexp)(tmp)
		}
	}
}

type ResCategories struct {
	Name string `json:"name"`
	Color string `json:"color"`
	Icon string `json:"icon"` // string bc rune can't have multi-bytes
}
func (r *ResCategories) DBCols() []string {
	return []string{`name`, `color`, `icon`}
}
func (r *ResCategories) DBColPointers() []any {
	return []any{&r.Name, &r.Color, &r.Icon}
}
func (r *ResCategories) Clone() ResourceValue {return &ResCategories{}}

func (a *API) ResourceInsert(ctx context.Context, table string, sample ResourceValue, authorID string) (string, error) {
	cols := sample.DBCols()

	args := ""
	for i := range len(cols) + 2 {
		args += fmt.Sprintf("$%d, ", i+1)
	}
	args = args[:len(args) - 2]
	id := snownode.NewID()

	vals := make([]any, len(sample.DBColPointers()))
	for i, p := range sample.DBColPointers() {
		vals[i] = *p.(*any)
	}

	_, err := a.db.Exec(ctx, fmt.Sprintf(
		`INSERT INTO %s (%s, id, author_id) VALUES (%s)`,
		table, strings.Join(cols, ", "), args,
	), append(vals, id, authorID)...)

	return id, err
}

func (a *API) ResourceUpdate(ctx context.Context, table string, sample ResourceValue, id string) error {
	cols := sample.DBCols()

	args := ""
	for i, c := range cols {
		args += fmt.Sprintf("%s = $%d, ", c, i+1)
	}
	args = args[:len(args)-2]

	vals := make([]any, len(sample.DBColPointers()))
	for i, p := range sample.DBColPointers() {
		vals[i] = *p.(*any)
	}

	_, err := a.db.Exec(ctx, fmt.Sprintf(
		`UPDATE %s SET %s WHERE id = $%d`,
		table, args, len(cols)+2,
	), append(vals, id)...)

	return err
}

func (a *API) ResourceVerifyExists(ctx context.Context, table string, authorID, id string) (bool, error) {
	var exist bool
	err := a.db.QueryRow(
		ctx,
		fmt.Sprintf(
			`SELECT EXISTS(SELECT 1 FROM %s WHERE author_id = $1 AND id = $2)`, table,
		),
		authorID, id,
	).Scan(&exist)
	if err != nil && db.NoRows(err) {
		err = nil
	}

	return exist, err
}

func (a *API) ResourceDrop(ctx context.Context, table string, id string) error {
	_, err := a.db.Exec(ctx, `DELETE FROM ` + table + ` WHERE id = $1`, id)
	return  err
}

// Gets a specific resource, but DOESN'T do an author check!
func (a *API) ResourceGetSpecific(ctx context.Context, table string, id string, target ResourceValue) error {
	row := a.db.QueryRow(
		ctx,
		fmt.Sprintf(
			`SELECT %s FROM %s WHERE id = $1`,
			strings.Join(target.DBCols(), ", "), table,
		),
		id,
	)
	return row.Scan(target.DBColPointers()...)
}

func (a *API) ResourceGetAll(ctx context.Context, table string, authorID string, sample ResourceValue) ([]ResourceValue, error) {
	rows, err := a.db.Query(
		ctx,
		fmt.Sprintf(
			`SELECT %s FROM %s WHERE author_id = $1`,
			strings.Join(sample.DBCols(), ", "), table,
		),
		authorID,
	)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (ResourceValue, error) {
		t := sample.Clone()
		err := row.Scan(t.DBColPointers()...)

		return t, err
	})
}

func (a *API) ResourceGetCount(ctx context.Context, table string, authorID string) (int, error) {
	c := 0
	err := a.db.QueryRow(
		ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE author_id = $1`, table),
		authorID,
	).Scan(&c)
	if err != nil {
		return 0, err
	}

	return c, err
}
