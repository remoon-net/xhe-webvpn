package main

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"syscall/js"

	"github.com/docker/go-units"
	"github.com/elazarl/goproxy"
	"github.com/maypok86/otter"
	promise "github.com/nlepage/go-js-promise"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	"golang.zx2c4.com/wireguard/device"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type Network struct {
	stk *stack.Stack
	nic tcpip.NICID
	dev *device.Device
}

func (net *Network) ToJS() js.Value {
	root := js.Global().Get("Object").New()
	root.Set("listen", js.FuncOf(net.Listen))
	root.Set("http_proxy", js.FuncOf(net.HTTPProxy))
	return root
}

type netListener = net.Listener

func (net *Network) Listen(this js.Value, args []js.Value) (p any) {
	p, resolve, reject := promise.New()
	go func() (err error) {
		defer err0.Then(&err, nil, func() {
			reject(err.Error())
		})
		if len(args) != 2 {
			reject("rqeuire listen addr and http server implement {fetch(Request):Response|Promise<Response>}")
			return
		}
		if args[0].Type() != js.TypeString {
			reject("addr is unknown")
			return
		}
		addr, cfg := args[0].String(), args[1]
		mux := NewHono(cfg)
		ctx := signal2ctx(cfg.Get("signal"))
		ctx, cancel := context.WithCancel(ctx)

		net.serveTry(ctx, addr, mux)

		root := js.Global().Get("Object").New()
		root.Set("close", js.FuncOf(func(this js.Value, args []js.Value) any {
			cancel()
			return js.Undefined()
		}))
		resolve(root)
		return nil
	}()
	return p
}

func (net *Network) serveTry(ctx context.Context, addrStr string, handler http.Handler) {
	ap := try.To1(netip.ParseAddrPort(addrStr))
	addr := ap.Addr()
	fa := tcpip.FullAddress{
		NIC:  net.nic,
		Port: ap.Port(),
	}
	if addr.Is6() {
		if !addr.IsUnspecified() {
			fa.Addr = tcpip.AddrFrom16(addr.As16())
		}
		l := try.To1(gonet.ListenTCP(net.stk, fa, ipv6.ProtocolNumber))
		go func() {
			<-ctx.Done()
			l.Close()
		}()
		go http.Serve(l, handler)
	}
	if addr.Is4() {
		if !addr.IsUnspecified() {
			fa.Addr = tcpip.AddrFrom4(addr.As4())
		}
		l := try.To1(gonet.ListenTCP(net.stk, fa, ipv4.ProtocolNumber))
		go func() {
			<-ctx.Done()
			l.Close()
		}()
		go http.Serve(l, handler)
	}
}

func (net *Network) HTTPProxy(this js.Value, args []js.Value) (p any) {
	p, resolve, reject := promise.New()
	go func() (err error) {
		defer err0.Then(&err, nil, func() {
			reject(err.Error())
		})
		if len(args) != 2 {
			reject("rqeuire listen addr and empty config {}")
			return
		}
		if args[0].Type() != js.TypeString {
			reject("addr is unknown")
			return
		}
		addr, cfg := args[0].String(), args[1]
		proxy := goproxy.NewProxyHttpServer()

		cache, err := otter.MustBuilder[string, *tls.Config](1 * units.GiB).Build()
		try.To(err)

		MitmConnectWithCache := &goproxy.ConnectAction{
			Action: goproxy.ConnectMitm,
			TLSConfig: func(host string, ctx *goproxy.ProxyCtx) (*tls.Config, error) {
				cfg, ok := cache.Get(host)
				if ok {
					return cfg, nil
				}
				cfg, err := goproxy.MitmConnect.TLSConfig(host, ctx)
				if err != nil {
					return nil, err
				}
				cache.Set(host, cfg)
				return cfg, err
			},
		}
		proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
			ctx.RoundTripper = goproxy.RoundTripperFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
				return http.DefaultClient.Do(req)
			})
			return MitmConnectWithCache, host
		})
		proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			injectJsFetchOptions(req)
			return req, nil
		})
		ctx := signal2ctx(cfg.Get("signal"))
		ctx, cancel := context.WithCancel(ctx)

		net.serveTry(ctx, addr, proxy)

		root := js.Global().Get("Object").New()
		root.Set("close", js.FuncOf(func(this js.Value, args []js.Value) any {
			cancel()
			return js.Undefined()
		}))
		resolve(root)
		return nil
	}()
	return p
}

const jsFetchOptInPrefix = "Js.fetch."
const jsFetchOptPrefix = "js.fetch:"

func injectJsFetchOptions(r *http.Request) {
	for k, vv := range r.Header {
		if strings.HasPrefix(k, jsFetchOptInPrefix) {
			r.Header.Del(k)
			k = jsFetchOptPrefix + k[len(jsFetchOptInPrefix):]
			r.Header[k] = vv
		}
	}
}
