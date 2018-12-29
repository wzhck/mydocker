package network

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/util"
	"net"
	"os"
	"strings"
)

func (ipam *IPAM) Init(nw *Network) error {
	if err := ipam.Load(); err != nil {
		return fmt.Errorf("failed to load IPAllocation info: %v", err)
	}

	// for subnet: 10.10.0.0/24, its mask is 255.255.255.0
	// so 'ones' will be 24 and 'bits' will be 32.
	ones, bits := nw.IPNet.Mask.Size()
	size := 1 << uint8(bits-ones)

	// will init subnet's configurations if ipam
	// allocated none ipaddr within this subnet.
	if _, exist := (*ipam.SubnetBitMap)[nw.IPNet.String()]; exist {
		return nil
	}

	// use "0" to fill the configurations of this subnet.
	// 1<<uint8(bits-ones) means the number of available
	// ip addresses in this subnet.
	// e.g. there are 1<<8 = 256 available ip addresses
	// for the subnet: 10.10.0.0/24
	(*ipam.SubnetBitMap)[nw.IPNet.String()] = strings.Repeat("0", size)
	return ipam.Dump()
}

func (ipam *IPAM) Allocate(nw *Network) (net.IP, error) {
	if err := ipam.Load(); err != nil {
		return nil, fmt.Errorf("failed to load IPAllocation info: %v", err)
	}

	if err := ipam.Init(nw); err != nil {
		return nil, err
	}

	// for subnet: 10.10.0/24, its mask is 255.255.255.0
	// so 'ones' will be 24 and 'bits' will be 32.
	ones, bits := nw.IPNet.Mask.Size()
	size := 1 << uint8(bits-ones)

	bitmapsStr := (*ipam.SubnetBitMap)[nw.IPNet.String()]
	for index, bit := range bitmapsStr {
		// the first ip address is kept for network
		// the second ip address is kept for gateway
		// the last ip address is kept for broadcast
		if index > 1 && index < size-1 && bit == '0' {
			bitmaps := []byte(bitmapsStr)
			bitmaps[index] = '1'
			(*ipam.SubnetBitMap)[nw.IPNet.String()] = string(bitmaps)

			subnetIPInt := IP2Int(nw.IPNet.IP)
			ip := Int2IP(subnetIPInt + uint32(index))
			log.Debugf("allocate a new ip address %s from subnet %s",
				ip, nw.IPNet.String())

			nw.Counts++
			if err := nw.Dump(); err != nil {
				return nil, err
			}

			return ip, ipam.Dump()
		}
	}

	return nil, fmt.Errorf("failed to allocate a new ip address")
}

func (ipam *IPAM) Release(nw *Network, ip *net.IP) error {
	if err := ipam.Load(); err != nil {
		return fmt.Errorf("failed to load IPAllocation info: %v", err)
	}

	if err := ipam.Init(nw); err != nil {
		return err
	}

	if len(*ipam.SubnetBitMap) == 0 {
		return fmt.Errorf("the subnets allocator is empty")
	}

	bitmaps := []byte((*ipam.SubnetBitMap)[nw.IPNet.String()])
	if len(bitmaps) == 0 {
		return fmt.Errorf("the subnet %s has not been initialized",
			nw.IPNet.String())
	}

	subnetIPInt := IP2Int(nw.IPNet.IP)
	releaseIPInt := IP2Int(*ip)
	index := int(releaseIPInt) - int(subnetIPInt)

	log.Debugf("first ip: %s => %d", nw.IPNet.IP, subnetIPInt)
	log.Debugf("release ip: %s => %d", *ip, releaseIPInt)

	if index <= 1 || index >= len(bitmaps) {
		return fmt.Errorf("the ip addr '%s' is out of iprange", ip)
	}

	// in case release same ip addr multiple times.
	if bitmaps[index] == '1' {
		bitmaps[index] = '0'
		(*ipam.SubnetBitMap)[nw.IPNet.String()] = string(bitmaps)

		nw.Counts--
		if err := nw.Dump(); err != nil {
			return err
		}
	}

	return ipam.Dump()
}

func (ipam *IPAM) Dump() error {
	if err := util.EnSureFileExists(ipam.Allocator); err != nil {
		return err
	}

	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	configFile, err := os.OpenFile(ipam.Allocator, int(flags), 0644)
	defer configFile.Close()
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(ipam.SubnetBitMap)
	if err != nil {
		return err
	}

	_, err = configFile.Write(jsonBytes)
	return err
}

func (ipam *IPAM) Load() error {
	if err := util.EnSureFileExists(ipam.Allocator); err != nil {
		return err
	}

	flags := os.O_RDONLY | os.O_CREATE
	configFile, err := os.OpenFile(ipam.Allocator, int(flags), 0644)
	defer configFile.Close()
	if err != nil {
		return err
	}

	jsonBytes := make([]byte, MaxBytes)
	n, err := configFile.Read(jsonBytes)
	if n == 0 {
		return nil
	}
	if err != nil {
		return err
	}

	ipam.SubnetBitMap = &map[string]string{}
	return json.Unmarshal(jsonBytes[:n], ipam.SubnetBitMap)
}
