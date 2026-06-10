package market

import (
	"context"
	"encoding/json"
	"fmt"
	"stock-sim/internal/domain"
	"stock-sim/internal/hub"
	"sync"

	"github.com/redis/go-redis/v9"
)

func StartRedisSubscriber(ctx context.Context, redisClient *redis.Client, hub *hub.Hub,wg *sync.WaitGroup) {
	defer wg.Done()
	pubSub := redisClient.Subscribe(ctx, "stocks")
	defer pubSub.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-pubSub.Channel():
			var stockData domain.StockData
			err := json.Unmarshal([]byte(msg.Payload), &stockData)
			if err != nil {
				fmt.Println("Error unmarshalling stock data:", err)
				continue
			}
			fmt.Println("stockData from redis: ", stockData.Symbol, " ", stockData.Price)
			hub.Broadcast <- domain.MarketEvent{
				StockName: stockData.Symbol,
				Data:      []byte(msg.Payload),
			}
		}
	}
}
