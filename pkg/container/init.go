package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {
	cmds := readUserCommand()
	if cmds == nil || len(cmds) == 0 {
		return fmt.Errorf("failed to run user's command in container, cmds is nil")
	}

	if err := mountVFS(); err != nil {
		return err
	}

	path, err := exec.LookPath(cmds[0])
	if err != nil {
		log.Errorf("failed to find the executable file's path of <%s>:", cmds[0])
		return err
	}

	log.Debugf("find the executable file's path: %s", path)

	if err := syscall.Exec(path, cmds, os.Environ()); err != nil {
		log.Errorf(err.Error())
	}

	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("failed to init read pipe: %v", err)
		return nil
	}

	cmdsStr := string(msg)
	log.Debugf("initCommand receives user-defined command: %s",
		strings.Replace(cmdsStr, "\u0000", " ", -1))
	return strings.Split(cmdsStr, "\u0000")
}

func mountVFS() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current work directory: %v", err)
	}

	log.Debugf("current work directory is %s", pwd)
	if err := pivotRoot(pwd); err != nil {
		return fmt.Errorf("failed to pivot_root for current container: %v", err)
	}

	procMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(procMountFlags), ""); err != nil {
		return fmt.Errorf("failed to mount procfs: %v", err)
	}

	tmpfsMountFlags := syscall.MS_NOSUID | syscall.MS_STRICTATIME
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", uintptr(tmpfsMountFlags), ""); err != nil {
		return fmt.Errorf("failed to mount tmpfs: %v", err)
	}

	return nil
}

func pivotRoot(root string) error {
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

	return os.RemoveAll(pivotDir)
}
