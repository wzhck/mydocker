package subsystems

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpusetSubSystem struct{}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}

func (s *CpusetSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CpuSet != "" {
			log.Debugf("setting cpuset.cpus: %s", res.CpuSet)
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"),
				[]byte(res.CpuSet), 0644); err != nil {
				return fmt.Errorf("failed to set cpuset %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpusetSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("failed to set cgroup tasks for cpuset subsystem: %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("failed to get cgroup %s: %v", cgroupPath, err)
	}
}
