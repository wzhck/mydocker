package network

import (
	"encoding/binary"
	"fmt"
	"github.com/weikeit/mydocker/util"
	"io/ioutil"
	"net"
	"os"
	"path"
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

func GetPhysicalNics() ([]string, error) {
	var physNics []string

	nics, err := ioutil.ReadDir(SysClassNet)
	if err != nil {
		return nil, err
	}

	for _, nic := range nics {
		nicName := nic.Name()
		nicPath := path.Join(SysClassNet, nicName)
		dest, _ := os.Readlink(nicPath)
		if !strings.Contains(dest, "/devices/virtual/net/") {
			physNics = append(physNics, nicName)
		}
	}

	if len(physNics) == 0 {
		return nil, fmt.Errorf("no physical nics")
	}

	return physNics, nil
}

func GetPhysicalIPs() ([]string, error) {
	var physIPs []string

	physNics, err := GetPhysicalNics()
	if err != nil {
		return nil, err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if !util.Contains(physNics, iface.Name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if strings.Contains(ip.String(), ":") {
				continue
			}
			physIPs = append(physIPs, ip.String())
		}
	}

	if len(physIPs) == 0 {
		return nil, fmt.Errorf("no physical ips")
	}

	return physIPs, nil
}
