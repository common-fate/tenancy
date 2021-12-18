// +build postgres

package tenancytests

import (
	"context"
	"testing"

	"github.com/common-fate/tenancy"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// tests that the tenanted connection is long lived and is not closed automatically by the underlying sql connection pool
// when the rootDb.SetConnMaxIdleTime() or rootDb.SetConnMaxLifetime() is set to a shorter duration that the connection is in use
func TestConfigurationOptions(t *testing.T) {
	ctx := context.Background()
	_, tenanted, err := getDB()

	assert.NoError(t, err)

	// Test that the configuration flag leads to a single connection being used for each request
	tc, ctx, err := tenancy.Open(ctx, tenanted.DB.DB, uuid.NewString(), tenancy.WithSingleConnection())
	conn, err := tc.Conn(ctx)
	assert.NoError(t, err)
	assert.NoError(t, conn.PingContext(ctx))

	conn2, err := tc.Conn(ctx)
	assert.NoError(t, err)

	assert.NoError(t, conn2.PingContext(ctx))

	assert.Equal(t, conn, conn2)
	assert.NoError(t, tenancy.Close(ctx, tc))

	// test that new connections are opened without by default
	tc, ctx, err = tenancy.Open(ctx, tenanted.DB.DB, uuid.NewString())
	conn, err = tc.Conn(ctx)
	assert.NoError(t, err)
	assert.NoError(t, conn.PingContext(ctx))

	conn2, err = tc.Conn(ctx)
	assert.NoError(t, err)

	assert.NoError(t, conn2.PingContext(ctx))

	assert.NotEqual(t, conn, conn2)
	assert.NoError(t, tenancy.Close(ctx, tc))
}
