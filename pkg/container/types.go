package container

import (
	"os"

	"weike.sh/mydocker/pkg/cgroups"
	"weike.sh/mydocker/pkg/network"
)

type Rootfs struct {
	ContainerDir string `json:"ContainerDir"`
	ImageDir     string `json:"ImageDir"`
	WriteDir     string `json:"WriteDir"`
	MergeDir     string `json:"MergeDir"`
}

type Container struct {
	Detach        bool                `json:"Detach"`
	Uuid          string              `json:"Uuid"`
	Name          string              `json:"Name"`
	Hostname      string              `json:"Hostname"`
	Dns           []string            `json:"Dns"`
	Image         string              `json:"Image"`
	CreateTime    string              `json:"CreateTime"`
	Status        string              `json:"Status"`
	StorageDriver string              `json:"StorageDriver"`
	Rootfs        *Rootfs             `json:"Rootfs"`
	Commands      []string            `json:"Commands"`
	Cgroups       *cgroups.Cgroups    `json:"Cgroups"`
	Volumes       map[string]string   `json:"Volumes"`
	Envs          map[string]string   `json:"Envs"`
	Ports         map[string]string   `json:"Ports"`
	Endpoints     []*network.Endpoint `json:"Endpoints"`
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
