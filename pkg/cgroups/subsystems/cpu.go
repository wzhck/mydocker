package subsystems

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpuSubSystem struct{}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}

func (s *CpuSubSystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CpuPeriod != "0" {
			log.Debugf("setting cpu.cfs_period_us: %s", res.CpuPeriod)
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpu.cfs_period_us"),
				[]byte(res.CpuPeriod), 0644); err != nil {
				return fmt.Errorf("failed to set cpuperiod %v", err)
			}
		}
		if res.CpuQuota != "0" {
			log.Debugf("setting cpu.cfs_quota_us: %s", res.CpuQuota)
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpu.cfs_quota_us"),
				[]byte(res.CpuQuota), 0644); err != nil {
				return fmt.Errorf("failed to set cpuquota %v", err)
			}
		}
		if res.CpuShare != "0" {
			log.Debugf("setting cpu.shares: %s", res.CpuShare)
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpu.shares"),
				[]byte(res.CpuShare), 0644); err != nil {
				return fmt.Errorf("failed to set cpushare %v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

func (s *CpuSubSystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	} else {
		return err
	}
}

func (s *CpuSubSystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("failed to set cgroup tasks for cpu subsystem: %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("failed to get cgroup %s: %v", cgroupPath, err)
	}
}
