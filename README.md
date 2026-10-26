# PreZero To-Do

A To-Do application built with Go and PostgreSQL. The current API serves static sample tasks; database-backed endpoints and the Nuxt frontend are not implemented yet.

## Current status

- Go HTTP API with health and sample task endpoints.
- PostgreSQL 17 running through Docker Compose with a persistent volume.
- Task schema verified manually, including inserts, defaults and nonblank title validation.
- pgx v5.11.0 registered through `database/sql`; startup verifies database connectivity with a five-second timeout.
- The `application` struct shares the connection pool with the task handler through a method receiver; task queries are not implemented yet.

## Requirements

- Go 1.27.3 or newer.
- Docker Engine with Docker Compose; Docker Desktop on macOS is supported.
- Available host ports: `8080` for the API and `5432` for PostgreSQL.

Run the following commands from the repository root unless stated otherwise.

## Local setup

### 1. Configure the database

On first setup, copy the environment template, set a local password and update the password in `DATABASE_URL` to match:

```bash
cp .env.example .env
```

Do not overwrite an existing `.env`. The file is ignored by Git.

| Variable | Purpose | Example |
| --- | --- | --- |
| `POSTGRES_USER` | Initial database user | `wehan` |
| `POSTGRES_PASSWORD` | Initial database password | Set locally |
| `POSTGRES_DB` | Initial database name | `wehan` |
| `DATABASE_URL` | Go API connection string | `postgres://wehan:<password>@localhost:5432/wehan?sslmode=disable` |

The commands below use `wehan` as both the user and database name. Adjust `-U` and `-d` if your configuration differs. The `POSTGRES_*` variables initialize a new data volume; changing `.env` does not change an existing database's credentials. `DATABASE_URL` must use the database's actual credentials. Quote the URL in `.env` as shown in `.env.example`; URL-encode reserved characters in its password component. `sslmode=disable` is intended for this local development connection.

### 2. Start PostgreSQL

Start Docker Desktop or Docker Engine, then run:

```bash
docker info
docker compose config -q
docker compose up -d db
docker compose exec db pg_isready -U wehan -d wehan
```

Wait for `accepting connections`. On first startup, initialization may take a few seconds; retry the readiness check if it initially reports `no response`.

`pg_isready` checks server readiness. Use a SQL query to verify the configured user and database:

```bash
docker compose exec db psql -U wehan -d wehan -v ON_ERROR_STOP=1 -c "SELECT current_user, current_database();"
```

### 3. Create the task table

For a new database, apply the schema once:

```bash
docker compose exec -T db psql -U wehan -d wehan -v ON_ERROR_STOP=1 < backend/schema.sql
```

Expected result: `CREATE TABLE`. If the table already exists, inspect it instead of recreating or deleting it:

```bash
docker compose exec db psql -U wehan -d wehan -c "\d tasks"
```

The schema defines an automatically generated `id`, a required title of up to 200 characters, and a `completed` flag defaulting to `false`. Empty titles and titles containing only ordinary spaces are rejected.

### 4. Run the API

In a separate terminal, from the repository root:

```bash
go -C backend mod download
set -a
source .env
set +a
go -C backend run ./cmd/api
```

Compose loads `.env` automatically; the Go process does not. The shell commands above export the variables from your local configuration before starting Go. Run them from the same terminal, using a `.env` file you control.

A successful database check logs `Database connection established`. The API then starts listening on port `8080`; confirm HTTP readiness with the requests below. If the connection string is missing or the database check fails, startup exits with an error. Restart the process after changing Go code.

For an existing setup, keep the actual working connection string in your ignored `.env`. A temporary terminal assignment is not saved to that file; sourcing an outdated `.env` will overwrite the terminal value.

## API

