package middleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/common-fate/tenancy"
)

// Tenancy is a middleware that injects a tenanted *sql.Conn into the context of each
// request. This middleware accepts a function which shouldextract the current tenant from context.
// The expectation is that you will have some auth middleware which can inject the current tenant into the ctx in a previous step.
//
//
// The getTenantIDFromCtx function should return the tenantID from some value in ctx,
// this enables the tenant middleware to be independant of your method of authenticating users or tenants
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
			next.ServeHTTP(w, r.WithContext(ctx))
			err = tenancy.Close(ctx, tc)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
		}
		return http.HandlerFunc(fn)
	}
}
