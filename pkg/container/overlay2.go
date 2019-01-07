package container

import (
	"fmt"
	"github.com/weikeit/mydocker/util"
	"os"
	"os/exec"
	"path"
)

type Overlay2Driver struct{}

func (overlay2 *Overlay2Driver) Name() string {
	return Overlay2
}

func (overlay2 *Overlay2Driver) Allowed() bool {
	// Note: overlay2 module name is "overlay"
	return util.ModuleIsLoaded("overlay")
}

func (overlay2 *Overlay2Driver) MountRootfs(c *Container) error {
	workdir := path.Join(c.Rootfs.ContainerDir, DriverConfigs[Overlay2]["workDir"])

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		c.Rootfs.ImageDir, c.Rootfs.WriteDir, workdir)
	cmd := exec.Command("mount", "-t", "overlay", "-o", options, "overlay", c.Rootfs.MergeDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to mount overlay2: %v", err)
	}
	return nil
}

func (overlay2 *Overlay2Driver) MountVolume(c *Container) error {
	for source, target := range c.Volumes {
		hashed := util.Sha256Sum(source)
		volumeDir := path.Join(c.Rootfs.ContainerDir, "volumes", hashed[:8])
		lowerDir := path.Join(volumeDir, "lower")
		workdir := path.Join(volumeDir, "work")

		for _, dir := range []string{lowerDir, workdir, source, target} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to mkdir %s: %v", dir, err)
			}
		}

		options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
			lowerDir, source, workdir)
		cmd := exec.Command("mount", "-t", "overlay", "-o", options, "overlay", target)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to mount local volume: %v", err)
		}
	}
	return nil
}
