package cgroups

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/util"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	cgroup         = "cgroup"
	cgroupProcs    = "cgroup.procs"
	cgroupInfoFile = "/proc/self/cgroup"
	mountInfoFile  = "/proc/self/mountinfo"
)

// Notes: all subsystemRootName parameter in this file must be one of the following:
// blkio; cpu,cpuacct; cpuset; devices; freezer; hugetlb; memory; net_cls,net_prio; pids

func subsystemIsMounted(subsystemRootName string) bool {
	contentsBytes, err := ioutil.ReadFile(cgroupInfoFile)
	if err != nil {
		return false
	}

	for _, subsystemInfo := range strings.Split(string(contentsBytes), "\n") {
		if strings.Split(subsystemInfo, ":")[1] == subsystemRootName {
			return true
		}
	}

	return false
}

func getSubsystemMountPoint(subsystemRootName string) (string, error) {
	contentsBytes, err := ioutil.ReadFile(mountInfoFile)
	if err != nil {
		return "", err
	}

	for _, mntInfo := range strings.Split(string(contentsBytes), "\n") {
		mntFields := strings.Split(mntInfo, " ")
		if mntFields[8] == cgroup && mntFields[9] == cgroup {
			if strings.HasSuffix(mntFields[4], subsystemRootName) {
				return mntFields[4], nil
			}
		}
	}

	return "", fmt.Errorf("subsystem %s not mounted", subsystemRootName)
}

func getSubsystemPath(subsystemRootName, cgPath string) (string, error) {
	rootMntPoint, err := getSubsystemMountPoint(subsystemRootName)
	if err != nil {
		return "", fmt.Errorf("failed to get root mountpoint of %s: %v",
			subsystemRootName, err)
	}

	subsystemPath := path.Join(rootMntPoint, cgPath)
	// ensure the subsystemPath always exists.
	if err := os.MkdirAll(subsystemPath, 0755); err != nil {
		return "", fmt.Errorf("failed to mkdir %s: %v",
			subsystemPath, err)
	}

	return subsystemPath, nil
}

func apply(subsystemRootName, cgPath string, pid int) error {
	subsystemPath, err := getSubsystemPath(subsystemRootName, cgPath)
	if err != nil {
		return err
	}

	confFile := path.Join(subsystemPath, cgroupProcs)
	confValue := []byte(strconv.Itoa(pid))
	if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
		return fmt.Errorf("failed to add process %d to subsystem %s: %v",
			pid, subsystemRootName, err)
	}

	return nil
}

func remove(subsystemRootName, cgPath string) error {
	subsystemPath, err := getSubsystemPath(subsystemRootName, cgPath)
	if err != nil {
		return err
	}

	cgroupProcsFile := path.Join(subsystemPath, cgroupProcs)
	procsBytes, err := ioutil.ReadFile(cgroupProcsFile)
	if err != nil {
		return err
	}
	if len(procsBytes) != 0 {
		// notes: if the cgroupProcs file of container's cgroup is still NOT empty,
		// which maybe contain zombie process, we must move these processes to the
		// container's parent cgroup of current subsystem before rmdir it.
		log.Debugf("the contents of %s is still NOT empty", cgroupProcsFile)

		processesStr := string(procsBytes[:len(procsBytes)-1])
		parentCgroupProcs := path.Join(path.Dir(subsystemPath), cgroupProcs)

		// NOTE: can only add ONE process to the file cgroup.procs once.
		processArray := strings.Split(processesStr, "\n")
		for i := len(processArray) - 1; i >= 0; i-- {
			process := processArray[i]
			processId, _ := strconv.Atoi(process)
			log.Debugf("killing the orphan process: %d", processId)

			// also need to kill the orphan process if they still exist
			// after the main process was killed in the reversed order.
			util.KillProcess(processId)
			time.Sleep(100 * time.Millisecond)

			// also need to recheck if the process exists after killing it.
			// because KillProcess() send SIGTERM signal first, the process
			// maybe just restart itself when receiving SIGTERM signal.
			processDir := fmt.Sprintf("/proc/%d", processId)
			if exist, _ := util.FileOrDirExists(processDir); !exist {
				continue
			}

			// if failed to kill orphan process, move them to parent cgroup.
			log.Debugf("moving the orphan process %s to its parent cgroup %s",
				process, parentCgroupProcs)
			if err := ioutil.WriteFile(parentCgroupProcs, []byte(process), 0644); err != nil {
				return err
			}
		}
	}

	// notes: we can't delete regular files in a cgroup directory,
	// just need to delete the directory named by container's name
	// ref: http://blog.tinola.com/?e=21
	log.Debugf("removing %s", subsystemPath)
	// `os.RemoveAll(subsystemPath)` doesn't work!
	// `exec.Command("rmdir", subsystemPath).Run()` is also ok.
	return os.Remove(subsystemPath)
}