| Method | Path | Current behavior |
| --- | --- | --- |
| `GET` | `/healthz` | Returns `{"status":"ok"}`; does not check the database. |
| `GET` | `/api/tasks` | Returns a static JSON task list. |
| `GET` | `/wehan` | Learning endpoint returning a fixed status object. |

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/api/tasks
```

Both return `200 OK`. The current task response is:

```json
[{"id":0,"title":"hackthon","completed":true},{"id":0,"title":"sleep","completed":false}]
```

Database rows do not affect this response yet. The sample tasks have zero-valued IDs because they are not populated from database rows.

## Verification

### Go checks

```bash
gofmt -l backend/cmd/api
go -C backend test ./...
go -C backend vet ./...
```

`gofmt -l` should print no filenames. There are currently no automated test files; `go test` verifies package compilation, while `go vet` performs static checks.

### Database smoke checks

Insert and read a sample task. Each execution of the insert adds a row:

```bash
docker compose exec db psql -U wehan -d wehan -v ON_ERROR_STOP=1 -c "INSERT INTO tasks (title) VALUES ('Learn PostgreSQL') RETURNING id, title, completed;"
docker compose exec db psql -U wehan -d wehan -c "SELECT id, title, completed FROM tasks ORDER BY id;"
```

The database generates the ID and defaults `completed` to `false` (`f` in `psql`).

To check title validation, record the row count, attempt an invalid insert, and check the count again:

```bash
docker compose exec db psql -U wehan -d wehan -c "SELECT COUNT(*) FROM tasks;"
docker compose exec db psql -U wehan -d wehan -v ON_ERROR_STOP=1 -c "INSERT INTO tasks (title) VALUES ('   ');"
docker compose exec db psql -U wehan -d wehan -c "SELECT COUNT(*) FROM tasks;"
```

The insert is expected to fail with `violates check constraint`; the row count should remain unchanged. Rejected inserts can consume identity values, so gaps in IDs are normal.

Manual verification on 2026-10-26 confirmed the schema, valid inserts and SQL reads, rejection of a space-only title without increasing the row count, and successful Go database connection and HTTP server startup. Task retrieval still uses static data, so database-backed task API behavior has not been verified.

## Project layout

```text
.
├── .env.example
├── .gitignore
├── README.md
├── compose.yaml
└── backend/
    ├── go.mod
    ├── go.sum
    ├── schema.sql
    └── cmd/api/
        ├── database.go
        └── main.go
```

The Go module is in `backend`. Compose maps host port `5432` to container port `5432` and mounts the `pgdata` volume at `/var/lib/postgresql/data`.

## Stopping services

Stop the Go process with `Control+C`. To stop PostgreSQL while preserving its container and data:

```bash
docker compose stop db
```

To remove Compose containers and the network while retaining the named volume:

```bash
docker compose down
```

Adding `-v` to `docker compose down` deletes the database volume and its data.

## Troubleshooting

| Symptom | Action |
| --- | --- |
| Cannot connect to `docker.sock` | Start Docker Desktop or Docker Engine and retry `docker info`. |
| Initial readiness check reports `no response` | Wait for initialization; inspect `docker compose logs --tail=50 db` if it persists. |
| `role "todo" does not exist` | Use the configured database user; this repository's examples use `-U wehan -d wehan`. |
| `relation "tasks" already exists` | Inspect the table with `\d tasks`; do not rerun the initial schema against an existing table. |
| `cannot find main module` | Use `go -C backend ...` from the repository root. |
| `DATABASE_URL is required` | Check the exact variable name, set it in `.env`, and export the file's variables in the terminal used to start Go. |
| `password authentication failed` | Replace any placeholder password in `DATABASE_URL` with the actual database password, encode reserved URL characters, and reload the environment. |
| HTTP `404` or `405` after editing routes | Verify the path and uppercase HTTP method, then restart the Go process. |
| Port already in use | Stop the previous process or select an unused host port. |

## Next milestone

Use `app.db.QueryContext` to read task rows, scan them into `Task` values, and replace the static response. Verify that `/api/tasks` returns the same records as the SQL query. Task creation, updates, deletion and the Nuxt frontend remain pending.

## References

- [PostgreSQL Docker image and Compose example](https://github.com/docker-library/docs/blob/master/postgres/README.md)
- [Docker Compose port mappings](https://docs.docker.com/reference/compose-file/services/#ports)
- [PostgreSQL 17 constraints](https://www.postgresql.org/docs/17/ddl-constraints.html)
- [pgx PostgreSQL driver](https://github.com/jackc/pgx)
