package network

import (
	"net"
	"testing"
	"time"
)

func TestNetwork_Create(t *testing.T) {
	_, ipnet, _ := net.ParseCIDR("10.10.0.0/24")
	gateway := GetIPFromSubnetByIndex(ipnet, 1)
	nw := &Network{
		Name:       "test-net",
		Counts:     0,
		Driver:     "bridge",
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Gateway:    gateway,
		IPNet:      ipnet,
	}
	nw.Create()
}

func TestNetwork_Delete(t *testing.T) {
	if err := Init(); err != nil {
		t.Logf("failed to init networks: %v", err)
		return
	}
	if nw, ok := Networks["test-net"]; ok {
		nw.Delete()
	}
}
