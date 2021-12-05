// +build postgres

package tenancytests

import (
	"context"
	"testing"

	"github.com/common-fate/tenancy"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

//TestTenantIsolationWorks ensures that RLS rules are correctly applied to tenanted connections
func TestTenantIsolationWorks(t *testing.T) {
	ctx := context.Background()
	rootDb, tenanted, err := getDB()
	assert.NoError(t, err)
	tenantId1 := uuid.NewString()
	tenantId2 := uuid.NewString()
	// Prepare the Database for the test
	assert.NoError(t, rootDb.seedTestData(ctx, tenantId1, tenantId2))

	// Open a connection for tenant 1 Expect that there will be 1 user returned
	tc, ctx, err := tenancy.Open(ctx, tenanted.DB.DB, tenantId1)
	rows, err := tc.QueryContext(ctx, "SELECT * FROM users")
	var tenant1UserID, tenantID string
	assert.True(t, rows.Next())
	assert.NoError(t, rows.Scan(&tenant1UserID, &tenantID))
	assert.False(t, rows.Next())
	assert.NoError(t, rows.Close())
	assert.Equal(t, tenantId1, tenantID)
	assert.NoError(t, tenancy.Close(ctx, tc))

	// Open a connection for tenant 2 Expect that there will be no users returned
	tc, ctx, err = tenancy.Open(ctx, tenanted.DB.DB, tenantId2)
	rows, err = tc.QueryContext(ctx, "SELECT * FROM users")
	assert.False(t, rows.Next())
	assert.NoError(t, rows.Close())

	// Just to be sure, quere by id also returns nothing
	rows, err = tc.QueryContext(ctx, "SELECT * FROM users WHERE id=$1", tenant1UserID)
	assert.False(t, rows.Next())
	assert.NoError(t, rows.Close())
	assert.NoError(t, tenancy.Close(ctx, tc))

}
