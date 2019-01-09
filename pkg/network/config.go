package network

import (
	"path"
)

const (
	MyDockerDir    = "/var/lib/mydocker"
	SysClassNet    = "/sys/class/net"
	DefaultNetwork = "mydocker0"
	DefaultCIDR    = "10.20.30.0/24"
)

const (
	Bridge = "bridge"
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

var bridgeIPTRules = map[string]string{
	"masq": "-w 5 -t nat {action} POSTROUTING -s {subnet} ! -o {bridge} -j MASQUERADE",
	"mark": "-w 5 -t mangle {action} PREROUTING -i {bridge} -j MARK --set-mark {mark}",
	"phys": "-w 5 -t mangle {action} POSTROUTING -o {physnic} -m mark --mark {mark} -j ACCEPT",
	"drop": "-w 5 -t mangle {action} POSTROUTING ! -o {bridge} -m mark --mark {mark} -j DROP",
}

var portMapsIPTRules = map[string]string{
	"dnat": "-w 5 -t nat {action} PREROUTING ! -s 127.0.0.1 ! -d 127.0.0.1 -p tcp -m tcp --dport {outPort} -j DNAT --to-destination {inIP}:{inPort}",
	"host": "-w 5 -t nat {action} OUTPUT -d {outIP} -p tcp -m tcp --dport {outPort} -j DNAT --to-destination {inIP}:{inPort}",
	"snat": "-w 5 -t nat {action} POSTROUTING -s 127.0.0.1 -d {inIP} -p tcp -m tcp --dport {inPort} -j SNAT --to-source {outIP}",
}
