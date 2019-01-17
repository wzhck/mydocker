package cgroups

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/util"
	"io/ioutil"
	"path"
	"strconv"
)

const (
	memory               = "memory"
	memoryLimit          = "memory.limit_in_bytes"
	memorySoftLimit      = "memory.soft_limit_in_bytes"
	memorySwapLimit      = "memory.memsw.limit_in_bytes"
	memorySwappiness     = "memory.swappiness"
	memoryOomControl     = "memory.oom_control"
	kernelMemoryLimit    = "memory.kmem.limit_in_bytes"
	kernelMemoryTCPLimit = "memory.kmem.tcp.limit_in_bytes"
)

type MemorySubsystem struct{}

func (_ *MemorySubsystem) Name() string {
	return memory
}

func (_ *MemorySubsystem) RootName() string {
	return memory
}

func (_ *MemorySubsystem) Apply(cgPath string, pid int) error {
	return apply(memory, cgPath, pid)
}

func (_ *MemorySubsystem) Remove(cgPath string) error {
	return remove(memory, cgPath)
}

func (_ *MemorySubsystem) Set(cgPath string, r *Resources) error {
	// NOTE: we can write "-1" to reset the *.limit_in_bytes(unlimited).
	// ref: https://unix.stackexchange.com/a/420911/264900
	// actually, kernel will set 9223372036854771712 (0x7FFFFFFFFFFFF000)
	// this value is the largest positive signed 64-bit integer (2^63-1),
	// rounded down to multiples of 4096 (2^12), the most common pagesize
	// i.e., 9223372036854771712 == (2 ^ 63 - 1) / 4096 * 4096

	memoryPath, err := getSubsystemPath(memory, cgPath)
	if err != nil {
		return err
	}

	// notes: ubuntu doesn't support this configuration.
	memswLimitFile := path.Join(memoryPath, memorySwapLimit)
	if exist, _ := util.FileOrDirExists(memswLimitFile); exist {
		// if the memoryLimit is set to `-1`, we should also
		// set swap to -1, which means unlimited swap memory
		if r.MemoryLimit == -1 {
			r.MemorySwapLimit = -1
		}
	} else {
		// if system doesn't support memsw, just ignore swap settings.
		r.MemorySwapLimit = -2
	}

	// TODO: compare current usages and values to be set.
	if r.MemoryLimit >= -1 {
		confFile := path.Join(memoryPath, memoryLimit)
		confValue := []byte(fmt.Sprintf("%d", r.MemoryLimit))
		log.Debugf("set %s => %s", memoryLimit, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.MemorySoftLimit >= -1 {
		confFile := path.Join(memoryPath, memorySoftLimit)
		confValue := []byte(fmt.Sprintf("%d", r.MemorySoftLimit))
		log.Debugf("set %s => %s", memorySoftLimit, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.MemorySwapLimit >= -1 {
		confFile := path.Join(memoryPath, memorySwapLimit)
		confValue := []byte(fmt.Sprintf("%d", r.MemorySwapLimit))
		log.Debugf("set %s => %s", memorySwapLimit, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.MemorySwappiness <= 100 {
		confFile := path.Join(memoryPath, memorySwappiness)
		confValue := []byte(fmt.Sprintf("%d", r.MemorySwappiness))
		log.Debugf("set %s => %s", memorySwappiness, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.OomKillDisable {
		confFile := path.Join(memoryPath, memoryOomControl)
		log.Debugf("set oom_kill_disable => true")
		if err := ioutil.WriteFile(confFile, []byte("1"), 0644); err != nil {
			return err
		}
	}

	if r.KernelMemoryLimit >= -1 {
		confFile := path.Join(memoryPath, kernelMemoryLimit)
		confValue := []byte(fmt.Sprintf("%d", r.KernelMemoryLimit))
		log.Debugf("set %s => %s", kernelMemoryLimit, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.KernelMemoryTCPLimit >= -1 {
		confFile := path.Join(memoryPath, kernelMemoryTCPLimit)
		confValue := []byte(fmt.Sprintf("%d", r.KernelMemoryTCPLimit))
		log.Debugf("set %s => %s", kernelMemoryTCPLimit, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}

// this function ignores errors on purpose.
func getDefaultSwappiness() uint64 {
	valueBytes, _ := ioutil.ReadFile("/proc/sys/vm/swappiness")
	value, _ := strconv.Atoi(string(valueBytes[:len(valueBytes)-1]))
	return uint64(value)
}
