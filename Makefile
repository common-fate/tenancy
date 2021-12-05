
migrate:
	migrate -source file://./tests/migrations -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up

migrate-down:
	migrate -source file://./tests/migrations -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" down

test-postgres:
	go test ./tests/...  -tags=postgres

