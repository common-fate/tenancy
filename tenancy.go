package tenancy

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

var ctxKey = &contextKey{"tenantedDatabaseConnection"}

type contextKey struct {
	name string
}

var ErrNoTenantSet = errors.New("tenant id not set in context")

// FromContext finds the tenanted database connection from the context. REQUIRES Middleware to have run.
func FromContext(ctx context.Context) *Conn {
	return ctx.Value(ctxKey).(*Conn)
}

type Conn struct {
	*sql.Conn
}

// Exec is implemented with background context to satisfy interface
func (tc *Conn) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tc.ExecContext(context.Background(), query, args...)
}

// Query is implemented with background context to satisfy interface
func (tc *Conn) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tc.QueryContext(context.Background(), query, args...)
}

// QueryRow is implemented with background context to satisfy interface
func (tc *Conn) QueryRow(query string, args ...interface{}) *sql.Row {
	return tc.QueryRowContext(context.Background(), query, args...)
}

// Open sets the tenant and returns a database connection scoped to the tenant
func Open(ctx context.Context, db *sql.DB, tenantID string) (*Conn, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	_, err = conn.ExecContext(ctx, "select set_tenant($1)", tenantID)
	if err != nil {
		closeError := conn.Close()
		return nil, errors.Wrap(closeError, err.Error())
	}
	return &Conn{conn}, nil
}

// Close unsets the current tenant before returning the connection to the pool
func Close(ctx context.Context, tc *Conn) error {
	_, err := tc.ExecContext(ctx, "select set_tenant('')")
	if err != nil {
		closeError := tc.Close()
		return errors.Wrap(closeError, err.Error())
	}
	return tc.Close()
}
