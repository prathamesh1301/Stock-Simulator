package market

import (
	"context"
	"encoding/json"
	"fmt"
	"stock-sim/internal/domain"
	"stock-sim/internal/metrics"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"sync"

	"github.com/redis/go-redis/v9"
)

func StartBinanceMarket(ctx context.Context, redisClient *redis.Client, wg *sync.WaitGroup, feedCommands chan domain.FeedCommand) {
	defer wg.Done()
	url := "wss://stream.binance.com:9443/ws"
	activeStocksMap := make(map[string]bool)
	for {
		select {
		case <-ctx.Done():
			return
		default:

		}
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}
		fmt.Println("Binance Connected")
		for k, _ := range activeStocksMap {
			metrics.IncrementTopSymbols(k)
			subscribeMsg := map[string]interface{}{
				"method": "SUBSCRIBE",
				"params": []string{
					strings.ToLower(k) + "@trade",
				},
				"id": 1,
			}
			err = conn.WriteJSON(subscribeMsg)
			if err != nil {
				fmt.Println("Resubscribe failed:", err)
				return
			}

		}
		disconnect := make(chan struct{})

		go func() {
			defer close(disconnect)

			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				_, data, err := conn.ReadMessage()

				if err != nil {
					fmt.Println(err)
					return
				}
				metrics.IncrementMessagesReceivedTotal()

				fmt.Println(string(data))
				var binanceEvent domain.BinanceEvent

				err = json.Unmarshal(data, &binanceEvent)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if binanceEvent.Symbol == "" {
					continue
				}

				stockData := binanceEvent.ToStockData()

				fmt.Println(stockData)
				dataBytes, err := json.Marshal(stockData)
				if err != nil {
					fmt.Println("Error marshaling stock data:", err)
					continue
				}
				err = redisClient.Publish(ctx, "stocks", dataBytes).Err()
				if err != nil {
					fmt.Println("Error publishing stock data:", err)
					continue
				}

			}
		}()

	connectedLoop:
		for {
			select {
			case <-ctx.Done():
				conn.Close()
				return
			case feed := <-feedCommands:
				fmt.Println("Feed command received for stock :", feed.Symbol)
				if feed.Action == "subscribe" {
					
					activeStocksMap[feed.Symbol] = true
					subscribeMsg := map[string]interface{}{
						"method": "SUBSCRIBE",
						"params": []string{
							strings.ToLower(feed.Symbol) + "@trade",
						},
						"id": 1,
					}
					err = conn.WriteJSON(subscribeMsg)
					if err != nil {
						fmt.Println(err)
						conn.Close()
						break connectedLoop
					}
				} else {
					delete(activeStocksMap, feed.Symbol)
					unsubscribeMsg := map[string]interface{}{
						"method": "UNSUBSCRIBE",
						"params": []string{
							strings.ToLower(feed.Symbol) + "@trade",
						},
						"id": time.Now().Unix(),
					}
					err = conn.WriteJSON(unsubscribeMsg)
					if err != nil {
						fmt.Println(err)
						conn.Close()
						break connectedLoop
					}
				}
			case <-disconnect:
				conn.Close()
				select {
				case <-time.After(5 * time.Second):

				case <-ctx.Done():
					return
				}
				for k:= range activeStocksMap {
					metrics.DecrementTopSymbols(k)
				}
				break connectedLoop
			}
		}

	}

}
