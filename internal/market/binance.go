package market

import (
	"context"
	"encoding/json"
	"fmt"
	"stock-sim/internal/domain"

	"github.com/gorilla/websocket"

	"sync"

	"github.com/redis/go-redis/v9"
)

func StartBinanceMarket(ctx context.Context, redisClient *redis.Client, wg *sync.WaitGroup) {
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
	subscribeMsg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{
			"btcusdt@trade",
			"ethusdt@trade",
		},
		"id": 1,
	}
	err = conn.WriteJSON(subscribeMsg)
	if err != nil {
		fmt.Println(err)
		return
	}
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
