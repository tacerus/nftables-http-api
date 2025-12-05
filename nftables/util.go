package nftables

import (
	"bytes"
	"fmt"
	"net"
	"net/netip"

	"github.com/google/nftables"
)

var (
	NullIPv4 = []byte{0, 0, 0, 0}
	NullIPv6 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

func parseAddrElements(elements []nftables.SetElement) (out []string) {
	// manual merging logic: https://github.com/google/nftables/issues/346
	var last nftables.SetElement

	// iterate in reverse: https://github.com/google/nftables/issues/320
	for i := len(elements) - 1; i >= 0; i-- {
		e := elements[i]

		if bytes.Compare(e.Key, NullIPv4) == 0 || bytes.Compare(e.Key, NullIPv6) == 0 {
			continue
		}

		ipCurrent, _ := netip.AddrFromSlice(e.Key)
		ipLast, _ := netip.AddrFromSlice(last.Key)
		ipLastNext := ipLast.Next()

		if e.IntervalEnd && ipCurrent == ipLastNext {
			out = append(out, ipLast.String())
			continue
		}

		if e.IntervalEnd {
			maxLen := 32
			if ipLastNext.Is6() {
				if !ipCurrent.Is6() {
					continue
				}
				maxLen = 128
			}
			for l := maxLen; l >= 0; l-- {
				mask := net.CIDRMask(l, maxLen)
				na := net.IP(ipLastNext.AsSlice()).Mask(mask)
				n := net.IPNet{IP: na, Mask: mask}
				if n.Contains(net.IP(ipCurrent.AsSlice())) {
					out = append(out, fmt.Sprintf("%s/%d", na, l+1))
					break
				}
			}
			continue
		}

		last = e
	}

	return out
}
