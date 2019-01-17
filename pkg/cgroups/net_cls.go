package cgroups

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
)

const (
	netcls        = "net_cls,net_prio"
	netclsClassid = "net_cls.classid"
)

type NetClsSubsystem struct{}

func (_ *NetClsSubsystem) Name() string {
	return "net_cls"
}

func (_ *NetClsSubsystem) RootName() string {
	return netcls
}

func (_ *NetClsSubsystem) Apply(cgPath string, pid int) error {
	return apply(netcls, cgPath, pid)
}

func (_ *NetClsSubsystem) Remove(cgPath string) error {
	return remove(netcls, cgPath)
}

func (_ *NetClsSubsystem) Set(cgPath string, r *Resources) error {
	netclsPath, err := getSubsystemPath(netcls, cgPath)
	if err != nil {
		return err
	}

	if r.NetClsClassid != 0 {
		confFile := path.Join(netclsPath, netclsClassid)
		confValue := []byte(fmt.Sprintf("%d", r.NetClsClassid))
		log.Debugf("set %s => %s", netclsClassid, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}
