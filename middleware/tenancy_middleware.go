package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/common-fate/tenancy"
)

// Tenancy is a middleware that injects a tenanted *sql.Conn into the context of each
// request. This middleware accepts a function which should extract the current tenant from context.
// The expectation is that you will have some auth middleware which can inject the current tenant into the ctx in a previous step.
//
//
func Tenancy(db *sql.DB, getTenantIDFromCtx func(context.Context) string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tc, ctx, err := tenancy.Open(ctx, db, getTenantIDFromCtx(ctx))
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
			// Close the connection after completing http handling
			defer func() {
				err = tenancy.Close(ctx, tc)
				if err != nil {
					panic(err)
				}
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// Tenancy is a middleware that injects a tenantedConnectionPool into the context of each
// request. This middleware accepts a function which should extract the current tenant from context.
// The expectation is that you will have some auth middleware which can inject the current tenant into the ctx in a previous step.
//
// When a query is made using the Pool, a new connection will be opened and tenancy applied to it.
// These connections are tracked internally and will be closed at the end of the request
//
// If you start a transaction from the Pool, a single connection will be opened and used for all queries for the duration of the transaction
func TenancyPool(db *sql.DB, getTenantIDFromCtx func(context.Context) string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tc, ctx, err := tenancy.OpenPool(ctx, db, getTenantIDFromCtx(ctx))
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
			// Close the connection after completing http handling
			defer func() {
				err = tenancy.ClosePool(ctx, tc)
				if err != nil {
					panic(err)
				}
			}()

			next.ServeHTTP(w, r.WithContext(ctx))

			err = tenancy.ClosePool(ctx, tc)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
		}
		return http.HandlerFunc(fn)
	}
}
