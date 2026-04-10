# HyperMotion League

A fantasy football web application for **La Liga Hypermotion** (Spanish Segunda Division). Create leagues with friends, sign real players, set your lineup each matchday, and compete for the top of the standings.

---

## Screenshots

### Login Screen

![Login Screen](docs/screenshots/login.png)

### Register Screen

![Register Screen](docs/screenshots/register.png)

### My Leagues

![My Leagues](docs/screenshots/myleagues.png)

More screens will be added soon!

## Features

### Authentication

- Email/password registration and login
- Google and Apple sign-in via Supabase OAuth
- JWT-based session management with automatic token handling
- Auth guard — unauthenticated users are redirected to the login screen

### Leagues

- Create private leagues with custom budget and member limits
- Join leagues via shareable invite codes
- View league members and settings
- Owner-only league deletion with confirmation

### Real Football Data

- 22 real teams from La Liga Hypermotion imported from API-Football
- 60+ real players with positions, photos, and nationalities
- 468 real fixtures with scores and match status
- Import script that can be re-run safely (upsert logic)

### Coming Soon

- Transfer market with bidding system (max 5 active bids per user)
- Lineup editor with formation selection (4-3-3, 4-4-2, etc.)
- League standings and matchday-by-matchday scoring
- Player points based on real match performance
- News feed from sports media

---

## Tech Stack

| Layer    | Technology                                  |
| -------- | ------------------------------------------- |
| Backend  | Go 1.25, Gin                                |
| Frontend | Vue 3, TypeScript, Vite                     |
| Styling  | Tailwind CSS                                |
| Database | PostgreSQL (Supabase)                       |
| Auth     | Custom JWT + Supabase OAuth (Google, Apple) |
| Testing  | Vitest, Playwright, Go test                 |

---

## Project Structure

```
hypermotionleague/
├── backend/
│   ├── cmd/server/          # Server entry point
│   ├── internal/
│   │   ├── auth/            # JWT generation and validation
│   │   ├── config/          # Environment variable loading
│   │   ├── db/              # Database connection pool
│   │   ├── handlers/        # HTTP request handlers
│   │   ├── middleware/       # JWT auth middleware
│   │   ├── models/          # Domain entities
│   │   └── repository/      # Data access layer (interfaces + PostgreSQL impl)
│   ├── db/
│   │   ├── migrations/      # SQL schema migrations (001, 002, 003)
│   │   └── seeds/           # Sample data for development
│   └── scripts/             # Data import from API-Football
├── frontend/
│   └── src/
│       ├── layouts/         # App shell with bottom navigation
│       ├── lib/             # API client, Supabase config
│       ├── router/          # Vue Router with auth guards
│       └── views/           # Page components
├── e2e/                     # Playwright end-to-end tests
├── .github/workflows/       # CI/CD pipelines
└── Makefile                 # Dev commands
```

---

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 20+
- A [Supabase](https://supabase.com) project (free tier works)
- An [API-Football](https://www.api-football.com/) key (free tier, optional — for importing real player data)

### 1. Clone and install dependencies

```bash
git clone https://github.com/isw2-unileon/hypermotionleague.git
cd hypermotionleague
make install
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and fill in:

```env
# Database
DATABASE_URL=postgresql://postgres:YOUR_PASSWORD@db.YOUR_PROJECT.supabase.co:5432/postgres

# Auth
JWT_SECRET=a-long-random-secret

# Supabase (for OAuth)
SUPABASE_URL=https://YOUR_PROJECT.supabase.co
SUPABASE_SERVICE_KEY=your-service-role-key

# API-Football (optional, for data import)
API_FOOTBALL_KEY=your-api-key
```

Create `frontend/.env`:

```env
VITE_SUPABASE_URL=https://YOUR_PROJECT.supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key
```

### 3. Set up the database

Run the migration files **in order** in Supabase Dashboard > SQL Editor:

1. `backend/db/migrations/001_initial_schema.up.sql` — Core tables
2. `backend/db/migrations/002_add_auth_provider.up.sql` — OAuth support
3. `backend/db/migrations/003_add_api_football_tables.up.sql` — Teams, fixtures, player metadata

### 4. Import real football data (optional)

```bash
go run backend/scripts/import_data.go
```

This imports teams, players, and fixtures from API-Football for La Liga Hypermotion. The script uses upsert logic, so it's safe to run multiple times.

### 5. Run the app

```bash
# Terminal 1 — Backend (port 8080)
make run-backend

# Terminal 2 — Frontend (port 5173)
make run-frontend
```

Open **http://localhost:5173**

---

## API Endpoints

### Public

| Method | Path                    | Description                  |
| ------ | ----------------------- | ---------------------------- |
| `GET`  | `/health`               | Health check                 |
| `GET`  | `/api/db-test`          | Database connection test     |
| `POST` | `/api/v1/auth/register` | Register with email/password |
| `POST` | `/api/v1/auth/login`    | Login with email/password    |
| `POST` | `/api/v1/auth/oauth`    | Supabase OAuth callback      |

### Protected (requires `Authorization: Bearer <token>`)

| Method   | Path                          | Description                |
| -------- | ----------------------------- | -------------------------- |
| `GET`    | `/api/v1/leagues`             | List user's leagues        |
| `POST`   | `/api/v1/leagues`             | Create a new league        |
| `POST`   | `/api/v1/leagues/join`        | Join league by invite code |
| `GET`    | `/api/v1/leagues/:id`         | Get league details         |
| `GET`    | `/api/v1/leagues/:id/members` | List league members        |
| `DELETE` | `/api/v1/leagues/:id`         | Delete league (owner only) |

---

## Development Commands

| Command               | Description                                    |
| --------------------- | ---------------------------------------------- |
| `make install`        | Install all dependencies (Go, npm, Playwright) |
| `make run-backend`    | Backend with hot reload via Air                |
| `make run-frontend`   | Frontend dev server (Vite)                     |
| `make test`           | Run all tests (Go + Vitest)                    |
| `make lint`           | Run linters (golangci-lint + ESLint)           |
| `make e2e`            | Run Playwright E2E tests                       |
| `make build-backend`  | Build backend binary                           |
| `make build-frontend` | Build frontend for production                  |

---

## Database Schema

The app uses 13 tables organized around fantasy football concepts:

**Core:** `users`, `teams`, `players`, `fixtures`

**Fantasy:** `leagues`, `league_members`, `team_players`, `matchdays`, `player_points`, `lineups`, `lineup_players`

**Market:** `market_listings`, `bids`

See `backend/db/migrations/` for the full schema with constraints, indexes, and triggers.

---

## Team

Built by the ISW2 team at Universidad de Leon.

## License

This project is for educational purposes as part of the Software Engineering II course.
