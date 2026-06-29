<div align="center">

# 📈 Stock Simulator

**A production-style, real-time market data streaming server**

*Built with Go · WebSockets · Redis Pub/Sub · Binance Streams*

<br/>

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Redis](https://img.shields.io/badge/Redis-Pub%2FSub-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![Binance](https://img.shields.io/badge/Binance-WebSocket-F0B90B?style=for-the-badge&logo=binance&logoColor=white)](https://binance.com/)
[![Render](https://img.shields.io/badge/Render-Deploy%20Ready-46E3B7?style=for-the-badge&logo=render&logoColor=white)](https://render.com/)

<br/>

> WebSocket clients subscribe to crypto symbols and receive **live market updates** with automatic feed management, reconnection handling, metrics, and horizontal scalability.

</div>

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 📡 Real-Time Market Data
- Live prices from Binance WebSocket streams
- Dynamic symbol subscriptions & unsubscriptions
- Feed activates **only when a client is listening**

### 🔌 WebSocket Server
- Multiple concurrent clients
- Per-client subscriptions
- Heartbeat ping/pong support
- Stale connection cleanup & backpressure protection

### 🔴 Redis Pub/Sub
- Decouples ingestion from delivery
- Enables **horizontal scaling**
- Multiple WS server instances share the same feed

</td>
<td width="50%">

### 🎛️ Feed Management
- Tracks active symbol counts
- Auto-subscribes to Binance on first client
- Auto-unsubscribes when last client leaves
- Prevents unnecessary Binance traffic

### 🛡️ Reliability
- Automatic Binance reconnection
- Stream re-subscription after reconnect
- Graceful shutdown (SIGINT/SIGTERM)
- Context-based cancellation

### 📊 Observability
- Connected clients, active symbols
- Messages received / sent totals
- Binance reconnect counter
- Top subscribed symbols

</td>
</tr>
</table>

---

## 🏗️ Architecture

```
                       ┌──────────────────┐
                       │ Binance WS Feed  │
                       └─────────┬────────┘
                                 │  live trade stream
                                 ▼
                 ┌────────────────────────────┐
                 │  Binance Ingestion Service │
                 │                            │
                 │  · Dynamic Feed Management │
                 │  · Reconnection Logic      │
                 └─────────────┬──────────────┘
                               │  publish
                               ▼
                      Redis Pub/Sub Channel
                               │  subscribe
                               ▼
                 ┌───────────────────────────┐
                 │       Redis Subscriber    │
                 └─────────────┬─────────────┘
                               │  broadcast
                               ▼
                 ┌───────────────────────────┐
                 │            Hub            │
                 │  · Register / Unregister  │
                 │  · Broadcast Events       │
                 │  · Symbol Tracking        │
                 └────────┬──────────┬───────┘
                          │          │
                          ▼          ▼
                      Client A    Client B
                      BTCUSDT     ETHUSDT
```

---

## 📦 Project Structure

```
stock-sim/
├── cmd/
│   └── main.go                  # Entry point
│
├── internal/
│   ├── domain/
│   │   ├── binance.go
│   │   ├── client.go
│   │   ├── feed_command.go
│   │   ├── market_event.go
│   │   ├── stock.go
│   │   └── subscription.go
│   │
│   ├── hub/
│   │   └── hub.go               # Client registration & broadcasting
│   │
│   ├── market/
│   │   ├── binance.go           # Binance WS ingestion
│   │   └── subscriber.go        # Redis subscriber
│   │
│   ├── metrics/
│   │   └── metrics.go           # Custom HTTP metrics
│   │
│   ├── redis/
│   │   └── client.go            # Redis client wrapper
│   │
│   └── websocket/
│       └── pumps.go             # Read/write pumps per client
│
├── Dockerfile
├── docker-compose.yml
├── .env
└── README.md
```

---

## 🔄 Event Flow

### Client Subscription

```
Client  ──►  WebSocket  ──►  Hub
                               │
                    SymbolCount == 1?
                               │
                               └──►  FeedCommand(subscribe)  ──►  Binance Feed
```

### Market Data Flow

```
Binance  ──►  Ingestion Service  ──►  Redis Publish
                                            │
                                      Redis Subscribe
                                            │
                                           Hub
                                            │
                                    Subscribed Clients
```

---

## 🔌 WebSocket API

### Connect

| Environment | URL |
|-------------|-----|
| Local | `ws://localhost:8080/ws` |
| Production | `wss://your-domain/ws` |

### Subscribe

```json
{
  "type": "subscribe",
  "symbol": ["BTCUSDT", "ETHUSDT"]
}
```

### Unsubscribe

```json
{
  "type": "unsubscribe",
  "symbol": ["ETHUSDT"]
}
```

### Market Update (Server → Client)

```json
{
  "symbol": "BTCUSDT",
  "price": 63542.12
}
```

---

## 📊 Metrics

```
GET /metrics
```

**Example Response:**

```json
{
  "active_symbols": 2,
  "binance_reconnects_total": 0,
  "connected_clients": 5,
  "messages_received_total": 1245,
  "messages_sent_total": 9860,
  "top_symbols": [
    { "symbol": "BTCUSDT", "count": 3 },
    { "symbol": "ETHUSDT", "count": 2 }
  ]
}
```

---

## ⚡ Load Testing & Performance

The Stock Simulator is optimized for high concurrency and has been load-tested with **10,000 concurrent WebSocket clients** to verify message delivery and system stability under high-throughput conditions.

### Load Test Overview
- **Concurrently Connected Clients**: 10,000
- **Subscription**: Every client successfully subscribes to the live `BTCUSDT` stream.
- **Heartbeat & Connection Management**: Keeps connections active without drops, responding to server-initiated pings within the timeout threshold.
- **Implementation**: Driven by the custom load-test script in [internal/test/load.go](internal/test/load.go).

### Performance Video

A visualization of the WebSocket load test running against the dashboard:

<video src="internal/assets/load_testing.mp4" width="100%" height="auto" controls autoplay loop muted></video>

---

## ⚙️ Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP/WS server port |
| `REDIS_ADDR` | `localhost:6379` | Redis connection address |

**Render example:**

```env
REDIS_ADDR=<render-redis-host>:6379
```

---

## 🐳 Running with Docker

```bash
# Build the images
docker compose build

# Start services
docker compose up
```

> Redis and the Go server both start automatically via Docker Compose.

---

## 🧪 Local Development

```bash
# Install dependencies
go mod tidy

# Start the server
go run ./cmd/main.go
```

| Endpoint | URL |
|----------|-----|
| Server | `http://localhost:8080` |
| Metrics | `http://localhost:8080/metrics` |
| WebSocket | `ws://localhost:8080/ws` |

---

## 🔐 Reliability

### Binance Reconnection

If the Binance connection drops:

1. Reconnect automatically
2. Restore the WebSocket connection
3. Re-subscribe to all active symbols
4. Continue streaming — **zero server restarts needed**

### Graceful Shutdown

On `SIGINT` / `SIGTERM`:

1. Stop accepting new requests
2. Cancel background goroutines
3. Close WebSocket connections
4. Close Redis connections
5. Exit cleanly ✅

---

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| 🐹 Language | Go |
| 🔌 WebSockets | Gorilla WebSocket |
| 📈 Market Data | Binance Streams |
| 📨 Message Broker | Redis Pub/Sub |
| ⚡ Concurrency | Goroutines + Channels |
| 🐳 Containerization | Docker |
| ☁️ Deployment | Render |
| 📊 Metrics | Custom HTTP Metrics |

---

<div align="center">

Built to learn and demonstrate **real-world event-driven backend architecture** using Go.

*⭐ Star this repo if you found it helpful!*

</div>
