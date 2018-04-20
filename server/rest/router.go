// https://www.bitmex.com/api/explorer

package rest

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"

	"github.com/trajanteam/offorder/db"
)

func RegisterRoute(r *mux.Router, app db.App) {
	r.HandleFunc("/order", getOrder(app)).Methods(http.MethodGet)   // query order
	r.HandleFunc("/order", postOrder(app)).Methods(http.MethodPost) // make order
	//	r.HandleFunc("/order", putOrder(app)).Methods(http.MethodPut)       // modify order
	r.HandleFunc("/order", deleteOrder(app)).Methods(http.MethodDelete) // cancel order

	r.HandleFunc("/orderbook", getOrderbook(app)).Methods(http.MethodGet) // query orderbook
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

func verifySignature(target []byte, data interface{}, sig []byte) int {
	bz, err := json.Marshal(data)
	if err != nil {
		return http.StatusInternalServerError
	}

	hash := crypto.Keccak256(bz)
	pubkey, err := secp256k1.RecoverPubkey(hash, sig)
	if err != nil {
		return http.StatusBadRequest
	}

	addr := common.LeftPadBytes(crypto.Keccak256(pubkey[1:])[12:], 32)

	if !bytes.Equal(target, addr) {
		return http.StatusUnauthorized
	}

	return http.StatusOK
}

type Handler func(http.ResponseWriter, *http.Request)

type GetOrderRequest struct {
	Base    string `json:"base"`
	Quote   string `json:"quote"`
	OrderID int64  `json:"order-id"`
}

/*

 */
func getOrder(a db.App) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		var req GetOrderRequest
		if !decodeRequest(w, r, &req) {
			return
		}

		m := a.Market(req.Base, req.Quote)

		order, err := m.GetOrder(req.OrderID)
		if err != nil {
			responseError(w, http.StatusNotFound, err.Error())
			return
		}

		response(w, order)
	}
}

type PostOrderRequest struct {
	db.OrderData
	Side    db.Side `json:"side"`
	Address []byte  `json:"address"`
	Sign    []byte  `json:"sign"`
	Nonce   int64   `json:"nonce"`
}

type PostOrderResponse struct {
	OrderID int64 `json:"order-id"`
}

func postOrder(a db.App) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		var req PostOrderRequest
		if !decodeRequest(w, r, &req) {
			return
		}

		u := a.User(req.Address)
		ok, err := u.Lock()

		if err != nil {
			responseError(w, http.StatusInternalServerError, err.Error())
		}
		if !ok {
			responseError(w, http.StatusConflict, "")
		}

		defer u.Unlock()

		var token string
		var bal int64
		var target int64

		switch req.Side {
		case db.Bid:
			token = req.OrderData.Base
			bal, err = u.GetBalance(token)
			target = req.Amount * req.Price
		case db.Ask:
			token = req.OrderData.Quote
			bal, err = u.GetBalance(token)
			target = req.Amount
		default:
			responseError(w, http.StatusBadRequest, "Invalid side")
			return
		}

		if err != nil {
			responseError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if bal < target {
			responseError(w, http.StatusNotAcceptable, "Not enought balance")
			return
		}

		m := a.Market(req.Base, req.Quote)

		id, err := m.IncrOrderLen()
		if err != nil {
			responseError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// TODO: check current balance

		verifySignature(req.Address, req.OrderData, req.Sign)

		order := db.Order{
			Data:  req.OrderData,
			Taken: 0,
		}

		err = m.SetOrder(id, req.Side, order)
		if err != nil {
			responseError(w, http.StatusInternalServerError, err.Error())
			return
		}

		u.SetBalance(token, bal-target)

		response(w, PostOrderResponse{id})
	}
}

/*
type PutOrderRequest struct {
	db.OrderData
	OrderID int64  `json:"order-id"`
	Address []byte `json:"address"`
	Sign    []byte `json:"sign"`
	Nonce   int64  `json:"nonce"`
}

type PutOrderResponse struct {
}

func putOrder(a db.App) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		var req PutOrderRequest
		if !decodeRequest(w, r, &req) {
			return
		}

		if code := verifySignature(req.Address, req.OrderData, req.Sign); code != http.StatusOK {
			responseError(w, code, "")
			return
		}

		order, err := a.GetOrder(req.OrderID)
		if err != nil {
			responseError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if order.Amount

		if err := a.SetOrder(a.OrderID, order)
	}
}
*/
type DeleteOrderRequest struct {
	Base    string `json:"base"`
	Quote   string `json:"quote"`
	OrderID int64  `json:"order-id"`
	Sign    []byte `json:"sign"`
}

func deleteOrder(a db.App) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		var req DeleteOrderRequest
		if !decodeRequest(w, r, &req) {
			return
		}

		m := a.Market(req.Base, req.Quote)

		ok, err := m.DelOrder(req.OrderID)
		if err != nil {
			responseError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if !ok {
			responseError(w, http.StatusBadRequest, "")
		}
	}
}

type GetOrderbookRequest struct {
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Side  db.Side `json:"side"`
}

type GetOrderbookResponse struct {
}

func getOrderbook(a db.App) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		var req GetOrderbookRequest
		if !decodeRequest(w, r, &req) {
			return
		}
	}
}
