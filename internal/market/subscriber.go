package market

import (
	"context"
	"encoding/json"
	"fmt"
	"stock-sim/internal/domain"
	"stock-sim/internal/hub"

	"github.com/redis/go-redis/v9"
)

func StartRedisSubscriber(redisClient *redis.Client,hub *hub.Hub){
	pubSub:=redisClient.Subscribe(context.Background(), "stocks")
	defer pubSub.Close()
	for msg:=range pubSub.Channel(){
		var stockData domain.StockData
		err:=json.Unmarshal([]byte(msg.Payload), &stockData)
		if err!=nil{
			fmt.Println("Error unmarshalling stock data:", err)
			continue
		}
		hub.Broadcast <- domain.MarketEvent{
			StockName: stockData.Symbol,
			Data:      []byte(msg.Payload),
		}
	}
}