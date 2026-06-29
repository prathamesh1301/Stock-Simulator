package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	const clients = 10000

	var wg sync.WaitGroup

	for i := 0; i < clients; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			conn, _, err := websocket.DefaultDialer.Dial(
				"ws://127.0.0.1:8080/ws", // use IPv4 explicitly — localhost resolves to ::1 on Windows which Docker handles unreliably
				nil,
			)
			if err != nil {
				log.Printf("client %d failed: %v", id, err)
				return
			}
			defer conn.Close()

			// Must respond to server Pings or the server drops us after pongWait (60s)
			conn.SetPongHandler(func(string) error {
				conn.SetReadDeadline(time.Now().Add(70 * time.Second))
				return nil
			})

			subscribe := map[string]interface{}{
				"type":    "subscribe",
				"symbols": []string{"BTCUSDT"},
			}

			if err := conn.WriteJSON(subscribe); err != nil {
				log.Printf("client %d subscribe failed: %v", id, err)
				return
			}

			count := 0
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					log.Printf("client %d disconnected after %d messages: %v", id, count, err)
					return
				}
				count++
				if count%50 == 0 {
					log.Printf("client %d received %d messages", id, count)
				}
			}
		}(i)

		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
}
