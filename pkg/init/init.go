package init

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/pkg/network"
	"github.com/weikeit/mydocker/util"
	"github.com/x-cray/logrus-prefixed-formatter"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var kernelNetConfs = []string{
	// enable iptables to forward packets between interfaces.
	"net.ipv4.ip_forward=1",
	// consider loopback addresses as normal source or destination while routing.
	"net.ipv4.conf.all.route_localnet=1",
	// enable iptables to hande bridged packets.
	"net.bridge.bridge-nf-call-iptables=1",
}

func init() {
	rand.Seed(time.Now().UnixNano())

	log.SetOutput(os.Stdout)
	log.SetFormatter(&prefixed.TextFormatter{
		ForceColors:     true,
		ForceFormatting: true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// need to reset the rule of iptables FORWARD chain to ACCEPT, because
	// docker 1.13+ changed the default iptables forwarding policy to DROP
	// https://github.com/moby/moby/pull/28257/files
	// https://github.com/kubernetes/kubernetes/issues/40182
	enableForwardCmd := exec.Command("iptables", "-P", "FORWARD", "ACCEPT")
	if err := enableForwardCmd.Run(); err != nil {
		log.Warnf("failed to execute command %s", enableForwardCmd.Args)
	}

	for _, netConf := range kernelNetConfs {
		if err := exec.Command("sysctl", "-w", netConf).Run(); err != nil {
			log.Errorf("failed to set net configuration %s: %v", netConf, err)
		}
	}

	for driverName := range network.Drivers {
		driverDir := path.Join(network.DriversDir, driverName)
		exist, _ := util.FileOrDirExists(driverDir)
		if !exist {
			if err := os.MkdirAll(driverDir, 0755); err != nil {
				log.Errorf("failed to create the dir %s", driverDir)
			}
		}

		nwConfigs, err := ioutil.ReadDir(driverDir)
		if err != nil {
			log.Errorf("failed to read network configs of driver %s: %v",
				driverName, err)
			continue
		}

		for _, nwConfig := range nwConfigs {
			nwName := strings.TrimRight(nwConfig.Name(), ".json")
			nw := &network.Network{
				Name:   nwName,
				Driver: driverName,
			}

			if err := nw.Load(); err != nil {
				log.Errorf("failed to load the network %s of driver %s: %v",
					nwName, driverName, err)
				continue
			}

			network.Networks[nw.Name] = nw
			if err := network.Drivers[driverName].Init(nw); err != nil {
				log.Errorf("failed to init the network %s of driver %s: %v",
					nwName, driverName, err)
			}
		}
	}

	if _, ok := network.Networks[network.DefaultNetwork]; !ok {
		_, ipNet, _ := net.ParseCIDR(network.DefaultCIDR)
		defaultNetwork := &network.Network{
			Name:       network.DefaultNetwork,
			Counts:     0,
			Driver:     network.Bridge,
			IPNet:      ipNet,
			Gateway:    network.GetIPFromSubnetByIndex(ipNet, 1),
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := defaultNetwork.Create(); err != nil {
			log.Errorf("failed to create default network %s: %v",
				network.DefaultNetwork, err)
		}

		network.Networks[network.DefaultNetwork] = defaultNetwork
	}

	jsonBytes, _ := json.MarshalIndent(network.Networks, "", "    ")
	log.Debugf("found existed networks:\n%s\n", string(jsonBytes))
}
