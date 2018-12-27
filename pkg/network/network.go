package network

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/util"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func NewNetwork(ctx *cli.Context) (*Network, error) {
	if len(ctx.Args()) < 1 {
		return nil, fmt.Errorf("missing network's name")
	}

	name := ctx.Args().Get(0)
	for nwName := range Networks {
		if nwName == name {
			return nil, fmt.Errorf("the network name %s already exists", name)
		}
	}

	driver := ctx.String("driver")
	if driver == "" {
		return nil, fmt.Errorf("missing --driver option")
	}

	subnet := ctx.String("subnet")
	if subnet == "" {
		return nil, fmt.Errorf("missing --subnet option")
	}

	// e.g. parse "10.20.30.1/24" to "10.20.30.0/24"
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	// set the gateway ip as the first ip addr of the subnet.
	// e.g. set gateway to 10.20.30.1 for subnet 10.20.30.0/24
	gateway := GetIPFromSubnetByIndex(ipNet, 1)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if addr.String() == gateway.String() {
			return nil, fmt.Errorf("the subnet %s already exists", ipNet)
		}
	}

	nw := &Network{
		Name:       name,
		Counts:     0,
		Driver:     driver,
		IPNet:      ipNet,
		Gateway:    gateway,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	Networks[name] = nw
	return nw, nil
}

func Init() error {
	for driverName := range Drivers {
		log.Debugf("init networks of %s driver...", driverName)
		driverDir := path.Join(DriversDir, driverName)
		exist, err := util.FileOrDirExists(driverDir)
		if err != nil {
			return err
		}
		if !exist {
			if err := os.MkdirAll(driverDir, 0755); err != nil {
				return fmt.Errorf("failed to create the dir %s", driverDir)
			}
		}

		if err := filepath.Walk(driverDir+"/",
			func(nwPath string, info os.FileInfo, err error) error {
				if strings.HasSuffix(nwPath, "/") {
					return nil
				}

				_, nwConfigName := path.Split(nwPath)
				nw := &Network{
					Name:   strings.Split(nwConfigName, ".")[0],
					Driver: driverName,
				}

				if err := nw.Load(); err != nil {
					return fmt.Errorf("failed to load the network %s of driver %s: %v",
						nw.Name, driverName, err)
				}
				log.Debugf("detect a %s network: %+v", driverName, nw)

				Networks[nw.Name] = nw
				return Drivers[driverName].Init(nw)
			}); err != nil {
			return err
		}
	}

	return nil
}

func (nw *Network) ConfigFileName() (string, error) {
	configDir := path.Join(DriversDir, nw.Driver)
	configFileName := path.Join(configDir, nw.Name+".json")
	if err := util.EnSureFileExists(configFileName); err != nil {
		return "", err
	}
	return configFileName, nil
}

func (nw *Network) Create() error {
	if err := Drivers[nw.Driver].Create(nw); err != nil {
		return err
	}
	if err := IPAllocator.Init(nw); err != nil {
		return err
	}
	return nw.Dump()
}

func (nw *Network) Delete() error {
	if nw.Counts > 0 {
		return fmt.Errorf("there still exist %d ips in subnet %s",
			nw.Counts, nw.IPNet)
	} else {
		if err := IPAllocator.Init(nw); err != nil {
			return err
		}
		delete(*IPAllocator.SubnetBitMap, nw.IPNet.String())
		if err := IPAllocator.Dump(); err != nil {
			return err
		}
	}

	if err := Drivers[nw.Driver].Delete(nw); err != nil {
		return err
	}

	if configFileName, err := nw.ConfigFileName(); err == nil {
		return os.Remove(configFileName)
	} else {
		return err
	}
}

func (nw *Network) Connect(uuid string, pid int, ipaddr string, portMaps []string) error {
	ep := &Endpoint{
		Uuid:     uuid,
		IPAddr:   net.ParseIP(ipaddr),
		PortMaps: portMaps,
		Network:  nw,
	}

	if err := Drivers[nw.Driver].Connect(nw, ep); err != nil {
		return fmt.Errorf("failed to init veth peers for container %s: %v", uuid, err)
	}

	if err := ep.AddIPAddrAndRoute(pid); err != nil {
		return fmt.Errorf("failed to config ipaddr and route for container %s: %v", uuid, err)
	}

	if err := ep.AddPortMaps(); err != nil {
		return fmt.Errorf("failed to config port maps for container %s: %v", uuid, err)
	}

	return ep.Dump()
}

func (nw *Network) DisConnect(uuid string, pid int) error {
	ep := &Endpoint{Uuid: uuid}
	if err := ep.Load(); err != nil {
		return fmt.Errorf("failed to parse endpoint of container %s: %v", uuid, err)
	}

	if err := ep.DelPortMaps(); err != nil {
		return fmt.Errorf("failed to delete port maps for container %s: %v", uuid, err)
	}

	if err := ep.DelIPAddrAndRoute(pid); err != nil {
		return fmt.Errorf("failed to delete ipaddr and route for container %s: %v", uuid, err)
	}

	if err := Drivers[nw.Driver].DisConnect(nw, ep); err != nil {
		return fmt.Errorf("failed to delete veth peers for container %s: %v", uuid, err)
	}

	if epConfigFileName, err := ep.ConfigFileName(); err == nil {
		return os.Remove(epConfigFileName)
	} else {
		return err
	}
}

func (nw *Network) Dump() error {
	configFileName, err := nw.ConfigFileName()
	if err != nil {
		return err
	}

	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	configFile, err := os.OpenFile(configFileName, int(flags), 0644)
	defer configFile.Close()
	if err != nil {
		return err
	}

	jsonBytes, err := nw.marshal()
	if err != nil {
		return err
	}

	_, err = configFile.Write(jsonBytes)
	return err
}

func (nw *Network) Load() error {
	configFileName, err := nw.ConfigFileName()
	if err != nil {
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

	return nw.unmarshal(jsonBytes[:n])
}

func (nw *Network) marshal() ([]byte, error) {
	type nwAlias Network
	return json.Marshal(&struct {
		IPNet   string `json:"IPNet"`
		Gateway string `json:"Gateway"`
		*nwAlias
	}{
		IPNet:   nw.IPNet.String(),
		Gateway: nw.Gateway.IP.String(),
		nwAlias: (*nwAlias)(nw),
	})
}

func (nw *Network) unmarshal(data []byte) error {
	type nwAlias Network
	aux := &struct {
		IPNet   string `json:"IPNet"`
		Gateway string `json:"Gateway"`
		*nwAlias
	}{
		nwAlias: (*nwAlias)(nw),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	_, ipNet, err := net.ParseCIDR(aux.IPNet)
	if err != nil {
		return err
	}

	nw.IPNet = ipNet
	nw.Gateway = GetIPFromSubnetByIndex(ipNet, 1)

	return nil
}
