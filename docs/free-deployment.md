# Free-tier deployment

Status: live public deployment verified on 28 October 2026.

- Application: https://wehan-todo.vercel.app/
- API: https://wehan-todo-api.onrender.com/
- Database: Neon Free, PostgreSQL 17, Frankfurt, database `todo`.

Public browser acceptance passed for create, reload persistence, completion in both directions, edit/save, edit/cancel, delete/reload and whitespace-title rejection. Only disposable test data was used and removed. Anonymous HTTP checks returned 200 for the frontend, health endpoint and task list. A complete idle/cold-start cycle was not timed. No Azure resource was created.

## Services

- Vercel Hobby: Nuxt frontend, repository root directory `frontend`.
- Render Free: Go API, Docker runtime, root directory `backend`, Dockerfile `Dockerfile`, Frankfurt region when available.
- Neon Free: PostgreSQL, a European region near the API when available.

Use free plans only. Do not upgrade or attach payment details as part of this workflow. Quotas can suspend services. Render Free sleeps after 15 idle minutes and may take about one minute to wake. This is a demonstration without authentication; visitors can modify all demo tasks. Do not upload local or personal task data.

## 1. Neon

Log in or register, create a Free project and choose the PostgreSQL version supported by the project (17 matches local development). Execute `backend/schema.sql` once in its SQL Editor against the new empty database. Do not execute it against a database containing an existing tasks table. Obtain the connection URL from Connect and preserve its SSL settings. Store it only as Render's secret `DATABASE_URL`; never paste it into documentation, chat or Git.

## 2. Render

Connect only the To-Do repository, then create the service from `render.yaml`, or manually select Docker / root `backend` / Free. Supply the Neon connection URL as `DATABASE_URL`. The API uses Render's `PORT` environment variable and defaults to 8080 locally. Verify `/healthz` and `/api/tasks` at the assigned HTTPS service address. Do not use Render's expiring Free PostgreSQL database.

## 3. Vercel

Import the same repository with root directory `frontend` and framework Nuxt. Use the Hobby plan. Before the production build, set `API_PROXY_TARGET` to the Render HTTPS origin, without a trailing slash or `/api` suffix. This is a non-secret API origin; the database URL does not belong in Vercel. Nuxt's route rule is generated at build time: redeploy after changing the API origin. Let Vercel detect its Nuxt build/output settings.

## 4. Public acceptance

Open the assigned Vercel URL and use a disposable task to check creation, reload persistence, both completion states, title save/cancel, deletion and invalid-title errors. Remove only the test task. Recheck after the API has slept so the first-visit behavior is known. A successful build is not sufficient evidence of a public deployment.

## Cleanup after two weeks

Export any required data first. Delete the demo projects/services from Vercel, Render and Neon, checking each project's name before deletion. Neon project deletion removes its database permanently. Do not delete unrelated account projects. This guide does not schedule automatic deletion.

## Official references

- https://render.com/docs/free
- https://render.com/docs/docker
- https://neon.com/pricing
- https://vercel.com/docs/frameworks/full-stack/nuxt
- https://vercel.com/docs/environment-variables

## Updates and schema changes

With Git-based automatic deployment enabled, a push to `main` rebuilds the affected Go and Nuxt services. Failed builds do not mean a new version is live. PostgreSQL data remains in Neon, separately from the application deployment.

Changes to `backend/schema.sql` are not automatically applied to Neon. For a future schema change, write a numbered SQL migration, back up important data, test it against a disposable database, and execute it once through Neon's SQL Editor against the correct project/database. Deploy compatible application code in a safe order. Do not rerun the original `CREATE TABLE tasks` script against an existing database or drop the table to update it. No migration was needed for the current deployment.
