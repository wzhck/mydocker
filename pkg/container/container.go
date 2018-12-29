package container

import (
	"encoding/json"
	"fmt"
	"github.com/Pallinder/go-randomdata"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/pkg/cgroups"
	"github.com/weikeit/mydocker/pkg/cgroups/subsystems"
	"github.com/weikeit/mydocker/pkg/network"
	_ "github.com/weikeit/mydocker/pkg/nsenter"
	"github.com/weikeit/mydocker/util"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func NewContainer(ctx *cli.Context) (*Container, error) {
	if len(ctx.Args()) < 1 {
		return nil, fmt.Errorf("missing container commands")
	}

	detach := ctx.Bool("detach")

	name := ctx.String("name")
	if name == "" {
		// generate a random name if necessary.
		name = strings.ToLower(randomdata.SillyName())
	}

	uuid, err := util.Uuid()
	if err != nil {
		return nil, err
	}

	// fetch the last 12 chars of standard uuid string.
	uuid = uuid[24:]

	dns := ctx.StringSlice("dns")

	if c, _ := GetContainerByNameOrUuid(name); c != nil {
		return nil, fmt.Errorf("the container name '%s' already exist", name)
	}

	imageName := ctx.String("image")
	if imageName == "" {
		return nil, fmt.Errorf("the image name is required")
	}

	imageTarFile, ok := Images[imageName]
	if !ok {
		return nil, fmt.Errorf("only support alpine and busybox")
	}

	image := &Image{
		Name:      imageName,
		TarFile:   imageTarFile,
		RootfsDir: path.Join(ImagesDir, imageName),
	}

	var commands []string
	for _, arg := range ctx.Args() {
		commands = append(commands, arg)
	}

	rootfs := &AufsStorage{
		ContainerDir: path.Join(ContainersDir, uuid),
		ReadOnlyDir:  image.RootfsDir,
		WriteDir:     path.Join(WriteLayterDir, uuid),
		MergeDir:     path.Join(ContainersDir, uuid, "mnt"),
	}

	var volumes []*Volume
	for _, volumeArg := range ctx.StringSlice("volume") {
		volumePeers := strings.Split(volumeArg, ":")
		if len(volumePeers) == 2 && volumePeers[0] != "" && volumePeers[1] != "" {
			volumes = append(volumes, &Volume{
				Source: volumePeers[0],
				Target: path.Join(rootfs.MergeDir, volumePeers[1]),
			})
		} else {
			return nil, fmt.Errorf("the argument of -v should be '-v /src:/dst'")
		}
	}

	var envs []*Env
	for _, envArg := range ctx.StringSlice("env") {
		envPeers := strings.Split(envArg, "=")
		if len(envPeers) == 2 && envPeers[0] != "" && envPeers[1] != "" {
			envs = append(envs, &Env{
				Key:   envPeers[0],
				Value: envPeers[1],
			})
		} else {
			return nil, fmt.Errorf("the argument of -e should be '-e key=value'")
		}
	}

	var ports []*Port
	for _, portArg := range ctx.StringSlice("publish") {
		portPeers := strings.Split(portArg, ":")
		if len(portPeers) == 2 && portPeers[0] != "" && portPeers[1] != "" {
			port := &Port{
				In:  portPeers[0],
				Out: portPeers[1],
			}

			for _, portStr := range []string{port.In, port.Out} {
				if portNum, err := strconv.Atoi(portStr); err != nil {
					return nil, fmt.Errorf("the port %s is not integer", portStr)
				} else if portNum < 0 || portNum > 65535 {
					return nil, fmt.Errorf("the port %s is out of [0, 65535]", portStr)
				}
			}

			if server, err := net.Listen("tcp", ":"+port.Out); err != nil {
				return nil, fmt.Errorf("the host port %s is already in use",
					port.Out)
			} else {
				server.Close()
			}

			ports = append(ports, port)
		} else {
			return nil, fmt.Errorf("the argument of -p should be '-p in:out'")
		}
	}

	nwName := ctx.String("network")
	var ipaddr string
	if nwName != "" {
		if err := network.Init(); err != nil {
			return nil, err
		}

		nw, ok := network.Networks[nwName]
		if !ok {
			return nil, fmt.Errorf("no such network %s, please create it first", nwName)
		}

		if ip, err := network.IPAllocator.Allocate(nw); err != nil {
			return nil, fmt.Errorf("failed to allocate new ip from network %s: %v", nwName, err)
		} else {
			ipaddr = ip.String()
		}
	}

	return &Container{
		Detach:     detach,
		Uuid:       uuid,
		Name:       name,
		Dns:        dns,
		Image:      image,
		Commands:   commands,
		Rootfs:     rootfs,
		Volumes:    volumes,
		Envs:       envs,
		Ports:      ports,
		Network:    nwName,
		IPAddr:     ipaddr,
		Status:     Creating,
		CgroupPath: Mydocker + "/" + uuid,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Resources: &subsystems.ResourceConfig{
			MemoryLimit: ctx.String("memory"),
			CpuPeriod:   ctx.String("cpu-period"),
			CpuQuota:    ctx.String("cpu-quota"),
			CpuShare:    ctx.String("cpu-share"),
			CpuSet:      ctx.String("cpuset"),
		},
	}, nil
}

func (c *Container) Run() error {
	parentCmd, writePipe, err := NewParentProcess(c)
	if err != nil {
		return err
	}

	if parentCmd == nil {
		return fmt.Errorf("failed to create parent process in container")
	}
	if err := parentCmd.Start(); err != nil {
		return err
	}

	c.Pid = parentCmd.Process.Pid
	c.Status = Running
	util.PrintExeFile(parentCmd.Process.Pid)

	// should call c.Record() after modifying c.Pid
	if err := c.Dump(); err != nil {
		return err
	}

	cm := cgroups.NewCgroupManager(c.CgroupPath)
	if !c.Detach {
		defer cm.Destroy()
	}

	cm.Set(c.Resources)
	cm.Apply(parentCmd.Process.Pid)

	SendInitCommand(c.Commands, writePipe)

	cm.Apply(parentCmd.Process.Pid)

	if err := c.handleNetwork(Create); err != nil {
		return err
	}

	if !c.Detach {
		parentCmd.Wait()
		c.handleNetwork(Delete)
		return deleteContainerRootfs(c)
	} else {
		_, err := fmt.Fprintln(os.Stdout, c.Uuid)
		return err
	}
}

func (c *Container) Logs() error {
	logFileName := path.Join(c.Rootfs.ContainerDir, LogName)
	logFile, err := os.Open(logFileName)
	defer logFile.Close()
	if err != nil {
		return fmt.Errorf("failed to open container log file %s: %v",
			logFileName, err)
	}

	contents, err := ioutil.ReadAll(logFile)
	if err != nil {
		return fmt.Errorf("failed to read container log file %s: %v",
			logFileName, err)
	}

	_, err = fmt.Fprintf(os.Stdout, string(contents))
	return err
}

func (c *Container) Exec(cmdArray []string) error {
	cmdStr := strings.Join(cmdArray, " ")
	log.Debugf("will execute command '%s' in the container with pid %d:",
		cmdStr, c.Pid)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ContainerPid, fmt.Sprintf("%d", c.Pid))
	os.Setenv(ContainerCmd, cmdStr)

	containerEnvs, err := util.GetEnvsByPid(c.Pid)
	if err != nil {
		return err
	}

	cmd.Env = append(os.Environ(), containerEnvs...)
	return cmd.Run()
}

