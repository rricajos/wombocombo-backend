# WomboCombo Backend — API Server

REST API for the WomboCombo multiplayer game. Built with **Go + Fiber + GORM + Redis**.

## Quick Start

```bash
cp .env.example .env
# Edit .env with your database credentials

# Run locally
go run main.go

# Or with Docker
docker build -t wombocombo-backend .
docker run --rm -p 3000:3000 --env-file .env wombocombo-backend
```

## Endpoints

| Method | Route | Auth | Description |
|--------|-------|------|-------------|
| GET | `/api/health` | No | Health check |
| POST | `/api/auth/register` | No | Create account |
| POST | `/api/auth/login` | No | Login |
| POST | `/api/auth/logout` | Yes | Logout |
| POST | `/api/auth/refresh` | No | Refresh JWT |
| POST | `/api/auth/forgot-password` | No | Request reset |
| POST | `/api/auth/reset-password` | No | Reset password |
| GET | `/api/players/me` | Yes | My profile |
| PATCH | `/api/players/me` | Yes | Update profile |
| GET | `/api/players/:id` | Yes | Public profile |
| GET | `/api/players/:id/stats` | Yes | Player stats |
| POST | `/api/rooms` | Yes | Create room |
| GET | `/api/rooms/public` | Yes | List public rooms |
| GET | `/api/rooms/:code` | Yes | Get room by code |

## Deploy

Push to `main` → GitHub Actions builds Docker image → pushes to GHCR.

Package needs to be **public** in GHCR: GitHub → Profile → Packages → Package settings → Change visibility → Public.

## Dependencies

- PostgreSQL 16
- Redis 7
- Go 1.22+
