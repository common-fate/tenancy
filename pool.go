package tenancy

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type Pool struct {
	db          *sql.DB
	connections []*sql.Conn
}

var poolCtxKey = &contextKey{"tenantedDatabaseConnectionPool"}

// meet the TExecutor interface
func (tc *Pool) isTenantScoped() {}

// Exec is implemented with background context to satisfy interface
func (tc *Pool) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tc.db.ExecContext(context.Background(), query, args...)
}

// Query is implemented with background context to satisfy interface
func (tc *Pool) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tc.db.QueryContext(context.Background(), query, args...)
}

// QueryRow is implemented with background context to satisfy interface
func (tc *Pool) QueryRow(query string, args ...interface{}) *sql.Row {
	return tc.db.QueryRowContext(context.Background(), query, args...)
}

// Exec is implemented with background context to satisfy interface
func (tc *Pool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	conn, err := tc.CheckoutConn(ctx)
	if err != nil {
		return nil, err
	}
	return conn.ExecContext(ctx, query, args...)
}

// Query is implemented with background context to satisfy interface
func (tc *Pool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	conn, err := tc.CheckoutConn(ctx)
	if err != nil {
		return nil, err
	}
	return conn.QueryContext(ctx, query, args...)
}

// QueryRow is implemented with background context to satisfy interface
func (tc *Pool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	conn, err := tc.CheckoutConn(ctx)
	if err != nil {
		// we panic here because we can't return an error from via the sql.Row type
		panic(err)
	}
	return conn.QueryRowContext(ctx, query, args...)
}

// checks out a new tenanted connection for this transaction
func (tc *Pool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*TTx, error) {
	conn, err := tc.CheckoutConn(ctx)
	if err != nil {
		return nil, err
	}
	tx, err := conn.BeginTx(ctx, opts)
	return &TTx{tx}, err
}

// Open sets the tenant and returns a database connection scoped to the tenant
// it returns a new context which contains the tenant ID and the connection
func OpenPool(ctx context.Context, db *sql.DB, tenantID string) (*Pool, context.Context, error) {
	tPool := &Pool{db: db, connections: []*sql.Conn{}}
	newCtx := context.WithValue(ctx, poolCtxKey, tPool)
	newCtx = context.WithValue(newCtx, tenantIDKey, tenantID)
	return tPool, newCtx, nil
}

// Close unsets the current tenant before returning the connection to the toplevel pool
func ClosePool(ctx context.Context, tc *Pool) error {
	var closingError error
	for _, conn := range tc.connections {
		// check for connections that have already be closed
		if conn != nil {
			_, err := conn.ExecContext(ctx, "select set_tenant('')")
			if err != nil {
				closingError = errors.Wrap(closingError, err.Error())
				err = conn.Close()
				if err != nil {
					closingError = errors.Wrap(closingError, err.Error())
				}
			}
		}
	}

	// will be nill if there were no errors while closing connections
	return closingError
}

// CheckoutConn will create a new connection and apply the current tenant
// The opened connection is registered to a pool internally which is closed when tenancy.Close(ctx, Tconn) is called
func (tc *Pool) CheckoutConn(ctx context.Context) (*sql.Conn, error) {
	tenantID := GetID(ctx)
	conn, err := tc.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	_, err = conn.ExecContext(ctx, "select set_tenant($1)", tenantID)
	if err != nil {
		closeError := conn.Close()
		return nil, errors.Wrap(closeError, err.Error())
	}
	tc.connections = append(tc.connections, conn)
	return conn, nil
}
