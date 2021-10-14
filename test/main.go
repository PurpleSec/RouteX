package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PurpleSec/routex"
	"github.com/PurpleSec/routex/val"
)

var jobVal = val.Set{
	val.Validator{Name: "services", Type: val.List, Optional: true},
	val.Validator{Name: "ping_sent", Type: val.Int, Rules: val.Rules{val.Min(0)}},
	val.Validator{Name: "ping_respond", Type: val.Int, Rules: val.Rules{val.Min(0)}},
}

type f bool

func (f) Println(v ...interface{}) {
	fmt.Println(v...)
}

func addHead(_ context.Context, w http.ResponseWriter, _ *routex.Request) bool {
	w.Header().Add("Hello", "World")
	return true
}

func alwaysJSON(_ context.Context, w http.ResponseWriter, _ *routex.Request) bool {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return true
}

func verify(_ context.Context, w http.ResponseWriter, r *routex.Request) bool {
	if r.Values.IntDefault("id", 0) < 10 {
		w.WriteHeader(http.StatusForbidden)
		return false
	}
	return true
}

func main() {
	var (
		m routex.Mux
		s = &http.Server{Addr: "127.0.0.1:8080", Handler: &m}
	)

	m.SetLog(f(true))

	m.Middleware(alwaysJSON)

	m.Must("^/(?P<name>[a-z]+)$", routex.Func(func1), http.MethodGet)
	m.Must("^/(?P<name>[a-z]+)/do$", routex.Func(func1))
	m.Must("^/derp/(?P<id>[0-9]+)$", routex.Func(func2), http.MethodPost).Middleware(verify)

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

func func1(_ context.Context, w http.ResponseWriter, r *routex.Request) {
	w.Write([]byte("hello there!"))
	v, err := r.Values.String("name")
	if err == nil {
		w.Write([]byte(" " + v + "!"))
	}
}

func func2(_ context.Context, w http.ResponseWriter, _ *routex.Request) {

	routex.JSON(w, 200, map[string]string{
		"value1": "1",
		"value2": "2",
		"value3": "3",
		"value4": "4",
	})

}
