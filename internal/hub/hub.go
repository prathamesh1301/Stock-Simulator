package hub

import (
	"fmt"
 "stock-sim/internal/domain"
)

type Hub struct {
	Clients map[*domain.Client]bool

	Register   chan *domain.Client
	Unregister chan *domain.Client

	Broadcast chan domain.MarketEvent
}

func NewHub() *Hub {

	return &Hub{
		Clients: make(map[*domain.Client]bool),

		Register:   make(chan *domain.Client),
		Unregister: make(chan *domain.Client),

		Broadcast: make(chan domain.MarketEvent),
	}
}

func (h *Hub) Run() {

	for {

		select {

		case client := <-h.Register:

			h.Clients[client] = true

			fmt.Println("Client registered")

		case client := <-h.Unregister:

			if _, ok := h.Clients[client]; ok {

				delete(h.Clients, client)

				close(client.Send)

				client.Conn.Close()

				fmt.Println("Client unregistered")
			}

		case message := <-h.Broadcast:

			for client := range h.Clients {

				if(!client.Subscriptions[message.StockName]) {
					continue
				}
				
				select {

				case client.Send <- message.Data:

				default:

					close(client.Send)

					delete(h.Clients, client)

					client.Conn.Close()
				}
			}
		}
	}
}