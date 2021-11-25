package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

var (
	host       = flag.String("host", "localhost", "postgres host")
	port       = flag.Int("port", 5432, "postgres port")
	user       = flag.String("user", "postgres", "postgres user")
	password   = flag.String("password", "postgres", "postgres password")
	database   = flag.String("database", "postgres", "postgres database")
	ssl        = flag.String("sslmode", "disable", "postgres sslmode")
	schema     = flag.String("schema", "public", "schema to scan for RLS")
	ignoredStr = flag.String("ignored-tables", "", "a comma-separated list of tables to ignore for RLS")
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	// for internal tables such as migrations tracking tables, we don't need RLS
	// this allows the user to specify tables to be ignored in scanning
	ignored := strings.Split(strings.Trim(*ignoredStr, " "), ",")

	connStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		*host, *port, *user, *password, *database, *ssl)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	rows, err := db.Query("select * from pg_tables where schemaname=$1", schema)
	if err != nil {
		return err
	}

	type row struct {
		SchemaName  string
		TableName   string
		TableOwner  string
		TableSpace  *string
		HasIndexes  bool
		HasRules    bool
		HasTriggers bool
		RowSecurity bool
	}

	insecureTables := []string{}

	for rows.Next() {
		var r row
		err = rows.Scan(&r.SchemaName, &r.TableName, &r.TableOwner, &r.TableSpace, &r.HasIndexes, &r.HasRules, &r.HasTriggers, &r.RowSecurity)

		if err != nil {
			return err
		}

		if !r.RowSecurity {
			var isIgnored bool
			// RLS isn't enabled for this table, but before warning the user about it let's check whether they have opted to ignore the table from our analysis.
			for _, i := range ignored {
				if r.TableName == i {
					isIgnored = true
					break
				}
			}
			if !isIgnored {
				insecureTables = append(insecureTables, r.TableName)
			}
		}
	}

	if len(insecureTables) > 0 {
		return fmt.Errorf("found the following tables that don't implement Row Level Security: %s", strings.Join(insecureTables, ", "))
	}

	fmt.Println("all tables are implementing Row Level Security")

	return nil
}
