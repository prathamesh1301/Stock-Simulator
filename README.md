# 📈 Stock Simulator

**A production-style, real-time market data streaming server**

Built with Go · WebSockets · Redis Pub/Sub · Binance Streams · AWS EC2

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![Redis](https://img.shields.io/badge/Redis-Pub%2FSub-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://www.docker.com/)
[![Binance](https://img.shields.io/badge/Binance-WebSocket-F0B90B?style=for-the-badge&logo=binance&logoColor=white)](https://binance.com/)
[![AWS EC2](https://img.shields.io/badge/AWS-EC2-FF9900?style=for-the-badge&logo=amazonaws&logoColor=white)](https://aws.amazon.com/ec2/)

A real-time market data streaming server where WebSocket clients subscribe to cryptocurrency symbols and receive live market updates. The system includes dynamic feed management, automatic reconnection, Redis Pub/Sub, connection monitoring, and backpressure handling.

## 🚀 Live Deployment

The application is deployed on an AWS EC2 instance and exposed through a DuckDNS domain.

| Component | Details |
|---|---|
| Cloud deployment | AWS EC2 |
| Domain | [prathamesh-stocksim.duckdns.org](http://prathamesh-stocksim.duckdns.org) |
| WebSocket endpoint | `ws://prathamesh-stocksim.duckdns.org/ws` |
| Metrics endpoint | `http://prathamesh-stocksim.duckdns.org/metrics` |
| Containerization | Docker Compose |
| Message broker | Redis Pub/Sub |

**Connect to the live server:**

```text
ws://prathamesh-stocksim.duckdns.org/ws
```

> The current endpoint uses `ws://`. Configure TLS and a reverse proxy before switching to `wss://`.

## ✨ Features

### 📡 Real-Time Market Data
- Live cryptocurrency trade data from Binance WebSocket streams
- Dynamic symbol subscriptions and unsubscriptions
- Feed connections activated based on active subscriptions

### 🔌 WebSocket Server
- Concurrent WebSocket clients
- Per-client symbol subscriptions
- Heartbeat ping/pong support
- Stale connection cleanup
- Backpressure handling for slow consumers

### 🔴 Redis Pub/Sub
- Decouples market data ingestion from client delivery
- Supports distributing market events across WebSocket server instances
- Separates exchange feed processing from client-facing delivery

### 🎛️ Dynamic Feed Management
- Tracks active subscriber counts per symbol
- Subscribes to Binance when the first client requests a symbol
- Unsubscribes when the last client disconnects or stops listening
- Avoids unnecessary upstream feed subscriptions

### 🛡️ Reliability
- Automatic Binance WebSocket reconnection
- Re-subscription to active symbols after reconnecting
- Graceful shutdown using SIGINT/SIGTERM
- Context-based cancellation and connection cleanup

### 📊 Observability
- Connected client count
- Active symbol count
- Message counters
- Binance reconnect counter
- Most subscribed symbols

## 🏗️ Architecture

```text
             Binance WebSocket Feed
                       |
                       v
          +---------------------------+
          | Binance Ingestion Service |
          | Dynamic subscriptions     |
          | Reconnection handling     |
          +-------------+-------------+
                        |
                   Redis Publish
                        |
                        v
                Redis Pub/Sub
                        |
                  Redis Subscribe
                        |
                        v
          +---------------------------+
          |     Redis Subscriber     |
          +-------------+-------------+
                        |
                    Broadcast
                        |
                        v
          +---------------------------+
          |            Hub            |
          | Client registration       |
          | Symbol subscription state |
          | Event routing             |
          +-------------+-------------+
                        |
                +-------+-------+
                |               |
                v               v
             Client A        Client B
             BTCUSDT         ETHUSDT
```

## 🔄 Event Flow

### Client Subscription

```text
Client
  |
  v
WebSocket Server
  |
  v
Hub updates symbol subscriber count
  |
  +-- First subscriber? --> Subscribe to Binance feed
  |
  +-- Existing subscribers? --> Reuse active feed
```

### Market Data Flow

```text
Binance
   |
   v
Ingestion Service
   |
   v
Redis Pub/Sub
   |
   v
Redis Subscriber
   |
   v
Hub
   |
   v
Subscribed WebSocket Clients
```

## 🔌 WebSocket API

### Connect

| Environment | URL |
|---|---|
| Local | `ws://localhost:8080/ws` |
| AWS EC2 deployment | `ws://prathamesh-stocksim.duckdns.org/ws` |

### Subscribe

Send a JSON message after connecting:

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

### Market Update

Example server-to-client message:

```json
{
  "symbol": "BTCUSDT",
  "price": 63542.12
}
```

The example values are illustrative; actual prices and message fields depend on the incoming market data.

## 📊 Metrics

The server exposes a metrics endpoint:

```text
GET /metrics
```

Example response:

```json
{
  "active_symbols": 2,
  "binance_reconnects_total": 0,
  "connected_clients": 5,
  "messages_received_total": 1245,
  "messages_sent_total": 9860,
  "top_symbols": [
    {
      "symbol": "BTCUSDT",
      "count": 3
    },
    {
      "symbol": "ETHUSDT",
      "count": 2
    }
  ]
}
```

The values above are illustrative.

## ⚡ Load Testing & Deployment Evidence

I built a custom Go load-testing script to evaluate concurrent WebSocket connections, subscription handling, and message delivery under load.

### AWS EC2 Deployment

The application runs in Docker containers on an AWS EC2 instance, with Redis supporting the market data distribution pipeline.

![Stock Simulator running on AWS EC2](assets/load_testing_video/stock-sim-aws.png)

*AWS EC2 deployment evidence.*

### WebSocket Testing with Postman

The live WebSocket endpoint can be tested using Postman's WebSocket client by connecting to the deployed domain and sending subscription messages.

![Stock Simulator WebSocket testing in Postman](assets/load_testing_video/stock-sim-postman.png)

*WebSocket connection and testing using Postman.*

### Load Testing

The project includes a custom Go load-testing script for simulating concurrent WebSocket clients and observing server behavior.

- Concurrent client connection testing
- Symbol subscription testing
- Heartbeat and connection lifecycle testing
- Message delivery and server metrics monitoring
- Backpressure handling for slow consumers

The project has been tested with up to **10,000 simulated concurrent WebSocket clients** in the load-testing environment.

## ⚙️ Configuration

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP and WebSocket server port |
| `REDIS_ADDR` | `localhost:6379` | Redis connection address |

Configure environment-specific values through your environment or deployment configuration. Do not commit credentials or secret tokens to the repository.

## 🐳 Running with Docker

Build and start the application and Redis:

```bash
docker compose build
docker compose up -d
```

Check running containers:

```bash
docker ps
```

View application logs:

```bash
docker logs -f stock-sim
```

Stop the services:

```bash
docker compose down
```

## 🧪 Local Development

Install dependencies:

```bash
go mod tidy
```

Start the application:

```bash
go run ./cmd/main.go
```

Local endpoints:

| Service | URL |
|---|---|
| Server | `http://localhost:8080` |
| Metrics | `http://localhost:8080/metrics` |
| WebSocket | `ws://localhost:8080/ws` |

## 🔐 Reliability & Connection Management

### Binance Reconnection

When the upstream Binance WebSocket connection drops, the ingestion service attempts to reconnect and restore subscriptions for active symbols.

1. Detect a connection failure.
2. Attempt reconnection.
3. Re-subscribe to active symbols.
4. Resume market data ingestion without requiring a manual server restart.

### Graceful Shutdown

On SIGINT or SIGTERM, the application is designed to:

1. Stop accepting new work.
2. Cancel background goroutines.
3. Close WebSocket connections.
4. Close Redis connections.
5. Exit cleanly.

### Backpressure Handling

Slow WebSocket consumers can cause outbound queues to fill when incoming market data arrives faster than clients can process it. The project explores bounded buffering and ring-buffer-based handling to improve behavior during bursts and avoid indefinite queue growth.

## 🛠️ Tech Stack

| Layer | Technology |
|---|---|
| Language | Go |
| WebSocket server | Gorilla WebSocket |
| Market data | Binance WebSocket Streams |
| Message broker | Redis Pub/Sub |
| Concurrency | Goroutines and channels |
| Burst handling | Ring buffer |
| Containerization | Docker and Docker Compose |
| Cloud deployment | AWS EC2 |
| Domain | DuckDNS |
| Monitoring | Custom HTTP metrics |
| Testing | Postman and custom Go load tester |

---

Built to explore real-world event-driven backend architecture, concurrent data streaming, and reliable market data distribution using Go.

⭐ Star the repository if you find it useful!
