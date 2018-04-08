// https://www.bitmex.com/api/explorer

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"

	"github.com/tendermint/go-amino"

	"github.com/trajanteam/offorder/db"
)

type app struct {
	cdc
	prefix string
	client *redis.Client
}

func RegisterRoute(r *mux.Router, prefix string, client *redis.Client) {
	a := app{prefix, client}

	r.HandleFunc("/order", a.getOrder).Methods(http.MethodGet)       // query order
	r.HandleFunc("/order", a.postOrder).Methods(http.MethodPost)     // make order
	r.HandleFunc("/order", a.putOrder).Methods(http.MethodPut)       // modify order
	r.HandleFunc("/order", a.deleteOrder).Methods(http.MethodDelete) // cancel order

	r.HandleFunc("/orderbook", a.getOrderbook).Methods(http.MethodGet) // query orderbook
}

type Error struct {
	Message string `json:"message"`
}

func responseError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(Error{message}); err != nil {
		panic(err)
	}
}

func response(w http.ResponseWriter, message interface{}) {
	if err := json.NewEncoder(w).Encode(message); err != nil {
		panic(err)
	}
}

func decodeRequest(w http.ResponseWriter, r *http.Request, ptr interface{}) bool {
	if r.Body == nil {
		responseError(w, http.StatusBadRequest, "Request body is nil")
		return false
	}

	if err := json.NewDecoder(r.Body).Decode(ptr); err != nil {
		responseError(w, http.StatusBadRequest, "Cannot decode request")
		return false
	}

	return true
}

func (a app) OrderKey(id uint64) string {
	return fmt.Sprintf("%s/order/%d", a.prefix, id)
}

type GetOrderRequest struct {
	OrderID uint64 `json:"order-id"`
}

func (a app) getOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req GetOrderRequest
	if !decodeRequest(w, r, &req) {
		return
	}

	order, err := a.client.Get(a.OrderKey(req.OrderID)).Result()
	if err != nil {
		responseError(w, http.StatusNotFound, err.Error())
		return
	}

	order, ok := order.(db.Order)
	response(w, order)
}

type PostOrderRequest struct {
	Get    string `json:"get"`
	Give   string `json:"give"`
	Amount uint64 `json:"amount"`
	Price  uint64 `json:"price"`
}

type PostOrderResponse struct {
	OrderID uint64 `json:"order-id"`
}

func (a app) postOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req PostOrderRequest
	if !decodeRequest(w, r, &req) {
		return
	}
}

type PutOrderRequest struct {
	OrderID uint64 `json:"order-id"`
}

type PutOrderResponse struct {
}

func (a app) putOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req PutOrderRequest
	if !decodeRequest(w, r, &req) {
		return
	}
}

type DeleteOrderRequest struct {
	OrderID uint64 `json:"order-id"`
	Sign    []byte `json:"sign"`
}

func (a app) deleteOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req GetOrderRequest
	if !decodeRequest(w, r, &req) {
		return
	}

}

type GetOrderbookRequest struct {
}

type GetOrderbookResponse struct {
}

func (a app) getOrderbook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var req GetOrderbookRequest
	if !decodeRequest(w, r, &req) {
		return
	}

}
