package network

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/util"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func Init() error {
	log.Debugf("initing networks only once before each mydocker command")
	// need to reset the rule of iptables FORWARD chain to ACCEPT, because
	// docker 1.13+ changed the default iptables forwarding policy to DROP
	// https://github.com/moby/moby/pull/28257/files
	// https://github.com/kubernetes/kubernetes/issues/40182
	enableForwardCmd := exec.Command("iptables", "-P", "FORWARD", "ACCEPT")
	if err := enableForwardCmd.Run(); err != nil {
		return fmt.Errorf("failed to execute cmd %s", enableForwardCmd.Args)
	}

	for key, value := range kernelNetConfs {
		confFile := path.Join("/proc/sys", strings.Replace(key, ".", "/", -1))
		contents, err := ioutil.ReadFile(confFile)
		if err != nil {
			return fmt.Errorf("failed to get value of %s: %v", confFile, err)
		}
		if string(contents[:len(contents)-1]) == value {
			continue
		}

		netConf := fmt.Sprintf("%s=%s", key, value)
		log.Debugf("set netconf %s to %s", key, value)
		if err := exec.Command("sysctl", "-w", netConf).Run(); err != nil {
			return fmt.Errorf("failed to set net configuration %s: %v",
				netConf, err)
		}
	}

	for driverName := range Drivers {
		driverDir := path.Join(DriversDir, driverName)
		exist, _ := util.FileOrDirExists(driverDir)
		if !exist {
			if err := os.MkdirAll(driverDir, 0755); err != nil {
				return fmt.Errorf("failed to create the dir %s", driverDir)
			}
		}

		nwConfigs, err := ioutil.ReadDir(driverDir)
		if err != nil {
			return fmt.Errorf("failed to read network config of driver %s: %v",
				driverName, err)
		}

		for _, nwConfig := range nwConfigs {
			nwName := strings.TrimRight(nwConfig.Name(), ".json")
			log.Debugf("found a network %s of driver %s", nwName, driverName)
			nw := &Network{
				Name:   nwName,
				Driver: driverName,
			}

			if err := nw.Load(); err != nil {
				return fmt.Errorf("failed to load network %s of driver %s: %v",
					nwName, driverName, err)
			}

			Networks[nw.Name] = nw
			log.Debugf("initing network %s of driver %s", nwName, driverName)
			if err := Drivers[driverName].Init(nw); err != nil {
				return fmt.Errorf("failed to init network %s of driver %s: %v",
					nwName, driverName, err)
			}
		}
	}

	if _, ok := Networks[DefaultNetwork]; !ok {
		log.Debugf("notes: the default network doesn't exist")
		_, ipNet, _ := net.ParseCIDR(DefaultCIDR)
		defaultNetwork := &Network{
			Name:       DefaultNetwork,
			Counts:     0,
			Driver:     Bridge,
			IPNet:      ipNet,
			Gateway:    GetIPFromSubnetByIndex(ipNet, 1),
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		}

		log.Debugf("create the default network %s", defaultNetwork.Name)
		if err := defaultNetwork.Create(); err != nil {
			return fmt.Errorf("failed to create default network %s: %v",
				DefaultNetwork, err)
		}

		Networks[DefaultNetwork] = defaultNetwork
	}

	jsonBytes, _ := json.MarshalIndent(Networks, "", "    ")
	log.Debugf("found existed networks:\n%s", string(jsonBytes))

	return nil
}
