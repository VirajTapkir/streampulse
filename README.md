# StreamPulse — Live Stream Analytics Dashboard

A full-stack creator analytics platform that simulates Twitch's streamer
monetization event pipeline and surfaces real-time engagement insights
through an interactive dashboard.

![Live](https://img.shields.io/badge/status-live-brightgreen)
![Go](https://img.shields.io/badge/backend-Go-00ADD8)
![CI](https://github.com/VirajTapkir/streampulse/actions/workflows/ci.yml/badge.svg)
![React](https://img.shields.io/badge/frontend-React-61DAFB)
![PostgreSQL](https://img.shields.io/badge/database-PostgreSQL-336791)
![Redis](https://img.shields.io/badge/cache-Redis-DC382D)

---

## What it does

- Streams live subscription, bits, and donation events to connected dashboards
  via WebSocket — no polling, sub-100ms delivery
- Aggregates live viewer metrics in Redis and persists earnings history in
  PostgreSQL
- Computes a **Stream Momentum Score** — an explainable engagement signal built
  from sub rate, bits/min, and chat density — updated every 5 seconds
- Provides a React dashboard with a live alert feed, emote leaderboard, and
  revenue chart
- Auto-reconnects to the backend with exponential backoff if the connection drops
- Structured logging throughout the backend with Go's slog package
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
│  ┌─────────────────────────────────────────────────┐ │
│  │         Stream Momentum Score (5s updates)      │ │
│  └─────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

---

## Tech stack

| Layer      | Technology                          |
|------------|-------------------------------------|
| Frontend   | React, Recharts, WebSocket API      |
| Backend    | Go (net/http, gorilla/websocket)    |
| Cache      | Redis (ElastiCache on AWS)          |
| Database   | PostgreSQL                          |
| Logging    | Go slog (structured, key-value)     |
| Testing    | Go testing package                  |
| Infra      | AWS ECS, S3, CloudFront             |
| Simulation | Mock event queue (Go channels)      |

---

## Stream Momentum Score

The momentum score is a weighted engagement signal computed every 5 seconds:

```
score = (subRate × 0.5) + (bitsPerMin × 0.3) + (chatDensity × 0.2)
```

Where each component is a per-minute rate derived from a 5-second
sampling window. The weights reflect the relative revenue signal strength
of each event type — subscriptions carry the most weight, chat activity
the least. Counters reset after each computation to prevent accumulation.

---

## API endpoints

| Method | Endpoint          | Description                        |
|--------|-------------------|------------------------------------|
| GET    | `/health`         | Health check — used by AWS ECS     |
| GET    | `/api/streamers`  | List all streamers                 |
| GET    | `/api/earnings`   | Last 50 earnings events            |
| GET    | `/api/counters`   | Live Redis event counters          |
| GET    | `/api/momentum`   | Latest momentum score from Redis   |
| WS     | `/ws`             | WebSocket — real-time event stream |

---

## Local setup

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (for Redis)
- [PostgreSQL 16](https://www.postgresql.org/download/)

### 1 — Clone the repo

```bash
git clone https://github.com/VirajTapkir/streampulse.git
cd streampulse
```

### 2 — Start Redis

```bash
docker run -d --name streampulse-redis -p 6379:6379 redis:7
```

### 3 — Set up PostgreSQL

Create the database and run the schema:

```bash
psql -U postgres -c "CREATE DATABASE streampulse;"
psql -U postgres -d streampulse -f backend/db/schema.sql
```

### 4 — Configure environment

Create `backend/.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=streampulse

REDIS_ADDR=localhost:6379
SERVER_PORT=8080
```

### 5 — Start the backend

```bash
cd backend
go run main.go
```

You should see:

```
time=... level=INFO msg="postgres connected" database=streampulse
time=... level=INFO msg="redis connected" addr=localhost:6379
time=... level=INFO msg="server started" addr=:8080
```

### 6 — Start the frontend

In a new terminal:

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
├── backend/
│   ├── main.go               # entry point — wires all components together
│   ├── api/
│   │   └── handlers.go       # REST API handlers
│   ├── db/
│   │   ├── postgres.go       # PostgreSQL connection pool
│   │   ├── redis.go          # Redis client
│   │   └── schema.sql        # table definitions + seed data
│   ├── events/
│   │   └── queue.go          # mock Twitch Pub/Sub event generator
│   ├── middleware/
│   │   └── cors.go           # CORS middleware
│   ├── scoring/
│   │   ├── momentum.go       # momentum score computation
│   │   └── momentum_test.go  # unit tests
│   └── ws/
│       └── hub.go            # WebSocket hub — client management + fan-out
└── frontend/
    └── src/
        ├── App.jsx                    # root component + layout
        ├── hooks/
        │   └── useWebSocket.js        # WS hook with auto-reconnect
        └── components/
            ├── AlertFeed.jsx          # live event feed
            ├── RevenueChart.jsx       # Recharts line graph
            ├── EmoteLeaderboard.jsx   # top chatters board
            └── MomentumScore.jsx      # engagement score display
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
- **Weighted momentum score** — a single explainable metric is more
  actionable for a creator than three separate counters. The weights
  were chosen to reflect Twitch's public revenue split between
  subscriptions, bits, and donations

---

## License

MIT
