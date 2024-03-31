package main

import (
	"fmt"
	"io"
	"net/http"
	"syscall/js"

	promise "github.com/nlepage/go-js-promise"
	"github.com/shynome/wahttp"
)

type Hono struct {
	handler js.Value
}

var _ http.Handler = (*Hono)(nil)

func NewHono(handler js.Value) *Hono {
	return &Hono{
		handler: handler,
	}
}

func (s *Hono) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		r := recover()
		switch {
		case err != nil:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		case r != nil:
			http.Error(w, fmt.Sprintf("painc: %v", r), http.StatusInternalServerError)
		}
	}()

	jsReq := wahttp.GoRequest(r)
	jsResp := s.handler.Call("fetch", jsReq)
	jsResp, err = promise.Await(jsResp)
	if err != nil {
		return
	}
	resp, err := wahttp.JsResponse(jsResp)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	h := w.Header()
	for k, vv := range resp.Header {
		for _, v := range vv {
			h.Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
