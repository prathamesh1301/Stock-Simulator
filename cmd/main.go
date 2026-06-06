package main

import (
	"fmt"
	"net/http"
	"stock-sim/internal/domain"
	"stock-sim/internal/hub"
	ws "stock-sim/internal/websocket" 
	"github.com/gorilla/websocket"
	"stock-sim/internal/market"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hubS = hub.NewHub()


func wsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		client := &domain.Client{Conn: conn, Send: make(chan []byte, 256), Subscriptions: make(map[string]bool)}
		hubS.Register <- client
		go ws.WritePump(client)
		fmt.Println("client connected")
		err = conn.WriteMessage(websocket.TextMessage, []byte("hello from websocket server"))
		if err != nil {
			fmt.Println("error writing to client", err)
			return
		}
		
		ws.ReadPump(client, hubS)
	}
}



func main() {
	fmt.Println("Hello World")
	go hubS.Run()
	go market.StockPriceGenerator(hubS)
	srv := http.Server{
		Addr:    ":8080",
		Handler: wsHandler(),
	}
	srv.ListenAndServe()
}
