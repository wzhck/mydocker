package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/util"
	"os"
	"os/exec"
	"path"
	"syscall"
)

func NewParentProcess(c *Container) (*exec.Cmd, *os.File, error) {
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

	if err := createContainerRootfs(c); err != nil {
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
		logFile, err = os.Open(logFileName)
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

	var envs []string
	for _, env := range c.Envs {
		envs = append(envs, fmt.Sprintf("%s=%s",
			env.Key, env.Value))
	}
	cmd.Env = append(os.Environ(), envs...)

	return cmd, writePipe, nil
}

func createContainerRootfs(c *Container) error {
	if err := createReadOnlyLayer(c); err != nil {
		return err
	}
	if err := createWriteLayer(c); err != nil {
		return err
	}
	if err := createMergeLayer(c); err != nil {
		return err
	}
	if err := mountMergeLayer(c); err != nil {
		return err
	}
	if len(c.Volumes) > 0 {
		if err := mountLocalVolumes(c); err != nil {
			return err
		}
	}
	return nil
}

func createReadOnlyLayer(c *Container) error {
	// NOTE: c.Rootfs.ReadOnlyDir equals c.Image.RootfsDir
	exist, err := util.FileOrDirExists(c.Rootfs.ReadOnlyDir)
	if err != nil {
		return fmt.Errorf("failed to check if the dir %s exists: %v",
			c.Rootfs.ReadOnlyDir, err)
	}
	if exist {
		return nil
	}

	if err := os.MkdirAll(c.Rootfs.ReadOnlyDir, 0777); err != nil {
		return fmt.Errorf("failed to mkdir %s: %v", c.Rootfs.ReadOnlyDir, err)
	}
	if _, err := exec.Command("tar", "-xvf", c.Image.TarFile,
		"-C", c.Image.RootfsDir).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to untar %s: %v", c.Rootfs.ReadOnlyDir, err)
	}
	return nil
}

func createWriteLayer(c *Container) error {
	exist, err := util.FileOrDirExists(c.Rootfs.WriteDir)
	if err != nil {
		return fmt.Errorf("failed to check if the dir %s exists: %v",
			c.Rootfs.WriteDir, err)
	}
	if exist {
		return nil
	}

	if err := os.MkdirAll(c.Rootfs.WriteDir, 0777); err != nil {
		return fmt.Errorf("failed to mkdir %s: %v", c.Rootfs.WriteDir, err)
	}
	return nil
}

func createMergeLayer(c *Container) error {
	exist, err := util.FileOrDirExists(c.Rootfs.MergeDir)
	if err != nil {
		return fmt.Errorf("failed to check if the dir %s exists: %v",
			c.Rootfs.MergeDir, err)
	}
	if exist {
		return nil
	}

	if err := os.MkdirAll(c.Rootfs.MergeDir, 0777); err != nil {
		return fmt.Errorf("failed to mkdir %s: %v", c.Rootfs.MergeDir, err)
	}
	return nil
}

func mountMergeLayer(c *Container) error {
	dirs := "dirs=" + c.Rootfs.WriteDir + ":" + c.Rootfs.ReadOnlyDir
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", c.Rootfs.MergeDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount aufs: %v", err)
	}
	return nil
}

func mountLocalVolumes(c *Container) error {
	for _, volume := range c.Volumes {
		if err := os.MkdirAll(volume.Source, 0777); err != nil {
			return fmt.Errorf("failed to mkdir %s: %v", volume.Source, err)
		}
		if err := os.MkdirAll(volume.Target, 0777); err != nil {
			return fmt.Errorf("failed to mkdir container volume dir %s: %v", volume.Target, err)
		}

		cmd := exec.Command("mount", "-t", "aufs", "-o", "dirs="+volume.Source, "none", volume.Target)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to mount local volume: %v", err)
		}
	}
	return nil
}

func deleteContainerRootfs(c *Container) error {
	log.Debug("cleaning container runtime files:")
	if len(c.Volumes) > 0 {
		if err := umountLocalVolumes(c); err != nil {
			return err
		}
	}
	if err := umountMergeLayer(c); err != nil {
		return err
	}
	if err := deleteMergeLayer(c); err != nil {
		return err
	}
	if err := deleteWriteLayer(c); err != nil {
		return err
	}
	return nil
}

func umountLocalVolumes(c *Container) error {
	for _, volume := range c.Volumes {
		if err := umount(volume.Target); err != nil {
			return err
		}
	}
	return nil
}

func umountMergeLayer(c *Container) error {
	return umount(c.Rootfs.MergeDir)
}

func deleteMergeLayer(c *Container) error {
	log.Debugf("removing the container rootfs: %s", c.Rootfs.MergeDir)
	// NOTE: c.Rootfs.ContainerDir which contains c.Rootfs.MergeDir
	if err := os.RemoveAll(c.Rootfs.ContainerDir); err != nil {
		return fmt.Errorf("failed to remove the dir %s: %v", c.Rootfs.MergeDir, err)
	}
	return nil
}

func deleteWriteLayer(c *Container) error {
	log.Debugf("removing the container writelayer: %s", c.Rootfs.WriteDir)
	if err := os.RemoveAll(c.Rootfs.WriteDir); err != nil {
		return fmt.Errorf("failed to remove dir %s: %v", c.Rootfs.WriteDir, err)
	}
	return nil
}

func umount(mntPoint string) error {
	if exist, _ := util.FileOrDirExists(mntPoint); !exist {
		return nil
	} else if !util.DirIsMounted(mntPoint) {
		return nil
	}

	cmd := exec.Command("umount", "-f", mntPoint)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Debugf("umounting the directory: %s", mntPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to umount the directory %s: %v", mntPoint, err)
	}
	return nil
}
