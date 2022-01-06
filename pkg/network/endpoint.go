package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/vishvananda/netlink"
	"weike.sh/mydocker/util"
)

func (ep *Endpoint) Connect(pid int) error {
	if err := Drivers[ep.Network.Driver].Connect(ep); err != nil {
		return fmt.Errorf("failed to init veth peers for container: %v", err)
	}

	if err := ep.addIPAddrAndRoute(pid); err != nil {
		return fmt.Errorf("failed to config ipaddr and route for container: %v", err)
	}

	if err := ep.handlePortMaps("create"); err != nil {
		return fmt.Errorf("failed to config port maps for container: %v", err)
	}

	return nil
}

func (ep *Endpoint) DisConnect(pid int) error {
	if err := ep.handlePortMaps("delete"); err != nil {
		return fmt.Errorf("failed to delete port maps for container: %v", err)
	}

	if err := ep.delIPAddrAndRoute(pid); err != nil {
		return fmt.Errorf("failed to delete ipaddr and route for container: %v", err)
	}

	if err := Drivers[ep.Network.Driver].DisConnect(ep); err != nil {
		return fmt.Errorf("failed to delete veth peers for container: %v", err)
	}

	return nil
}

func (ep *Endpoint) addIPAddrAndRoute(pid int) error {
	netnsFileName := fmt.Sprintf("/proc/%d/ns/net", pid)
	netnsFile, err := os.OpenFile(netnsFileName, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("the container is not running")
	} else {
		defer netnsFile.Close()
	}

	// container's nentns fd
	fd := int(netnsFile.Fd())

	containerVeth, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("failed to find container veth %s: %v",
			ep.Device.PeerName, err)
	}

	// move the veth into container netns
	// ip link set cif-<uuid> netns $netns
	if err := netlink.LinkSetNsFd(containerVeth, fd); err != nil {
		return fmt.Errorf("failed to move veth %s into container"+
			"netns: %v", ep.Device.PeerName, err)
	}

	defer EnterContainerNetns(pid)()

	////////////////////////////////////////////////////////////////
	// all the following operations will be executed in the netns //
	// of container (pid) before current function exits finally.  //
	////////////////////////////////////////////////////////////////

	// for the ipaddr 10.20.1.2, which belongs to
	// network 10.20.1.0/24, the containerIP will
	// be 10.20.1.2/24 and set in container.
	ipNet := *ep.Network.IPNet
	ipNet.IP = ep.IPAddr

	// [ip netns exec $netns] ip addr add $addr dev cif-<uuid>
	if err := setInterfaceIP(ep.Device.PeerName, &ipNet); err != nil {
		return fmt.Errorf("failed to set ip for container veth %s: %v",
			ep.Device.PeerName, err)
	}

	// [ip netns exec $netns] ip link set cif-<uuid> up
	for _, ifaceName := range []string{ep.Device.PeerName, "lo"} {
		if err := setInterfaceUP(ifaceName); err != nil {
			return fmt.Errorf("failed to set interface %s up: %v",
				ifaceName, err)
		}
	}

	_, dstNet, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: containerVeth.Attrs().Index,
		Gw:        ep.Network.Gateway.IP,
		Dst:       dstNet,
	}

	// [ip netns exec $netns] ip route add default [dev cif-<uuid>] via $gateway
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		if !strings.Contains(err.Error(), "file exists") {
			// in case the default route has been added before.
			return fmt.Errorf("failed to set default route: %v", err)
		}
	}

	return nil
}

func (ep *Endpoint) delIPAddrAndRoute(pid int) error {
	pidFile := fmt.Sprintf("/proc/%d", pid)
	if exist, _ := util.FileOrDirExists(pidFile); !exist {
		return fmt.Errorf("container (pid: %d) is not running", pid)
	}

	netnsFile, err := os.OpenFile("/proc/1/ns/net", os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to get netns of current process")
	} else {
		defer netnsFile.Close()
	}

	// the host's nentns fd
	fd := int(netnsFile.Fd())

	defer EnterContainerNetns(pid)()

	////////////////////////////////////////////////////////////////
	// all the following operations will be executed in the netns //
	// of container (pid) before current function exits finally.  //
	////////////////////////////////////////////////////////////////

	containerVeth, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("failed to find container veth %s: %v",
			ep.Device.PeerName, err)
	}

	_, dstNet, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: containerVeth.Attrs().Index,
		Gw:        ep.Network.Gateway.IP,
		Dst:       dstNet,
	}

	if err := netlink.RouteDel(defaultRoute); err != nil {
		if !strings.Contains(err.Error(), "no such process") {
			// in case the default route has been removed before.
			return fmt.Errorf("failed to del default route: %v", err)
		}
	}

	// move the veth out of container netns
	// ip link set cif-<uuid> netns $netns
	if err := netlink.LinkSetNsFd(containerVeth, fd); err != nil {
		return fmt.Errorf("failed to move veth %s out of container"+
			"netns: %v", ep.Device.PeerName, err)
	}

	return nil
}

func (ep *Endpoint) handlePortMaps(action string) error {
	var err error
	for out, in := range ep.Ports {
		switch action {
		case "create":
			err = setPortMap(out, ep.IPAddr.String(), in)
		case "delete":
			err = delPortMap(out, ep.IPAddr.String(), in)
		default:
			err = fmt.Errorf("unknown action %s", action)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (ep *Endpoint) MarshalJSON() ([]byte, error) {
	type epAlias Endpoint
	return json.Marshal(&struct {
		IPAddr  string            `json:"IPAddr"`
		Device  string            `json:"Device"`
		Network map[string]string `json:"Network"`
		*epAlias
	}{
		IPAddr: ep.IPAddr.String(),
		Device: ep.Device.Name + "@" + ep.Device.PeerName,
		Network: map[string]string{
			"name":   ep.Network.Name,
			"driver": ep.Network.Driver,
		},
		epAlias: (*epAlias)(ep),
	})
}

func (ep *Endpoint) UnmarshalJSON(data []byte) error {
	type epAlias Endpoint
	aux := &struct {
		IPAddr  string            `json:"IPAddr"`
		Device  string            `json:"Device"`
		Network map[string]string `json:"Network"`
		*epAlias
	}{
		epAlias: (*epAlias)(ep),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	nw := &Network{
		Name:   aux.Network["name"],
		Driver: aux.Network["driver"],
	}
	if err := nw.Load(); err != nil {
		return fmt.Errorf("failed to load network %s: %v",
			nw.Name, err)
	}

	ep.IPAddr = net.ParseIP(aux.IPAddr)
	ep.Network = nw

	vethPeers := strings.Split(aux.Device, "@")

	// this requires the network bridge exists.
	br, err := netlink.LinkByName(nw.Name)
	if err != nil {
		return err
	}

	la := netlink.NewLinkAttrs()
	la.Name = vethPeers[0]
	la.MasterIndex = br.Attrs().Index

	ep.Device = &netlink.Veth{
		LinkAttrs: la,
		PeerName:  vethPeers[1],
	}

	return nil
}
