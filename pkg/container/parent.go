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

	if err := c.prepareRootfs(); err != nil {
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
