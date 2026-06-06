package hub

import (
	"fmt"
	"stock-sim/internal/domain"
)

type Hub struct {
	Clients map[*domain.Client]bool

	Register   chan *domain.Client
	Unregister chan *domain.Client

	Broadcast     chan domain.MarketEvent
	Subscriptions map[string]map[*domain.Client]bool
}

func NewHub() *Hub {

	return &Hub{
		Clients: make(map[*domain.Client]bool),

		Register:   make(chan *domain.Client),
		Unregister: make(chan *domain.Client),

		Broadcast:     make(chan domain.MarketEvent),
		Subscriptions: make(map[string]map[*domain.Client]bool),
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

				for symbol := range client.Subscriptions {

					if subscribers, ok := h.Subscriptions[symbol]; ok {

						delete(subscribers, client)

						if len(subscribers) == 0 {
							delete(h.Subscriptions, symbol)
						}
					}
				}

				delete(h.Clients, client)

				close(client.Send)

				client.Conn.Close()

				fmt.Println("Client unregistered")
			}

		case event := <-h.Broadcast:

			subscribers, ok := h.Subscriptions[event.StockName]

			if !ok {
				continue
			}

			for client := range subscribers {

				select {

				case client.Send <- event.Data:

				default:

					h.Unregister <- client
				}
			}
		}
	}
}
