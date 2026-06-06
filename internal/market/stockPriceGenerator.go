package market

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"stock-sim/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

func StockPriceGenerator(redisClient *redis.Client) {
	symbols := map[string]float64{
		"GOOG": 100.0,
		"AAPL": 150.0,
		"MSFT": 200.0,
	}
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for range ticker.C {
		for symbol, price := range symbols {
			price += rand.Float64()*10 - 5
			eventData := domain.StockData{
				Symbol: symbol,
				Price:  price,
			}
			symbols[symbol] = price
			dataBytes, err := json.Marshal(eventData)
			if err != nil {
				fmt.Println("Error marshaling stock data:", err)
				continue
			}
			err = redisClient.Publish(context.Background(), "stocks", dataBytes).Err()
			if err != nil {
				fmt.Println("Error publishing stock data:", err)
				continue
			}
		}
	}
}