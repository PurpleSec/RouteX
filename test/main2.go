package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/PurpleSec/routex/val"

	"github.com/PurpleSec/routex"
)

var jobVal = val.Set{
	val.Validator{Name: "services", Type: val.List, Optional: true},
	val.Validator{Name: "ping_sent", Type: val.Int, Rules: val.Rules{val.Min(0)}},
	val.Validator{Name: "ping_respond", Type: val.Int, Rules: val.Rules{val.Min(0)}},
}

type f bool

func (f) Print(v ...interface{}) {
	fmt.Println(v...)
}

func main() {
	var (
		m routex.Mux
		s = &http.Server{Addr: "127.0.0.1:8080", Handler: &m}
	)

	m.SetLog(f(true))

	m.Error = routex.FuncError(err)

	m.MustMethod("job", http.MethodGet, `^/job$`, routex.WrapEx(nil, m.Error, routex.FuncWrap(job)))
	m.MustMethod("job", http.MethodPost, `^/job/(?P<job_id>[0-9]+)$`, routex.WrapEx(jobVal, m.Error, routex.FuncWrap(job)))

	if err := s.ListenAndServe(); err != nil {
		panic(err)
	}
}

//routex.Wrap(v routex.Validator, h routex.Wrapper)

func err(c int, m string, w http.ResponseWriter, r *routex.Request) {
	w.WriteHeader(c)
	w.Write([]byte(
		`{"error": "` + m + `", "code": ` + strconv.Itoa(c) + `, "req": "` + r.Route + `"}`,
	))
}
func job(x context.Context, w http.ResponseWriter, r *routex.Request, d routex.Content) {
	v, err := r.Values.Int64("job_id")
	if err == nil {
		w.Write([]byte("type: " + strconv.FormatInt(v, 10) + "req"))
	}
}
