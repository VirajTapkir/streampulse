# StreamPulse — Live Stream Analytics Dashboard

A full-stack creator analytics platform that simulates Twitch's streamer
monetization event pipeline and surfaces real-time engagement insights
through an interactive dashboard.

![CI](https://github.com/VirajTapkir/streampulse/actions/workflows/ci.yml/badge.svg?branch=main)
![Go](https://img.shields.io/badge/backend-Go-00ADD8)
![React](https://img.shields.io/badge/frontend-React-61DAFB)
![PostgreSQL](https://img.shields.io/badge/database-PostgreSQL-336791)
![Redis](https://img.shields.io/badge/cache-Redis-DC382D)
![Docker](https://img.shields.io/badge/containers-Docker-2496ED)

---

## What it does

- Streams live subscription, bits, and donation events to connected dashboards
  via WebSocket — no polling, sub-100ms delivery
- Simulates real Twitch EventSub payloads — `channel.subscribe`,
  `channel.cheer`, and `channel.charity_campaign.donate` event schemas
- Supports multiple streamers — each streamer gets their own independent
  event stream and Redis namespace
- Aggregates live viewer metrics in Redis and persists earnings history in
  PostgreSQL
- Computes a **Stream Momentum Score** — an explainable engagement signal built
  from sub rate, bits/min, and chat density — updated every 5 seconds
- Fires configurable alerts when momentum drops below a user-defined threshold
- Provides a React dashboard with a live alert feed, emote leaderboard,
  real-time revenue chart, and 7-day historical revenue breakdown
- Auto-reconnects to the backend with exponential backoff if the connection drops
- Structured logging throughout the backend with Go's `slog` package
- Rate limiting — token bucket algorithm, 10 requests/sec per IP with burst of 20
- Graceful shutdown — drains in-flight requests before exiting on SIGTERM

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Go Backend (:8080)                 │
│                                                      │
│  ┌─────────────┐    ┌──────────────┐                │
│  │  REST API   │    │  Event Queue │                │
│  │ /api/*      │    │ (Go channel) │                │
│  └──────┬──────┘    └──────┬───────┘                │
│         │                  │                         │
│         │           ┌──────▼───────┐                │
│         │           │ WebSocket Hub│                │
│         │           │   fan-out    │                │
│         │           └──────┬───────┘                │
│         │                  │                         │
│  ┌──────▼──────┐    ┌──────▼───────┐                │
│  │  PostgreSQL │    │    Redis     │                │
│  │  earnings   │    │  counters +  │                │
│  │  history    │    │  momentum    │                │
│  └─────────────┘    └──────────────┘                │
└─────────────────────────────────────────────────────┘
                          │ WebSocket
┌─────────────────────────▼───────────────────────────┐
│              React Dashboard (:3000)                 │
│                                                      │
│  ┌───────────┐ ┌──────────────┐ ┌────────────────┐  │
│  │ Alert Feed│ │Revenue Chart │ │ Emote Board    │  │
│  └───────────┘ └──────────────┘ └────────────────┘  │
│  ┌──────────────────┐ ┌───────────────────────────┐  │
│  │ Momentum Score   │ │ 7-day Revenue Breakdown   │  │
│  │ + alert system   │ │ (bar chart by event type) │  │
│  └──────────────────┘ └───────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

---

## Tech stack

| Layer        | Technology                              |
|--------------|-----------------------------------------|
| Frontend     | React, Recharts, WebSocket API          |
| Backend      | Go (net/http, gorilla/websocket)        |
| Cache        | Redis 7                                 |
| Database     | PostgreSQL 16                           |
| Logging      | Go slog (structured, key-value)         |
| Testing      | Go testing package                      |
| Rate limiting| Token bucket (golang.org/x/time/rate)   |
| Containers   | Docker, Docker Compose                  |
| CI/CD        | GitHub Actions                          |
| Simulation   | Mock Twitch EventSub event queue        |

---

## Stream Momentum Score

The momentum score is a weighted engagement signal computed every 5 seconds
per streamer:

```
score = (subRate × 0.5) + (bitsPerMin × 0.3) + (chatDensity × 0.2)
```

Where each component is a per-minute rate derived from a 5-second sampling
window. The weights reflect the relative revenue signal strength of each
event type — subscriptions carry the most weight, chat activity the least.
Counters reset after each computation to prevent accumulation.

Configurable alerting thresholds trigger warning and critical banners on
the dashboard when engagement drops below a user-defined level.

---

## API endpoints

| Method | Endpoint            | Description                              |
|--------|---------------------|------------------------------------------|
| GET    | `/health`           | Health check                             |
| GET    | `/api/streamers`    | List all streamers                       |
| GET    | `/api/earnings`     | Last 50 earnings events per streamer     |
| GET    | `/api/counters`     | Live Redis event counters per streamer   |
| GET    | `/api/momentum`     | Latest momentum score from Redis         |
| GET    | `/api/analytics`    | 7-day daily revenue breakdown            |
| WS     | `/ws`               | WebSocket — real-time event stream       |

All endpoints that return per-streamer data accept a `?streamer_id=` query
parameter. The WebSocket endpoint accepts `?streamer_id=` to subscribe to
a specific streamer's event stream.

---

## Local setup

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (for Redis + one-command startup)
- [PostgreSQL 16](https://www.postgresql.org/download/)

### Option A — Docker Compose (recommended)

One command starts all four services:

```bash
git clone https://github.com/VirajTapkir/streampulse.git
cd streampulse
cp .env.example .env   # then fill in your DB_PASSWORD
docker-compose up --build
```

Visit `http://localhost:3000` — the dashboard is live.

### Option B — Manual setup

**1 — Start Redis via Docker:**
```bash
docker run -d --name streampulse-redis -p 6379:6379 redis:7
```

**2 — Set up PostgreSQL:**
```bash
psql -U postgres -c "CREATE DATABASE streampulse;"
psql -U postgres -d streampulse -f backend/db/schema.sql
```

**3 — Configure environment — create `backend/.env`:**
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=streampulse

REDIS_ADDR=localhost:6379
SERVER_PORT=8080
```

**4 — Start the backend:**
```bash
cd backend
go run main.go
```

**5 — Start the frontend:**
```bash
cd frontend
npm install
npm start
```

Visit `http://localhost:3000` — the dashboard should be live.

---

## Running tests

```bash
cd backend
go test ./... -v
```

Expected output:

```
=== RUN   TestComputeZeroCounters
--- PASS: TestComputeZeroCounters (0.00s)
=== RUN   TestComputeSubsOnly
--- PASS: TestComputeSubsOnly (0.00s)
=== RUN   TestComputeWeightedFormula
--- PASS: TestComputeWeightedFormula (0.00s)
=== RUN   TestScoreWeights
--- PASS: TestScoreWeights (0.00s)
=== RUN   TestComputeResetsCounters
--- PASS: TestComputeResetsCounters (0.00s)
ok  github.com/VirajTapkir/streampulse/scoring
```

---

## Project structure

```
streampulse/
├── docker-compose.yml              # starts all four services
├── .env.example                    # copy to .env and fill in secrets
├── backend/
│   ├── main.go                     # entry point — wires all components
│   ├── api/
│   │   └── handlers.go             # REST API handlers
│   ├── db/
│   │   ├── postgres.go             # PostgreSQL connection pool
│   │   ├── redis.go                # Redis client
│   │   └── schema.sql              # table definitions + seed data
│   ├── events/
│   │   └── queue.go                # mock Twitch EventSub event generator
│   ├── middleware/
│   │   ├── cors.go                 # CORS middleware
│   │   └── ratelimit.go            # token bucket rate limiter
│   ├── scoring/
│   │   ├── momentum.go             # momentum score computation
│   │   └── momentum_test.go        # unit tests
│   └── ws/
│       └── hub.go                  # WebSocket hub + fan-out
└── frontend/
    └── src/
        ├── App.jsx                         # root component + layout
        ├── hooks/
        │   └── useWebSocket.js             # WS hook with auto-reconnect
        └── components/
            ├── AlertFeed.jsx               # live event feed
            ├── RevenueChart.jsx            # real-time revenue line chart
            ├── EmoteLeaderboard.jsx        # top chatters leaderboard
            ├── MomentumScore.jsx           # engagement score + alerting
            └── HistoricalChart.jsx         # 7-day revenue bar chart
```

---

## Why I built this

This project was designed to mirror the engineering challenges in
Twitch's Streamer Monetization Experience team — specifically around
real-time event delivery at scale, creator-facing metrics, and
full-stack Go + React development.

The core technical decisions reflect real production constraints:

- **Go channels over message brokers** — for a single-node deployment,
  Go's built-in channels give you a Twitch Pub/Sub-like architecture
  without the operational overhead of Kafka or RabbitMQ
- **Redis for hot data, PostgreSQL for cold data** — live counters that
  update 10+ times per second belong in memory; earnings history that
  powers charts belongs on disk
- **WebSocket fan-out over polling** — polling at 1-second intervals
  for 1000 concurrent users means 1000 requests/second; a single
  WebSocket connection per user means 0 polling requests and sub-100ms
  delivery
- **Token bucket rate limiting** — allows legitimate burst traffic
  (a browser firing multiple requests on page load) while protecting
  against abusive clients, without the thundering herd problem of
  fixed-window limiting
- **Per-streamer Redis namespacing** — prefixing all keys with
  `streamer:{id}:` gives each streamer completely independent counters
  and momentum scores with zero cross-contamination, at no extra
  infrastructure cost
- **Weighted momentum score** — a single explainable metric is more
  actionable for a creator than three separate counters. The weights
  were chosen to reflect Twitch's public revenue split between
  subscriptions, bits, and donations

---

## License

MIT
