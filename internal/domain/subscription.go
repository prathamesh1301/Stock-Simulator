package domain

type Subscription struct {
	Type   string `json:"type"`
	Symbol []string `json:"symbols"`
}
