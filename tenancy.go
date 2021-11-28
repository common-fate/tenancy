package tenancy

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

var ctxKey = &contextKey{"tenantedDatabaseConnection"}
var tenantIDKey = &contextKey{"tenantID"}

type contextKey struct {
	name string
}

var ErrNoTenantSet = errors.New("tenant id not set in context")

// FromContext finds the tenanted database connection from the context. REQUIRES tenancy.Open() to have run.
func FromContext(ctx context.Context) *Conn {
	return ctx.Value(ctxKey).(*Conn)
}

// GetID finds the tenant ID from the context. REQUIRES tenancy.Open() to have run.
func GetID(ctx context.Context) string {
	return ctx.Value(tenantIDKey).(string)
}

type Conn struct {
	*sql.Conn
}

// meet the TExecutor interface
func (tc *Conn) isTenantScoped() {}

// TTx is a tenancy-scoped database transaction.
// This type is exported so that you can write methods in your application
// which enforce that a tenancy-scoped transaction must be passed to them.
type TTx struct {
	*sql.Tx
}

// meet the TExecutor interface
func (t *TTx) isTenantScoped() {}

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

func (tc *Conn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*TTx, error) {
	tx, err := tc.Conn.BeginTx(ctx, opts)
	return &TTx{tx}, err
}

// Open sets the tenant and returns a database connection scoped to the tenant
// it returns a new context which contains the tenant ID and the connection
func Open(ctx context.Context, db *sql.DB, tenantID string) (*Conn, context.Context, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, nil, err
	}
	_, err = conn.ExecContext(ctx, "select set_tenant($1)", tenantID)
	if err != nil {
		closeError := conn.Close()
		return nil, nil, errors.Wrap(closeError, err.Error())
	}

	tConn := &Conn{conn}

	newCtx := context.WithValue(ctx, ctxKey, tConn)
	newCtx = context.WithValue(newCtx, tenantIDKey, tenantID)

	return tConn, newCtx, nil
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
