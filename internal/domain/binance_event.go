package domain

import "strconv"

type BinanceEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	TradeTime int64  `json:"T"`
	Maker     bool   `json:"M"`
	Price     string `json:"p"`
	Quantity  string `json:"q"`
	TradeID   int64  `json:"t"`
	IsBuyerMarketMaker bool `json:"m"`
}

func (b *BinanceEvent) ToStockData() StockData {
	price, _ := strconv.ParseFloat(b.Price, 64)
	return StockData{
		Symbol: b.Symbol,
		Price:  price,
	}
}