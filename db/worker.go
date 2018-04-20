package db

import (
	"github.com/go-redis/redis"
)

type Work interface {
	Run(*redis.Tx) error
	RunKeys() []string
	HasRevert() bool
	Revert(*redis.Tx) error
	RevertKeys() []string
}

type Worker struct {
	client *redis.Client
	Chan   chan Work
}

func (w Worker) Start() {
	go func() {
		for {
			select {
			case work := <-w.Chan:
				err := w.client.Watch(work.Run, work.RunKeys()...)
				if err != nil && work.HasRevert() {
					go w.revert(work)
				}
			}
		}
	}()
}

func (w Worker) revert(work Work) {
	for {
		err := w.client.Watch(work.Revert, work.RevertKeys()...)
		if err == nil {
			break
		}
	}
}
