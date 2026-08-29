# Agentic_Build

A small full-stack app for tracking agent build pipelines. It demonstrates a
complete local development experience: an Express JSON API and a modern
Vite + React + TypeScript UI, wired together for Cloud Agents.

## Stack

- **Server** (`server/`): Node.js + Express REST API with an in-memory store.
- **Web** (`web/`): Vite + React + TypeScript single-page app.
- **Tooling**: npm workspaces, ESLint (flat config), TypeScript, `node:test`.

## Prerequisites

- Node.js >= 20 (repo is developed against Node 22)
- npm >= 10

## Getting started

```bash
npm install        # install all workspace dependencies
npm run dev        # start API (:3001) and web dev server (:5173) together
```

Then open http://localhost:5173. The Vite dev server proxies `/api/*` to the
API on port 3001.

## Common commands

| Command | Description |
| --- | --- |
| `npm run dev` | Run the API and web dev server concurrently. |
| `npm run dev:server` | Run only the API (port 3001). |
| `npm run dev:web` | Run only the web dev server (port 5173). |
| `npm test` | Run the API test suite (`node:test`). |
| `npm run typecheck` | Type-check the web app (`tsc --noEmit`). |
| `npm run lint` | Lint the whole repo with ESLint. |
| `npm run build` | Type-check and build the web app to `web/dist`. |
| `npm start` | Serve the built web app + API from the server (port 3001). |

## API

- `GET /api/health` — service health.
- `GET /api/builds` — list build records.
- `POST /api/builds` — create a build `{ name, branch }`.
- `POST /api/builds/:id/advance` — advance a build's status
  (`queued → running → succeeded`).

## Cloud Agent environment

`.cursor/environment.json` installs dependencies via `npm ci` and launches both
dev servers in named terminals so the app is ready to use when an agent starts.
