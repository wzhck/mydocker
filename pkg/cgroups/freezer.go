package cgroups

import (
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/util"
	"io/ioutil"
	"path"
)

const (
	freezer      = "freezer"
	freezerState = "freezer.state"
)

var freezerStates = []string{"FROZEN", "THAWED"}

type FreezerSubsystem struct{}

func (_ *FreezerSubsystem) Name() string {
	return freezer
}

func (_ *FreezerSubsystem) RootName() string {
	return freezer
}

func (_ *FreezerSubsystem) Apply(cgPath string, pid int) error {
	return apply(freezer, cgPath, pid)
}

func (_ *FreezerSubsystem) Remove(cgPath string) error {
	return remove(freezer, cgPath)
}

func (_ *FreezerSubsystem) Set(cgPath string, r *Resources) error {
	freezerPath, err := getSubsystemPath(freezer, cgPath)
	if err != nil {
		return err
	}

	if util.Contains(freezerStates, r.FreezerState) {
		confFile := path.Join(freezerPath, freezerState)
		confValue := []byte(r.FreezerState)
		log.Debugf("set %s => %s", freezerState, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}
