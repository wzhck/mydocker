package container

import (
	"fmt"
	"github.com/weikeit/mydocker/util"
	"os"
	"os/exec"
)

type AufsDriver struct{}

func (aufs *AufsDriver) Name() string {
	return Aufs
}

func (aufs *AufsDriver) Allowed() bool {
	return util.ModuleIsLoaded("aufs")
}

func (aufs *AufsDriver) MountRootfs(c *Container) error {
	if err := aufs.mountTmpfsForXino(); err != nil {
		return err
	}

	options := fmt.Sprintf("xino=%s/.xino,dirs=%s:%s",
		XinoTmpfs, c.Rootfs.WriteDir, c.Rootfs.ImageDir)
	cmd := exec.Command("mount", "-t", "aufs", "-o", options, "none", c.Rootfs.MergeDir)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount aufs: %v", err)
	}
	return nil
}

func (aufs *AufsDriver) MountVolume(c *Container) error {
	for source, target := range c.Volumes {
		if err := os.MkdirAll(source, 0755); err != nil {
			return fmt.Errorf("failed to mkdir %s: %v", source, err)
		}
		if err := os.MkdirAll(target, 0755); err != nil {
			return fmt.Errorf("failed to mkdir container volume dir %s: %v", target, err)
		}

		options := fmt.Sprintf("xino=%s/.xino,dirs=%s", XinoTmpfs, source)
		cmd := exec.Command("mount", "-t", "aufs", "-o", options, "none", target)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to mount local volume: %v", err)
		}
	}
	return nil
}

func (aufs *AufsDriver) mountTmpfsForXino() error {
	if util.DirIsMounted(XinoTmpfs) {
		return nil
	}

	if exist, _ := util.FileOrDirExists(XinoTmpfs); !exist {
		if err := os.MkdirAll(XinoTmpfs, 0755); err != nil {
			return err
		}
	}

	// aufs mount option xino=/path/to/.xino can't be xfs.
	// ref: https://sourceforge.net/p/aufs/mailman/message/25083283/
	cmd := exec.Command("mount", "-t", "tmpfs", "-o", "size=100M", "tmpfs", XinoTmpfs)
	return cmd.Run()
}
