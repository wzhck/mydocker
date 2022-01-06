package container

import (
	"fmt"
	"os"
	"path"

	"weike.sh/mydocker/util"
)

func (c *Container) prepareRootfs() error {
	if err := c.createRootfs(); err != nil {
		return err
	}

	if err := c.mountRootfsVolume(); err != nil {
		return err
	}

	if err := c.configHostname(); err != nil {
		return err
	}

	if err := c.configDNS(); err != nil {
		return err
	}

	return nil
}

func (c *Container) cleanupRootfs() error {
	if err := c.umountRootfsVolume(); err != nil {
		return err
	}
	if err := c.deleteRootfs(); err != nil {
		return err
	}
	return nil
}

func (c *Container) createRootfs() error {
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

func (c *Container) deleteRootfs() error {
	if err := os.RemoveAll(c.Rootfs.ContainerDir); err != nil {
		return fmt.Errorf("failed to remove the dir %s: %v",
			c.Rootfs.MergeDir, err)
	}
	return nil
}

func (c *Container) mountRootfsVolume() error {
	if err := Drivers[c.StorageDriver].MountRootfs(c); err != nil {
		return err
	}
	if err := Drivers[c.StorageDriver].MountVolume(c); err != nil {
		return err
	}
	return nil
}

func (c *Container) umountRootfsVolume() error {
	for _, target := range c.Volumes {
		if err := util.Umount(target); err != nil {
			return err
		}
	}

	return util.Umount(c.Rootfs.MergeDir)
}
