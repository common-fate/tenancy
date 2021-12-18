# tenancy

A Go library for multitenancy in Postgres using Row Level Security (RLS).

# Usage

Tenancy as a connection pool.

By default, tenancy.Open() begins a connection pool which will open a new database connection for each database operation.
The purpose of this pool is to support concurrent API layers such as GraphQL where a single request may be served by concurrent database operations.

For uses that do not require concurrent database connections per request, a single connection can be opened with either BeginTx() or Conn()
When the request has finished being served, tenancy.Close() should be called which will close any open connections.

## Tenancy as middleware

auth middleware -> tenancy.Open() -> request handlers -> tenancy.Close()

## tenancy for background tasks

for each tenant -> tenancy.Open() -> process requiring tenanted database connection, you will likely use Conn() to get a single connection for the whole process -> tenancy.Close()

## Interfaces

The tenancy package provides some interfaces which you can use in your applications where you want to enforce a tenanted database connection.
The TContextExecutor interface satisfies the requirements of ORM libraries such as SQLBoiler which was the basis for creating this package.

## Acknowledgements

`tenancy` was created by [Joshua Wilkes](https://github.com/JoshuaWilkes) and is maintained by [Common Fate](commonfate.io).
