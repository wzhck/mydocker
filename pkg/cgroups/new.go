package cgroups

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/util"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func NewResources(ctx *cli.Context) (*Resources, error) {
	parseFlagsFuncs := []func(*cli.Context, *Resources) error{
		parseCpuFlags,
		parseCpusetFlags,
		parseMemoryFlags,
		parseBlkioFlags,
		parseDevicesFlags,
		parsePidsFlags,
		parseNetClsFlags,
		parseNetPrioFlags,
		parseFreezerFlags,
		parseHugetlbFlags,
	}

	resources := &Resources{}
	for _, parseFlagsFunc := range parseFlagsFuncs {
		if err := parseFlagsFunc(ctx, resources); err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func parseCpuFlags(ctx *cli.Context, r *Resources) error {
	numCPU := uint64(runtime.NumCPU())
	maxRate := ctx.Float64("cpu-exceed-rate")
	if maxRate <= 0 {
		return fmt.Errorf("--cpu-exceed-rate must be positive")
	}

	cpuCfsPeriod := ctx.Uint64("cpu-cfs-period")
	if cpuCfsPeriod > 0 && cpuCfsPeriod < 1000 || cpuCfsPeriod > 1000000 {
		return fmt.Errorf("--cpu-cfs-period requires [1000, 1000000]")
	}
	r.CpuCfsPeriod = cpuCfsPeriod

	cpuCfsQuota := ctx.Uint64("cpu-cfs-quota")
	if cpuCfsQuota > uint64(float64(cpuCfsPeriod*numCPU)*maxRate) {
		return fmt.Errorf("--cpu-cfs-quota can't exceed cpuCfsPeriod*numCPU*maxRate")
	}
	r.CpuCfsQuota = cpuCfsQuota

	cpuRtPeriod := ctx.Uint64("cpu-rt-period")
	if cpuRtPeriod > 2000000 {
		return fmt.Errorf("--cpu-rt-period can't exceed 2000000")
	}
	r.CpuRtPeriod = cpuRtPeriod

	cpuRtRuntime := ctx.Uint64("cpu-rt-runtime")
	if cpuRtRuntime > uint64(float64(cpuRtPeriod*numCPU)*maxRate) {
		return fmt.Errorf("--cpu-rt-runtime can't exceed cpuRtPeriod*numCPU*maxRate")
	}
	r.CpuRtRuntime = cpuRtRuntime

	cpuShares := ctx.Uint64("cpu-shares")
	if cpuShares > 0 && cpuShares < 2 {
		return fmt.Errorf("--cpu-shares requires >= 2")
	}
	r.CpuShares = cpuShares

	return nil
}

func parseCpusetFlags(ctx *cli.Context, r *Resources) error {
	numCPU := runtime.NumCPU()
	numMem := getMemNodesNum()

	cpusetCpus := ctx.String("cpuset-cpus")
	if err := validateCpusetArgs(cpusetCpus, "cpu", numCPU); err != nil {
		return err
	}
	r.CpusetCpus = cpusetCpus

	cpusetMems := ctx.String("cpuset-mems")
	if err := validateCpusetArgs(cpusetMems, "mem", numMem); err != nil {
		return err
	}
	r.CpusetMems = cpusetMems

	return nil
}

func validateCpusetArgs(args, cls string, maxNum int) error {
	if args == "" {
		return nil
	}

	err := fmt.Errorf("--cpuset-%ss requires a-b, "+
		"and a must be less or equal to b, and both "+
		"must be in the range [0, %d)", cls, maxNum)

	for _, arg := range strings.Split(args, ",") {
		if strings.Contains(arg, "-") {
			re, _ := regexp.Compile(`(\d+)-(\d+)`)
			if !re.MatchString(args) {
				return err
			}

			pairs := re.FindStringSubmatch(arg)
			a, _ := strconv.Atoi(pairs[1])
			b, _ := strconv.Atoi(pairs[2])

			if a >= 0 && b < maxNum && a <= b {
				continue
			} else {
				return err
			}
		} else {
			numArg, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("%s is not a number", arg)
			}
			if numArg < 0 || numArg >= maxNum {
				return fmt.Errorf("value of --cpuset-%ss must "+
					"be in the range [0, %d)", cls, maxNum)
			}
		}
	}

	return nil
}

