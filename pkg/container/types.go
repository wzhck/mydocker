package container

import (
	"github.com/weikeit/mydocker/pkg/cgroups/subsystems"
	"os"
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
	Allowed() bool
	MountRootfs(*Container) error
	MountVolume(*Container) error
}

type Device struct {
	Type     rune        `json:"Type"`
	Path     string      `json:"Path"`
	Major    int64       `json:"Major"`
	Minor    int64       `json:"Minor"`
	FileMode os.FileMode `json:"FileMode"`
	Uid      uint32      `json:"Uid"`
	Gid      uint32      `json:"Gid"`
}

type Mount struct {
	Source string `json:"Source"`
	Target string `json:"Target"`
	Fstype string `json:"Fstype"`
	Flags  int    `json:"Flags"`
	Data   string `json:"Data"`
}
