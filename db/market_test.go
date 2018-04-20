package db_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-redis/redis"

	"github.com/trajanteam/offorder/db"
)

func defaultApp(t *testing.T) db.App {
	c := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	pong, err := c.Ping().Result()
	require.Nil(t, err)
	require.Equal(t, "PONG", pong)

	c.FlushDB()

	return db.NewApp(c, "test-api")
}

type Orders []db.Order

func (os Orders) Len() int {
	return len(os)
}

func (os Orders) Less(i, j int) bool {
	return os[i].Data.Price < os[j].Data.Price
}

func (os Orders) Swap(i, j int) {
	temp := os[i]
	os[i] = os[j]
	os[j] = temp
}

func TestOrder(t *testing.T) {
	a := defaultApp(t)

	base := "base"
	quote := "quote"

	m := a.Market(base, quote)
	/*
		init, err := m.Initialized()
		assert.False(t, init)
		assert.Nil(t, err)

		err = m.Initialize()
		assert.Nil(t, err)

		init, err = m.Initialized()
		assert.True(t, init)
		assert.Nil(t, err)
	*/
	addr := []byte("addr")

	orders := make([]db.Order, 5)
	for i := int64(0); i < 5; i++ {
		id, err := m.IncrOrderLen()
		assert.Nil(t, err)
		assert.Equal(t, i+1, id)

		amt := rand.Int31()
		price := rand.Int31()
		order := db.Order{
			Data: db.OrderData{
				Address: addr,
				Base:    base,
				Quote:   quote,
				Amount:  int64(amt),
				Price:   int64(price),
			},
			Taken: 0,
		}

		err = m.SetOrder(i, db.Bid, order)
		assert.Nil(t, err)

		orders[i] = order
	}

	len, err := m.GetOrderLen()
	assert.Nil(t, err)
	assert.Equal(t, int64(5), len)

	for i := 0; i < 5; i++ {
		order, err := m.GetOrder(int64(i))
		assert.Nil(t, err)
		assert.Equal(t, orders[i], order)
	}

	deleted := make(map[int]bool)
	for i := 0; i < 5; i++ {
		if rand.Int()%2 == 0 {
			res, err := m.DelOrder(int64(i))
			assert.Nil(t, err)
			assert.True(t, res)
			deleted[i] = true
		}
	}

	for i := 0; i < 5; i++ {
		ok, err := m.HasOrder(int64(i))
		assert.Nil(t, err)
		_, del := deleted[i]
		if del {
			assert.False(t, ok)
		} else {
			assert.True(t, ok)
		}
	}
}

func TestOrderbook(t *testing.T) {
	a := defaultApp(t)

	base := "base"
	quote := "quote"

	m := a.Market(base, quote)

	addr := []byte("addr")

	orders := make([]db.Order, 5)
	for i := 0; i < 5; i++ {
		amt := rand.Int31()
		price := rand.Int31()
		order := db.Order{
			Data: db.OrderData{
				Address: addr,
				Base:    base,
				Quote:   quote,
				Amount:  int64(amt),
				Price:   int64(price),
			},
			Taken: 0,
		}

		err := m.SetOrder(int64(i), db.Bid, order)
		assert.Nil(t, err)

		orders[i] = order
	}

	sort.Sort(sort.Reverse(Orders(orders)))

	book, err := m.GetOrderbook(db.Bid, 5)
	assert.Nil(t, err)

	for i, o := range book {
		assert.Equal(t, orders[i], o)
	}
}
