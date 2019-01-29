package container

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/pkg/image"
	"github.com/weikeit/mydocker/pkg/network"
	_ "github.com/weikeit/mydocker/pkg/nsenter"
	"github.com/weikeit/mydocker/util"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

func (c *Container) Run() error {
	parentCmd, writePipe, err := c.NewParentProcess()
	if err != nil {
		return err
	}

	if parentCmd == nil {
		return fmt.Errorf("failed to create parent process in container")
	}

	sendInitCommand(c.Commands, writePipe)
	if err := parentCmd.Start(); err != nil {
		return err
	}

	c.Cgroups.Pid = parentCmd.Process.Pid
	c.Status = Running
	// util.PrintExeFile(parentCmd.Process.Pid)

	// MUST call c.Dump() after modifying c.Pid
	if err := c.Dump(); err != nil {
		return err
	}

	if err := c.Cgroups.Set(); err != nil {
		return err
	}

	if err := c.Cgroups.Apply(); err != nil {
		return err
	}

	if err := c.handleNetwork(Create); err != nil {
		if err := image.ChangeCounts(c.Image, "delete"); err != nil {
			log.Debugf("failed to recover image counts: %v", err)
		}
		return err
	}

	if !c.Detach {
		parentCmd.Wait()
		c.handleNetwork(Delete)
		c.cleanNetworkImage()
		c.Cgroups.Destory()
		return c.cleanupRootfs()
	} else {
		fmt.Println(c.Uuid)
		return nil
	}
}

func (c *Container) Logs(ctx *cli.Context) error {
	logFileName := path.Join(c.Rootfs.ContainerDir, LogName)
	if ctx.Bool("follow") {
		// third-party go library:
		// https://github.com/hpcloud/tail
		// https://github.com/fsnotify/fsnotify
		// but call tailf command is the easiest way :)
		cmd := exec.Command("tail", "-f", logFileName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

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

	fmt.Println(string(contents))
	return nil
}

func (c *Container) Exec(cmdArray []string) error {
	cmdStr := strings.Join(cmdArray, " ")
	log.Debugf("will execute command <%s> in the container "+
		"(pid: %d)", cmdStr, c.Cgroups.Pid)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// pass these environment variables to the nsenter.go
	// note: need to rename the env "debug", or nsenter won't
	// receive any change of the env "debug", don't know why?
	os.Setenv("debug_nsenter", os.Getenv("debug"))
	os.Setenv("container_pid", fmt.Sprintf("%d", c.Cgroups.Pid))
	os.Setenv("container_cmd", cmdStr)
	os.Setenv("cgroup_root", "/sys/fs/cgroup")
	os.Setenv("cgroup_path", c.Cgroups.Path)

	containerEnvs, err := util.GetEnvsByPid(c.Cgroups.Pid)
	if err != nil {
		return err
	}

	cmd.Env = append(os.Environ(), containerEnvs...)
	return cmd.Run()
}

func (c *Container) Stop() error {
	if c.Status == Stopped {
		return nil
	}

	if len(c.Endpoints) > 0 {
		if err := c.handleNetwork(Delete); err != nil {
			// just need to record error logs if failed.
			log.Debugf("failed to cleanup networks of container %s: %v",
				c.Uuid, err)
		}
	}

	if err := util.KillProcess(c.Cgroups.Pid); err != nil {
		return err
	}

	if err := c.umountRootfsVolume(); err != nil {
		return err
	}

	c.Cgroups.Pid = 0
	c.Status = Stopped
	if err := c.Dump(); err != nil {
		return fmt.Errorf("failed to modify the status of container %s : %v",
			c.Uuid, err)
	}

	c.Cgroups.Destory()
	fmt.Println(c.Uuid)

	return nil
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

	c.cleanNetworkImage()
	return c.cleanupRootfs()
}

func (c *Container) Dump() error {
	configFileName := path.Join(c.Rootfs.ContainerDir, ConfigName)
	if err := util.EnSureFileExists(configFileName); err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to json-encode container %s: %v",
			c.Uuid, err)
	}

	// WriteFile will create the file if it doesn't exist,
	// otherwise WriteFile will truncate it before writing
	if err := ioutil.WriteFile(configFileName, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write container config to file %s: %v",
			configFileName, err)
	}

	return nil
}

