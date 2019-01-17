package cgroups

type Resources struct {
	/////////////////////////////////////////////////
	// configurations for cpu subsystem of cgroups //
	/////////////////////////////////////////////////

	// since Linux 2.6.24; CONFIG_CGROUP_SCHED
	// https://www.kernel.org/doc/Documentation/scheduler/sched-bwc.txt
	// https://www.kernel.org/doc/Documentation/scheduler/sched-design-CFS.txt
	// https://www.kernel.org/doc/Documentation/scheduler/sched-rt-group.txt
	// https://github.com/digoal/blog/blob/master/201606/20160613_01.md

	// cpu period (1ms-1s, i.e., 1000-1000000) to be used for CFS scheduling.
	// `0` to use system default.
	// cpu.cfs_quota_us: the total available run-time within a period (in us)
	// default: 100ms = 100000
	CpuCfsPeriod uint64 `json:"CpuCfsPeriod"`

	// how many time cpu will use in CFS scheduling.
	// cpu.cfs_period_us: the length of a period (in us)
	// default: -1 indicates unlimited.
	CpuCfsQuota uint64 `json:"CpuCfsQuota"`

	// how many time cpu will use in realtime scheduling.
	// cpu.rt_period_us
	// default: 1s, i.e., 1000000
	CpuRtPeriod uint64 `json:"CpuRtPeriod"`

	// cpu period to be used for realtime scheduling
	// cpu.rt_runtime_us
	// default: 0.95s, i.e., 950000
	CpuRtRuntime uint64 `json:"CpuRtRuntime"`

	// relative weight vs. other containers
	// cpu.shares
	CpuShares uint64 `json:"CpuShares"`

	////////////////////////////////////////////////////
	// configurations for cpuset subsystem of cgroups //
	////////////////////////////////////////////////////

	// since Linux 2.6.24; CONFIG_CPUSETS
	// https://www.kernel.org/doc/Documentation/cgroup-v1/cpusets.txt

	// cpus can be used in this cgroup, e.g., 0,2-4,6
	// cpuset.cpus
	CpusetCpus string `json:"CpusetCpus"`

	// cpuset.mems
	CpusetMems string `json:"CpusetMems"`

	////////////////////////////////////////////////////
	// configurations for memory subsystem of cgroups //
	////////////////////////////////////////////////////

	// since Linux 2.6.25; CONFIG_MEMCG
	// https://www.kernel.org/doc/Documentation/cgroup-v1/memory.txt
	// https://github.com/digoal/blog/blob/master/201701/20170111_02.md

	// memory.limit_in_bytes
	MemoryLimit int64 `json:"MemoryLimit"`

	// memory.soft_limit_in_bytes
	MemorySoftLimit int64 `json:"MemorySoftLimit"`

	// total memory usage (memory + swap); `-1` to enable unlimited swap.
	// memory.memsw.limit_in_bytes
	MemorySwapLimit int64 `json:"MemorySwapLimit"`

	// swappiness (0 to 100) controls how aggressive the kernel will swap memory pages,
	// higher value will increase aggressiveness, lower values decrease the amount of swap.
	// memory.swappiness
	MemorySwappiness uint64 `json:"MemorySwappiness"`

	// `false` means process will be OOMKilled if memory usages exceed MemoryLimit,
	// while `true` won't.
	// memory.oom_control
	OomKillDisable bool `json:"OomKillDisable"`

	// memory.kmem.limit_in_bytes
	KernelMemoryLimit int64 `json:"KernelMemoryLimit"`

	// memory.kmem.tcp.limit_in_bytes
	KernelMemoryTCPLimit int64 `json:"KernelMemoryTCPLimit"`

	///////////////////////////////////////////////////
	// configurations for blkio subsystem of cgroups //
	///////////////////////////////////////////////////

	// since Linux 2.6.33; CONFIG_BLK_CGROUP
	// https://www.kernel.org/doc/Documentation/cgroup-v1/blkio-controller.txt

	// specifies per cgroup weight, range is from 10 to 1000.
	// blkio.weight
	BlkioWeight uint64 `json:"BlkioWeight"`

	// specifies tasks' weight in the given cgroup while competing with the
	// cgroup's child cgroups, range is from 10 to 1000, cfq scheduler only
	// blkio.leaf_weight
	BlkioLeafWeight uint64 `json:"BlkioLeafWeight"`

	// echo "majar:minor weight" > blkio.weight_device
	// blkio.weight_device
	BlkioWeightDevice []*WeightDevice `json:"BlkioWeightDevice"`

	// echo "majar:minor leaf_weight" > blkio.leaf_weight_device
	// blkio.leaf_weight_device
	BlkioLeafWeightDevice []*WeightDevice `json:"BlkioLeafWeightDevice"`

	// echo "majar:minor rate_bytes_per_second" > blkio.throttle.read_bps_device
	// blkio.throttle.read_bps_device
	BlkioThrottleReadBpsDevice []*ThrottleDevice `json:"BlkioThrottleReadBpsDevice"`

	// echo "majar:minor rate_bytes_per_second" > blkio.throttle.write_bps_device
	// blkio.throttle.write_bps_device
	BlkioThrottleWriteBpsDevice []*ThrottleDevice `json:"BlkioThrottleWriteBpsDevice"`

	// echo "majar:minor rate_io_per_second" > blkio.throttle.read_iops_device
	// blkio.throttle.read_iops_device
	BlkioThrottleReadIOPSDevice []*ThrottleDevice `json:"BlkioThrottleReadIOPSDevice"`

	// echo "majar:minor rate_io_per_second" > blkio.throttle.write_iops_device
	// blkio.throttle.write_iops_device
	BlkioThrottleWriteIOPSDevice []*ThrottleDevice `json:"BlkioThrottleWriteIOPSDevice"`

	/////////////////////////////////////////////////////
	// configurations for devices subsystem of cgroups //
	/////////////////////////////////////////////////////

	// since Linux 2.6.26; CONFIG_CGROUP_DEVICE
	// https://www.kernel.org/doc/Documentation/cgroup-v1/devices.txt

	// echo "c 1:3 rwm" > devices.allow
	Device []*Device `json:"Device"`

	//////////////////////////////////////////////////
	// configurations for pids subsystem of cgroups //
	//////////////////////////////////////////////////

	// since Linux 4.3; CONFIG_CGROUP_PIDS
	// https://www.kernel.org/doc/Documentation/cgroup-v1/pids.txt

	// process limit, set `0` to disable limit.
	// pids.max
	PidsMax uint64 `json:"PidsMax"`

	/////////////////////////////////////////////////////
	// configurations for net_cls subsystem of cgroups //
	/////////////////////////////////////////////////////

	// since Linux 2.6.29; CONFIG_CGROUP_NET_CLASSID
	// https://www.kernel.org/doc/Documentation/cgroup-v1/net_cls.txt

	// set class identifier for container's network packets.
	// net_cls.classid
	NetClsClassid uint64 `json:"NetClsClassid"`

	//////////////////////////////////////////////////////
	// configurations for net_prio subsystem of cgroups //
	//////////////////////////////////////////////////////

	// since Linux 3.3; CONFIG_CGROUP_NET_PRIO
	// https://www.kernel.org/doc/Documentation/cgroup-v1/net_prio.txt

	// echo "eth0 5" > net_prio.ifpriomap
	// net_prio.ifpriomap
	NetPrioIfpriomap []*IfPrioMap `json:"NetPrioIfpriomap"`

	/////////////////////////////////////////////////////
	// configurations for freezer subsystem of cgroups //
	/////////////////////////////////////////////////////

	// since Linux 2.6.28; CONFIG_CGROUP_FREEZER
	// https://www.kernel.org/doc/Documentation/cgroup-v1/freezer-subsystem.txt

	// can only be "FROZEN" or "THAWED"
	// freezer.state
	FreezerState string `json:"Freezer"`

	/////////////////////////////////////////////////////
	// configurations for hugetlb subsystem of cgroups //
	/////////////////////////////////////////////////////

	// since Linux 3.5; CONFIG_CGROUP_HUGETLB
	// https://www.kernel.org/doc/Documentation/cgroup-v1/hugetlb.txt

	// for &Hugepage{PageSize: "2MB", Limit: 2000}
	// echo 2000 > hugetlb.2MB.limit_in_bytes
	HugepagesLimit []*Hugepage `json:"HugepagesLimit"`
}
