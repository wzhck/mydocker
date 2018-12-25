package network

import (
	"net"
	"testing"
)

func TestIPAM_Allocate(t *testing.T) {
	if err := Init(); err != nil {
		t.Logf("failed to init networks: %v", err)
		return
	}
	if nw, ok := Networks["test-net"]; ok {
		ipAddr, _ := IPAllocator.Allocate(nw)
		t.Logf("allocate a new ip address %s from subnet %s",
			ipAddr, nw.IPNet)
	}
}

func TestIPAM_Release(t *testing.T) {
	ip, _, _ := net.ParseCIDR("10.10.0.2/24")
	if err := Init(); err != nil {
		t.Logf("failed to init networks: %v", err)
		return
	}
	if nw, ok := Networks["test-net"]; ok {
		IPAllocator.Release(nw, &ip)
	}

}
