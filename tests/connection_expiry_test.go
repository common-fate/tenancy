// +build postgres

package tenancytests

import (
	"context"
	"testing"
	"time"

	"github.com/common-fate/tenancy"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// tests that the tenanted connection is long lived and is not closed automatically by the underlying sql connection pool
// when the rootDb.SetConnMaxIdleTime() or rootDb.SetConnMaxLifetime() is set to a shorter duration that the connection is in use
func TestConnectionExpiry(t *testing.T) {
	ctx := context.Background()
	_, tenanted, err := getDB()

	assert.NoError(t, err)

	tenanted.SetConnMaxIdleTime(time.Second)
	tenanted.SetConnMaxLifetime(time.Second)

	// We use a random uuid here because we are just testing that PingContext is successfull
	tc, ctx, err := tenancy.Open(ctx, tenanted.DB.DB, uuid.NewString())

	assert.NoError(t, tc.PingContext(ctx))

	time.Sleep(time.Second * 2)

	assert.NoError(t, tc.PingContext(ctx))
	assert.NoError(t, tenancy.Close(ctx, tc))
}
