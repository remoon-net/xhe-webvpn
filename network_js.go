package main

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"syscall/js"

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
		ap := try.To1(netip.ParseAddrPort(args[0].String()))
		cfg := args[1]
		mux := NewHono(cfg)
		ctx := signal2ctx(cfg.Get("signal"))
		ctx, cancel := context.WithCancel(ctx)
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
			go http.Serve(l, mux)
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
			go http.Serve(l, mux)
		}

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
