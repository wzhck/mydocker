package network

import (
	"path"
)

const (
	MyDockerDir    = "/var/lib/mydocker"
	DefaultNetwork = "mydocker0"
	DefaultCIDR    = "10.20.30.0/24"
	MaxBytes       = 2000
)

const (
	Bridge = "bridge"
)

const (
	Delete = "delete"
)

var (
	NetworksDir      = path.Join(MyDockerDir, "networks")
	DriversDir       = path.Join(NetworksDir, "drivers")
	EndpointDir      = path.Join(NetworksDir, "endpoints")
	IPAMDir          = path.Join(NetworksDir, "ipam")
	DefaultAllocator = path.Join(IPAMDir, "subnets.json")
)

var (
	// key is driver's name, value is a Driver implements.
	// should register all network drivers here.
	Drivers = map[string]Driver{
		Bridge: &BridgeDriver{},
	}
	// key is network's name, value is a Network instance.
	Networks = map[string]*Network{}
)

var IPAllocator = &IPAM{
	Allocator:    DefaultAllocator,
	SubnetBitMap: &map[string]string{},
}

var kernelNetConfs = []string{
	"net.ipv4.ip_forward=1",
	"net.ipv4.conf.all.route_localnet=1",
	"net.bridge.bridge-nf-call-iptables=0",
}

var iptablesRules = map[string]string{
	"masq": "-t nat {action} POSTROUTING -s {subnet} ! -o {bridge} -j MASQUERADE",
	"dnat": "-t nat {action} PREROUTING ! -s 127.0.0.1 ! -d 127.0.0.1 -p tcp -m tcp --dport {outPort} -j DNAT --to-destination {inIP}:{inPort}",
	"host": "-t nat {action} OUTPUT -d {outIP} -p tcp -m tcp --dport {outPort} -j DNAT --to-destination {inIP}:{inPort}",
	"snat": "-t nat {action} POSTROUTING -s 127.0.0.1 -d {inIP} -p tcp -m tcp --dport {inPort} -j SNAT --to-source {outIP}",
}