func (c *Container) Load() error {
	configFileName := path.Join(ContainersDir, c.Uuid, ConfigName)
	if err := util.EnSureFileExists(configFileName); err != nil {
		return err
	}

	jsonBytes, err := ioutil.ReadFile(configFileName)
	if len(jsonBytes) == 0 {
		return fmt.Errorf("the configFile of container %s is empty",
			configFileName)
	}
	if err != nil {
		return fmt.Errorf("failed to read container config %s: %v",
			configFileName, err)
	}

	if err := json.Unmarshal(jsonBytes, c); err != nil {
		return fmt.Errorf("failed to json-decode container %s: %v",
			c.Uuid, err)
	}

	if c.Cgroups.Pid > 0 {
		processDir := fmt.Sprintf("/proc/%d", c.Cgroups.Pid)
		if exist, _ := util.FileOrDirExists(processDir); !exist {
			c.Cgroups.Pid = 0
			c.Status = Stopped
			if err := c.Dump(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Container) cleanNetworkImage() {
	for _, ep := range c.Endpoints {
		nw := ep.Network
		ip := ep.IPAddr
		if err := network.IPAllocator.Release(nw, &ip); err != nil {
			log.Errorf("failed to release ip %s of container %s: %v",
				ip.String(), c.Uuid, err)
		}
	}

	if err := image.ChangeCounts(c.Image, "delete"); err != nil {
		log.Errorf("failed to decrease counts of image %s: %v",
			c.Image, err)
	}
}

func (c *Container) handleNetwork(action string) error {
	var err error
	unknownErr := fmt.Errorf("unknown action: %s", action)
	for _, ep := range c.Endpoints {
		switch action {
		case Create:
			err = ep.Connect(c.Cgroups.Pid)
		case Delete:
			err = ep.DisConnect(c.Cgroups.Pid)
		default:
			err = unknownErr
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Container) ConfigHosts() error {
	etcHosts := path.Join(c.Rootfs.WriteDir, "/etc/hosts")
	etcHostsLines := append(defaultHostsLines[:0:0], defaultHostsLines...)
	for _, ep := range c.Endpoints {
		etcHostsLines = append(etcHostsLines, fmt.Sprintf("%s %s", ep.IPAddr, c.Hostname))
	}

	etcHostsBytes := []byte(strings.Join(etcHostsLines, "\n") + "\n")
	if err := util.EnSureFileExists(etcHosts); err != nil {
		return fmt.Errorf("failed to create %s in container: %v", etcHosts, err)
	}
	if err := ioutil.WriteFile(etcHosts, etcHostsBytes, 0644); err != nil {
		return fmt.Errorf("failed to write contents into %s: %v", etcHosts, err)
	}

	return nil
}

func (c *Container) configHostname() error {
	etcHostname := path.Join(c.Rootfs.WriteDir, "/etc/hostname")
	if err := util.EnSureFileExists(etcHostname); err != nil {
		return fmt.Errorf("failed to create %s in container: %v", etcHostname, err)
	}
	if err := ioutil.WriteFile(etcHostname, []byte(c.Hostname+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write contents into %s: %v", etcHostname, err)
	}

	return c.ConfigHosts()
}

func (c *Container) configDNS() error {
	var nameservers []string
	for _, dns := range c.Dns {
		nameservers = append(nameservers, fmt.Sprintf("nameserver %s", dns))
	}

	resolvConf := path.Join(c.Rootfs.WriteDir, "/etc/resolv.conf")
	resolvConfBytes := []byte(strings.Join(nameservers, "\n") + "\n")

	if err := util.EnSureFileExists(resolvConf); err != nil {
		return fmt.Errorf("failed to create %s in container: %v", resolvConf, err)
	}
	if err := ioutil.WriteFile(resolvConf, resolvConfBytes, 0644); err != nil {
		return fmt.Errorf("failed to write contents into %s: %v", resolvConf, err)
	}

	return nil
}
