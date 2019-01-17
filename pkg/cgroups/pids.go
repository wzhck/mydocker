package cgroups

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
)

const (
	pids    = "pids"
	pidsMax = "pids.max"
)

type PidsSubsystem struct{}

func (_ *PidsSubsystem) Name() string {
	return pids
}

func (_ *PidsSubsystem) RootName() string {
	return pids
}

func (_ *PidsSubsystem) Apply(cgPath string, pid int) error {
	return apply(pids, cgPath, pid)
}

func (_ *PidsSubsystem) Remove(cgPath string) error {
	return remove(pids, cgPath)
}

func (_ *PidsSubsystem) Set(cgPath string, r *Resources) error {
	pidsPath, err := getSubsystemPath(pids, cgPath)
	if err != nil {
		return err
	}

	if r.PidsMax != 0 {
		confFile := path.Join(pidsPath, pidsMax)
		confValue := []byte(fmt.Sprintf("%d", r.PidsMax))
		log.Debugf("set %s => %s", pidsMax, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}
