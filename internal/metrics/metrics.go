package metrics

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
)

var ConnectedClients int64
var ActiveSymbols int64
var MessagesReceivedTotal int64
var MessagesSentTotal int64
var BinanceReconnectsTotal int64
var (
	TopSymbols   = make(map[string]int64)
	TopSymbolsMu sync.Mutex
)

func IncrementConnectedClients() {
	atomic.AddInt64(&ConnectedClients, 1)
}
func DecrementConnectedClients() {
	atomic.AddInt64(&ConnectedClients, -1)
}

func IncrementMessagesReceivedTotal() {
	atomic.AddInt64(&MessagesReceivedTotal, 1)
}
func IncrementMessagesSentTotal() {
	atomic.AddInt64(&MessagesSentTotal, 1)
}
func IncrementBinanceReconnectsTotal() {
	atomic.AddInt64(&BinanceReconnectsTotal, 1)
}

func IncrementTopSymbols(symbol string) {
	TopSymbolsMu.Lock()
	TopSymbols[symbol]++
	TopSymbolsMu.Unlock()
}

func DecrementTopSymbols(symbol string) {
	TopSymbolsMu.Lock()
	TopSymbols[symbol]--
	if TopSymbols[symbol] == 0 {
		delete(TopSymbols, symbol)
	}
	TopSymbolsMu.Unlock()
}

type SymbolCount struct {
	Symbol string `json:"symbol"`
	Count  int64  `json:"count"`
}

// getTop5Symbols returns the top 5 symbols by count (caller must hold TopSymbolsMu).
func getTop5Symbols() []SymbolCount {
	pairs := make([]SymbolCount, 0, len(TopSymbols))
	for sym, cnt := range TopSymbols {
		pairs = append(pairs, SymbolCount{Symbol: sym, Count: cnt})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})
	if len(pairs) > 5 {
		pairs = pairs[:5]
	}
	return pairs
}

// GetTopSymbols returns the top 5 symbols by subscription count.
func GetTopSymbols() []SymbolCount {
	TopSymbolsMu.Lock()
	defer TopSymbolsMu.Unlock()
	return getTop5Symbols()
}
func GetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	TopSymbolsMu.Lock()
	ActiveSymbols = int64(len(TopSymbols))
	top5 := getTop5Symbols()
	TopSymbolsMu.Unlock()

	snap := map[string]interface{}{
		"connected_clients":        atomic.LoadInt64(&ConnectedClients),
		"active_symbols":           atomic.LoadInt64(&ActiveSymbols),
		"messages_received_total":  atomic.LoadInt64(&MessagesReceivedTotal),
		"messages_sent_total":      atomic.LoadInt64(&MessagesSentTotal),
		"binance_reconnects_total": atomic.LoadInt64(&BinanceReconnectsTotal),
		"top_symbols":              top5,
		"goroutines":               runtime.NumGoroutine(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snap)
}
