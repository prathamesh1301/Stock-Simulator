# 📈 Stock Simulator

A real-time stock price streaming server built with **Go**, **WebSockets**, and **Redis Pub/Sub** — designed to be horizontally scalable from day one.

> Simulates live market data for multiple stock symbols and streams price updates to subscribed WebSocket clients in real time.

---

## 🚀 Features

- ⚡ **Real-time price streaming** — stock prices update every second with simulated market fluctuations
- 🔌 **WebSocket server** — clients connect and subscribe to specific stock symbols
- 📡 **Redis Pub/Sub** — decouples price generation from delivery, enabling multiple server instances
- 🎯 **Per-symbol subscriptions** — clients only receive data for the stocks they care about
- 💓 **Heartbeat / ping-pong** — automatic connection health checks keep stale clients from piling up
- 🧹 **Clean client lifecycle** — clients are auto-unregistered when they disconnect

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Go Server                          │
│                                                         │
│  ┌──────────────────┐      ┌───────────────────────┐   │
│  │ StockPriceGen    │─────▶│  Redis (Pub/Sub)      │   │
│  │ (goroutine)      │      │  channel: "stocks"    │   │
│  └──────────────────┘      └──────────┬────────────┘   │
│                                       │                 │
│                            ┌──────────▼────────────┐   │
│                            │  RedisSubscriber      │   │
│                            │  (goroutine)          │   │
│                            └──────────┬────────────┘   │
│                                       │                 │
│                            ┌──────────▼────────────┐   │
│                            │  Hub                  │   │
│                            │  (event router)       │   │
│                            └──────────┬────────────┘   │
│                                       │                 │
│                   ┌───────────────────┼───────────┐    │
│                   ▼                   ▼           ▼    │
│              [Client A]          [Client B]  [Client N] │
│              GOOG, AAPL          MSFT         GOOG      │
└─────────────────────────────────────────────────────────┘
```

**Flow:**
1. `StockPriceGenerator` ticks every second, simulates price changes, and **publishes** JSON payloads to Redis
2. `StartRedisSubscriber` **subscribes** to Redis and forwards events to the Hub's broadcast channel
3. The `Hub` routes each `MarketEvent` only to clients that have subscribed to that symbol
4. Each client has a `WritePump` goroutine that flushes the send channel over WebSocket, and a `ReadPump` that handles subscription messages and pongs

---

## 📦 Project Structure

```
stock-sim/
├── cmd/
│   └── main.go                  # Entry point — wires everything together
├── internal/
│   ├── domain/
│   │   ├── client.go            # Client struct (WebSocket conn + send channel + subscriptions)
│   │   ├── stock.go             # StockData model
│   │   ├── market_event.go      # MarketEvent (symbol + raw bytes)
│   │   └── subscription.go      # Subscription request from client
│   ├── hub/
│   │   └── hub.go               # Central event router — register, unregister, broadcast
│   ├── market/
│   │   ├── stockPriceGenerator.go  # Generates fake prices, publishes to Redis
│   │   └── subscriber.go           # Consumes Redis channel, feeds the Hub
│   ├── redis/
│   │   └── client.go            # Redis client factory
│   └── websocket/
│       └── pumps.go             # ReadPump & WritePump per client
├── go.mod
└── go.sum
```

---

## 🛠️ Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| WebSockets | [gorilla/websocket](https://github.com/gorilla/websocket) |
| Message Broker | [Redis Pub/Sub](https://redis.io/docs/manual/pubsub/) via [go-redis/v9](https://github.com/redis/go-redis) |
| Concurrency | Goroutines + Channels |

---

## 🧑‍💻 Getting Started

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Redis](https://redis.io/download) running on `localhost:6379`

### Run

```bash
# Clone the repo
git clone https://github.com/prathamesh1301/Stock-Simulator.git
cd Stock-Simulator

# Install dependencies
go mod tidy

# Start the server
go run ./cmd/main.go
```

Server starts on **`:8080`**.

---

## 🔌 WebSocket API

Connect to `ws://localhost:8080/`.

### Subscribe to symbols

Send a JSON message to start receiving price updates:

```json
{
  "type": "subscribe",
  "symbol": ["GOOG", "AAPL"]
}
```

### Unsubscribe

```json
{
  "type": "unsubscribe",
  "symbol": ["AAPL"]
}
```

### Incoming price update (server → client)

```json
{
  "symbol": "GOOG",
  "price": 102.47
}
```

---

## 📊 Simulated Stocks

| Symbol | Starting Price |
|--------|---------------|
| GOOG   | $100.00       |
| AAPL   | $150.00       |
| MSFT   | $200.00       |

Prices drift by a random value in the range `[-5, +5]` each second.

---

## ⚙️ Why Redis Pub/Sub?

Without Redis, price generation and delivery are coupled to a single process — you can't scale horizontally. By publishing to a Redis channel:

- Multiple server instances can run behind a load balancer
- Each instance subscribes to the same channel and fans out to its local clients
- The price generator only needs to run **once** (or be made leader-elected), not once per server

---

