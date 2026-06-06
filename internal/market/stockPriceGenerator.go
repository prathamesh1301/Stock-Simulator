package market

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"stock-sim/internal/domain"
	"stock-sim/internal/hub"
	"time"
)

func StockPriceGenerator(hubS *hub.Hub) {
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
			event := domain.MarketEvent{
				StockName: symbol,
				Data:      dataBytes,
			}
			hubS.Broadcast <- event
		}
	}
}