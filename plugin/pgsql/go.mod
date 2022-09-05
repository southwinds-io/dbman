module southwinds.dev/dbman/plugins/pgsql

go 1.13

replace (
	southwinds.dev/dbman => ../../
	southwinds.dev/http => ../../../http
)

require (
	github.com/jackc/pgconn v1.10.1
	github.com/jackc/pgtype v1.9.1
	github.com/jackc/pgx/v4 v4.14.1
	southwinds.dev/dbman v0.0.0-00010101000000-000000000000
)
