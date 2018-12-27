package network

import (
	"encoding/json"
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/weikeit/mydocker/util"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
)

func (ep *Endpoint) ConfigFileName() (string, error) {
	if ep.Uuid == "" {
		return "", fmt.Errorf("endpoint uuid is empty")
	}
	return path.Join(EndpointDir, ep.Uuid+".json"), nil
}

func (ep *Endpoint) Delete() error {
	configFileName, err := ep.ConfigFileName()
	if err != nil {
		return err
	}
	if err := util.EnSureFileExists(configFileName); err != nil {
		return err
	}
	return os.Remove(configFileName)
}

func (ep *Endpoint) AddIPAddrAndRoute(pid int) error {
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

	if err := setInterfaceIP(ep.Device.PeerName, &ipNet); err != nil {
		return fmt.Errorf("failed to set ip for container veth %s: %v",
			ep.Device.PeerName, err)
	}

	// ip link set cif-<uuid> up
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

	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("failed to set default route: %v", err)
	}

	return nil
}

func (ep *Endpoint) DelIPAddrAndRoute(pid int) error {
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
		return fmt.Errorf("failed to del default route: %v", err)
	}

	// move the veth out of container netns
	// ip link set cif-<uuid> netns $netns
	if err := netlink.LinkSetNsFd(containerVeth, fd); err != nil {
		return fmt.Errorf("failed to move veth %s out of container"+
			"netns: %v", ep.Device.PeerName, err)
	}

	return nil
}

func (ep *Endpoint) AddPortMaps() error {
	var cmd *exec.Cmd
	ipaddr := ep.IPAddr.String()

	for _, portMap := range ep.PortMaps {
		port := strings.Split(portMap, ":")
		cmd = GetDnatIPTablesCmd("-C", port[0], ipaddr, port[1])
		if _, err := cmd.Output(); err == nil {
			continue
		}
		cmd = GetDnatIPTablesCmd("-A", port[0], ipaddr, port[1])
		if _, err := cmd.Output(); err != nil {
			return fmt.Errorf("failed to set iptables: %v", err)
		}
	}

	return nil
}

func (ep *Endpoint) DelPortMaps() error {
	var cmd *exec.Cmd
	ipaddr := ep.IPAddr.String()

	for _, portMap := range ep.PortMaps {
		port := strings.Split(portMap, ":")
		cmd = GetDnatIPTablesCmd("-C", port[0], ipaddr, port[1])
		if _, err := cmd.Output(); err != nil {
			continue
		}
		cmd = GetDnatIPTablesCmd("-D", port[0], ipaddr, port[1])
		if _, err := cmd.Output(); err != nil {
			return fmt.Errorf("failed to del iptables: %v", err)
		}
	}

	return nil
}

func (ep *Endpoint) Dump() error {
	configFileName, err := ep.ConfigFileName()
	if err != nil {
		return err
	}
	if err := util.EnSureFileExists(configFileName); err != nil {
		return err
	}

	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	configFile, err := os.OpenFile(configFileName, int(flags), 0644)
	defer configFile.Close()
	if err != nil {
		return err
	}

	jsonBytes, err := ep.marshal()
	if err != nil {
		return err
	}

	_, err = configFile.Write(jsonBytes)
	return err
}

func (ep *Endpoint) Load() error {
	configFileName, err := ep.ConfigFileName()
	if err != nil {
		return err
	}
	if err := util.EnSureFileExists(configFileName); err != nil {
		return err
	}

	flags := os.O_RDONLY | os.O_CREATE
	configFile, err := os.OpenFile(configFileName, int(flags), 0644)
	defer configFile.Close()
	if err != nil {
		return err
	}

	jsonBytes := make([]byte, MaxBytes)
	n, err := configFile.Read(jsonBytes)
	if err != nil {
		return err
	}

	return ep.unmarshal(jsonBytes[:n])
}

func (ep *Endpoint) marshal() ([]byte, error) {
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

func (ep *Endpoint) unmarshal(data []byte) error {
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
		return fmt.Errorf("failed to load network: %v", err)
	}

	ep.IPAddr = net.ParseIP(aux.IPAddr)
	ep.Network = nw

	vethPeers := strings.Split(aux.Device, "@")
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
