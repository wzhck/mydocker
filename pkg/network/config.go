package network

import (
	"path"
)

const (
	MyDockerDir = "/var/lib/mydocker"
	MaxBytes    = 2000
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

var iptablesRules = map[string]string{
	"masq": "-t nat {action} POSTROUTING -s {src} ! -o {out} -j MASQUERADE",
}
