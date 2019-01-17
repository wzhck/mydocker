package cgroups

import (
	"fmt"
	"github.com/urfave/cli"
	"runtime"
)

var Flags = []cli.Flag{
	cli.Float64Flag{
		Name:   "cpu-exceed-rate",
		Usage:  "Limit each cpu usages won't exceed this rate",
		Value:  2.5,
		Hidden: true,
	},
	cli.Uint64Flag{
		Name:  "cpu-cfs-period",
		Usage: "Limit CPU CFS (Completely Fair Scheduler) period in us",
		Value: 200000,
	},
	cli.Uint64Flag{
		Name:  "cpu-cfs-quota",
		Usage: "Limit CPU CFS (Completely Fair Scheduler) quota in us",
		Value: 200000,
	},
	cli.Uint64Flag{
		Name:  "cpu-rt-period",
		Usage: "Limit CPU Real-Time Scheduler period in us",
		Value: 1000000,
	},
	cli.Uint64Flag{
		Name:  "cpu-rt-runtime",
		Usage: "Limit CPU Real-Time Scheduler runtime in us",
		Value: 950000,
	},
	cli.Uint64Flag{
		Name:  "cpu-shares,c",
		Usage: "CPU shares (relative weight)",
		Value: 1024,
	},
	cli.StringFlag{
		Name:  "cpuset-cpus",
		Usage: "CPUs in which to allow execution (0-3, 0,1)",
		Value: fmt.Sprintf("0-%d", runtime.NumCPU()-1),
	},
	cli.StringFlag{
		Name:  "cpuset-mems",
		Usage: "MEMs in which to allow execution (0-3, 0,1)",
		Value: fmt.Sprintf("0-%d", getMemNodesNum()-1),
	},
	cli.Int64Flag{
		Name:  "memory-limit",
		Usage: "Memory limit in bytes; -1 indicates unlimited",
		Value: -1,
	},
	cli.Int64Flag{
		Name:  "memory-soft-limit",
		Usage: "Memory soft limit in bytes; -1 indicates unlimited",
		Value: -1,
	},
	cli.Int64Flag{
		Name:  "memory-swap-limit",
		Usage: "Swap limit equals to memory plus swap; -1 indicates unlimited",
		Value: -1,
	},
	cli.Uint64Flag{
		Name:  "memory-swappiness",
		Usage: "Tune container memory swappiness (range [0, 100])",
		Value: getDefaultSwappiness(),
	},
	cli.BoolFlag{
		Name:  "oom-kill-disable",
		Usage: "Disable oom killer, i.e., process will be hung if oom, NOT killed",
	},
	cli.Int64Flag{
		Name:  "kernel-memory-limit",
		Usage: "Kernel memory limit in bytes; -1 indicates unlimited",
		Value: -1,
	},
	cli.Int64Flag{
		Name:  "kernel-memory-tcp-limit",
		Usage: "Kernel memory tcp limit in bytes; -1 indicates unlimited",
		Value: -1,
	},
	cli.Uint64Flag{
		Name:   "blkio-weight",
		Usage:  "Block IO relative weight (range [10, 1000])",
		Value:  1000,
		Hidden: true,
	},
	cli.Uint64Flag{
		Name:   "blkio-leaf-weight",
		Usage:  "Block IO leaf weight (range [10, 1000])",
		Value:  1000,
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:   "blkio-weight-device",
		Usage:  "Block IO devices with relative weight, format: 'major:minor:weight'",
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:   "blkio-leaf-weight-device",
		Usage:  "Block IO devices with leaf weight, format: 'major:minor:weight'",
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:   "device-read-bps",
		Usage:  "Limit read bps rate of a device, format: 'major:minor:rate'",
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:   "device-write-bps",
		Usage:  "Limit write bps rate of a device, format: 'major:minor:rate'",
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:   "device-read-iops",
		Usage:  "Limit read iops rate of a device, format: 'major:minor:rate'",
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:   "device-write-iops",
		Usage:  "Limit write iops rate of a device, format: 'major:minor:rate'",
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:   "device",
		Usage:  "Add a host device to the container, format: '/src:/dst:rwm'",
		Hidden: true,
	},
	cli.Uint64Flag{
		Name:  "pids-max",
		Usage: "Limit pids number in container; 0 indicates unlimited",
	},
	cli.Uint64Flag{
		Name:   "net-classid",
		Usage:  "Set class identifier for container's network packets",
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:   "net-prio",
		Usage:  "Set priority for container's interface, format: 'ifacename:priority'",
		Hidden: true,
	},
	cli.StringFlag{
		Name:   "freezer-state",
		Usage:  "Set freezer state for the container, must be 'FROZEN' or 'THAWED'",
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:   "hugepages-limit",
		Usage:  "Limit usages by bytes for hugepages, format: 'PageSize:Limit'",
		Hidden: true,
	},
}
