package db

import (
	"github.com/go-redis/redis"
)

type App struct {
	client *redis.Client
	prefix string
}

func NewApp(client *redis.Client, prefix string) App {
	return App{
		client: client,
		prefix: prefix,
	}
}

func (a App) Pipe() Pipe {
	return Pipe{
		pipe:   a.client.TxPipeline(),
		prefix: a.prefix,
	}
}

type Pipe struct {
	pipe   redis.Pipeliner
	prefix string
}

func (p Pipe) Exec() {
	p.pipe.Exec()
}
