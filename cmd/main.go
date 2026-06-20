package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"stock-sim/internal/domain"
	"stock-sim/internal/hub"
	"stock-sim/internal/market"
	"stock-sim/internal/metrics"
	"stock-sim/internal/redis"
	ws "stock-sim/internal/websocket"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hubS = hub.NewHub()

func wsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade to websockets failed ", err)
			fmt.Println(err)
			return
		}

		client := &domain.Client{Conn: conn, Send: make(chan []byte, 256), Subscriptions: make(map[string]bool)}
		hubS.Register <- client
		go ws.WritePump(client)
		fmt.Println("client connected")
		err = conn.WriteMessage(websocket.TextMessage, []byte("hello from websocket server"))
		if err != nil {
			fmt.Println("error writing to client", err)
			return
		}

		ws.ReadPump(client, hubS)
	}
}

func main() {
	fmt.Println("Hello World")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	feedCommands := make(chan domain.FeedCommand)
	wg.Add(1)
	go hubS.Run(ctx, &wg, feedCommands)
	redisClient := redis.NewClient()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("REDIS CONNECTION FAILED: %v", err)
	}

	wg.Add(1)
	// go market.StockPriceGenerator(ctx, redisClient, &wg)
	go market.StartBinanceMarket(ctx, redisClient, &wg, feedCommands)
	wg.Add(1)
	go market.StartRedisSubscriber(ctx, redisClient, hubS, &wg)
	mux := http.NewServeMux()

	// Health check / homepage
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Stock Simulator Running"))
	})

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsHandler())

	// Metrics endpoint
	mux.HandleFunc("/metrics", metrics.GetCurrentMetrics)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("PORT =", port)

	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		fmt.Println("HTTP server listening on", port)

		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			log.Fatal(err)
		}
	}()

	// Signal handling
	sigChan := make(chan os.Signal, 1)

	signal.Notify(
		sigChan,
		os.Interrupt,
		syscall.SIGTERM,
	)

	sig := <-sigChan

	fmt.Println("Received signal:", sig)

	// Trigger cancellation
	cancel()

	// Shutdown HTTP server gracefully
	shutdownCtx, shutdownCancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Println("server shutdown error:", err)
	}

	fmt.Println("Waiting for goroutines to exit...")
	wg.Wait()

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		fmt.Println("redis close error:", err)
	}

	fmt.Println("Shutdown complete")

}