func parseMemoryFlags(ctx *cli.Context, r *Resources) error {
	memoryLimit := ctx.Int64("memory-limit")
	if memoryLimit < 0 {
		memoryLimit = -1
	}
	r.MemoryLimit = memoryLimit

	memorySoftLimit := ctx.Int64("memory-soft-limit")
	if memorySoftLimit < 0 {
		memorySoftLimit = -1
	}
	if memorySoftLimit > -1 && memoryLimit > -1 && memorySoftLimit < memoryLimit {
		return fmt.Errorf("memorySoftLimit requires >= memoryLimit")
	}
	r.MemorySoftLimit = memorySoftLimit

	memorySwapLimit := ctx.Int64("memory-swap-limit")
	if memorySwapLimit < 0 {
		memorySwapLimit = -1
	}
	if memorySwapLimit > -1 && memoryLimit > -1 && memorySwapLimit < memoryLimit {
		return fmt.Errorf("memorySwapLimit requires >= memoryLimit")
	}
	r.MemorySwapLimit = memorySwapLimit

	memorySwappiness := ctx.Uint64("memory-swappiness")
	if memorySwappiness > 100 {
		memorySwappiness = 100
	}
	r.MemorySwappiness = memorySwappiness

	r.OomKillDisable = ctx.Bool("oom-kill-disable")

	kernelMemoryLimit := ctx.Int64("kernel-memory-limit")
	if kernelMemoryLimit < 0 {
		kernelMemoryLimit = -1
	}
	r.KernelMemoryLimit = kernelMemoryLimit

	kernelMemoryTCPLimit := ctx.Int64("kernel-memory-tcp-limit")
	if kernelMemoryTCPLimit < 0 {
		kernelMemoryTCPLimit = -1
	}
	r.KernelMemoryTCPLimit = kernelMemoryTCPLimit

	return nil
}

