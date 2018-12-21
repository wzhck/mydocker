package container

import "github.com/weikeit/mydocker/pkg/cgroups/subsystems"

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

type Image struct {
	Name      string `json:"Name"`
	TarFile   string `json:"TarFile"`
	RootfsDir string `json:"RootfsDir"`
}

type Container struct {
	Detach     bool         `json:"Detach"`
	Name       string       `json:"Name"`
	Image      *Image       `json:"Image"`
	Uuid       string       `json:"Uuid"`
	Pid        int          `json:"Pid"`
	CgroupPath string       `json:"CgroupPath"`
	CreateTime string       `json:"CreateTime"`
	Status     string       `json:"Status"`
	Commands   []string     `json:"Commands"`
	Rootfs     *AufsStorage `json:"Rootfs"`
	Volumes    []*Volume    `json:"Volumes"`
	Envs       []*Env       `json:"Envs"`

	// the resources limits.
	Resources *subsystems.ResourceConfig `json:"Resources"`
}
