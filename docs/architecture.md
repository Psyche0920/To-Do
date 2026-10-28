# Architecture

Nuxt renders the page and proxies API requests to Go. Go validates requests and reads or writes PostgreSQL.

```text
Browser → Nuxt → Go API → PostgreSQL
```

| Component | Local | Public demo |
| --- | --- | --- |
| Nuxt | localhost:3000 | Vercel Hobby |
| Go | localhost:8080 | Render Free |
| PostgreSQL 17 | Docker, localhost:5432 | Neon Free |

## Files

| File | Purpose |
| --- | --- |
| `frontend/app/app.vue` | Task display, forms, page state and responses to user actions |
| `frontend/nuxt.config.ts` | API proxy configuration |
| `backend/cmd/api/main.go` | Request routing, input checks, and functions to list, create, update and delete tasks |
| `backend/cmd/api/database.go` | Database pool and startup connection check |
| `backend/cmd/api/main_test.go` | Health and invalid-request tests |
| `backend/schema.sql` | Initial table definition |
| `compose.yaml` | Local database and persistent volume |
| `backend/Dockerfile`, `render.yaml` | Backend deployment |

## Data

A task has a generated `id`, a `title` and a boolean `completed` value. Titles are trimmed and limited to 1–200 Unicode code points. PostgreSQL stores the records; Docker's `pgdata` volume preserves local data across container recreation.

## Request flow

On page load, `useFetch` retrieves the task list. Nuxt can fetch during server rendering and pass the data to the browser.

| Action | Function | Request |
| --- | --- | --- |
| Create | `addTask` | POST `/api/tasks` |
| Toggle completion | `toggleTask` | PATCH `/api/tasks/{id}` |
| Start editing | `startEditingTask` | None; copy the ID and title into draft state |
| Cancel editing | `cancelEditingTask` | None; discard the draft |
| Save title | `editTaskTitle` | PATCH `/api/tasks/{id}` |
| Delete | `deleteTask` | DELETE `/api/tasks/{id}` |

After creating, updating or deleting a task, the frontend requests the latest list. While a request is running, state variables temporarily disable repeated or conflicting actions. Each operation has its own error message.

Go matches the HTTP method and path, validates input and executes parameterized SQL through a shared `database/sql` pool with pgx. PATCH preserves omitted fields. Database operations have a five-second timeout.

## Configuration

- `DATABASE_URL`: Go database connection string; keep it secret.
- `PORT`: Go listening port, default 8080.
- `API_PROXY_TARGET`: Go origin used when building Nuxt; defaults to the local API.
- Local PostgreSQL is published only on the host loopback interface. Containers on the Compose network use `db:5432`.

## Scope

The demo has no authentication, pagination, migration runner or graceful shutdown. Application deployments do not apply SQL schema changes. See the [deployment guide](free-deployment.md).

## Development stages

1. Set up the Go server and routes.
2. Add PostgreSQL and store, list, update and delete tasks.
3. Connect Nuxt and display saved tasks.
4. Add creation, completion, deletion and title editing.
5. Verify persistence and error handling, deploy, and add validation tests.
