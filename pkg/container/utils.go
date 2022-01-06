package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"weike.sh/mydocker/util"
)

func sendInitCommand(cmds []string, writePipe *os.File) {
	cmdsStr := strings.Join(cmds, "\u0000")
	log.Debugf("runCommand sends user-defined command: %s",
		strings.Replace(cmdsStr, "\u0000", " ", -1))
	writePipe.WriteString(cmdsStr)
	writePipe.Close()
}

func receiveInitCommand() []string {
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

func GetAllContainers() ([]*Container, error) {
	exist, _ := util.FileOrDirExists(ContainersDir)
	if !exist {
		if err := os.MkdirAll(ContainersDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to mkdir %s: %v", ContainersDir, err)
		}
		return nil, nil
	}

	containerDirs, err := ioutil.ReadDir(ContainersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %v", ContainersDir, err)
	}

	var containers []*Container
	for _, containerDir := range containerDirs {
		uuid := containerDir.Name()
		c := &Container{Uuid: uuid}
		if err := c.Load(); err != nil {
			log.Errorf("failed to get the info of container %s: %v", uuid, err)
			continue
		}
		containers = append(containers, c)
	}

	return containers, nil
}

func GetContainerByNameOrUuid(identifier string) (*Container, error) {
	allContainers, err := GetAllContainers()
	if err != nil {
		return nil, err
	}

	var c *Container
	for _, container := range allContainers {
		if identifier == container.Name || identifier == container.Uuid {
			c = container
			break
		}
	}

	if c == nil {
		return nil, fmt.Errorf("no such container: %s", identifier)
	}

	return c, nil
}

// cgroupPaths = {"cpu,cpuacct": "/mydocker/c627f5f58033", ..., }
func getContainerCgroupPaths() (map[string]string, error) {
	contentsBytes, err := ioutil.ReadFile("/proc/self/cgroup")
	if err != nil {
		return nil, err
	}

	cgroupPaths := make(map[string]string)
	// the info should be `9:cpuset:/mydocker/c627f5f58033`
	restr := fmt.Sprintf(`(\d+):([^:]+):(/%s/[0-9a-f]+)`, MyDocker)
	re, _ := regexp.Compile(restr)
	for _, line := range strings.Split(string(contentsBytes), "\n") {
		if !re.MatchString(line) {
			continue
		}
		parts := re.FindStringSubmatch(line)
		cgroupPaths[parts[2]] = parts[3]
	}

	return cgroupPaths, nil
}
