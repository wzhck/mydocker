package container

import (
	"fmt"
	"github.com/weikeit/mydocker/pkg/image"
	"github.com/weikeit/mydocker/util"
	"os"
	"os/exec"
	"path"
	"syscall"
)

func (c *Container) NewParentProcess() (*exec.Cmd, *os.File, error) {
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create pipe: %v", err)
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}

	if err := c.PrepareRootfs(); err != nil {
		return nil, nil, err
	}

	logFileName := path.Join(c.Rootfs.ContainerDir, LogName)
	var logFile *os.File

	exist, _ := util.FileOrDirExists(logFileName)
	if !exist {
		logFile, err = os.Create(logFileName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create container log file %s: %v",
				logFileName, err)
		}
	} else {
		flags := syscall.O_WRONLY | syscall.O_APPEND
		logFile, err = os.OpenFile(logFileName, int(flags), 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open container log file %s: %v",
				logFileName, err)
		}
	}

	if c.Detach {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	} else {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.Dir = c.Rootfs.MergeDir
	cmd.ExtraFiles = []*os.File{readPipe}

	img, err := image.GetImageByNameOrUuid(c.Image)
	if err != nil {
		return nil, nil, err
	}
	envs := img.Envs

	for _, env := range c.Envs {
		envs = append(envs, fmt.Sprintf("%s=%s",
			env.Key, env.Value))
	}
	cmd.Env = append(os.Environ(), envs...)

	return cmd, writePipe, nil
}

func (c *Container) PrepareRootfs() error {
	if err := c.CreateRootfs(); err != nil {
		return err
	}
	if err := c.MountRootfsVolume(); err != nil {
		return err
	}
	return c.SetDNS()
}

func (c *Container) CleanupRootfs() error {
	if err := c.UmountRootfsVolume(); err != nil {
		return err
	}
	if err := c.DeleteRootfs(); err != nil {
		return err
	}
	return nil
}

func (c *Container) CreateRootfs() error {
	for _, value := range DriverConfigs[c.StorageDriver] {
		dir := path.Join(c.Rootfs.ContainerDir, value)

		if exist, err := util.FileOrDirExists(dir); err != nil {
			return fmt.Errorf("failed to check if the dir %s exists: %v",
				dir, err)
		} else if exist {
			continue
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to mkdir %s: %v", dir, err)
		}
	}

	return nil
}

func (c *Container) DeleteRootfs() error {
	if err := os.RemoveAll(c.Rootfs.ContainerDir); err != nil {
		return fmt.Errorf("failed to remove the dir %s: %v",
			c.Rootfs.MergeDir, err)
	}
	return nil
}

func (c *Container) MountRootfsVolume() error {
	if err := Drivers[c.StorageDriver].MountRootfs(c); err != nil {
		return err
	}
	if err := Drivers[c.StorageDriver].MountVolume(c); err != nil {
		return err
	}
	return nil
}

func (c *Container) UmountRootfsVolume() error {
	for _, volume := range c.Volumes {
		if err := util.Umount(volume.Target); err != nil {
			return err
		}
	}

	return util.Umount(c.Rootfs.MergeDir)
}
