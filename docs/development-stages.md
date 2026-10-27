# Development and Learning Stages

This document records the major implementation sequence and what each phase was intended to teach. It groups the work into six phases rather than reproducing every question or terminal attempt. The application was built incrementally, with manual commands and browser checks between stages.

## Phase 1 — Go API foundation

**Work:** Create the backend module, define the task structure and JSON tags, register health/practice/list routes, configure the server and test with curl. The first task list used in-memory sample values.

**Concepts:** Go packages and entry point, structure types and pointers, handler parameters, method/path routing, status/header/body separation and server logs versus HTTP responses.

**Checkpoint:** The health endpoint returned 200 and JSON; routes were verified with the correct uppercase method and path. Static data established the HTTP contract before adding persistence.

## Phase 2 — PostgreSQL and persistent storage

**Work:** Start PostgreSQL with Compose, configure local credentials, attach pgdata, create the tasks table and insert/read sample rows with psql.

**Concepts:** Docker engine versus container, host-to-container port mapping, named volumes, database roles, identity IDs, default values and SQL constraints.

**Checkpoint:** SQL reads returned stored tasks; empty/space-only title attempts were rejected. Existing volumes retained data and credentials independently of later edits to environment templates.

## Phase 3 — Go database connection and CRUD

**Work:** Register pgx, prepare and verify the connection pool, place it in application, replace static list data with SQL rows, then implement POST, completion PATCH and DELETE.

**Concepts:** DSN versus driver versus pool, context deadlines, cleanup, method receivers, SQL placeholders, row iteration/Scan, RETURNING, RowsAffected and error responses.

**Checkpoint:** HTTP list results matched SQL. Creation returned 201, completion updates returned 200 and deletion returned 204 then 404 on repetition. Invalid input was rejected before storage.

## Phase 4 — Nuxt initialization and real list data

**Work:** Initialize the official Nuxt starter, configure forwarding from port 3000 to Go on port 8080, then replace the welcome screen with a list using useFetch.

**Concepts:** Script versus template, TypeScript interface versus runtime validation, data/error refs, page requests versus API requests and server-rendered initial fetching.

**Checkpoint:** Curl through the Nuxt port returned the Go task JSON. The page displayed actual database tasks and completion state. The list-only version was committed before adding controls.

## Phase 5 — Browser mutations and title editing

**Work:** Add creation form, completion checkbox and delete button; refresh the existing list after successful mutations. Then extend PATCH to accept a title, with frontend Edit, Save/Submit and Cancel controls.

**Concepts:** Ref/value and const bindings, busy flags, async/await sequencing, try/catch/finally, form submit versus change/click events, independent error messages, conditional/list rendering and editable drafts versus stored records.

**Checkpoint:** Browser actions persisted after reload. Cancel left the saved title unchanged. Temporary acceptance tasks were removed without changing original tasks. Editing reused the existing table and PATCH endpoint rather than adding another database structure.

## Phase 6 — Acceptance and handoff

**Work:** Verify mutation persistence, invalid titles and database-failure messages, then organize setup instructions, architecture and this development record.

**Observed evidence:** Local browser checks covered create, completion changes, delete and title editing. An unavailable-database check showed create/list error messages; restoring the database and reloading recovered the list. Go compilation/static checks and Nuxt builds passed.

**Remaining handoff boundary:** Fresh operating-system installation and public deployment are separate from those existing-machine checks. Do not describe them as verified until actually performed. No automated test files currently exist.

## Original milestone map

The learning plan counted 18 checkpoints. Percentages count checkpoints, not hours, difficulty or production readiness. Title editing was an additional feature rather than another percentage denominator.

| Steps | Group | Outcome |
| --- | --- | --- |
| 1–5 | API basics and JSON | Running Go skeleton, health/practice routes, task shape and sample list. |
| 6–7 | Local database | Compose database and schema/constraint verification. |
| 8–11 | Persistent CRUD | Connection/list, creation, completion update and deletion. |
| 12–13 | Frontend foundation | Nuxt starter/proxy and real task display. |
| 14–16 | Frontend mutations | Creation, completion and deletion connected to Go. |
| 17 | Acceptance | Main flows, persistence and error scenarios. |
| 18 | Delivery | Reproducible documentation and an accessible delivery destination; public app URL remains pending. |

## Verification discipline

Each phase separated code compilation from runtime proof. A successful build cannot prove SQL correctness or browser behavior. A successful local flow cannot prove a clean-machine install or public deployment. Keep the README and setup evidence aligned with what has actually run.
