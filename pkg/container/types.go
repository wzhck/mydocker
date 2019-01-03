package container

import (
	"github.com/weikeit/mydocker/pkg/cgroups/subsystems"
)

type Rootfs struct {
	ContainerDir string `json:"ContainerDir"`
	ImageDir     string `json:"ImageDir"`
	WriteDir     string `json:"WriteDir"`
	MergeDir     string `json:"MergeDir"`
}

type Volume struct {
	Source string `json:"Source"`
	Target string `json:"Target"`
}

type Env struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type Port struct {
	Out string `json:"Out"`
	In  string `json:"In"`
}

type Container struct {
	Detach        bool      `json:"Detach"`
	Uuid          string    `json:"Uuid"`
	Name          string    `json:"Name"`
	Dns           []string  `json:"Dns"`
	Pid           int       `json:"Pid"`
	Image         string    `json:"Image"`
	CgroupPath    string    `json:"CgroupPath"`
	CreateTime    string    `json:"CreateTime"`
	Status        string    `json:"Status"`
	StorageDriver string    `json:"StorageDriver"`
	Rootfs        *Rootfs   `json:"Rootfs"`
	Commands      []string  `json:"Commands"`
	Volumes       []*Volume `json:"Volumes"`
	Envs          []*Env    `json:"Envs"`
	Network       string    `json:"Network"`
	IPAddr        string    `json:"IPAddr"`
	Ports         []*Port   `json:"Ports"`

	// the resources limits.
	Resources *subsystems.ResourceConfig `json:"Resources"`
}

type Driver interface {
	Name() string
	Module() string
	MountRootfs(*Container) error
	MountVolume(*Container) error
}
