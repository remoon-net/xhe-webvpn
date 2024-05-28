package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os/exec"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/shynome/err0/try"
	"github.com/stretchr/testify/assert"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	link_server "remoon.net/wslink/server"
	"remoon.net/xhe-vpn/config"
	"remoon.net/xhe-vpn/vpn"
	"remoon.net/xhe-vpn/vpn/vtun"
)

var testLinkAddr string

var fileSrvAddr string = "127.0.0.1:61111"

func TestMain(m *testing.M) {
	buildTry()

	{
		srv := link_server.New(0)
		l := try.To1(net.Listen("tcp", "127.0.0.1:2234"))
		defer l.Close()
		go http.Serve(l, srv)
	}

	caddy := exec.Command("caddy", "file-server", "--listen", fileSrvAddr)
	try.To(caddy.Start())
	defer caddy.Process.Kill()

	m.Run()
}

func TestReqBrowserAtServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Config{
		Key:   "003ed5d73b55806c30de3f8a7bdab38af13539220533055e635690b8b87ad641",
		Link:  []string{"http://127.0.0.1:2233"},
		VTun:  true,
		Route: []string{"192.168.4.29/24"},
		Peer: []config.Peer{
			{
				Pubkey: "f928d4f6c1b86c12f2562c10b07c555c5c57fd00f59e90c8d8d88767271cbf7c",
				PSK:    "ba3ef732682972723e233daf6daaa748a6641e4c22b0bc726d94ca03b35055bb",
				Allow:  []string{"192.168.4.28/32"},
			},
		},
	}
	hc := startSrv(ctx, cfg)

	openTry(ctx, fmt.Sprintf("http://%s/testdata/client.html", fileSrvAddr))

	resp := try.To1(hc.Get("http://192.168.4.28:80"))
	assert.Equal(t, resp.StatusCode, 200)
	defer resp.Body.Close()
	body := try.To1(io.ReadAll(resp.Body))
	assert.Equal(t, string(body), "ok")
}

func TestReqBrowserAtClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Config{
		Key:   "003ed5d73b55806c30de3f8a7bdab38af13539220533055e635690b8b87ad641",
		Link:  []string{"http://127.0.0.1:2233"},
		VTun:  true,
		Route: []string{"192.168.4.29/24"},
		Peer: []config.Peer{
			{
				Pubkey: "f928d4f6c1b86c12f2562c10b07c555c5c57fd00f59e90c8d8d88767271cbf7c",
				PSK:    "ba3ef732682972723e233daf6daaa748a6641e4c22b0bc726d94ca03b35055bb",
				WHIP:   []string{"http://f928d4f6c1b86c12f2562c10b07c555c5c57fd00f59e90c8d8d88767271cbf7c@127.0.0.1:2234"},
				Allow:  []string{"192.168.4.28/32"},
			},
		},
	}
	hc := startSrv(ctx, cfg)

	openTry(ctx, fmt.Sprintf("http://%s/testdata/server.html", fileSrvAddr))

	resp := try.To1(hc.Get("http://192.168.4.28:80"))
	assert.Equal(t, resp.StatusCode, 200)
	defer resp.Body.Close()
	body := try.To1(io.ReadAll(resp.Body))
	assert.Equal(t, string(body), "ok")
}

func openTry(ctx context.Context, link string) {
	opts := chromedp.DefaultExecAllocatorOptions[:]
	// opts = append(opts, chromedp.Flag("headless", false))
	ctx, _ = chromedp.NewExecAllocator(ctx, opts...)
	ctx, _ = chromedp.NewContext(ctx)
	tasks := []chromedp.Action{
		chromedp.Navigate(link),
	}
	try.To(chromedp.Run(ctx, tasks...))
}

func startSrv(ctx context.Context, cfg config.Config) *http.Client {
	_, tdev := try.To2(vpn.Connect(ctx, cfg))
	tun := tdev.(vtun.GetStack)
	stk, nic := tun.GetStack(), tun.NIC()
	hc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				ap, err := netip.ParseAddrPort(addr)
				if err != nil {
					return nil, err
				}
				fa, pn := convertToFullAddr(nic, ap)
				return gonet.DialContextTCP(ctx, stk, fa, pn)
			},
		},
	}
	return hc
}
