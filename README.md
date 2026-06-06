<div align="center">

# ⚽ HyperMotion League

**Fantasy football for La Liga Hypermotion — the Spanish Segunda División.**

Build a squad of real players, run the transfer market, set your lineup every matchday, and climb the league standings against your friends.

### 🔗 Live app → **[hypermotionleague.vercel.app](https://hypermotionleague.vercel.app)**

[![Backend](https://img.shields.io/badge/backend-Go%201.25%20%2B%20Gin-00ADD8)](#tech-stack)
[![Frontend](https://img.shields.io/badge/frontend-Vue%203%20%2B%20Vite-42b883)](#tech-stack)
[![Database](https://img.shields.io/badge/database-PostgreSQL%20%2F%20Supabase-3ECF8E)](#tech-stack)

</div>

---

## What you can do

HyperMotion League turns the real Segunda División season into a fantasy competition. Sign up, create or join a private league, and manage your own team across the campaign.

- **Create & join leagues** — spin up a private league with a custom budget and member limit, then invite friends with a shareable join code.
- **Sign real players** — pick from real La Liga Hypermotion squads imported from API-Football, complete with positions, photos, and nationalities.
- **Run the transfer market** — browse available players, place bids (with anti-abuse limits on active bids), follow a live activity feed, and track listing status.
- **Set your lineup** — choose a formation and your starters for each matchday from the players you own.
- **Score real points** — player performance is scored from real match results and rolled up matchday by matchday.
- **Compete on the table** — follow overall and per-matchday standings to see who's winning the league.
- **Sign in your way** — email/password or Google / Apple OAuth via Supabase, with JWT-based sessions.

---

## Tech stack

| Layer         | Technology                                                       |
| ------------- | ---------------------------------------------------------------- |
| Backend       | Go 1.25, Gin, pgx (PostgreSQL driver)                            |
| Frontend      | Vue 3, TypeScript, Vite, Vue Router                              |
| Styling       | Tailwind CSS v4                                                  |
| Database      | PostgreSQL (Supabase)                                            |
| Auth          | Custom JWT + Supabase OAuth (Google, Apple)                      |
| External data | API-Football (teams, players, fixtures)                          |
| Testing       | Go test, Vitest, Playwright                                      |
| Hosting       | **Vercel** (frontend) · **Render** (backend) · **Supabase** (DB) |

---

## Run it locally

Clone the repo and you can have the full stack running with a handful of `make` commands.

### Prerequisites

- **Go** 1.25+
- **Node.js** 20+
- **PostgreSQL** — either a free [Supabase](https://supabase.com) project or a local instance (a `docker-compose.yml` is included)
- An [API-Football](https://www.api-football.com/) key (free tier) — _only_ needed if you want to re-import real data

### 1. Clone and install

```bash
git clone git@github.com:isw2-unileon/hypermotionleague.git
cd hypermotionleague
make install
```

`make install` pulls Go modules, installs the dev tooling (Air for hot reload, golangci-lint), and runs `npm ci` for both the frontend and the e2e suite.

### 2. Configure environment variables

**Backend** — copy the template and fill in your values:

```bash
cp .env.example .env
```

```env
PORT=8080
GIN_MODE=debug
CORS_ALLOW_ORIGINS=http://localhost:5173,http://localhost:3000

DATABASE_URL=postgresql://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres
JWT_SECRET=a-long-random-secret

SUPABASE_URL=https://[PROJECT-REF].supabase.co
SUPABASE_SERVICE_KEY=your-service-role-key

# Optional — only for re-importing real data
API_FOOTBALL_KEY=your-api-key
```

**Frontend** — create `frontend/.env`:

```env
# Leave VITE_API_URL empty in local dev so requests go through the Vite proxy
VITE_API_URL=

VITE_SUPABASE_URL=https://[PROJECT-REF].supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key
```

> 💡 Prefer a fully local database? Run `docker compose up -d` to start PostgreSQL on port `5433`, and point `DATABASE_URL` at it.

### 3. Set up the database

Apply the schema migrations and load seed data. Both targets are idempotent, so they're safe to re-run:

```bash
make migrate DB_URL="<your-database-url>"
make seed    DB_URL="<your-database-url>"
```

Migrations live in `backend/db/migrations/` and the seed data in `backend/db/seeds/`.

### 4. (Optional) Import real football data

To refresh teams, players, and fixtures straight from API-Football (upsert logic — safe to re-run):

```bash
go run ./backend/cmd/ingest
```

### 5. Start the app

```bash
# Terminal 1 — backend with hot reload (port 8080)
make run-backend

# Terminal 2 — frontend dev server (port 5173)
make run-frontend
```

Open **http://localhost:5173** and you're in. 🎉

---

## Deploying

The production app runs across three managed services. The repo is set up so each piece deploys independently from the same monorepo.

### Frontend → Vercel

- Connect the repo and set the project root to `frontend/`.
- Build command `npm run build`, output directory `dist`. `frontend/vercel.json` already handles SPA routing rewrites.
- Set environment variables in the Vercel project: `VITE_API_URL` (the full public Render URL, no trailing slash), `VITE_SUPABASE_URL`, and `VITE_SUPABASE_ANON_KEY`.

### Backend → Render

- Deploy as a Docker service using `backend/Dockerfile`. **Keep the build context at the repo root** — the monorepo is a single Go module rooted there.
- On Render set _Dockerfile Path_ = `backend/Dockerfile`.
- Provide the env vars: `DATABASE_URL`, `JWT_SECRET`, `SUPABASE_URL`, `SUPABASE_SERVICE_KEY`, and `CORS_ALLOW_ORIGINS` (include your Vercel domain). Render injects `PORT` automatically.

### Database → Supabase

A hosted PostgreSQL project. Run the migrations (`make migrate`) against it and configure Supabase Auth for Google/Apple OAuth.

---

## Project layout

```
hypermotionleague/
├── backend/
│   ├── cmd/
│   │   ├── server/          # HTTP API entry point
│   │   ├── ingest/          # Import real data from API-Football
│   │   ├── resolve/         # Data resolution job
│   │   └── sync-players/    # Player sync job
│   ├── internal/
│   │   ├── auth/            # JWT generation & validation
│   │   ├── handlers/        # HTTP handlers (auth, leagues, market, lineup, …)
│   │   ├── middleware/      # JWT auth middleware
│   │   ├── repository/      # Data access (interfaces + PostgreSQL impl)
│   │   ├── market/          # Transfer market logic
│   │   ├── scoring/         # Matchday scoring engine
│   │   ├── apifootball/     # API-Football client
│   │   └── models/          # Domain entities
│   └── db/
│       ├── migrations/      # Versioned SQL schema
│       └── seeds/           # Development seed data
├── frontend/
│   └── src/
│       ├── views/          # Pages (Leagues, Market, Lineup, Standings, Team, …)
│       ├── layouts/        # App shell + navigation
│       ├── router/         # Vue Router with auth guards
│       ├── lib/            # API client & Supabase config
│       └── design-system/  # Shared UI components
├── e2e/                     # Playwright end-to-end tests
├── .github/workflows/       # CI pipelines (backend, frontend, e2e, CodeQL)
├── docker-compose.yml       # Local PostgreSQL
└── Makefile                 # Dev commands
```

---

## API reference

### Public

| Method | Path                         | Description                  |
| ------ | ---------------------------- | ---------------------------- |
| `GET`  | `/health`                    | Health check                 |
| `POST` | `/api/v1/auth/register`      | Register with email/password |
| `POST` | `/api/v1/auth/login`         | Login with email/password    |
| `POST` | `/api/v1/auth/oauth`         | Supabase OAuth callback      |
| `GET`  | `/api/v1/players`            | List players                 |
| `GET`  | `/api/v1/players/:id`        | Player details               |
| `GET`  | `/api/v1/players/:id/points` | Player points by matchday    |

### Protected — `Authorization: Bearer <token>`

| Method   | Path                                           | Description                |
| -------- | ---------------------------------------------- | -------------------------- |
| `GET`    | `/api/v1/leagues`                              | List your leagues          |
| `POST`   | `/api/v1/leagues`                              | Create a league            |
| `POST`   | `/api/v1/leagues/join`                         | Join by invite code        |
| `GET`    | `/api/v1/leagues/:id`                          | League details             |
| `GET`    | `/api/v1/leagues/:id/members`                  | League members             |
| `DELETE` | `/api/v1/leagues/:id`                          | Delete league (owner only) |
| `GET`    | `/api/v1/leagues/:id/standings`                | League standings           |
| `GET`    | `/api/v1/leagues/:id/matchdays`                | List matchdays             |
| `GET`    | `/api/v1/leagues/:id/matchdays/current`        | Current matchday           |
| `GET`    | `/api/v1/leagues/:id/team`                     | Your team in the league    |
| `GET`    | `/api/v1/leagues/:id/matchdays/:number/lineup` | Get lineup                 |
| `PUT`    | `/api/v1/leagues/:id/matchdays/:number/lineup` | Save lineup                |
| `GET`    | `/api/v1/leagues/:id/market/players`           | Available players          |
| `GET`    | `/api/v1/leagues/:id/market/listings`          | Active listings            |
| `POST`   | `/api/v1/leagues/:id/market/bids`              | Place a bid                |
| `GET`    | `/api/v1/leagues/:id/market/bids`              | Your bids                  |
| `DELETE` | `/api/v1/leagues/:id/market/bids/:bid_id`      | Cancel a bid               |
| `GET`    | `/api/v1/leagues/:id/market/feed`              | Recent market activity     |

---

## Development commands

| Command               | Description                                    |
| --------------------- | ---------------------------------------------- |
| `make install`        | Install all dependencies (Go, npm, Playwright) |
| `make run-backend`    | Backend with hot reload (Air)                  |
| `make run-frontend`   | Frontend dev server (Vite)                     |
| `make migrate`        | Apply pending DB migrations (idempotent)       |
| `make seed`           | Load seed data (idempotent)                    |
| `make test`           | Run all tests (Go + Vitest)                    |
| `make lint`           | Run linters (golangci-lint + ESLint)           |
| `make e2e`            | Run Playwright E2E tests                       |
| `make build-backend`  | Build the backend binary                       |
| `make build-frontend` | Build the frontend for production              |

---

## Team & license

Built by the **ISW2 Group 9 at Universidad de León** as part of the Software Engineering II course. For educational purposes.
