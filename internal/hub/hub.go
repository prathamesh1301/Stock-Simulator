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
	SubscribeCmds chan domain.SubscribeCmd // ReadPump sends here; only hub.Run touches the maps
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
		SubscribeCmds: make(chan domain.SubscribeCmd, 256),
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

		case cmd := <-h.SubscribeCmds:
			// All map mutations happen here — single goroutine, no race.
			if cmd.Action == "subscribe" {
				cmd.Client.Subscriptions[cmd.Symbol] = true
				if _, ok := h.Subscriptions[cmd.Symbol]; !ok {
					h.Subscriptions[cmd.Symbol] = make(map[*domain.Client]bool)
				}
				h.Subscriptions[cmd.Symbol][cmd.Client] = true
				h.SymbolCounts[cmd.Symbol]++
				if h.SymbolCounts[cmd.Symbol] == 1 {
					feedCommands <- domain.FeedCommand{Symbol: cmd.Symbol, Action: "subscribe"}
				}
			} else {
				delete(cmd.Client.Subscriptions, cmd.Symbol)
				metrics.DecrementTopSymbols(cmd.Symbol)
				h.SymbolCounts[cmd.Symbol]--
				if h.SymbolCounts[cmd.Symbol] == 0 {
					delete(h.SymbolCounts, cmd.Symbol)
					feedCommands <- domain.FeedCommand{Symbol: cmd.Symbol, Action: "unsubscribe"}
				}
				if subscribers, ok := h.Subscriptions[cmd.Symbol]; ok {
					delete(subscribers, cmd.Client)
					if len(subscribers) == 0 {
						delete(h.Subscriptions, cmd.Symbol)
					}
				}
			}

		case feed := <-h.FeedCommands:
			fmt.Println("Feed command received in hub for stock :", feed.Symbol)
			feedCommands <- feed
		}
	}
}
