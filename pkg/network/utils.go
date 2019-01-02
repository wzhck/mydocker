package network

import (
	"encoding/binary"
	"net"
	"os/exec"
	"strings"
)

func IP2Int(ip net.IP) uint32 {
	if len(ip) == 4 {
		return binary.BigEndian.Uint32(ip)
	}
	return binary.BigEndian.Uint32(ip[12:16])
}

func Int2IP(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func GetIPFromSubnetByIndex(subnet *net.IPNet, index int) *net.IPNet {
	ones, bits := subnet.Mask.Size()
	size := 1 << uint8(bits-ones)
	if index < 0 {
		// for subnet 10.20.30.0/24 and index -1
		// the result should be 10.20.30.255/24
		index = (size + index - 1) % size
	}
	subnetIPInt := IP2Int(subnet.IP)
	return &net.IPNet{
		IP:   Int2IP(subnetIPInt + uint32(index)),
		Mask: subnet.Mask,
	}
}

func GetMasqIPTablesCmd(action, subnet, bridge string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{subnet}", subnet,
		"{bridge}", bridge)
	args := argsReplacer.Replace(iptablesRules["masq"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

func GetDnatIPTablesCmd(action, outPort, inIP, inPort string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{outPort}", outPort,
		"{inIP}", inIP,
		"{inPort}", inPort)
	args := argsReplacer.Replace(iptablesRules["dnat"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

func GetHostIPTablesCmd(action, outIP, outPort, inIP, inPort string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{outIP}", outIP,
		"{outPort}", outPort,
		"{inIP}", inIP,
		"{inPort}", inPort)
	args := argsReplacer.Replace(iptablesRules["host"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}

func GetSnatIPTablesCmd(action, outIP, inIP, inPort string) *exec.Cmd {
	argsReplacer := strings.NewReplacer(
		"{action}", action,
		"{outIP}", outIP,
		"{inIP}", inIP,
		"{inPort}", inPort)
	args := argsReplacer.Replace(iptablesRules["snat"])
	return exec.Command("iptables", strings.Split(args, " ")...)
}
