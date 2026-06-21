package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"stock-sim/internal/domain"
	"stock-sim/internal/hub"
	"stock-sim/internal/metrics"

	gorilla "github.com/gorilla/websocket"
)

const (
	pingPeriod = 54 * time.Second
	pongWait   = 60 * time.Second
)

func ReadPump(client *domain.Client, h *hub.Hub) {

	defer func() {

		h.Unregister <- client
	}()

	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		fmt.Println("pong received")
		return nil
	})

	for {

		_, msg, err := client.Conn.ReadMessage()

		if err != nil {
			break
		}
		var sub domain.Subscription
		err = json.Unmarshal(msg, &sub)
		if err != nil {
			fmt.Println("Error unmarshaling subscription:", err)
			continue
		}
		for _, symbol := range sub.Symbol {
			metrics.IncrementMessagesSentTotal()
			if sub.Type == "subscribe" {
				metrics.IncrementTopSymbols(symbol)
				client.Subscriptions[symbol] = true
				if _, ok := h.Subscriptions[symbol]; !ok {
					h.Subscriptions[symbol] = make(map[*domain.Client]bool)
				}
				h.Subscriptions[symbol][client] = true
				h.SymbolCounts[symbol]++
				if h.SymbolCounts[symbol] == 1 {
					h.FeedCommands <- domain.FeedCommand{
						Symbol: symbol,
						Action: "subscribe",
					}
				}
			} else {
				delete(client.Subscriptions, symbol)
				metrics.DecrementTopSymbols(symbol)
				h.SymbolCounts[symbol]--
				if h.SymbolCounts[symbol] == 0 {
					delete(h.SymbolCounts, symbol)
					h.FeedCommands <- domain.FeedCommand{
						Symbol: symbol,
						Action: "unsubscribe",
					}
				}
				if subscribers, ok := h.Subscriptions[symbol]; ok {

					delete(subscribers, client)

					if len(subscribers) == 0 {
						delete(h.Subscriptions, symbol)
					}
				}
			}
		}
	}
}

func WritePump(client *domain.Client) {

	ticker := time.NewTicker(pingPeriod)

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				// Hub closed the channel; send a close frame and exit.
				client.Conn.WriteMessage(gorilla.CloseMessage, []byte{})
				return
			}
			err := client.Conn.WriteMessage(
				gorilla.TextMessage,
				message,
			)
			if err != nil {
				return
			}
		case <-ticker.C:
			err := client.Conn.WriteMessage(
				gorilla.PingMessage,
				[]byte(""),
			)
			if err != nil {
				fmt.Println("error pinging client", err)
				return
			}
			fmt.Println("ping sent")
		}
	}
}
