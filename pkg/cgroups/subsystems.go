package cgroups

type Subsystem interface {
	Name() string
	RootName() string
	Remove(cgPath string) error
	Apply(cgPath string, pid int) error
	Set(cgPath string, resources *Resources) error
}

// register all subsystems here.
var Subsystems = []Subsystem{
	&CpuSubsystem{},
	&CpusetSubsystem{},
	&MemorySubsystem{},
	&BlkioSubsystem{},
	&DevicesSubsystem{},
	&PidsSubsystem{},
	&NetClsSubsystem{},
	&NetPrioSubsystem{},
	&FreezerSubsystem{},
	&HugetlbSubsystem{},
}
