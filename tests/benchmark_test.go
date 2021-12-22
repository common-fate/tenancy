// +build postgres

package tenancytests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/common-fate/tenancy"
)

type ContextExecutor interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func Query(ctx context.Context, db ContextExecutor) error {
	rows, err := db.QueryContext(ctx, "SELECT * FROM users")
	if err != nil {
		return err
	}
	var tenant1UserID, tenantID string
	hasNext := rows.Next()
	if hasNext != true {
		return err
	}
	err = rows.Scan(&tenant1UserID, &tenantID)
	if err != nil {
		return err
	}
	hasNext = rows.Next()
	if hasNext == true {
		return err
	}
	err = rows.Close()
	if err != nil {
		return err
	}
	return nil
}

func QueryWithWhere(ctx context.Context, db ContextExecutor) error {
	rows, err := db.QueryContext(ctx, "SELECT * FROM users where tenant_id=$1", TenantId1)
	if err != nil {
		return err
	}
	var tenant1UserID, tenantID string
	hasNext := rows.Next()
	if hasNext != true {
		return err
	}
	err = rows.Scan(&tenant1UserID, &tenantID)
	if err != nil {
		return err
	}
	hasNext = rows.Next()
	if hasNext == true {
		return err
	}
	err = rows.Close()
	if err != nil {
		return err
	}
	return nil
}

func BenchmarkTenancyWithSingleConnectionSelectAll(b *testing.B) {
	ctx := context.Background()
	rootDB, tenantDB, err := getDB()
	if err != nil {
		panic(err)
	}
	// Prepare the Database for the test
	err = rootDB.seedTestData(ctx, TenantId1, TenantId2)
	if err != nil {
		panic(err)
	}
	// Open a connection for tenant 1 Expect that there will be 1 user returned
	tc, ctx, err := tenancy.Open(ctx, tenantDB.DB.DB, TenantId1, tenancy.WithSingleConnection())
	if err != nil {
		b.Fatal(err)
	}
	b.Run("tenanted", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = Query(ctx, tc)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	err = tenancy.Close(ctx, tc)
	if err != nil {
		panic(err)
	}
}
func BenchmarkRawSelectAll(b *testing.B) {
	ctx := context.Background()
	rootDB, _, err := getDB()
	if err != nil {
		panic(err)
	}
	// Prepare the Database for the test
	err = rootDB.seedTestData(ctx, TenantId1, TenantId2)
	if err != nil {
		panic(err)
	}

	b.Run("raw", func(b *testing.B) {

		for i := 0; i < b.N; i++ {
			err = QueryWithWhere(ctx, rootDB)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
func BenchmarkTenancyWithPoolSelectAll(b *testing.B) {
	ctx := context.Background()
	rootDB, tenantDB, err := getDB()
	if err != nil {
		panic(err)
	}
	// Prepare the Database for the test
	err = rootDB.seedTestData(ctx, TenantId1, TenantId2)
	if err != nil {
		panic(err)
	}

	b.Run("tenanted", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Open a connection for tenant 1 Expect that there will be 1 user returned
			tc, ctx, err := tenancy.Open(ctx, tenantDB.DB.DB, TenantId1)
			if err != nil {
				b.Fatal(err)
			}
			err = Query(ctx, tc)
			if err != nil {
				b.Fatal(err)
			}
			err = tenancy.Close(ctx, tc)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

}
