package container

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/weikeit/mydocker/util"
	"io/ioutil"
	"os"
	"strings"
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
	if ! exist {
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
