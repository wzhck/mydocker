package image

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/Pallinder/go-randomdata"
	"weike.sh/mydocker/util"
)

func (img *Image) RootDir() string {
	return path.Join(ImagesDir, img.Uuid)
}

func (img *Image) MakeRootfs() error {
	var cmd *exec.Cmd

	containerName := randomdata.SillyName()
	cmd = exec.Command("docker", "run", "-d", "--name", containerName, img.Uuid)
	defer func() {
		exec.Command("docker", "rm", "-f", containerName).Run()
	}()

	if err := cmd.Run(); err != nil {
		return err
	}

	rootfsTarFile := fmt.Sprintf("/tmp/%s.tar", containerName)
	cmd = exec.Command("docker", "export", "-o", rootfsTarFile, containerName)
	defer func() {
		os.Remove(rootfsTarFile)
	}()

	if err := cmd.Run(); err != nil {
		return err
	}

	if exist, _ := util.FileOrDirExists(img.RootDir()); !exist {
		if err := os.MkdirAll(img.RootDir(), 0755); err != nil {
			return fmt.Errorf("failed to mkdir %s for image %s: %v",
				img.RootDir(), img.RepoTag, err)
		}
	}

	cmd = exec.Command("tar", "-xf", rootfsTarFile, "-C", img.RootDir())
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
