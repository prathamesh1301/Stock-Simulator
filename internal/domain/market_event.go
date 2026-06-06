package domain

type MarketEvent struct {
	StockName string  `json:"stock_name"`
	Data []byte `json:"data"`
}