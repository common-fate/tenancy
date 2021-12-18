package tenancy

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type contextKey struct {
	name string
}

var ctxKey = &contextKey{"tenantedDatabaseConnection"}
var tenantIDKey = &contextKey{"tenantID"}

var ErrNoTenantSet = errors.New("tenant id not set in context")

type Conn struct {
	*sql.Conn
}

// TTx is a tenancy-scoped database transaction.
// This type is exported so that you can write methods in your application
// which enforce that a tenancy-scoped transaction must be passed to them.
type TTx struct {
	*sql.Tx
}
type Pool struct {
	db          *sql.DB
	connections []*sql.Conn
}

// Open sets the tenant and returns a database connection scoped to the tenant
// it returns a new context which contains the tenant ID and the connection
func Open(ctx context.Context, db *sql.DB, tenantID string) (*Pool, context.Context, error) {
	tPool := &Pool{db: db, connections: []*sql.Conn{}}
	newCtx := context.WithValue(ctx, ctxKey, tPool)
	newCtx = context.WithValue(newCtx, tenantIDKey, tenantID)
	return tPool, newCtx, nil
}

// Close unsets the current tenant before returning the connections
func Close(ctx context.Context, p *Pool) error {
	var closingError error
	for _, conn := range p.connections {
		// check for connections that have already be closed
		if conn != nil {
			_, err := conn.ExecContext(ctx, "select set_tenant('')")
			if err != nil {
				closingError = errors.Wrap(closingError, err.Error())
			}
			err = conn.Close()
			if err != nil {
				closingError = errors.Wrap(closingError, err.Error())
			}
		}
	}
	// will be nil if there were no errors while closing connections
	return closingError
}

// FromContext finds the tenanted database connection pool from the context. REQUIRES tenancy.Open() to have run.
func FromContext(ctx context.Context) *Pool {
	return ctx.Value(ctxKey).(*Pool)
}

// GetID finds the tenant ID from the context. REQUIRES tenancy.Open() to have run.
func GetID(ctx context.Context) string {
	return ctx.Value(tenantIDKey).(string)
}

// meet the TExecutor interface
func (tc *Conn) isTenantScoped() {}

// meet the TExecutor interface
func (t *TTx) isTenantScoped() {}

// meet the TExecutor interface
func (p *Pool) isTenantScoped() {}

// Exec is implemented with background context to satisfy interface
func (p *Pool) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(context.Background(), query, args...)
}

// Query is implemented with background context to satisfy interface
func (p *Pool) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(context.Background(), query, args...)
}

// QueryRow is implemented with background context to satisfy interface
func (p *Pool) QueryRow(query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(context.Background(), query, args...)
}

// Exec is implemented with background context to satisfy interface
func (p *Pool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	conn, err := p.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return conn.ExecContext(ctx, query, args...)
}

// Query is implemented with background context to satisfy interface
func (p *Pool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	conn, err := p.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return conn.QueryContext(ctx, query, args...)
}

// QueryRow is implemented with background context to satisfy interface
func (p *Pool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	conn, err := p.Conn(ctx)
	if err != nil {
		// we panic here because we can't return an error from via the sql.Row type
		panic(err)
	}
	return conn.QueryRowContext(ctx, query, args...)
}

// checks out a new tenanted connection for this transaction
func (p *Pool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*TTx, error) {
	conn, err := p.Conn(ctx)
	if err != nil {
		return nil, err
	}
	tx, err := conn.BeginTx(ctx, opts)
	return &TTx{tx}, err
}

// Conn will create a new connection and apply the current tenant
// The opened connection is registered to a pool internally which is closed when tenancy.Close(ctx, Tconn) is called
func (p *Pool) Conn(ctx context.Context) (*sql.Conn, error) {
	tenantID := GetID(ctx)
	conn, err := p.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	_, err = conn.ExecContext(ctx, "select set_tenant($1)", tenantID)
	if err != nil {
		closeError := conn.Close()
		return nil, errors.Wrap(closeError, err.Error())
	}
	p.connections = append(p.connections, conn)
	return conn, nil
}

// Exec is implemented with background context to satisfy interface
func (c *Conn) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.ExecContext(context.Background(), query, args...)
}

// Query is implemented with background context to satisfy interface
func (c *Conn) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.QueryContext(context.Background(), query, args...)
}

// QueryRow is implemented with background context to satisfy interface
func (c *Conn) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.QueryRowContext(context.Background(), query, args...)
}

func (c *Conn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*TTx, error) {
	tx, err := c.Conn.BeginTx(ctx, opts)
	return &TTx{tx}, err
}
