package network

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
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
	cmd := getIPTablesCmd("-D", nw.IPNet.String(), nw.Name)
	if _, err := cmd.Output(); err != nil {
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

func (bd *BridgeDriver) Connect(nw *Network, ep *Endpoint) error {
	return nil
}

func (bd *BridgeDriver) DisConnect(nw *Network, ep *Endpoint) error {
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
		return fmt.Errorf("failed to set ip for the bridge '%s': %v",
			bridgeName, err)
	}

	// step3: enable the new bridge.
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("failed to set the bridge '%s' up: %v",
			bridgeName, err)
	}

	// step4: set iptables for the bridge.
	if err := setIPTables(bridgeName, nw.IPNet); err != nil {
		return fmt.Errorf("failed to set iptables for bridge '%s': %v",
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

	// i.e. `ip link add $br`
	log.Debugf("create the bridge %s using the command: `ip link add %s`",
		bridgeName, bridgeName)
	if err = netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("failed to create bridge %s: %v", bridgeName, err)
	}

	return nil
}

// set the ip addr of a netlink interface
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

	// i.e. `ip addr add $addr dev $iface`
	log.Debugf("set the ip address %s on the bridge %s", ipNet.IP, ifaceName)
	return netlink.AddrAdd(iface, addr)
}

func setInterfaceUP(ifaceName string) error {
	iface, err := netlink.LinkByName(ifaceName)
	if err != nil {
		return fmt.Errorf("failed to get interface '%s': %v", ifaceName, err)
	}

	// i.e. `ip link set $iface up`
	log.Debugf("set the bridge %s up", ifaceName)
	if err = netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("failed to enable interface '%s': %v", ifaceName, err)
	}

	return nil
}

func setIPTables(bridgeName string, subnet *net.IPNet) error {
	var cmd *exec.Cmd

	cmd = getIPTablesCmd("-C", subnet.String(), bridgeName)
	if _, err := cmd.Output(); err == nil {
		return nil
	}

	cmd = getIPTablesCmd("-A", subnet.String(), bridgeName)
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("failed to set iptables: %v", err)
	}

	return nil
}

func getIPTablesCmd(action, src, device string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{src}", src,
		"{out}", device)
	args := argsReplacer.Replace(iptablesRules["masq"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}
