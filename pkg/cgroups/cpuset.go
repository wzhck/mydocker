package cgroups

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
	"regexp"
	"strconv"
)

const (
	cpuset     = "cpuset"
	cpusetCpus = "cpuset.cpus"
	cpusetMems = "cpuset.mems"
)

type CpusetSubsystem struct{}

func (_ *CpusetSubsystem) Name() string {
	return cpuset
}

func (_ *CpusetSubsystem) RootName() string {
	return cpuset
}

func (_ *CpusetSubsystem) Apply(cgPath string, pid int) error {
	return apply(cpuset, cgPath, pid)
}

func (_ *CpusetSubsystem) Remove(cgPath string) error {
	return remove(cpuset, cgPath)
}

func (_ *CpusetSubsystem) Set(cgPath string, r *Resources) error {
	cpusetPath, err := getSubsystemPath(cpuset, cgPath)
	if err != nil {
		return err
	}

	// notes: cpuset.cpus and cpuset.mems must be set before
	// adding container's process id to cpuset cgroup.procs
	for _, dir := range []string{path.Dir(cpusetPath), cpusetPath} {
		if err := setParentCpuset(dir); err != nil {
			return err
		}
	}

	if r.CpusetCpus != "" {
		confFile := path.Join(cpusetPath, cpusetCpus)
		confValue := []byte(r.CpusetCpus)
		log.Debugf("set %s => %s", cpusetCpus, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.CpusetMems != "" {
		confFile := path.Join(cpusetPath, cpusetMems)
		confValue := []byte(r.CpusetMems)
		log.Debugf("set %s => %s", cpusetMems, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}

func setParentCpuset(currentDir string) error {
	parentDir := path.Dir(currentDir)
	for _, conf := range []string{cpusetCpus, cpusetMems} {
		confFile := path.Join(parentDir, conf)
		bytes, err := ioutil.ReadFile(confFile)
		if err != nil {
			return err
		}
		confFile = path.Join(currentDir, conf)
		err = ioutil.WriteFile(confFile, bytes, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// get value from /sys/fs/cgroup/cpuset/cpuset.mems
// or use the command: `numactl --hardware`
// this function ignores errors on purpose.
func getMemNodesNum() int {
	cpusetRoot, _ := getSubsystemMountPoint(cpuset)
	confFile := path.Join(cpusetRoot, cpusetMems)

	valueBytes, _ := ioutil.ReadFile(confFile)
	value := string(valueBytes[:len(valueBytes)-1])

	re, _ := regexp.Compile(`^[\d,-]*(\d+)$`)
	results := re.FindStringSubmatch(value)

	memNodesNum, _ := strconv.Atoi(results[1])
	return memNodesNum + 1
}
