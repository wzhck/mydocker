package cgroups

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type Cgroups struct {
	Pid       int        `json:"Pid"`
	Path      string     `json:"Path"`
	Resources *Resources `json:"Resources"`
}

func (cg *Cgroups) Set() error {
	for _, subsystem := range Subsystems {
		if !subsystemIsMounted(subsystem.RootName()) {
			log.Warnf("subsystem %s is not mounted", subsystem.Name())
			continue
		}
		if err := subsystem.Set(cg.Path, cg.Resources); err != nil {
			return fmt.Errorf("failed to set subsystem %s: %v",
				subsystem.Name(), err)
		}
	}

	return nil
}

func (cg *Cgroups) Apply() error {
	for _, subsystem := range Subsystems {
		if !subsystemIsMounted(subsystem.RootName()) {
			log.Warnf("subsystem %s is not mounted", subsystem.Name())
			continue
		}
		if err := subsystem.Apply(cg.Path, cg.Pid); err != nil {
			return fmt.Errorf("failed to apply container process %d "+
				"to subsystem %s: %v", cg.Pid, subsystem.Name(), err)
		}
	}

	return nil
}

func (cg *Cgroups) Destory() error {
	// sleep for a while is necessary!
	time.Sleep(500 * time.Millisecond)

	for _, subsystem := range Subsystems {
		if !subsystemIsMounted(subsystem.RootName()) {
			log.Warnf("subsystem %s is not mounted", subsystem.Name())
			continue
		}
		if err := subsystem.Remove(cg.Path); err != nil {
			log.Debugf("failed to remove directory of subsystem %s: %v",
				subsystem.Name(), err)
		}
	}

	return nil
}
