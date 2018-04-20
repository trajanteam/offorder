package db

type Order struct {
	Data  OrderData `json:"data"`
	Taken int64     `json:"taken"`
}

type Side string

const (
	Bid = "bid"
	Ask = "ask"
)

type OrderData struct {
	Address []byte `json:"address"`
	Base    string `json:"base"`
	Quote   string `json:"quote"`
	Amount  int64  `json:"amount"`
	Price   int64  `json:"price"`
}
