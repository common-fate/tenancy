package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/common-fate/tenancy"
)

type Logger interface {
	Error(err error)
}

// Tenancy is a middleware that injects a tenanted *sql.Conn into the context of each
// request. This middleware accepts a function which should extract the current tenant from context.
// The expectation is that you will have some auth middleware which can inject the current tenant into the ctx in a previous step.
//
// When a query is made using the Pool, a new connection will be opened and tenancy applied to it.
// These connections are tracked internally and will be closed at the end of the request
//
// If you start a transaction from the Pool, a single connection will be opened and used for all queries for the duration of the transaction
//
// You may also open a single Conn from the pool which will be closed when the pool is closed
func Tenancy(db *sql.DB, log Logger, getTenantIDFromCtx func(context.Context) string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tc, ctx, err := tenancy.Open(ctx, db, getTenantIDFromCtx(ctx))
			// Close the connection after completing http handling
			defer func() {
				if tc != nil {
					err = tenancy.Close(ctx, tc)
					if err != nil && log != nil {
						log.Error(err)
					}
				}
			}()

			if err != nil {
				if log != nil {
					log.Error(err)
				}
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}
