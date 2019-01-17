package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/util"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
)

func RunContainerInitProcess() error {
	cmds := receiveInitCommand()
	if cmds == nil || len(cmds) == 0 {
		return fmt.Errorf("failed to run user's command in container: cmd is nil")
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current work directory: %v", err)
	}

	if err := pivotRoot(root); err != nil {
		return err
	}

	if err := mountVFS(); err != nil {
		return err
	}

	if err := createDevices(); err != nil {
		return err
	}

	if err := setupDevSymlinks(); err != nil {
		return err
	}

	path, err := exec.LookPath(cmds[0])
	if err != nil {
		return fmt.Errorf("failed to find the executable file's path of %s: %v",
			cmds[0], err)
	} else {
		log.Debugf("find the executable file's path: %s", path)
	}

	if err := syscall.Exec(path, cmds, os.Environ()); err != nil {
		log.Errorf(err.Error())
	}

	return nil
}

func pivotRoot(root string) error {
	// note: runc use the flags MS_SLAVE and MS_REC.
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to enable the mount namespace work properly: %v", err)
	}

	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to mount rootfs to itself: %v", err)
	}

	pivotDir := filepath.Join(root, ".oldroot")
	if err := os.Mkdir(pivotDir, 0700); err != nil {
		return fmt.Errorf("failed to mkdir old_root %s: %v", pivotDir, err)
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("failed to syscall pivot_root: %v", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("failed to syscall chdir /: %v", err)
	}

	pivotDir = filepath.Join("/", ".oldroot")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount pivot_root dir: %v", err)
	}

	// note: need to delete the origin directory /dev
	for _, dir := range []string{pivotDir, "/dev",} {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}

	return nil
}

func mountVFS() error {
	for _, m := range Mounts {
		if m.Data != "" {
			re, err := regexp.Compile(`(?:^|\W+)mode=(\d{4})`)
			if err != nil {
				panic(err)
			}

			results := re.FindAllStringSubmatch(m.Data, -1)
			if len(results) == 1 {
				mode, _ := strconv.ParseInt(results[0][1], 8, 32)
				// log.Debugf("create the dir %s with mode %o", m.Target, mode)
				if err := os.MkdirAll(m.Target, os.FileMode(mode)); err != nil {
					return fmt.Errorf("failed to mkdir %s: %v", m.Target, err)
				}
			}
		}

		flags := uintptr(m.Flags)
		if err := syscall.Mount(m.Source, m.Target, m.Fstype, flags, m.Data); err != nil {
			return fmt.Errorf("failed to mount %s: %v", m.Target, err)
		}
	}

	return nil
}

func setupDevSymlinks() error {
	// kcore support can be toggled with CONFIG_PROC_KCORE;
	// only create a symlink in /dev if it exists in /proc.
	if exist, _ := util.FileOrDirExists("/proc/kcore"); exist {
		DevSymlinks["/proc/kcore"] = "/dev/core"
	}

	// /dev/ptmx is a character device file that is used
	// to create a pseudo terminal master-slave pair.
	if exist, _ := util.FileOrDirExists("/dev/pts/ptmx"); exist {
		DevSymlinks["/dev/pts/ptmx"] = "/dev/ptmx"
	}

	for src, dst := range DevSymlinks {
		if err := os.Symlink(src, dst); err != nil && !os.IsExist(err) {
			return fmt.Errorf("failed to create symlink %s -> %s: %v",
				dst, src, err)
		}
	}

	return nil
}
