package ember

import (
	"context"
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	goqusqlite3 "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	"github.com/jmoiron/sqlx"
)

type Query interface {
	ToSQL() (string, []any, error)
}

type DB interface {
	Query(ctx context.Context, query Query) (*sql.Rows, error)
	QueryRow(ctx context.Context, query Query) (*sql.Row, error)
	Exec(ctx context.Context, query Query) (sql.Result, error)

	Single(ctx context.Context, query Query, dest any) error
	Multiple(ctx context.Context, query Query, dest any) error
}

type ErrorHandler func(err error) error

var _ DB = (*Database)(nil)

type Database struct {
	*sqlx.DB

	ErrorHandler ErrorHandler
}

func OpenDatabase(driver, dataSourceName string) (*Database, error) {
	db, err := sql.Open(driver, dataSourceName)
	if err != nil {
		return nil, err
	}

	return &Database{
		DB: sqlx.NewDb(db, driver),
	}, nil
}

func (s *Database) Begin() (*Tx, error) {
	tx, err := s.DB.Beginx()
	if err != nil {
		return nil, err
	}

	return &Tx{
		Tx:           tx,
		ErrorHandler: s.ErrorHandler,
	}, nil
}

func (s *Database) handleErr(err error) error {
	if s.ErrorHandler != nil {
		return s.ErrorHandler(err)
	}

	return err
}

func (s *Database) Exec(ctx context.Context, query Query) (sql.Result, error) {
	sql, params, err := query.ToSQL()
	if err != nil {
		return nil, s.handleErr(err)
	}

	return s.ExecContext(ctx, sql, params...)
}

func (s *Database) Multiple(ctx context.Context, query Query, dest any) error {
	sql, params, err := query.ToSQL()
	if err != nil {
		return s.handleErr(err)
	}

	return s.SelectContext(ctx, dest, sql, params...)
}

// Query implements Database.
func (s *Database) Query(ctx context.Context, query Query) (*sql.Rows, error) {
	sql, params, err := query.ToSQL()
	if err != nil {
		return nil, s.handleErr(err)
	}

	return s.QueryContext(ctx, sql, params...)
}

// QueryRow implements Database.
func (s *Database) QueryRow(ctx context.Context, query Query) (*sql.Row, error) {
	sql, params, err := query.ToSQL()
	if err != nil {
		return nil, s.handleErr(err)
	}

	return s.QueryRowContext(ctx, sql, params...), nil
}

// Single implements Database.
func (s *Database) Single(ctx context.Context, query Query, dest any) error {
	sql, params, err := query.ToSQL()
	if err != nil {
		return s.handleErr(err)
	}

	return s.GetContext(ctx, dest, sql, params...)
}

var _ DB = (*Tx)(nil)

type Tx struct {
	*sqlx.Tx

	ErrorHandler ErrorHandler
}

func (s *Tx) handleErr(err error) error {
	if s.ErrorHandler != nil {
		return s.ErrorHandler(err)
	}

	return err
}

func (s *Tx) Commit() error {
	return s.Tx.Commit()
}

func (s *Tx) Rollback() error {
	return s.Tx.Rollback()
}

func (s *Tx) Exec(ctx context.Context, query Query) (sql.Result, error) {
	sql, params, err := query.ToSQL()
	if err != nil {
		return nil, s.handleErr(err)
	}

	return s.ExecContext(ctx, sql, params...)
}

func (s *Tx) Multiple(ctx context.Context, query Query, dest any) error {
	sql, params, err := query.ToSQL()
	if err != nil {
		return s.handleErr(err)
	}

	return s.SelectContext(ctx, dest, sql, params...)
}

// Query implements Database.
func (s *Tx) Query(ctx context.Context, query Query) (*sql.Rows, error) {
	sql, params, err := query.ToSQL()
	if err != nil {
		return nil, s.handleErr(err)
	}

	return s.QueryContext(ctx, sql, params...)
}

// QueryRow implements Database.
func (s *Tx) QueryRow(ctx context.Context, query Query) (*sql.Row, error) {
	sql, params, err := query.ToSQL()
	if err != nil {
		return nil, s.handleErr(err)
	}

	return s.QueryRowContext(ctx, sql, params...), nil
}

// Single implements Database.
func (s *Tx) Single(ctx context.Context, query Query, dest any) error {
	sql, params, err := query.ToSQL()
	if err != nil {
		return s.handleErr(err)
	}

	return s.GetContext(ctx, dest, sql, params...)
}

// Simple wrappers for retriving single row from the database
func Single[T any](db DB, ctx context.Context, query Query) (T, error) {
	var res T
	err := db.Single(ctx, query, &res)
	return res, err
}

// Simple wrappers for retriving multiple row from the database
func Multiple[T any](db DB, ctx context.Context, query Query) ([]T, error) {
	var res []T
	err := db.Multiple(ctx, query, &res)
	return res, err
}

type RawQuery struct {
	Sql    string
	Params []any
}

func (q RawQuery) ToSQL() (string, []any, error) {
	return q.Sql, q.Params, nil
}

type DialectWrapper struct {
	goqu.DialectWrapper
}

func (dw DialectWrapper) From(table ...any) *goqu.SelectDataset {
	return dw.DialectWrapper.From(table...).Prepared(true)
}

func (dw DialectWrapper) Select(cols ...any) *goqu.SelectDataset {
	return dw.DialectWrapper.Select(cols...).Prepared(true)
}

func (dw DialectWrapper) Update(table any) *goqu.UpdateDataset {
	return dw.DialectWrapper.Update(table).Prepared(true)
}

func (dw DialectWrapper) Insert(table any) *goqu.InsertDataset {
	return dw.DialectWrapper.Insert(table).Prepared(true)
}

func (dw DialectWrapper) Delete(table any) *goqu.DeleteDataset {
	return dw.DialectWrapper.Delete(table).Prepared(true)
}

func (dw DialectWrapper) Truncate(table ...any) *goqu.TruncateDataset {
	return dw.DialectWrapper.Truncate(table...).Prepared(true)
}

func SqliteDialect() DialectWrapper {
	return DialectWrapper{
		DialectWrapper: goqu.Dialect("pyrin_custom_sqlite"),
	}
}

func init() {
	opts := goqusqlite3.DialectOptions()
	opts.SupportsReturn = true
	goqu.RegisterDialect("pyrin_custom_sqlite", opts)
}