// TODO: to be implemented or verified in detail.
func parseBlkioFlags(ctx *cli.Context, r *Resources) error {
	return nil

	blkioWeight := ctx.Uint64("blkio-weight")
	if blkioWeight < 10 || blkioWeight > 1000 {
		return fmt.Errorf("--blkio-weight requires [10, 1000]")
	}
	r.BlkioWeight = blkioWeight

	blkioLeafWeight := ctx.Uint64("blkio-leaf-weight")
	if blkioLeafWeight < 10 || blkioLeafWeight > 1000 {
		return fmt.Errorf("--blkio-leaf-weight requires [10, 1000]")
	}
	r.BlkioLeafWeight = blkioLeafWeight

	var blkioWeightDevice []*WeightDevice
	for _, deviceArg := range ctx.StringSlice("blkio-weight-device") {
		if device, err := parseWeightDevice(deviceArg, "weight"); err == nil {
			blkioWeightDevice = append(blkioWeightDevice, device)
		} else {
			return err
		}
	}
	r.BlkioWeightDevice = blkioWeightDevice

	var blkioLeafWeightDevice []*WeightDevice
	for _, deviceArg := range ctx.StringSlice("blkio-leaf-weight-device") {
		if device, err := parseWeightDevice(deviceArg, "leafWeight"); err == nil {
			blkioLeafWeightDevice = append(blkioLeafWeightDevice, device)
		} else {
			return err
		}
	}
	r.BlkioLeafWeightDevice = blkioLeafWeightDevice

	var blkioThrottleReadBpsDevice []*ThrottleDevice
	for _, deviceArg := range ctx.StringSlice("device-read-bps") {
		if device, err := parseThrottleDevice(deviceArg, r); err == nil {
			blkioThrottleReadBpsDevice = append(blkioThrottleReadBpsDevice, device)
		} else {
			return err
		}
	}
	r.BlkioThrottleReadBpsDevice = blkioThrottleReadBpsDevice

	var blkioThrottleWriteBpsDevice []*ThrottleDevice
	for _, deviceArg := range ctx.StringSlice("device-write-bps") {
		if device, err := parseThrottleDevice(deviceArg, r); err == nil {
			blkioThrottleWriteBpsDevice = append(blkioThrottleWriteBpsDevice, device)
		} else {
			return err
		}
	}
	r.BlkioThrottleWriteBpsDevice = blkioThrottleWriteBpsDevice

	var blkioThrottleReadIOPSDevice []*ThrottleDevice
	for _, deviceArg := range ctx.StringSlice("device-read-iops") {
		if device, err := parseThrottleDevice(deviceArg, r); err == nil {
			blkioThrottleReadIOPSDevice = append(blkioThrottleReadIOPSDevice, device)
		} else {
			return err
		}
	}
	r.BlkioThrottleReadIOPSDevice = blkioThrottleReadIOPSDevice

	var blkioThrottleWriteIOPSDevice []*ThrottleDevice
	for _, deviceArg := range ctx.StringSlice("device-write-iops") {
		if device, err := parseThrottleDevice(deviceArg, r); err == nil {
			blkioThrottleWriteIOPSDevice = append(blkioThrottleWriteIOPSDevice, device)
		} else {
			return err
		}
	}
	r.BlkioThrottleWriteIOPSDevice = blkioThrottleWriteIOPSDevice

	return nil
}

func parseWeightDevice(arg, cls string) (*WeightDevice, error) {
	// format: 'major:minor:weight'
	re, _ := regexp.Compile(`(\d+):(\d+):(\d+)`)
	if !re.MatchString(arg) {
		return nil, fmt.Errorf("the format must be 'major:minor:weight'")
	}

	results := re.FindStringSubmatch(arg)
	major, _ := strconv.Atoi(results[1])
	minor, _ := strconv.Atoi(results[2])
	weight, _ := strconv.Atoi(results[3])

	if !blockDeviceExists(major, minor) {
		return nil, fmt.Errorf("the block device %d:%d doesn't exist",
			major, minor)
	}
	if weight < 10 || weight > 1000 {
		return nil, fmt.Errorf("the weight must be in the range [10, 1000]")
	}

	device := &WeightDevice{}
	device.Major = uint64(major)
	device.Minor = uint64(minor)

	switch cls {
	case "weight":
		device.Weight = uint64(weight)
	case "leafWeight":
		device.LeafWeight = uint64(weight)
	default:
		return nil, fmt.Errorf("cls must be weight or leafWeight")
	}

	return device, nil
}

func parseThrottleDevice(arg string, r *Resources) (*ThrottleDevice, error) {
	// format: 'major:minor:rate'
	re, _ := regexp.Compile(`(\d+):(\d+):(\d+)`)
	if !re.MatchString(arg) {
		return nil, fmt.Errorf("the format must be 'major:minor:rates'")
	}

	results := re.FindStringSubmatch(arg)
	major, _ := strconv.Atoi(results[1])
	minor, _ := strconv.Atoi(results[2])
	rates, _ := strconv.Atoi(results[3])

	deviceFound := false
	for _, device := range append(r.BlkioWeightDevice, r.BlkioLeafWeightDevice...) {
		if device.Major == uint64(major) && device.Minor == uint64(minor) {
			deviceFound = true
		}
	}

	if !deviceFound {
		return nil, fmt.Errorf("the device %d:%d has not been added to container",
			major, minor)
	}
	if rates <= 0 {
		return nil, fmt.Errorf("the rates must be positive")
	}

	device := &ThrottleDevice{}
	device.Major = uint64(major)
	device.Minor = uint64(minor)
	device.Rates = uint64(rates)

	return device, nil
}

