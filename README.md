# PoloBomba Backend — API Server

REST API for the PoloBomba multiplayer game. Built with **Go + Fiber + GORM + Redis**.

## Quick Start

```bash
cp .env.example .env
# Edit .env with your database credentials

# Run locally
go run main.go

# Or with Docker
docker build -t polobomba-backend .
docker run --rm -p 3000:3000 --env-file .env polobomba-backend
```

## Endpoints

### Auth (public)

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/api/auth/register` | Create account |
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/refresh` | Refresh JWT |
| POST | `/api/auth/forgot-password` | Request reset |
| POST | `/api/auth/reset-password` | Reset password |
| POST | `/api/auth/logout` | Logout (auth required) |

### Players (auth required)

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/players/me` | My profile |
| PATCH | `/api/players/me` | Update profile |
| GET | `/api/players/:id` | Public profile |
| GET | `/api/players/:id/stats` | Player stats |

### Rooms (auth required)

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/api/rooms` | Create room |
| GET | `/api/rooms/public` | List public rooms |
| GET | `/api/rooms/:code` | Get room by code |

### Stats (auth required)

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/stats/history/:id` | Match history (`:id` or `me`) |
| GET | `/api/stats/leaderboard` | Leaderboard (`?sort=kills&page=1&per_page=20`) |

**Leaderboard sort options:** `kills`, `deaths`, `rounds`, `survived`, `playtime`, `revives`

### Friends (auth required)

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/friends` | List accepted friends (with online status) |
| GET | `/api/friends/pending` | List incoming friend requests |
| POST | `/api/friends/request` | Send friend request `{"friend_id": "uuid"}` |
| POST | `/api/friends/accept` | Accept friend request `{"friend_id": "uuid"}` |
| DELETE | `/api/friends/:id` | Remove friend |
| POST | `/api/friends/block` | Block player `{"friend_id": "uuid"}` |

### Inventory (auth required)

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/inventory` | List my items |
| POST | `/api/inventory/unlock` | Unlock item `{"item_type": "skin_zombie_01"}` |
| GET | `/api/inventory/catalog` | Get item catalog with prices |

### Admin (auth + admin required)

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/api/admin/ban` | Ban player `{"player_id": "uuid"}` |
| POST | `/api/admin/unban` | Unban player `{"player_id": "uuid"}` |
| GET | `/api/admin/stats` | Server stats (players, sessions, rooms, matches) |

### Utility

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| GET | `/api/health` | No | Health check |

## Architecture

```
├── config/         # Environment config loader
├── database/       # Postgres + Redis connections
├── dto/            # Request/response structs
├── errors/         # AppError type + helpers
├── handlers/       # HTTP handlers (Fiber)
├── middleware/      # Auth, CORS, rate limit, admin
├── models/         # GORM models
├── routes/         # Route registration
├── services/       # Business logic
├── utils/          # JWT, hashing, pagination, random
└── workers/        # Background workers (cleanup, match processor)
```

## Deploy

Push to `main` → GitHub Actions builds Docker image → pushes to GHCR.

## Dependencies

- PostgreSQL 16
- Redis 7
- Go 1.22+
