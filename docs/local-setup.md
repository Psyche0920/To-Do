# Local Setup and Testing

This guide targets a new **macOS machine using zsh and VS Code**. Commands are shell commands unless labelled SQL. Copy commands only, without terminal prompts or example output. Execute each numbered stage and confirm its checkpoint before continuing.

## 1. Install tools

Install the tools below using their official installers. Choose Apple silicon/ARM64 for an Apple silicon Mac and Intel/x64 for an Intel Mac. Reopen the VS Code terminal after installation so PATH changes are loaded.

| Tool | Installation | Required for |
| --- | --- | --- |
| Git | [Official downloads](https://git-scm.com/downloads/) | Cloning the repository. On macOS, `xcode-select --install` can install Apple's command-line tools, including Git. Complete the installer before continuing. |
| Go | [Official installation](https://go.dev/doc/install) | Go 1.27.3 or newer, matching the module's `go` directive. If that version is unavailable in your distribution, resolve the toolchain requirement before proceeding rather than silently downgrading the module. |
| Node.js and npm | [Official downloads](https://nodejs.org/en/download) | Install a supported current LTS release with npm. The existing Mac was verified with Node 26.10.0 and npm 11.19.1; other versions still need the build check below. |
| Docker Desktop | [Official Mac installation](https://docs.docker.com/desktop/setup/install/mac-install/) | Runs PostgreSQL and includes Docker Compose. Installing the application alone does not start its engine. |

Go and Vue editor extensions are optional; they do not replace the runtimes. PostgreSQL and psql do **not** need a separate host installation: both run in the database container. Homebrew is optional, not a prerequisite for these instructions.

Open **Terminal → New Terminal** in VS Code and check:

```bash
git --version
go version
node --version
npm --version
docker --version
docker compose version
```

Checkpoint: every command prints a version, not `command not found`. Reserve ports 3000, 8080 and 5432. Avoid starting multiple copies of the same service.

## 2. Clone and open the repository

In a terminal, choose a parent directory where you keep projects:

```bash
git clone https://github.com/Psyche0920/To-Do.git
cd To-Do
pwd
```

In VS Code, use **File → Open Folder** and select this `To-Do` directory. All subsequent root-relative commands assume the terminal is in this directory. In each new terminal, navigate back to the same checkout. Do not copy another person's absolute home directory.

If Git reports authentication or repository-access errors, resolve access first. If `To-Do` already exists, open the existing checkout rather than cloning into it.

## 3. Configure local credentials

For a fresh clone only:

```bash
cp .env.example .env
```

Open `.env` in VS Code. Choose a local development password and replace **both** password placeholders with the same value:

```dotenv
POSTGRES_USER=wehan
POSTGRES_PASSWORD=YOUR_LOCAL_PASSWORD
POSTGRES_DB=wehan
DATABASE_URL='postgres://wehan:YOUR_LOCAL_PASSWORD@localhost:5432/wehan?sslmode=disable'
```

`YOUR_LOCAL_PASSWORD` is a placeholder, not a working password. For the simplest first setup, choose a password containing only letters and numbers. Other characters may require shell quoting and percent-encoding in the URL password component. Save the file before continuing. Do not commit it.

The user and database names in this guide are `wehan`. If you change them, adjust both `.env` and every `psql -U ... -d ...` command consistently. Disabling TLS here is for a local development connection only.

**Existing database:** PostgreSQL initialization variables only apply when the data volume is first created. Changing `.env` does not change credentials inside an existing volume. Use the actual existing credentials; never delete a volume just to resolve a password mismatch.

## 4. Start Docker, then PostgreSQL

Open Docker Desktop from Applications, or on macOS run:

```bash
open -a Docker
```

Wait until Docker Desktop says its engine is running. Then, in the repository root:

```bash
docker info
docker compose config -q
docker compose up -d db
docker compose ps db
docker compose exec db pg_isready -U wehan -d wehan
```

Checkpoints:

- `docker info` connects to the daemon. If it does not, stop here and start Docker Desktop.
- `docker compose config -q` exits successfully without printing credentials.
- The database container is running and readiness reports `accepting connections`.
- On first startup, `no response` can mean initialization is still in progress. Retry the readiness command; if it persists, run `docker compose logs --tail=50 db`.

Readiness is not a password test. Check the actual database/user with SQL:

```bash
docker compose exec db psql -U wehan -d wehan -v ON_ERROR_STOP=1 -c "SELECT current_user, current_database();"
```

Container-local psql can authenticate differently from Go's TCP connection. The Go startup check in stage 6 verifies the configured connection URL.

## 5. Create the schema once

Only on a **new database**:

```bash
docker compose exec -T db psql -U wehan -d wehan -v ON_ERROR_STOP=1 < backend/schema.sql
```

Expected: `CREATE TABLE`. Inspect it:

```bash
docker compose exec db psql -U wehan -d wehan -c "\d tasks"
```

The table has generated bigint `id`, required `title` up to 200 characters, and boolean `completed` defaulting to false. The check rejects empty or ordinary-space-only titles. The API additionally trims surrounding whitespace.

If `relation tasks already exists` appears, do not rerun or drop the table. Inspect it and proceed when the expected schema is present. The schema file is not a repeatable migration script.

## 6. Terminal A: start Go

From the repository root:

```bash
go -C backend mod download
set -a
source .env
set +a
go -C backend run ./cmd/api
```

Expected startup messages include `Database connection established` and `API server listening on :8080`. Keep this terminal running. The listening message announces the attempt: a subsequent bind error still means startup failed. Go does not automatically load `.env` or reload saved source changes.

If the URL password is outdated, edit and save `.env`, then reload it. Alternatively, use this **zsh-only** password prompt from the root, assuming the user/database names above and a URL-safe password:

```zsh
read -s "DB_PASSWORD?Database password: "
echo
export DATABASE_URL="postgres://wehan:${DB_PASSWORD}@localhost:5432/wehan?sslmode=disable"
unset DB_PASSWORD
go -C backend run ./cmd/api
```

Enter the existing database password and press Enter; typing is hidden. This sets the URL in the current terminal only. It does not save `.env` or change the database password. Do not source an outdated `.env` afterward. New terminals need their own environment loading.

## 7. Terminal B: start Nuxt

Open a **second VS Code terminal**, navigate to the same checkout, then:

```bash
cd frontend
npm ci
npm run dev
```

`npm ci` uses the committed lockfile. Run it on first setup and after dependency-lock changes; do not initialize another Nuxt app in this clone.

Keep this terminal running. Open the URL printed by Nuxt, normally `http://localhost:3000/`. If a different port is printed, use it consistently in subsequent checks. Nuxt's configuration forwards `/api/**` to the Go service on port 8080. Do not run only `npm run build` and expect a running development server: build verifies/produces output, dev starts the server.

## 8. Terminal C: verify the running system

Open a third terminal in the repository root. These read-only requests should return `200`:

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/api/tasks
curl -i http://localhost:3000/api/tasks
```

On a fresh database, both task endpoints should return `[]`. Existing checkouts may have rows. The two task responses should agree. The Nuxt response can include additional connection/transfer headers; those are not task fields.

### Browser acceptance

Use only a task created specifically for this test. Deletion removes its database row permanently.

| Test | Action | Expected |
| --- | --- | --- |
| Create | Enter `Local smoke test` and select Add task. | Task appears automatically; input clears. |
| Persistence | Reload the browser. | Same task remains. |
| Completion | Check the task, reload; uncheck, reload. | Completed and incomplete states both persist. |
| Cancel edit | Select Edit, change the draft, select Cancel. | Saved title stays unchanged. |
| Save edit | Select Edit, enter `Edited smoke test`, submit. | New title appears; remains after reload. The submit button may read Save or Submit. |
| Invalid title | Try an ordinary-space-only new title or edit. | Error message; no invalid database change. |
| Delete | Delete this disposable task, then reload. | Task stays absent. |

### API mutation check (alternative to browser actions)

Each POST creates a row. Copy its generated ID; replace `123` below with that actual ID. Do not paste the placeholders literally or operate on an existing personal task.

```bash
curl -i -X POST http://localhost:3000/api/tasks -H 'Content-Type: application/json' -d '{"title":"API smoke test"}'
curl -i -X PATCH http://localhost:3000/api/tasks/123 -H 'Content-Type: application/json' -d '{"completed":true}'
curl -i -X PATCH http://localhost:3000/api/tasks/123 -H 'Content-Type: application/json' -d '{"title":"Edited API smoke test"}'
curl -i -X DELETE http://localhost:3000/api/tasks/123
curl -i -X DELETE http://localhost:3000/api/tasks/123
```

Expected statuses in order: 201, 200, 200, 204, 404. A 204 response has no body.

```bash
curl -i -X POST http://localhost:3000/api/tasks -H 'Content-Type: application/json' -d '{"title":"   "}'
curl -i -X POST http://localhost:3000/api/tasks -H 'Content-Type: application/json' -d '{"title":"Task","extra":true}'
curl -i -X POST http://localhost:3000/api/tasks -H 'Content-Type: application/json' -d '{"title":"Task"} {"title":"Other"}'
```

Each should return 400 without creating a task.

### Inspect stored records

```bash
docker compose exec db psql -U wehan -d wehan -c "SELECT id, title, completed FROM tasks ORDER BY id;"
```

In psql, `t` means true, `f` means false. IDs can have gaps after rejected inserts or deleted rows. For interactive inspection:

```bash
docker compose exec db psql -U wehan -d wehan
```

Use `\dt` to list tables, `\d tasks` for schema, SQL SELECT to read records, and `\q` to exit. Files are stored in Docker's named volume, not the source folder. Docker Desktop shows the container and volume under Containers and Volumes; names commonly include `to-do-db-1` and `to-do_pgdata`, depending on the Compose project name. Never edit raw database volume files.

## 9. Build and static checks

From the repository root, with no need to stop running services:

```bash
gofmt -l backend/cmd/api
go -C backend test ./...
go -C backend vet ./...
npm --prefix frontend run build
```

Expected: no filenames from gofmt, successful Go commands and a successful Nuxt build. There are currently no automated Go test files; `go test` compiles packages rather than establishing HTTP behavior. Building Nuxt is not the same as a dedicated TypeScript typecheck. Browser/API checks above are separate acceptance evidence.

## 10. Later starts and stopping

After a reboot, open Docker Desktop and wait for its engine. In the repository root:

```bash
docker compose up -d db
docker compose exec db pg_isready -U wehan -d wehan
```

After readiness, repeat stage 6 in Terminal A and `cd frontend` / `npm run dev` in Terminal B. Do **not** copy the environment template over your working `.env`, recreate the schema, or initialize another frontend.

Stop Go and Nuxt with Control+C in their respective terminals. From the repository root, stop the database while preserving data:

```bash
docker compose stop db
```

`docker compose down` removes containers/network but retains the named volume. **Adding `-v` deletes the database data.** It is not part of this guide's normal workflow.

## Troubleshooting

| Symptom | Check / action |
| --- | --- |
| command not found | Install the corresponding tool, reopen the terminal, rerun the version check. |
| Cannot connect to Docker daemon | Open Docker Desktop; wait for engine readiness, then retry docker info. |
| connection refused on 5432 | Start db; inspect compose ps and logs. Confirm port mapping and readiness. |
| DATABASE_URL is required | Load/export `.env` in the same terminal where Go starts. |
| password authentication failed | Confirm the URL is not a placeholder and matches the existing database password. Changing initialization variables does not update existing volumes. |
| role/database does not exist | Use the actual initialized user/database names. Do not assume readiness proves authentication. |
| port 8080 already in use | Stop the old Go service in its terminal before starting another. Do not kill unknown processes blindly. |
| Nuxt chooses another port | Close the earlier dev server or use the printed port for the page and curl checks. |
| tasks already exists | Schema was applied earlier; inspect it rather than recreating it. |
| relation tasks does not exist | Apply schema once to the database referenced by DATABASE_URL. |
| cannot find main module | Run Go with `-C backend` from the repository root, or enter backend first. |
| zsh tries to run HTTP/1.1 or Date | You pasted output as commands. Copy command blocks only. |
| Page shows load failure | Check Go terminal, db readiness and direct 8080 request before testing the proxy. |
| Changes not reflected in Go | Save source, stop the old process and restart Go. |

## Evidence boundary

This guide was checked against source, configuration, tracked dependencies, local build results and browser/API acceptance on the existing development Mac. It is not a claim of installation testing on a brand-new Mac, Windows machine or Linux VM. An isolated tracked-file copy was also tested without copying `.env`, node_modules or build output: fresh `npm ci`, Nuxt build, Go compilation and static checks passed on 2026-10-27. This still uses the existing host toolchain and is not a fresh-OS or empty-database end-to-end test. A genuinely clean-machine run is a separate verification step.