func (c *Container) Stop() error {
	if c.Network != "" {
		if err := c.handleNetwork(Delete); err != nil {
			// just need to record error logs if failed.
			log.Errorf("failed to cleanup networks of container %s: %v",
				c.Uuid, err)
		}
	}

	msg := "failed to stop container %s by sending %s signal"

	if exist, _ := util.FileOrDirExists(fmt.Sprintf("/proc/%d", c.Pid)); exist {
		if err := syscall.Kill(c.Pid, syscall.SIGTERM); err != nil {
			log.Warnf(msg, c.Uuid, "SIGTERM")
			if err := syscall.Kill(c.Pid, syscall.SIGKILL); err != nil {
				return fmt.Errorf(msg, c.Uuid, "SIGKILL")
			}
		}
	}

	if err := umountLocalVolumes(c); err != nil {
		return err
	}

	if err := umountMergeLayer(c); err != nil {
		return err
	}

	c.Pid = 0
	c.Status = Stopped
	if err := c.Dump(); err != nil {
		return fmt.Errorf("failed to modify the status of container %s : %v",
			c.Uuid, err)
	}

	_, err := fmt.Fprintln(os.Stdout, c.Uuid)
	return err
}

func (c *Container) Start() error {
	if c.Status != Running {
		return c.Run()
	}
	return nil
}

func (c *Container) Restart() error {
	if c.Status == Running {
		if err := c.Stop(); err != nil {
			return err
		}
	}
	return c.Start()
}

func (c *Container) Delete() error {
	if c.Status == Running {
		if err := c.Stop(); err != nil {
			return err
		}
	}

	if c.Network != "" {
		if err := network.Init(); err != nil {
			return err
		}

		nw := network.Networks[c.Network]
		ip := net.ParseIP(c.IPAddr)
		if err := network.IPAllocator.Release(nw, &ip); err != nil {
			log.Errorf("failed to release ipaddr of container %s: %v",
				c.Uuid, err)
		}

		ep := &network.Endpoint{Uuid: c.Uuid}
		// delete the endpoint config file.
		ep.Delete()
	}

	return deleteContainerRootfs(c)
}

func (c *Container) Dump() error {
	containerBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to encode container object using json: %v", err)
	}

	configFileName := path.Join(c.Rootfs.ContainerDir, ConfigName)
	configFile, err := os.Create(configFileName)
	defer configFile.Close()
	if err != nil {
		return fmt.Errorf("failed to create container config file: %v", err)
	}

	if err := ioutil.WriteFile(configFileName, containerBytes, 0644); err != nil {
		return fmt.Errorf("failed to write container configs to file %s: %v",
			configFileName, err)
	}
	return nil
}

func (c *Container) handleNetwork(action string) error {
	if c.Network == "" {
		return nil
	}

	var portMaps []string
	for _, port := range c.Ports {
		portMaps = append(portMaps, fmt.Sprintf("%s:%s",
			port.In, port.Out))
	}

	if err := network.Init(); err != nil {
		return err
	}

	nw := network.Networks[c.Network]

	switch action {
	case Create:
		return nw.Connect(c.Uuid, c.Pid, c.IPAddr, portMaps)
	case Delete:
		return nw.DisConnect(c.Uuid, c.Pid)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}
