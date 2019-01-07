package container

import (
	"path"
	"syscall"
)

const (
	MyDocker    = "mydocker"
	MyDockerDir = "/var/lib/mydocker"
	ConfigName  = "config.json"
	LogName     = "container.log"
	XinoTmpfs   = "/var/local/xino"
	MaxBytes    = 10000
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

	defaultMountFlags = syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID

	// for more mount informations, please see documents of runc:
	// https://github.com/opencontainers/runc/blob/master/libcontainer/README.md
	// https://github.com/opencontainers/runc/blob/master/libcontainer/SPEC.md#filesystem
	Mounts = []*Mount{
		{
			Source: "proc",
			Target: "/proc",
			Fstype: "proc",
			Flags:  defaultMountFlags,
		},
		{
			Source: "sysfs",
			Target: "/sys",
			Fstype: "sysfs",
			Flags:  defaultMountFlags | syscall.MS_RDONLY,
		},
		{
			Source: "tmpfs",
			Target: "/dev",
			Fstype: "tmpfs",
			Flags:  syscall.MS_NOSUID | syscall.MS_STRICTATIME,
			Data:   "mode=0755,size=200M",
		},
		{
			Source: "devpts",
			Target: "/dev/pts",
			Fstype: "devpts",
			Flags:  syscall.MS_NOEXEC | syscall.MS_NOSUID,
			Data:   "mode=0620,newinstance,ptmxmode=0666,gid=5",
		},
		{
			Source: "shm",
			Target: "/dev/shm",
			Fstype: "tmpfs",
			Flags:  defaultMountFlags,
			Data:   "mode=1777,size=100M",
		},
		{
			Source: "mqueue",
			Target: "/dev/mqueue",
			Fstype: "mqueue",
			Flags:  defaultMountFlags,
			Data:   "mode=1777,size=100M",
		},
	}

	// e.g. use command `file /dev/null` to get major:minor
	Devices = []*Device{
		{
			Type:     'c',
			Path:     "/dev/null",
			Major:    1,
			Minor:    3,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/zero",
			Major:    1,
			Minor:    5,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/full",
			Major:    1,
			Minor:    7,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/random",
			Major:    1,
			Minor:    8,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/urandom",
			Major:    1,
			Minor:    9,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/tty",
			Major:    5,
			Minor:    0,
			FileMode: 0666,
			Uid:      0,
			Gid:      0,
		},
		{
			Type:     'c',
			Path:     "/dev/console",
			Major:    5,
			Minor:    1,
			FileMode: 0620,
			Uid:      0,
			Gid:      0,
		},
	}

	DevSymlinks = map[string]string{
		"/proc/self/fd":   "/dev/fd",
		"/proc/self/fd/0": "/dev/stdin",
		"/proc/self/fd/1": "/dev/stdout",
		"/proc/self/fd/2": "/dev/stderr",
	}
)
