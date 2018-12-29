package container

import (
	"github.com/weikeit/mydocker/pkg/cgroups/subsystems"
)

type AufsStorage struct {
	ContainerDir string `json:"ContainerDir"`
	ReadOnlyDir  string `json:"ReadOnlyDir"`
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
	In  string `json:"In"`
	Out string `json:"Out"`
}

type Container struct {
	Detach     bool         `json:"Detach"`
	Uuid       string       `json:"Uuid"`
	Name       string       `json:"Name"`
	Dns        []string     `json:"Dns"`
	Pid        int          `json:"Pid"`
	Image      string       `json:"Image"`
	CgroupPath string       `json:"CgroupPath"`
	CreateTime string       `json:"CreateTime"`
	Status     string       `json:"Status"`
	Commands   []string     `json:"Commands"`
	Rootfs     *AufsStorage `json:"Rootfs"`
	Volumes    []*Volume    `json:"Volumes"`
	Envs       []*Env       `json:"Envs"`
	Network    string       `json:"Network"`
	IPAddr     string       `json:"IPAddr"`
	Ports      []*Port      `json:"Ports"`

	// the resources limits.
	Resources *subsystems.ResourceConfig `json:"Resources"`
}
