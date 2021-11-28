package tenancy

import (
	"context"
	"database/sql"
)

// Executor can perform SQL queries and is tenant-scoped.
type TExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	isTenantScoped()
}

// TContextExecutor can perform SQL queries with context and is tenant-scoped.
type TContextExecutor interface {
	TExecutor

	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Transactor can commit and rollback, on top of being able to execute queries and is tenant-scoped.
type TTransactor interface {
	Commit() error
	Rollback() error

	TExecutor
}
