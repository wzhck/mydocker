package cgroups

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
)

type IfPrioMap struct {
	Interface string `json:"Interface"`
	Priority  uint64 `json:"Priority"`
}

const (
	netprio          = "net_cls,net_prio"
	netprioIfpriomap = "net_prio.ifpriomap"
)

type NetPrioSubsystem struct{}

func (_ *NetPrioSubsystem) Name() string {
	return "net_prio"
}

func (_ *NetPrioSubsystem) RootName() string {
	return netprio
}

func (_ *NetPrioSubsystem) Apply(cgPath string, pid int) error {
	return apply(netprio, cgPath, pid)
}

func (_ *NetPrioSubsystem) Remove(cgPath string) error {
	return remove(netprio, cgPath)
}

func (_ *NetPrioSubsystem) Set(cgPath string, r *Resources) error {
	netprioPath, err := getSubsystemPath(netprio, cgPath)
	if err != nil {
		return err
	}

	for _, ifpriomap := range r.NetPrioIfpriomap {
		confFile := path.Join(netprioPath, netprioIfpriomap)
		confValue := []byte(fmt.Sprintf("%s %d", ifpriomap.Interface, ifpriomap.Priority))
		logrus.Debugf("set %s => %s", netprioIfpriomap, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}
