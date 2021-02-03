package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/PurpleSec/routex"
	"github.com/PurpleSec/routex/val"
)

type item struct {
	ID   uint64
	Name string
	Desc string
}

func (i item) json() string {
	return `{"id": ` + strconv.FormatUint(i.ID, 10) + `, "name": "` + i.Name + `", "desc": "` + i.Desc + `"}`
}

var items = make(map[uint64]*item)

var itemPostVal = val.Set{
	val.Validator{Name: "id", Type: val.Int, Rules: val.ID, Optional: true},
	val.Validator{Name: "name", Type: val.String, Rules: val.Rules{val.Length{Min: 6, Max: 64}}},
	val.Validator{Name: "desc", Type: val.String, Rules: val.Rules{val.Length{Min: 0, Max: 255}}, Optional: true},
}

func main() {
	var (
		h routex.Mux
		s = &http.Server{Addr: "127.0.0.1:8080", Handler: &h}
	)

	if err := h.HandleFuncMethod("item_list", http.MethodGet, "^/item/$", httpItemGetAll); err != nil {
		panic(err)
	}
	if err := h.HandleFuncMethod("item_get", http.MethodGet, "^/item/(?P<item_id>[0-9]+)$", httpItemGet); err != nil {
		panic(err)
	}
	if err := h.HandleFuncMethod("item_post", http.MethodPost, "^/item/(?P<item_id>[0-9]+)$", httpItemPost); err != nil {
		panic(err)
	}

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

func httpItemGet(x context.Context, w http.ResponseWriter, r *routex.Request) {
	n, err := r.Values.Uint64("item_id")
	if err != nil {
		http.Error(w, `{"error": "missing or invalid item id", "type": "invalid request"}`, http.StatusNotFound)
		return
	}
	i, ok := items[n]
	if !ok {
		http.Error(w, `{"error": "unknown item id", "type": "bad request"}`, http.StatusNotFound)
		return
	}
	http.Error(w, i.json(), http.StatusOK)
}
func httpItemPost(x context.Context, w http.ResponseWriter, r *routex.Request) {
	c, err := r.ContentValidate(itemPostVal)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`", "type": "invalid request"}`, http.StatusBadRequest)
		return
	}
	if c == nil {
		http.Error(w, `{"error": "no data", "type": "empty request"}`, http.StatusBadRequest)
		return
	}
	n, err := r.Values.Uint64("item_id")
	if err != nil {
		http.Error(w, `{"error": "missing or invalid item id", "type": "invalid request"}`, http.StatusNotFound)
		return
	}
	i, ok := items[n]
	if !ok {
		i = &item{ID: n}
		items[n] = i
	}
	i.Name, _ = c.String("name")
	if v, err := c.String("desc"); err == nil {
		i.Desc = v
	}
	http.Error(w, i.json(), http.StatusOK)
}
func httpItemGetAll(x context.Context, w http.ResponseWriter, r *routex.Request) {
	var (
		s = `{"items": [`
		c = len(items)
	)
	for _, v := range items {
		s += v.json()
		if c--; c > 0 {
			s += ", "
		}
	}
	http.Error(w, s+"]}", http.StatusOK)
}
