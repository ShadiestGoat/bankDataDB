package db

import (
	"context"
	"errors"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shadiestgoat/bankDataDB/log"
)

type DBQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type genericDBWithLog[T DBQuerier] struct {
	conn T
	log  log.CtxLogger
}

var regWhitespace = regexp.MustCompile(`\s{2,}`)

func (db *genericDBWithLog[T]) logErr(ctx context.Context, err error, method string, sql string) {
	if err != nil {
		db.log(ctx).Errorw(
			"Error when doing "+method,
			"sql", regWhitespace.ReplaceAllString(sql, " "),
			"error", err,
		)
	}
}

type fakeRowsScanner struct {
	pgx.Rows
	s *fakeRowScanner
}

func (f *fakeRowsScanner) Scan(dst ...any) error {
	return f.s.Scan(dst...)
}

func (f *fakeRowsScanner) Values() ([]any, error) {
	d, err := f.Rows.Values()
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		f.s.log.Errorw("Error doing Query Values", "sql", f.s.sql, "error", err)
	}

	return d, err
}

func (db *genericDBWithLog[T]) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	rows, err := db.conn.Query(ctx, sql, args...)
	db.logErr(ctx, err, "Query", sql)

	return &fakeRowsScanner{
		Rows: rows,
		s: &fakeRowScanner{
			row: rows,
			sql: sql,
			log: db.log(ctx),
			n:   "Query",
		},
	}, err
}

func (db *genericDBWithLog[T]) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return db.conn.SendBatch(ctx, b)
}

type fakeRowScanner struct {
	row pgx.Row
	sql string
	log log.Logger
	n   string
}

func (r *fakeRowScanner) Scan(dst ...any) error {
	err := r.row.Scan(dst...)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		r.log.Errorw("Error doing "+r.n+" Scan", "sql", r.sql, "error", err)
	}

	return err
}

func (db *genericDBWithLog[T]) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	row := db.conn.QueryRow(ctx, sql, args...)

	return &fakeRowScanner{row, sql, db.log(ctx), "QueryRow"}
}

func (db *genericDBWithLog[T]) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	resp, err := db.conn.Exec(ctx, sql, arguments...)
	db.logErr(ctx, err, "(sql) Exec", sql)

	return resp, err
}

// CopyFrom implements pgx.Tx.
func (t *genericDBWithLog[T]) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	resp, err := t.conn.CopyFrom(ctx, tableName, columnNames, rowSrc)
	if err != nil {
		t.logErr(ctx, err, "CopyFrom", "table:"+tableName.Sanitize())
	}

	return resp, err
}

func (db *genericDBWithLog[T]) Begin(ctx context.Context) (pgx.Tx, error) {
	ogTx, err := db.conn.Begin(ctx)
	if err != nil {
		db.log(ctx).Errorf("Failed to make tx", "error", err)
	}

	return &tx{
		genericDBWithLog: genericDBWithLog[pgx.Tx]{
			conn: ogTx,
			log:  db.log,
		},
	}, err
}

type tx struct {
	genericDBWithLog[pgx.Tx]
}

// Commit implements pgx.Tx.
func (t *tx) Commit(ctx context.Context) error {
	err := t.conn.Commit(ctx)
	if err != nil {
		t.log(ctx).Errorw("Error while doing commit in tx", "error", err)
	}

	return err
}

// Conn implements pgx.Tx.
func (t *tx) Conn() *pgx.Conn {
	return t.conn.Conn()
}

// LargeObjects implements pgx.Tx.
func (t *tx) LargeObjects() pgx.LargeObjects {
	return t.conn.LargeObjects()
}

// Prepare implements pgx.Tx.
func (t *tx) Prepare(ctx context.Context, name string, sql string) (*pgconn.StatementDescription, error) {
	stmt, err := t.conn.Prepare(ctx, name, sql)
	if err != nil {
		t.logErr(ctx, err, "Prepare", sql)
	}

	return stmt, err
}

// Rollback implements pgx.Tx.
func (t *tx) Rollback(ctx context.Context) error {
	err := t.conn.Rollback(ctx)
	if err != nil {
		t.logErr(ctx, err, "Rollback", "")
	}

	return err
}

// SendBatch implements pgx.Tx.
func (t *tx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return t.conn.SendBatch(ctx, b)
}
