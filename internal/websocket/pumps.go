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
		if err = json.Unmarshal(msg, &sub); err != nil {
			fmt.Println("Error unmarshaling subscription:", err)
			continue
		}

		for _, symbol := range sub.Symbol {
			metrics.IncrementMessagesSentTotal()
			// Send to hub — hub.Run is the only goroutine that touches the maps.
			h.SubscribeCmds <- domain.SubscribeCmd{
				Client: client,
				Symbol: symbol,
				Action: sub.Type,
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
