package db

import (
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	App
	addr []byte
}

func (a App) User(addr []byte) User {
	return User{
		App:  a,
		addr: addr,
	}
}

func (u User) BalanceKey(token string) string {
	return fmt.Sprintf("%s:user:%s:balance:%s", u.prefix, string(u.addr), token)
}

func (u User) MutexKey() string {
	return fmt.Sprintf("%s:user:%s:lock", u.prefix, string(u.addr))
}

func (u User) Lock() (bool, error) {
	return u.client.SetNX(u.MutexKey(), "1", time.Minute).Result()
}

func (u User) Unlock() (bool, error) {
	del, err := u.client.Del(u.MutexKey()).Result()
	return del == 1, err
}

func (u User) GetBalance(token string) (res int64, err error) {
	bz, err := u.client.Get(u.BalanceKey(token)).Result()
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(bz), &res)
	return
}

func (u User) GetBalanceWithSide(base string, quote string, side Side) (res int64, err error) {
	var bz string

	switch side {
	case Bid:
		bz, err = u.client.Get(u.BalanceKey(base)).Result()
	case Ask:
		bz, err = u.client.Get(u.BalanceKey(quote)).Result()
	default:
		return 0, fmt.Errorf("Invalid side")
	}

	err = json.Unmarshal([]byte(bz), &res)
	return
}

func (u User) SetBalance(token string, amt int64) (err error) {
	bz, err := json.Marshal(amt)
	if err != nil {
		return
	}

	err = u.client.Set(u.BalanceKey(token), bz, 0).Err()
	return
}
