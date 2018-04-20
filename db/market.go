package db

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
)

type Market struct {
	App
	base  string
	quote string
}

func (a App) Market(base string, quote string) Market {
	return Market{
		App:   a,
		base:  base,
		quote: quote,
	}
}

func (m Market) InitKey() string {
	return fmt.Sprintf("%s:init", m.prefix)
}

func (m Market) Initialize() (err error) {
	err = m.client.Set(m.OrderLenKey(), "0", 0).Err()
	if err != nil {
		return
	}
	err = m.client.Set(m.InitKey(), "1", 0).Err()
	return
}

func (m Market) Initialized() (res bool, err error) {
	bz, err := m.client.Get(m.InitKey()).Result()
	if err != nil {
		return false, err
	}
	return bz == "1", nil
}

func (m Market) OrderKey(id int64) string {
	return fmt.Sprintf("%s:%s:%s:order:%d", m.prefix, m.base, m.quote, id)
}

func (m Market) GetOrder(id int64) (res Order, err error) {
	bz, err := m.client.Get(m.OrderKey(id)).Result()
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(bz), &res)
	return
}

func (m Market) SetOrder(id int64, side Side, o Order) (err error) {
	bz, err := json.Marshal(o)
	if err != nil {
		return
	}

	err = m.client.Set(m.OrderKey(id), bz, 0).Err()
	if err != nil {
		return
	}

	ok, err := m.client.ZAdd(m.OrderbookKey(side), redis.Z{float64(o.Data.Price), id}).Result()
	if err != nil {
		m.client.Del(m.OrderKey(id))
		return
	}

	if ok != 1 {
		err = fmt.Errorf("Failed to add order")
		m.client.Del(m.OrderKey(id))
		return
	}

	return
}

func (m Market) HasOrder(id int64) (res bool, err error) {
	ok, err := m.client.Exists(m.OrderKey(id)).Result()
	if err != nil {
		return
	}
	res = ok == 1
	return

}

func (m Market) DelOrder(id int64) (res bool, err error) {
	ok, err := m.client.Del(m.OrderKey(id)).Result()
	if err != nil {
		return
	}
	res = ok == 1
	return
}

func (m Market) OrderLenKey() string {
	return fmt.Sprintf("%s:%s:%s:orderlen", m.prefix, m.base, m.quote)
}

func (m Market) GetOrderLen() (res int64, err error) {
	bz, err := m.client.Get(m.OrderLenKey()).Result()
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(bz), &res)
	return
}

func (m Market) IncrOrderLen() (res int64, err error) {
	return m.client.Incr(m.OrderLenKey()).Result()
}

func (m Market) NonceKey(addr []byte) string {
	return fmt.Sprintf("%s:nonce:%s", m.prefix, string(addr))
}

func (m Market) GetNonce(addr []byte) (res int64, err error) {
	bz, err := m.client.Get(m.NonceKey(addr)).Result()
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(bz), &res)
	return
}

func (m Market) OrderbookKey(side Side) string {
	return fmt.Sprintf("%s:%s:%s:orderbook:%s", m.prefix, m.base, m.quote, side)
}

func (m Market) GetOrderbook(side Side, depth int64) (res []Order, err error) {
	var bzz []string

	switch side {
	case Bid:
		bzz, err = m.client.ZRange(m.OrderbookKey(side), depth, -1).Result()
	case Ask:
		bzz, err = m.client.ZRevRange(m.OrderbookKey(side), depth, -1).Result()
	default:
		return nil, fmt.Errorf("Invalid side")
	}

	if err != nil {
		return
	}

	res = make([]Order, len(bzz))
	for i, bz := range bzz {
		var id int64
		err = json.Unmarshal([]byte(bz), &id)
		if err != nil {
			return
		}

		var order Order
		order, err = m.GetOrder(id)
		if err != nil {
			return
		}

		res[i] = order
	}

	return
}
