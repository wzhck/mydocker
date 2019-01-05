package container

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// create all the device nodes in the container.
func createDevices() error {
	oldMask := syscall.Umask(0000)
	for _, device := range Devices {
		// TODO: containers running in a user namespace are not allowed
		// to mknod devices, so we can just bind mount it from the host
		if err := device.create(); err != nil {
			syscall.Umask(oldMask)
			return err
		}
	}
	syscall.Umask(oldMask)
	return nil
}

func (d *Device) mkdev() int {
	return int((d.Major << 8) | (d.Minor & 0xff) | ((d.Minor & 0xfff00) << 12))
}

func (d *Device) create() error {
	fileMode := d.FileMode
	switch d.Type {
	case 'c', 'u':
		fileMode |= syscall.S_IFCHR
	case 'b':
		fileMode |= syscall.S_IFBLK
	case 'p':
		fileMode |= syscall.S_IFIFO
	default:
		return fmt.Errorf("unknown type '%c' for device %s", d.Type, d.Path)
	}

	if err := os.MkdirAll(filepath.Dir(d.Path), 0755); err != nil {
		return err
	}
	if err := syscall.Mknod(d.Path, uint32(fileMode), d.mkdev()); err != nil {
		return err
	}
	return syscall.Chown(d.Path, int(d.Uid), int(d.Gid))
}
