package market

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"stock-sim/internal/domain"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

func StockPriceGenerator(ctx context.Context, redisClient *redis.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	symbols := map[string]float64{
		"GOOG": 100.0,
		"AAPL": 150.0,
		"MSFT": 200.0,
	}
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping stock price generator")
			return
		case <-ticker.C:
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
				err = redisClient.Publish(ctx, "stocks", dataBytes).Err()
				if err != nil {
					fmt.Println("Error publishing stock data:", err)
					continue
				}
			}
		}
	}
}
