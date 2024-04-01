package main

import (
	"net/netip"
	"os"
	"os/exec"

	"github.com/shynome/err0/try"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
)

func buildTry() {
	cmd := exec.Command("npm", "run", "build")
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	try.To(err)
}

func convertToFullAddr(NICID tcpip.NICID, endpoint netip.AddrPort) (tcpip.FullAddress, tcpip.NetworkProtocolNumber) {
	var protoNumber tcpip.NetworkProtocolNumber
	if endpoint.Addr().Is4() {
		protoNumber = ipv4.ProtocolNumber
	} else {
		protoNumber = ipv6.ProtocolNumber
	}
	return tcpip.FullAddress{
		NIC:  NICID,
		Addr: tcpip.AddrFromSlice(endpoint.Addr().AsSlice()),
		Port: endpoint.Port(),
	}, protoNumber
}
