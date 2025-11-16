# Expenses API (Go + Fiber)

This folder contains a new implementation of the Expenses backend written with **Go 1.25**, [Fiber](https://gofiber.io/) and GORM. It mirrors the most important endpoints from the original TypeScript service so both codebases can be compared or used interchangeably.

## Features

- Session based authentication backed by Redis (same cookie name as the TS project).
- User registration and login/logout flows that return the same `CustomResponse` shape.
- CRUD endpoints for categories plus the "delete transactions" safeguard.
- Transaction endpoints supporting bulk inserts, filtering, balances, months-by-year and total savings calculations.
- Health route (`/api/health`) for quick checks.

## Project structure

```
cmd/api            # Application entrypoint
internal/config    # Runtime configuration loader
internal/server    # Fiber bootstrap & routing
internal/service   # Domain logic (users, auth, categories, transactions)
internal/http      # Handlers & middleware
internal/domain    # Database models & enums
pkg                # Shared helpers (responses, errors, date helpers)
```

## Configuration

The server reads the following environment variables (defaults in parentheses):

| Variable | Description |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string (required). |
| `REDIS_URL` | Redis connection string (required). |
| `PORT` (`3000`) | HTTP port. |
| `ENV` (`DEV`) | Environment label, used for logging/cookie flags. |
| `CORS_ORIGINS` (`["*"]`) | JSON array (or comma separated list) with the allowed origins. |
| `SESSION_COOKIE_NAME` (`sessionID`) | Cookie used to keep the session id (matches the TS backend). |
| `SESSION_TTL_HOURS` (`720`, 30 days) | Session lifetime in hours. |

You can reuse the `.env` from `expenses-ts` or create a new one next to this README.

## Running locally

```
# Populate dependencies
cd expenses-go
go mod tidy

# Launch the API (Postgres & Redis can be the ones from docker-compose)
go run ./cmd/api
```

The API will auto-migrate the `users`, `categories` and `transactions` tables on start. If you already ran the TypeScript migrations, both services can share the same database.

> **Note:** The TypeScript project exposes more domains (financial goals, shopping lists, etc.). This Go version currently focuses on auth, categories and transactions, which were the most used flows.
