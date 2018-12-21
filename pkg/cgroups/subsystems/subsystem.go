package subsystems

type ResourceConfig struct {
	MemoryLimit string `json:"MemoryLimit"`
	CpuPeriod   string `json:"CpuPeriod"`
	CpuQuota    string `json:"CpuQuota"`
	CpuShare    string `json:"CpuShare"`
	CpuSet      string `json:"CpuSet"`
}

type Subsystem interface {
	Name() string
	Set(path string, res *ResourceConfig) error
	Apply(path string, pid int) error
	Remove(path string) error
}

var (
	Ins = []Subsystem{
		&MemorySubSystem{},
		&CpuSubSystem{},
		&CpusetSubSystem{},
	}
)
