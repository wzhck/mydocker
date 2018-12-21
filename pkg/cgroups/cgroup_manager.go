package cgroups

import (
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/pkg/cgroups/subsystems"
)

type CgroupManager struct {
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (c *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range subsystems.Ins {
		subSysIns.Apply(c.Path, pid)
	}
	return nil
}

func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subSysIns := range subsystems.Ins {
		subSysIns.Set(c.Path, res)
	}
	return nil
}

func (c *CgroupManager) Destroy() error {
	for _, subSysIns := range subsystems.Ins {
		if err := subSysIns.Remove(c.Path); err != nil {
			log.Warnf("failed to remove cgroup: %v", err)
		}
	}
	return nil
}
