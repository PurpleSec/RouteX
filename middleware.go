package routex

import (
	"context"
	"net/http"
)

// Middleware is a function alias that can be used to handle and work on a request
// before it is handled to the assigned Mux function (or default function).
//
// Middlewares applied to Routes will be applied AFTER the global Mux Middleware.
//
// The returned boolean can be used to interrupt the call stack before handeling
// back control to implement features such as redirects or authentication.
type Middleware func(context.Context, http.ResponseWriter, *Request) bool

// Middleware adds the supplied Middleware functions to the Mux.
// These are ran before control is passed until the Handler.
//
// And empty function is considered a NOP.
func (m *Mux) Middleware(w ...Middleware) {
	if len(w) == 0 {
		return
	}
	if m.wares == nil {
		m.wares = &wares{w: w}
		return
	}
	m.wares.lock.Lock()
	m.wares.w = append(m.wares.w, w...)
	m.wares.lock.Unlock()
}
func (h *handler) Middleware(w ...Middleware) Route {
	if len(w) == 0 {
		return h
	}
	if h.wares == nil {
		h.wares = &wares{w: w}
		return h
	}
	h.wares.lock.Lock()
	h.wares.w = append(h.wares.w, w...)
	h.wares.lock.Unlock()
	return h
}
