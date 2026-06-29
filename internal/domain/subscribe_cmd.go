package domain

// SubscribeCmd is sent from ReadPump to the Hub to request a
// subscribe/unsubscribe action on behalf of a client.
// This avoids ReadPump directly mutating hub maps from multiple goroutines.
type SubscribeCmd struct {
	Client *Client
	Symbol string
	Action string // "subscribe" | "unsubscribe"
}
