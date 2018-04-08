package db

type Order struct {
	Address string `json:"address"`
	Get     string `json:"get"`
	Give    string `json:"give"`
	Amount  uint64 `json:"amount"`
	Taken   uint64 `json:"taken"`
	Price   uint64 `json:"price"`
}
