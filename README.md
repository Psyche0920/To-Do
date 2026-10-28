# To-Do

A full-stack task manager built with Nuxt, Go and PostgreSQL for an internship exercise. Create tasks, edit titles, toggle completion and delete tasks. Data persists in PostgreSQL.

[Live demo](https://wehan-todo.vercel.app/)

The demo uses shared data without authentication. The free backend may take about a minute to wake after inactivity.

## Getting started

Follow the [local setup and testing guide](docs/local-setup.md) to install dependencies, configure the database and start both services. Instructions target macOS with zsh.

- [Architecture and request flow](docs/architecture.md)
- [Deployment and database migrations](docs/free-deployment.md)

## Stack

- **Frontend:** Nuxt 4, Vue 3, TypeScript
- **Backend:** Go, `net/http`, `database/sql`, pgx
- **Database:** PostgreSQL 17; Docker Compose for local development
- **Hosting:** Vercel, Render, Neon

## API

| Method | Path | Action |
| --- | --- | --- |
| GET | `/api/tasks` | List tasks |
| POST | `/api/tasks` | Create a task |
| PATCH | `/api/tasks/{id}` | Update title and/or completion |
| DELETE | `/api/tasks/{id}` | Delete a task |
| GET | `/healthz` | Check API availability, without querying the database |

Titles are trimmed and must contain 1–200 Unicode code points. Nuxt proxies `/api/**` to Go.

## Checks

```sh
go -C backend test ./...
go -C backend vet ./...
cd frontend
npm ci
npm run build
```

Automated tests cover health and invalid requests. Task creation, listing, updates and deletion were checked manually, along with editing cancellation and persistence after reloading. The frontend build does not include a dedicated TypeScript type check. Local setup was also verified with a clean project copy and a new database on macOS.

## Structure

```text
backend/       Go API, tests, database schema and Dockerfile
frontend/      Nuxt application
docs/         Setup, architecture and deployment guides
compose.yaml   Local PostgreSQL
render.yaml    Backend deployment configuration
```

Keep credentials in the ignored `.env`; use `.env.example` as a template.
