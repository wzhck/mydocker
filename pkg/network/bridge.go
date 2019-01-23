package network

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"syscall"
)

type BridgeDriver struct{}

func (bd *BridgeDriver) Name() string {
	return Bridge
}

func (bd *BridgeDriver) Init(nw *Network) error {
	return bd.Create(nw)
}

func (bd *BridgeDriver) Create(nw *Network) error {
	log.Debugf("create the network %s of %s driver with iprange %s",
		nw.Name, nw.Driver, nw.IPNet.String())
	if err := bd.initBridge(nw); err != nil {
		return fmt.Errorf("failed to init bridge %s: %v", nw.Name, err)
	}
	return nil
}

func (bd *BridgeDriver) Delete(nw *Network) error {
	if err := delBridgeIptablesRules(nw.Name, nw.IPNet); err != nil {
		return fmt.Errorf("failed to remove iptables: %v", err)
	}

	log.Debugf("delete the network %s of bridge %s with iprange %s",
		nw.Name, nw.Driver, nw.IPNet.String())
	br, err := netlink.LinkByName(nw.Name)
	if err != nil {
		return err
	}

	return netlink.LinkDel(br)
}

func (bd *BridgeDriver) Connect(ep *Endpoint) error {
	// ip link add veth-<uuid> type veth peer name cif-<uuid>
	// ip link set veth-<uuid> master ${bridgeName} or
	// brctl addif ${bridgeName} veth-<uuid>
	if err := netlink.LinkAdd(ep.Device); err != nil {
		return fmt.Errorf("failed to add endpoint device: %v", err)
	}

	// ip link set veth-<uuid> up
	if err := netlink.LinkSetUp(ep.Device); err != nil {
		return fmt.Errorf("failed to set veth device %s up: %v",
			ep.Device.Name, err)
	}

	return nil
}

func (bd *BridgeDriver) DisConnect(ep *Endpoint) error {
	// ip link set veth-<uuid> down
	if err := netlink.LinkSetDown(ep.Device); err != nil {
		return fmt.Errorf("failed to set veth device %s down: %v",
			ep.Device.Name, err)
	}

	// brctl delif ${bridgeName} veth-<uuid>
	if err := netlink.LinkDel(ep.Device); err != nil {
		return fmt.Errorf("failed to del endpoint device: %v", err)
	}

	return nil
}

func (bd *BridgeDriver) initBridge(nw *Network) error {
	bridgeName := nw.Name

	// step1: create a new bridge virtual network device.
	if err := createBridgeInterface(bridgeName); err != nil {
		return err
	}

	// step2: set ip addr and ip route of the bridge.
	if err := setInterfaceIP(bridgeName, nw.Gateway); err != nil {
		return fmt.Errorf("failed to set ip for the bridge %s: %v",
			bridgeName, err)
	}

	// step3: enable the new bridge.
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("failed to set the bridge %s up: %v",
			bridgeName, err)
	}

	// step4: set iptables for the bridge.
	if err := setBridgeIptablesRules(bridgeName, nw.IPNet); err != nil {
		return fmt.Errorf("failed to set iptables for bridge %s: %v",
			bridgeName, err)
	}

	return nil
}

func createBridgeInterface(bridgeName string) error {
	// check if the bridge with name bridgeName exists.
	_, err := net.InterfaceByName(bridgeName)
	// if bridge with name bridgeName exists or throw other errors, just return.
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{}
	br.LinkAttrs = la

	// brctl addbr $bridgeName or
	// ip link add $bridgeName type bridge
	log.Debugf("create a new linux bridge %s: `ip link add %s type bridge`",
		bridgeName, bridgeName)
	if err = netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("failed to create bridge %s: %v", bridgeName, err)
	}

	return nil
}

func setInterfaceIP(ifaceName string, ipNet *net.IPNet) error {
	iface, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get the interface '%s': %v",
			ifaceName, err)
	}

	addrs, err := netlink.AddrList(iface, syscall.AF_INET)
	if err != nil {
		return fmt.Errorf("failed to get ip addrs of the interface '%s': %v",
			ifaceName, err)
	}

	for _, addr := range addrs {
		if addr.IP.String() == ipNet.IP.String() {
			log.Debugf("the ip addr '%s' on the interface '%s' already exists",
				ipNet, ifaceName)
			return nil
		}
	}

	addr := &netlink.Addr{}
	addr.IPNet = ipNet

	// broadcast := GetIPFromSubnetByIndex(ipNet, -1)
	// log.Debugf("get the broadcast %s of subnet %s", broadcast, ipNet)
	// addr.Broadcast = broadcast.IP

	// ip addr add $addr dev $ifaceName
	log.Debugf("set the ip addr %s on the interface %s", ipNet.IP, ifaceName)
	return netlink.AddrAdd(iface, addr)
}

func setInterfaceUP(ifaceName string) error {
	iface, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface '%s': %v", ifaceName, err)
	}

	// ip link set $ifaceName up
	log.Debugf("set the interface %s up", ifaceName)
	if err = netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("failed to enable interface '%s': %v", ifaceName, err)
	}

	return nil
}
