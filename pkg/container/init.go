package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"weike.sh/mydocker/util"
)

func RunContainerInitProcess() error {
	cmds := receiveInitCommand()
	if cmds == nil || len(cmds) == 0 {
		return fmt.Errorf("missing command to be executed in container")
	}

	initFuncs := []func() error{
		pivotRoot,
		mountVFS,
		createDevices,
		createDevSymlinks,
		mountCgroups,
		setHostname,
	}

	for _, initFunc := range initFuncs {
		if err := initFunc(); err != nil {
			return err
		}
	}

	cmdPath, err := exec.LookPath(cmds[0])
	if err != nil {
		return fmt.Errorf("failed to find the executable file's "+
			"path of %s: %v", cmds[0], err)
	} else {
		log.Debugf("find the executable file's path: %s", cmdPath)
	}

	if err := syscall.Exec(cmdPath, cmds, os.Environ()); err != nil {
		return fmt.Errorf("failed to call execve syscall: %v", err)
	}

	return nil
}

func pivotRoot() error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current work directory: %v", err)
	}

	// note: runc use the flags MS_SLAVE and MS_REC.
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to enable the mount namespace work properly: %v", err)
	}

	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to mount rootfs to itself: %v", err)
	}

	pivotDir := path.Join(root, ".oldroot")
	if err := os.Mkdir(pivotDir, 0700); err != nil {
		return fmt.Errorf("failed to mkdir old_root %s: %v", pivotDir, err)
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("failed to syscall pivot_root: %v", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("failed to syscall chdir /: %v", err)
	}

	pivotDir = path.Join("/", ".oldroot")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount old root dir: %v", err)
	}

	// note: need to delete the origin directory /dev
	for _, dir := range []string{pivotDir, "/dev"} {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}

	return nil
}

func mountVFS() error {
	for _, m := range Mounts {
		if m.Data != "" {
			re, _ := regexp.Compile(`(?:^|\W+)mode=(\d+)`)
			results := re.FindStringSubmatch(m.Data)
			if len(results) == 2 {
				mode, _ := strconv.ParseInt(results[1], 8, 32)
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

func createDevSymlinks() error {
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

func mountCgroups() error {
	const (
		tmpCgroup  = "/tmp/cgroup"
		cgroupRoot = "/sys/fs/cgroup"
	)

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(cgroupRoot); err != nil {
		return err
	}

	if err := os.MkdirAll(tmpCgroup, 0755); err != nil {
		return err
	}

	defaultFlags := uintptr(defaultMountFlags)
	if err := syscall.Mount("tmpfs", tmpCgroup, "tmpfs", defaultFlags, ""); err != nil {
		return fmt.Errorf("failed to mount %s: %v", tmpCgroup, err)
	}

	defer func() {
		exec.Command("umount", tmpCgroup).Run()
		os.RemoveAll(tmpCgroup)
	}()

	cgroupPaths, err := getContainerCgroupPaths()
	if err != nil {
		return err
	}

	for subsystemName, cgroupPath := range cgroupPaths {
		tmpSubsystemDir := path.Join(tmpCgroup, subsystemName)
		subsystemDir := path.Join(cgroupRoot, subsystemName)

		for _, dir := range []string{tmpSubsystemDir, subsystemDir} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to mkdir %s: %v", dir, err)
			}
		}

		options := fmt.Sprintf("-t cgroup -o nosuid,nodev,noexec,ro,%s cgroup %s",
			subsystemName, tmpSubsystemDir)
		if err := exec.Command("mount", strings.Split(options, " ")...).Run(); err != nil {
			return fmt.Errorf("failed to execute `mount %s`: %v", options, err)
		}

		containerSubsystemDir := path.Join(tmpSubsystemDir, cgroupPath)
		options = fmt.Sprintf("-o ro,bind %s %s", containerSubsystemDir, subsystemDir)
		if err := exec.Command("mount", strings.Split(options, " ")...).Run(); err != nil {
			return fmt.Errorf("failed to execute `mount %s`: %v", options, err)
		}

		if err := exec.Command("umount", tmpSubsystemDir).Run(); err != nil {
			return fmt.Errorf("failed to umount %s: %v", tmpSubsystemDir, err)
		}

		if err := os.RemoveAll(tmpSubsystemDir); err != nil {
			return fmt.Errorf("failed to remove %s: %v", tmpSubsystemDir, err)
		}

		if strings.Contains(subsystemName, ",") {
			for _, subs := range strings.Split(subsystemName, ",") {
				if err := os.Symlink(subsystemName, subs); err != nil {
					return fmt.Errorf("failed to create symlink %s => %s: %v",
						subs, subsystemName, err)
				}
			}
		}
	}

	return nil
}

func setHostname() error {
	hostname, err := ioutil.ReadFile("/etc/hostname")
	if err != nil {
		return fmt.Errorf("failed to read /etc/hostname")
	}

	if err := syscall.Sethostname(hostname[:len(hostname)-1]); err != nil {
		return fmt.Errorf("failed to sethostname in container: %s", err)
	}

	return nil
}
