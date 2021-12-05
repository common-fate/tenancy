// +build postgres

package tenancytests

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type rootDbUser struct{ *sqlx.DB }
type tenantedDbUser struct{ *sqlx.DB }

var rootDatabase *rootDbUser
var tenantedDatabase *tenantedDbUser
var dbName string

func init() {
	_, _, err := getDB()
	if err != nil {
		panic(err)
	}
}

// seedTestData is a util to wipe the test db and provision some test data.
func (db *rootDbUser) seedTestData(ctx context.Context, tenantId1 string, tenantId2 string) error {
	_, err := db.ExecContext(ctx, "DELETE FROM tenants")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "INSERT INTO tenants (id) VALUES ($1)", tenantId1)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO tenants (id) VALUES ($1)", tenantId2)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "INSERT INTO users (id, tenant_id) VALUES ($1, $2)", uuid.NewString(), tenantId1)
	if err != nil {
		return err
	}

	return nil
}

// getDB connects to a localhost database and runs migrations
func getDB() (*rootDbUser, *tenantedDbUser, error) {
	var err error
	if rootDatabase != nil && tenantedDatabase != nil {
		return rootDatabase, tenantedDatabase, nil
	}

	// Root user ignores RLS rules
	psqlString := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	rootDB, err := sqlx.Connect("postgres", psqlString)
	if err != nil {
		return nil, nil, err
	}

	// Tenant user must conform to RLS rules
	psqlString = "host=localhost port=5432 user=tenant password=postgres dbname=postgres sslmode=disable"
	tenantedDB, err := sqlx.Connect("postgres", psqlString)
	if err != nil {
		return nil, nil, err
	}

	rootDatabase = &rootDbUser{rootDB}
	tenantedDatabase = &tenantedDbUser{tenantedDB}

	driver, err := postgres.WithInstance(rootDatabase.DB.DB, &postgres.Config{})
	if err != nil {
		return nil, nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)

	if err != nil {
		return nil, nil, errors.Wrap(err, "error connecting to database while running migrations")
	}
	err = m.Up()
	if err != migrate.ErrNoChange {
		return nil, nil, errors.Wrap(err, "applying migrations")
	}

	return rootDatabase, tenantedDatabase, nil
}

// closeDB closes the connections
func closeDB() {
	if rootDatabase != nil {
		rootDatabase.Close()
	}
	if tenantedDatabase != nil {
		tenantedDatabase.Close()
	}
}
