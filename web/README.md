# Konatsu Mailer — Frontend

The web client for **Konatsu Mailer**, an AI-driven rich web mail client.
Built with React 18 + TypeScript + Vite, TanStack Query, Zustand and Tailwind CSS,
shipped as an installable PWA. It talks to the Go backend over the REST API,
WebSocket and SSE described in the project's design docs (`基本設計.md` / `詳細設計.md`).

## Features

- **三ペインUI (案B)** — sidebar + virtualized mail list + reading pane, responsive
  to a single column on narrow / mobile screens.
- **テーマ & キーカラー** — light/dark/system themes and a user-selectable brand
  key color (default `#ffd20a`), all driven by CSS variables with no flash on load.
- **AI連携** — AI summary cards, priority badges, "AIで返信案 / 下書き" via streaming SSE.
- **リアルタイム** — WebSocket events update the cache in place (no list flicker).
- **PWA** — installable, offline app shell, and Web Push notifications carrying the
  LLM summary of important mail.

## Requirements

- Node.js 20+ (22 recommended)
- The Konatsu Mailer backend running (default `http://localhost:8080`)

## Local development

```bash
cd web
cp .env.example .env        # optional; defaults work out of the box
npm install
npm run dev                 # http://localhost:5173
```

The Vite dev server proxies `/api` (REST + WebSocket) to the backend. Point it at a
different backend with `VITE_DEV_API_TARGET`:

```bash
VITE_DEV_API_TARGET=http://localhost:8080 npm run dev
```

## Scripts

| Command             | Description                              |
| ------------------- | ---------------------------------------- |
| `npm run dev`       | Start the Vite dev server                |
| `npm run build`     | Type-check and build to `dist/`          |
| `npm run preview`   | Preview the production build locally      |
| `npm run typecheck` | Type-check only                          |

## Configuration

| Variable               | Default                  | Description                                          |
| ---------------------- | ------------------------ | ---------------------------------------------------- |
| `VITE_API_BASE_URL`    | `/api`                   | API base URL as seen by the browser                  |
| `VITE_DEV_API_TARGET`  | `http://localhost:8080`  | Dev-server proxy target (dev only)                   |

## Docker

The frontend is part of the root `docker-compose.yml`. From the repository root:

```bash
docker compose up --build
```

This starts PostgreSQL, the backend (`:8080`) and the frontend (`:3000`). The
frontend container is nginx serving the built SPA and reverse-proxying `/api`
(including the WebSocket and SSE streams) to the backend, so everything is
same-origin. Override the published port with `FRONTEND_PORT`.

To build the image on its own:

```bash
cd web
docker build -t konatsu-mailer-web .
```

## PWA icons

The PNG icons in `public/icons/` are generated procedurally (no binary blobs in
review). Regenerate them after changing the brand mark:

```bash
node scripts/generate-icons.mjs
```

## License

Apache License 2.0. See the repository root for details.
