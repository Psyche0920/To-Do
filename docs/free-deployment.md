# Deployment

[Live application](https://wehan-todo.vercel.app/) · [API health](https://wehan-todo-api.onrender.com/healthz)

The demo uses Vercel Hobby, Render Free and Neon Free. It has no authentication; visitors share demonstration data. The free API can sleep when idle and take about a minute to wake.

## 1. Database: Neon

1. Create a Free PostgreSQL 17 project, near the API's region.
2. Run `backend/schema.sql` once in the SQL Editor against the empty database.
3. Copy the connection URL, preserving its SSL settings. Store it as Render's secret `DATABASE_URL`, never in Git or the frontend.

## 2. API: Render

Connect the repository and use `render.yaml`, or configure:

| Setting | Value |
| --- | --- |
| Plan | Free |
| Runtime | Docker |
| Root directory | `backend` |
| Dockerfile | `Dockerfile` |
| Region | Frankfurt |
| Environment | `DATABASE_URL` from Neon |
| Health check | `/healthz` |

The API uses Render's `PORT`. After deployment, check both `/healthz` and `/api/tasks`: health confirms the server responds; the task endpoint also checks database access.

## 3. Frontend: Vercel

Import the repository, select the Hobby plan, set root directory to `frontend` and framework to Nuxt. Let Vercel detect build settings.

Set `API_PROXY_TARGET` to the Render HTTPS origin, without a trailing slash or `/api` suffix. Redeploy after changing it: the proxy target is read at build time.

## Updates

With automatic deployment enabled, pushes to `main` trigger application builds. Check deployment status before assuming changes are live. Database records remain in Neon.

Schema changes are separate: back up required data, write and test a numbered SQL migration, then run it once in the correct Neon database. Deploy compatible code in the appropriate order. Do not rerun the initial schema or drop the tasks table to update it.

## Verify

Create a disposable task. Check completion in both directions, edit/save, edit/cancel and deletion; reload after each saved change. Confirm blank titles are rejected, then remove the test task. Also check first-load behavior after the backend has been idle.

## Remove the demo

Export required data, then delete this project's Vercel deployment, Render service and Neon project. Neon project deletion permanently removes the database. Cleanup is manual.

## References

- [Render free services](https://render.com/docs/free)
- [Render Docker deployments](https://render.com/docs/docker)
- [Neon plans](https://neon.com/pricing)
- [Nuxt on Vercel](https://vercel.com/docs/frameworks/full-stack/nuxt)
