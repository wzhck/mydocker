package cgroups

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/util"
	"io/ioutil"
	"path"
)

const (
	cpu          = "cpu,cpuacct"
	cpuCfsPeriod = "cpu.cfs_period_us"
	cpuCfsQuota  = "cpu.cfs_quota_us"
	cpuRtPeriod  = "cpu.rt_period_us"
	cpuRtRuntime = "cpu.rt_runtime_us"
	cpuShares    = "cpu.shares"
)

type CpuSubsystem struct{}

func (_ *CpuSubsystem) Name() string {
	return "cpu"
}

func (_ *CpuSubsystem) RootName() string {
	return cpu
}

func (_ *CpuSubsystem) Apply(cgPath string, pid int) error {
	return apply(cpu, cgPath, pid)
}

func (_ *CpuSubsystem) Remove(cgPath string) error {
	return remove(cpu, cgPath)
}

func (_ *CpuSubsystem) Set(cgPath string, r *Resources) error {
	cpuPath, err := getSubsystemPath(cpu, cgPath)
	if err != nil {
		return err
	}

	// notes: ubuntu doesn't support cpu.rt_xxx.
	cpuRtPeriodFile := path.Join(cpuPath, cpuRtPeriod)
	if exist, _ := util.FileOrDirExists(cpuRtPeriodFile); !exist {
		log.Debugf("current machine doesn't support cpu.rt_xxx")
		r.CpuRtPeriod, r.CpuRtRuntime = 0, 0
	}

	if r.CpuCfsPeriod > 0 {
		confFile := path.Join(cpuPath, cpuCfsPeriod)
		confValue := []byte(fmt.Sprintf("%d", r.CpuCfsPeriod))
		log.Debugf("set %s => %s", cpuCfsPeriod, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.CpuCfsQuota > 0 {
		confFile := path.Join(cpuPath, cpuCfsQuota)
		confValue := []byte(fmt.Sprintf("%d", r.CpuCfsQuota))
		log.Debugf("set %s => %s", cpuCfsQuota, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.CpuRtPeriod > 0 {
		confFile := path.Join(cpuPath, cpuRtPeriod)
		confValue := []byte(fmt.Sprintf("%d", r.CpuRtPeriod))
		log.Debugf("set %s => %s", cpuRtPeriod, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.CpuRtRuntime > 0 {
		confFile := path.Join(cpuPath, cpuRtRuntime)
		confValue := []byte(fmt.Sprintf("%d", r.CpuRtRuntime))
		log.Debugf("set %s => %s", cpuRtRuntime, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.CpuShares > 0 {
		confFile := path.Join(cpuPath, cpuShares)
		confValue := []byte(fmt.Sprintf("%d", r.CpuShares))
		log.Debugf("set %s => %s", cpuShares, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}
