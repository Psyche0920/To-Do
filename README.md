# PreZero To-Do

A basic full-stack task manager built to explore the Go, Nuxt and PostgreSQL stack for an internship exercise. Users can create tasks, view saved tasks, mark them complete or incomplete, edit titles and delete tasks. Records persist in PostgreSQL across page reloads and container restarts.

[Open the live demo](https://wehan-todo.vercel.app/) · [Deployment guide](docs/free-deployment.md)

The application keeps its scope small: one Nuxt page, a Go HTTP API and a PostgreSQL service managed by Docker Compose. The public demo uses Vercel Hobby, Render Free and Neon Free. It has no authentication: visitors share disposable demonstration data. The backend may take about a minute to wake after inactivity.

## Start here

| Document | Purpose |
| --- | --- |
| [Local setup and testing](docs/local-setup.md) | Start from a new Mac with VS Code: install tools, clone, configure, run and test. Includes repeat-start instructions and troubleshooting. |
| [Architecture and workflow](docs/architecture.md) | Technology stack, file responsibilities, ports, database design and request flows. |
| [Development and learning stages](docs/development-stages.md) | The major implementation phases, learning objectives and verification checkpoints. |
| [Free-tier deployment](docs/free-deployment.md) | Cloud configuration, automatic deployments, schema changes and teardown. |
| [Concept reference](docs/learning-notes.md) | Short explanations of the Go, Vue and database concepts discussed during implementation. |

**New machine:** follow the setup guide in order. VS Code alone does not install Git, Go, Node.js or Docker. The commands target macOS with zsh; use the official installation instructions for other operating systems rather than copying shell-specific commands unchanged.

## Stack

| Layer | Technology |
| --- | --- |
| Frontend | Nuxt 4.6.2, Vue 3.5, TypeScript |
| API | Go standard `net/http`, `database/sql`, pgx v5 |
| Database | PostgreSQL 17 Alpine, Docker Compose named volume |
| Local addresses | Nuxt `http://localhost:3000`, Go `http://localhost:8080`, PostgreSQL `localhost:5432` |

The frontend dependency lockfile fixes the installed dependency tree. See the setup guide for recorded tool versions and installation requirements.

## API

| Method | Path | Result |
| --- | --- | --- |
| GET | `/healthz` | `200` with a fixed status object; not a database health check. |
| GET | `/api/tasks` | `200` with tasks ordered by ID, or `[]` when empty. |
| POST | `/api/tasks` | `201` with the created task. Body: `{"title":"Learn Nuxt"}`. |
| PATCH | `/api/tasks/{id}` | `200` with the updated task. Accepts title and/or completed. |
| DELETE | `/api/tasks/{id}` | `204` without a body; `404` if absent. |
| GET | `/wehan` | Practice endpoint with a fixed status object. |

Titles are trimmed and must contain 1–200 Unicode code points. Invalid requests return `400`; absent task IDs return `404`; database failures return `500`. Mutation errors use plain text; successful task responses use JSON. Nuxt forwards `/api/**` to Go; health and practice routes remain directly accessible on port 8080.

## Verification and limits

Local browser checks confirmed creation, completion changes, deletion and title editing, including reload persistence and cancellation. Database-unavailability checks confirmed creation and list error messages and subsequent recovery. Go compilation/static checks and the Nuxt production build passed. These are manual and build checks; there are no automated test files in the repository.

A clean-machine installation has not been performed on a newly provisioned operating system. The setup guide distinguishes these evidence limits from instructions a new developer can follow. The public demo was manually checked on 28 October 2026: create, completion changes, title save/cancel, delete, reload persistence and blank-title rejection passed. Test data was removed. A full idle/cold-start cycle was not timed. Localhost links work only on the machine running the services.

## Repository layout

```text
.
├── README.md
├── .env.example
├── compose.yaml
├── backend/
│   ├── go.mod
│   ├── go.sum
│   ├── schema.sql
│   └── cmd/api/
│       ├── main.go
│       └── database.go
├── frontend/
│   ├── app/app.vue
│   ├── nuxt.config.ts
│   ├── package.json
│   └── package-lock.json
└── docs/
    ├── local-setup.md
    ├── architecture.md
    ├── development-stages.md
    └── learning-notes.md
```

Real credentials belong in the ignored `.env`, never in source control. Dependency directories, build outputs and local credentials are excluded from Git.
