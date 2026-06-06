package domain


type StockData struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}
