package main

import (
	"context"

	promise "github.com/nlepage/go-js-promise"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
	gojs "github.com/shynome/hack-gojs"
	"remoon.net/xhe-vpn/config"
	"remoon.net/xhe-vpn/vpn"
	"remoon.net/xhe-vpn/vpn/vtun"
)

func main() {
	jsVPN := gojs.JSGo.Get("importObject").Get("vpn")
	cfg := jsVPN.Get("config")
	ctx := signal2ctx(cfg.Get("signal"))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	p, resolve, reject := promise.New()
	go func() (err error) {
		defer err0.Then(&err, nil, func() {
			cancel()
			reject(err.Error())
		})

		config := try.To1(getConfig[config.Config](jsVPN.Get("config")))
		dev, tdev := try.To2(vpn.Connect(ctx, config))
		tun := tdev.(vtun.GetStack)
		stk, nic := tun.GetStack(), tun.NIC()
		net := &Network{stk: stk, nic: nic, dev: dev}
		resolve(net.ToJS())
		return
	}()
	jsVPN.Set("connect_result", p)
	<-ctx.Done()
}
