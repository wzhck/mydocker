package network

import (
	"github.com/vishvananda/netlink"
	"net"
)

type Network struct {
	Name       string     `json:"Name"`
	Counts     uint32     `json:"Counts"`
	Driver     string     `json:"Driver"`
	CreateTime string     `json:"CreateTime"`
	IPNet      *net.IPNet `json:"IPNet"`
	Gateway    *net.IPNet `json:"Gateway"`
}

type Endpoint struct {
	// same with the container's uuid
	Uuid     string        `json:"Uuid"`
	IPAddr   net.IP        `json:"IPAddr"`
	Network  *Network      `json:"Network"`
	Device   *netlink.Veth `json:"Device"`
	PortMaps []string      `json:"PortMaps"`
}

type IPAM struct {
	// the path of ip allocator file.
	Allocator string
	// key is subnet's cidr, value is the bitmap of ipaddr.
	SubnetBitMap *map[string]string
}

type Driver interface {
	Name() string
	Init(nw *Network) error
	Create(nw *Network) error
	Delete(nw *Network) error
	Connect(nw *Network, ep *Endpoint) error
	DisConnect(nw *Network, ep *Endpoint) error
}