// TODO: to be implemented or verified in detail.
func parseDevicesFlags(ctx *cli.Context, r *Resources) error {
	return nil

	var devices []*Device
	for _, devArg := range ctx.StringSlice("device") {
		if device, err := parseDevice(devArg); err == nil {
			devices = append(devices, device)
		} else {
			return err
		}
	}
	r.Device = devices
	return nil
}

func parseDevice(arg string) (*Device, error) {
	argre, _ := regexp.Compile(`(\w+):(\w+):(\w+)`)
	if !argre.MatchString(arg) {
		return nil, fmt.Errorf("the format must be '/src/dev:/dst/dev:rwm'")
	}

	results := argre.FindStringSubmatch(arg)
	src, dst, access := results[1], results[2], results[3]

	for _, devStr := range []string{src, dst} {
		devRe, _ := regexp.Compile(`(/\w+)+`)
		if !devRe.MatchString(devStr) {
			return nil, fmt.Errorf("invalid device %s", devStr)
		}
	}

	accessRe, _ := regexp.Compile(`[rwm]{1,3}`)
	if !accessRe.MatchString(access) {
		return nil, fmt.Errorf("invalid access, must be [rwm]{1,3}")
	}

	return &Device{Path: src, Access: access,}, nil
}

func parsePidsFlags(ctx *cli.Context, r *Resources) error {
	r.PidsMax = ctx.Uint64("pids-max")
	return nil
}

// TODO: to be implemented or verified in detail.
func parseNetClsFlags(ctx *cli.Context, r *Resources) error {
	r.NetClsClassid = ctx.Uint64("net-classid")
	return nil
}

// TODO: to be implemented or verified in detail.
func parseNetPrioFlags(ctx *cli.Context, r *Resources) error {
	return nil

	var netPrioIfpriomap []*IfPrioMap

	for _, netPrio := range ctx.StringSlice("net-prio") {
		re, _ := regexp.Compile(`(\w+):(\d+)`)
		if !re.MatchString(netPrio) {
			return fmt.Errorf("--net-prio requires 'ifacename:priority'")
		}
		results := re.FindStringSubmatch(netPrio)
		priority, _ := strconv.Atoi(results[2])
		netPrioIfpriomap = append(netPrioIfpriomap, &IfPrioMap{
			Interface: results[1],
			Priority:  uint64(priority),
		})
	}
	r.NetPrioIfpriomap = netPrioIfpriomap

	return nil
}

// TODO: to be implemented or verified in detail.
func parseFreezerFlags(ctx *cli.Context, r *Resources) error {
	return nil

	freezerState := ctx.String("freezer-state")
	if freezerState != "" && !util.Contains(freezerStates, freezerState) {
		return fmt.Errorf("--freezer-state must be 'FROZEN' or 'THAWED'")
	}
	r.FreezerState = freezerState

	return nil
}

// TODO: to be implemented or verified in detail.
func parseHugetlbFlags(ctx *cli.Context, r *Resources) error {
	return nil

	var hugepagesLimit []*Hugepage

	for _, hugepage := range ctx.StringSlice("hugepages-limit") {
		re, _ := regexp.Compile(`(\d+[KMGT]B):(\d+)`)
		if !re.MatchString(hugepage) {
			return fmt.Errorf("--hugepages-limit requires PageSize:Limit, e.g., 2MB:10000")
		}

		results := re.FindStringSubmatch(hugepage)
		limitBytes, _ := strconv.Atoi(results[2])
		if limitBytes == 0 {
			return fmt.Errorf("limits of hugePage must be positive")
		}

		hugepagesLimit = append(hugepagesLimit, &Hugepage{
			PageSize: results[1],
			Limit:    uint64(limitBytes),
		})
	}
	r.HugepagesLimit = hugepagesLimit

	return nil
}
