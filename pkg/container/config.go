package container

import "path"

const (
	MyDocker    = "mydocker"
	MyDockerDir = "/var/lib/mydocker"
	ConfigName  = "config.json"
	LogName     = "container.log"
	XinoTmpfs   = "/var/local/xino"
)

const (
	ContainerPid = "ContainerPid"
	ContainerCmd = "ContainerCmd"
)

const (
	Creating = "creating"
	Running  = "running"
	Stopped  = "stopped"
	Exited   = "exited"
)

const (
	Stop    = "stop"
	Start   = "start"
	Restart = "restart"
	Create  = "create"
	Delete  = "delete"
)

const (
	Aufs     = "aufs"
	Overlay2 = "overlay2"
)

var (
	ContainersDir = path.Join(MyDockerDir, "containers")

	// each driver MUST contain writeDir and mergeDir
	DriverConfigs = map[string]map[string]string{
		Aufs: {
			"writeDir": "diff",
			"mergeDir": "merged",
		},
		Overlay2: {
			"writeDir": "diff",
			"mergeDir": "merged",
			"workDir":  "work",
		},
	}

	// key is driver's name, value is a Driver implements.
	// should register all storage drivers here.
	Drivers = map[string]Driver{
		Aufs:     &AufsDriver{},
		Overlay2: &Overlay2Driver{},
	}
)
