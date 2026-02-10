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

### Internal (game server only)

These endpoints are called by the C++ game server and authenticated via `X-Server-Key` header.

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/internal/player/validate` | Validate JWT + check ban status |
| GET | `/internal/player/:id` | Get player info (cached) |
| POST | `/internal/player/heartbeat` | Batch heartbeat for connected players |
| POST | `/internal/match/start` | Register match start |
| POST | `/internal/match/end` | Submit match result (immediate processing) |
| GET | `/internal/match/:id` | Get active match info |
| PATCH | `/internal/room/status` | Update room status |

### Utility

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| GET | `/api/health` | No | Health check |

## Game Server Integration

The C++ game server communicates with this API via the `/internal/*` endpoints.

### Authentication

All internal endpoints require the `X-Server-Key` header with the shared secret configured in `GAME_SERVER_SECRET`.

```
X-Server-Key: your-shared-secret
```

### Match Lifecycle

```
1. Player connects to game server
   → Game server calls POST /internal/player/validate with JWT
   → Gets back player info + ban status

2. Match starts
   → Game server calls POST /internal/match/start
   → API tracks active match in Redis

3. During match
   → Game server calls POST /internal/player/heartbeat periodically
   → Keeps sessions alive and tracks online players

4. Match ends
   → Game server calls POST /internal/match/end with full results
   → API persists to Postgres, awards currency, updates stats
   → Cleans up Redis tracking data
```

### Match Result Format (POST /internal/match/end)

```json
{
  "match_id": "uuid",
  "room_id": "room-id",
  "started_at": "2025-01-01T12:00:00Z",
  "ended_at": "2025-01-01T12:15:00Z",
  "rounds_completed": 5,
  "map_id": "arena_01",
  "players": [
    {
      "player_id": "uuid",
      "kills": 12,
      "deaths": 3,
      "score": 1500,
      "rounds_survived": 4,
      "coop_revives": 2
    }
  ]
}
```

### Currency Rewards

Players earn currency automatically after each match:
- **10** per round completed
- **5** per kill
- **2** per coop revive
- **15** per round survived

### Alternative: Redis Worker Path

If the game server writes match results to Redis key `match:<id>:result`, the match_processor worker picks them up within 5 seconds. The `/internal/match/end` endpoint is preferred for lower latency.

## Background Workers

Workers are enabled by default (`WORKERS_ENABLED=true`) and can be disabled for testing.

| Worker | Interval | Description |
|--------|----------|-------------|
| `cleanup` | 30s | Removes expired rooms from public rooms set |
| `match_processor` | 5s | Processes match results written to Redis (fallback path) |
| `session_heartbeat` | 60s | Counts online players, publishes live stats to Redis |

## Architecture

```
├── config/         # Environment config loader
├── database/       # Postgres + Redis connections
├── dto/            # Request/response structs (incl. internal API)
├── errors/         # AppError type + helpers
├── handlers/       # HTTP handlers (public + internal)
├── middleware/      # Auth, CORS, rate limit, admin, internal auth
├── models/         # GORM models
├── routes/         # Route registration (public + internal)
├── services/       # Business logic (incl. game server service)
├── utils/          # JWT, hashing, pagination, random
└── workers/        # Background workers (cleanup, match, heartbeat)
```

## Redis Key Reference

| Pattern | TTL | Description |
|---------|-----|-------------|
| `session:<player_id>` | 24h | Active session token |
| `player_cache:<player_id>` | 24h | Cached player info hash |
| `room:<room_id>` | 2h | Room data (JSON) |
| `room_code:<code>` | 2h | Room code → room ID mapping |
| `rooms:public` | 2h | Set of public room IDs |
| `active_match:<match_id>` | 4h | Active match tracking (JSON) |
| `player_match:<player_id>` | 4h | Player's current match ID |
| `heartbeat:<player_id>` | 5m | Game server heartbeat |
| `match:<id>:result` | ∞ | Match result pending processing |
| `reset:<token>` | 1h | Password reset token |
| `jwt:secret` | ∞ | JWT secret for game server |
| `server:stats:live` | 5m | Live server stats hash |

## Deploy

Push to `main` → GitHub Actions builds Docker image → pushes to GHCR.

## Dependencies

- PostgreSQL 16
- Redis 7
- Go 1.22+
