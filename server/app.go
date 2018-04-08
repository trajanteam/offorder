package server

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"

	"github.com/trajanteam/offorder/server/rest"
)

func NewRouter(v uint, opts *redis.Options) *mux.Router {
	client := redis.NewClient(opts)

	r := mux.NewRouter()

	rest.RegisterRoute(r.PathPrefix(fmt.Sprintf("/api/v%d", v)).Subrouter(), client)

	return r
}
