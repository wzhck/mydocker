package util

import (
	"crypto/sha256"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
)

func PrintExeFile(pid int) {
	output, err := exec.Command("/bin/readlink",
		fmt.Sprintf("/proc/%d/exe", pid)).Output()
	if err != nil {
		log.Warnf("failed to readlink the /proc/%d/exe", pid)
	}
	log.Debugf("the executable file of pid [%d] is %s", pid,
		strings.Trim(string(output), "\n"))
}

func FileOrDirExists(fileOrDir string) (bool, error) {
	_, err := os.Stat(fileOrDir)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func EnSureFileExists(fileName string) error {
	dir, _ := path.Split(fileName)
	exist, err := FileOrDirExists(dir)
	if err != nil {
		return err
	}
	if !exist {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
	defer file.Close()
	return err
}

func DirIsMounted(dir string) bool {
	args := fmt.Sprintf("mount | grep -qw %s", dir)
	return exec.Command("bash", "-c", args).Run() == nil
}

func ModuleIsLoaded(module string) bool {
	args := fmt.Sprintf("lsmod | grep -qw %s", module)
	return exec.Command("bash", "-c", args).Run() == nil
}

func Umount(mntPoint string) error {
	if exist, _ := FileOrDirExists(mntPoint); !exist {
		return nil
	} else if !DirIsMounted(mntPoint) {
		return nil
	}

	log.Debugf("umounting the dir: %s", mntPoint)
	cmd := exec.Command("umount", "-f", mntPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to umount the dir %s: %v",
			mntPoint, err)
	}
	return nil
}

func GetEnvsByPid(pid int) ([]string, error) {
	envFile := fmt.Sprintf("/proc/%d/environ", pid)
	envsBytes, err := ioutil.ReadFile(envFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read envfile %s: %v",
			envFile, err)
	}
	return strings.Split(string(envsBytes), "\u0000"), nil
}

func Sha256Sum(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func Uniq(items []string) []string {
	sort.Strings(items)
	j := 0
	for i := 0; i < len(items); i++ {
		if items[j] == items[i] {
			continue
		}
		j++
		items[j] = items[i]
	}
	return items[:j+1]
}

func Contains(items []string, ele string) bool {
	for _, value := range items {
		if value == ele {
			return true
		}
	}
	return false
}
