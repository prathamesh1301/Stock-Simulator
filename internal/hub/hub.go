package hub

import (
	"context"
	"fmt"
	"stock-sim/internal/domain"
	"stock-sim/internal/metrics"
	"sync"
)

type Hub struct {
	Clients       map[*domain.Client]bool
	Register      chan *domain.Client
	Unregister    chan *domain.Client
	Broadcast     chan domain.MarketEvent
	Subscriptions map[string]map[*domain.Client]bool
	SymbolCounts  map[string]int
	FeedCommands  chan domain.FeedCommand
}

func NewHub() *Hub {

	return &Hub{
		Clients:       make(map[*domain.Client]bool),
		Register:      make(chan *domain.Client),
		Unregister:    make(chan *domain.Client, 64),
		Broadcast:     make(chan domain.MarketEvent, 256),
		Subscriptions: make(map[string]map[*domain.Client]bool),
		SymbolCounts:  make(map[string]int),
		FeedCommands:  make(chan domain.FeedCommand, 64),
	}
}

func (h *Hub) Run(ctx context.Context, wg *sync.WaitGroup, feedCommands chan domain.FeedCommand) {
	defer wg.Done()
	for {

		select {
		case <-ctx.Done():
			fmt.Println("Stopping hub")
			return

		case client := <-h.Register:
			h.Clients[client] = true
			fmt.Println("Client registered")
			metrics.IncrementConnectedClients()

		case client := <-h.Unregister:

			if _, ok := h.Clients[client]; ok {
				for symbol := range client.Subscriptions {
					if subscribers, ok := h.Subscriptions[symbol]; ok {
						delete(subscribers, client)

						h.SymbolCounts[symbol]--

						metrics.DecrementTopSymbols(symbol)

						if h.SymbolCounts[symbol] == 0 {
							delete(h.SymbolCounts, symbol)

							h.FeedCommands <- domain.FeedCommand{
								Symbol: symbol,
								Action: "unsubscribe",
							}
						}

						if len(subscribers) == 0 {
							delete(h.Subscriptions, symbol)
						}
					}
				}
				metrics.DecrementConnectedClients()
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
		case feed := <-h.FeedCommands:
			fmt.Println("Feed command received in hub for stock :", feed.Symbol)
			feedCommands <- feed
		}
	}
}
