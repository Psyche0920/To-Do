# Local setup and testing

For macOS with zsh and VS Code. Run commands from the repository root unless stated otherwise.

## 1. Install tools

| Tool | Install |
| --- | --- |
| Git | [Git downloads](https://git-scm.com/downloads/) or `xcode-select --install` |
| Go 1.27.3 or newer | [Go installation](https://go.dev/doc/install) |
| Node.js and npm | [Node.js downloads](https://nodejs.org/en/download); use a supported LTS release |
| Docker Desktop | [Mac installation](https://docs.docker.com/desktop/setup/install/mac-install/) |

Reopen the VS Code terminal after installation. PostgreSQL runs in Docker; no separate installation is needed.

```sh
git --version
go version
node --version
npm --version
docker compose version
```

## 2. Clone and configure

```sh
git clone https://github.com/Psyche0920/To-Do.git
cd To-Do
cp .env.example .env
```

Open the folder in VS Code. Edit `.env`, replacing both password placeholders with the same local password:

```dotenv
POSTGRES_USER=wehan
POSTGRES_PASSWORD=YOUR_LOCAL_PASSWORD
POSTGRES_DB=wehan
DATABASE_URL='postgres://wehan:YOUR_LOCAL_PASSWORD@localhost:5432/wehan?sslmode=disable'
```

Use letters and numbers for a simple URL-safe local password. Keep `.env` out of Git. Existing database volumes retain their original credentials; editing `.env` does not change them.

## 3. Start the database

```sh
open -a Docker
```

Wait for Docker Desktop's engine to start, then run:

```sh
docker info
docker compose config -q
docker compose up -d db
docker compose exec db pg_isready -U wehan -d wehan
```

Continue when readiness reports `accepting connections`. If initialization is still running, retry the readiness command.

On a **new database only**, create the table:

```sh
docker compose exec -T db psql -U wehan -d wehan -v ON_ERROR_STOP=1 < backend/schema.sql
```

Do not rerun this schema on later starts.

## 4. Start Go — terminal A

```sh
go -C backend mod download
set -a
source .env
set +a
go -C backend run ./cmd/api
```

Keep the terminal open. The API should connect to the database and listen on port 8080 without an error.

To enter the existing database password interactively instead of loading `.env`, use this zsh alternative:

```zsh
read -s "DB_PASSWORD?Database password: "
echo
export DATABASE_URL="postgres://wehan:${DB_PASSWORD}@localhost:5432/wehan?sslmode=disable"
unset DB_PASSWORD
go -C backend run ./cmd/api
```

Typing is hidden. This assumes the same user/database names and a URL-safe password. It sets the connection URL only in this terminal; do not load an outdated `.env` afterward.

## 5. Start Nuxt — terminal B

Open another terminal in the same repository:

```sh
cd frontend
npm ci
npm run dev
```

Open [localhost:3000](http://localhost:3000/), or the address Nuxt prints. Keep both service terminals open. `npm run build` builds the application; it does not start the development server.

## 6. Test

In another terminal:

```sh
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/api/tasks
curl -i http://localhost:3000/api/tasks
```

All should return 200. The task endpoints should return the same list, initially `[]` on an empty database.

Use a disposable task in the browser:

| Action | Expected |
| --- | --- |
| Add a task | Appears in the list; input clears |
| Check, reload; uncheck, reload | Both completion states persist |
| Edit, change title, cancel | Original title remains |
| Edit, change title, save, reload | New title persists |
| Submit a space-only title | Error; no invalid change |
| Delete, reload | Task stays absent |

### Inspect the database

```sh
docker compose exec db psql -U wehan -d wehan -c "SELECT id, title, completed FROM tasks ORDER BY id;"
```

For interactive inspection:

```sh
docker compose exec db psql -U wehan -d wehan
```

`tasks` is the database table that stores each task's ID, title and completion status.

- `\dt`: list the database tables.
- `\d tasks`: show the task table's columns and rules, not its saved records.
- `SELECT id, title, completed FROM tasks ORDER BY id;`: show the saved tasks.
- `\q`: exit the database terminal.

Completion values display as `t` (true) and `f` (false). Database files live in Docker's `pgdata` volume.

### Automated checks

From the repository root:

```sh
# List Go files that need formatting.
gofmt -l backend/cmd/api

# Run the tests defined in Go test files.
go -C backend test ./...

# Check Go code for common mistakes without running the server.
go -C backend vet ./...

# Build the frontend; this does not start the development server.
npm --prefix frontend run build
```


## Restart and stop

After a reboot, start Docker Desktop and the database, then repeat the Go and Nuxt startup steps. Run `npm ci` again when dependencies change. Keep your existing `.env` and database schema.

Stop Go and Nuxt with Control+C. Stop PostgreSQL without deleting data:

```sh
docker compose stop db
```

`docker compose down` preserves the named volume; adding `-v` deletes database data.

## Troubleshooting

| Problem | Action |
| --- | --- |
| Command not found | Install the tool and reopen the terminal |
| Docker daemon unavailable | Open Docker Desktop and wait for its engine |
| Connection refused on 5432 | Start the database; check `docker compose ps` and `docker compose logs --tail=50 db` |
| Password authentication failed | Use the existing database password; changing `.env` does not reset a volume's credentials |
| DATABASE_URL is required | Load `.env` in the terminal starting Go |
| Tasks table already exists | Skip schema creation |
| Tasks table does not exist | Apply the schema to the database used by Go |
| Port already in use | Stop the earlier service or use Nuxt's printed address |
| Page cannot load tasks | Check Go's terminal and the direct API request before the proxy |
| Go changes are not visible | Save, stop and restart Go |

Setup was verified on macOS using an isolated source copy, a fresh dependency installation and a new database. Checks covered service startup, the rendered page, API forwarding, task operations and the production build.
