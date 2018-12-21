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
	// init函数会一直阻塞在这里，直到父进程把容器的启动命令和参数写入管道
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

	// 此时，pid=1的进程并不是用户的进程，而是init初始化的进程。
	// syscall.Exec()系统调用其实是调用了Linux的execve()这个系统函数，它的作用是：
	// 执行当前filename对应的程序，它会覆盖当前进程的镜像，数据和堆栈等信息，包括pid
	// 这些都会被将要运行的进程覆盖掉。
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
	msgStr := string(msg)
	log.Debugf("initCommand receives user-defined command: %s", msgStr)
	return strings.Split(msgStr, " ")
}

func mountVFS() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current work directory: %v", err)
	}

	// enable the mount namespace works properly on archlinux computer
	// syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")

	log.Debugf("current work directory is %s", pwd)
	if err := pivotRoot(pwd); err != nil {
		return fmt.Errorf("failed to pivot_root for current container: %v", err)
	}

	procMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(procMountFlags), ""); err != nil {
		return fmt.Errorf("failed to mount procfs: %v", err)
	}

	// tmpfs是一种基于内存的文件系统，可以使用RAM或者swap分区来存储。
	tmpfsMountFlags := syscall.MS_NOSUID | syscall.MS_STRICTATIME
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", uintptr(tmpfsMountFlags), ""); err != nil {
		return fmt.Errorf("failed to mount tmpfs: %v", err)
	}

	return nil
}

// pivot_root是一个系统调用，主要功能是改变当前的root文件系统。
// pivot_root可以将当前进程的root文件系统移到old_root目录，然后使new_root成为新的root文件系统。
// new_root和old_root必须不能同时在当前的root文件系统中。
// pivot_root和chroot的区别是：pivot_root是把整个系统切换到一个新的root目录下，而移除对之前root文件系统
// 的依赖，这样就能够umount原来的root文件系统；而chroot是针对某个进程，系统的其它进程依然依赖原来的root文件系统。
func pivotRoot(root string) error {
	// 为了使当前root的old_root和new_root不在同一个文件系统下，我们把root重新mount了一次，
	// bind mount是把相同的内容换了一个挂载点的挂载方法
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to mount rootfs to itself: %v", err)
	}

	// 创建rootfs/.pivot_root存储old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0700); err != nil {
		return fmt.Errorf("failed to mkdir old_root %s: %v", pivotDir, err)
	}

	// pivot_root到新的rootfs,现在老的old_root是挂载在rootfs/.pivot_root
	// 挂载点现在依然可以在mount命令中看到
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("failed to syscall pivot_root: %v", err)
	}

	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("failed to syscall chdir /: %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount pivot_root dir: %v", err)
	}

	// 删除临时目录
	return os.RemoveAll(pivotDir)
}
