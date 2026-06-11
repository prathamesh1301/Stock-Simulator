package market

import (
	"context"
	"encoding/json"
	"fmt"
	"stock-sim/internal/domain"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"sync"

	"github.com/redis/go-redis/v9"
)

func StartBinanceMarket(ctx context.Context, redisClient *redis.Client, wg *sync.WaitGroup, feedCommands chan domain.FeedCommand) {
	defer wg.Done()
	url := "wss://stream.binance.com:9443/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	fmt.Println("Binance Connected")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case feed := <-feedCommands:
				fmt.Println("Feed command received for stock :", feed.Symbol)
				if feed.Action == "subscribe" {
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
						return
					}
				} else {
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
						return
					}
				}
			}
		}
	}()
	for {

		_, data, err := conn.ReadMessage()

		if err != nil {
			fmt.Println(err)
			return
		}

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

}
