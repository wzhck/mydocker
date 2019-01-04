package image

import (
	"encoding/json"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/weikeit/mydocker/util"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func Exist(imageName string) bool {
	if err := Load(); err != nil {
		return false
	}

	for _, img := range Images {
		if img.RepoTag == imageName {
			return true
		}
	}
	return false
}

func Pull(imageName string) error {
	if err := Load(); err != nil {
		return err
	}

	var cmd *exec.Cmd

	cmd = exec.Command("docker", "inspect", imageName)
	imageExist := cmd.Run() == nil

	cmd = exec.Command("docker", "pull", imageName)
	if !imageExist {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); !imageExist && err != nil {
		return fmt.Errorf("failed to pull image %s: %v",
			imageName, err)
	}

	fmtArgs := []string{
		"{{.Id}}",
		"{{.Size}}",
		"{{.Config.WorkingDir}}",
		"{{json .RepoTags}}",
		"{{json .Config.Entrypoint}}",
		"{{json .Config.Cmd}}",
		"{{json .Config.Env}}",
	}

	format := strings.Join(fmtArgs, "#")
	cmd = exec.Command("docker", "inspect", "-f", format, imageName)
	outBytes, err := cmd.Output()
	if err != nil {
		return err
	}

	// notes: need to remove the tailing newline \n.
	outs := strings.Split(strings.Trim(string(outBytes), "\n"), "#")

	byteSize, err := strconv.Atoi(outs[1])
	if err != nil {
		return err
	}

	size := datasize.ByteSize(byteSize).HumanReadable()

	var tags, epts, cmds, envs []string
	if err := json.Unmarshal([]byte(outs[3]), &tags); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(outs[4]), &epts); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(outs[5]), &cmds); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(outs[6]), &envs); err != nil {
		return err
	}

	// same image maybe have multiple repotags.
	for _, repoTag := range tags {
		img := &Image{
			// fetch the first 12 chars of sha256 checksum of image.
			Uuid:       outs[0][7:19],
			Size:       size,
			Counts:     0,
			WorkingDir: outs[2],
			RepoTag:    repoTag,
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
			Entrypoint: epts,
			Command:    cmds,
			Envs:       envs,
		}

		if err := img.MakeRootfs(); err != nil {
			return fmt.Errorf("failed to make rootfs for image %s: %v",
				imageName, err)
		}
		Images = append(Images, img)
	}

	return Dump()
}

func Delete(identifier string) error {
	if err := Load(); err != nil {
		return err
	}

	img, err := GetImageByNameOrUuid(identifier)
	if err != nil {
		return err
	}

	if img.Counts > 0 {
		return fmt.Errorf("there still exist %d containers using the image %s",
			img.Counts, img.RepoTag)
	}

	for idx, tmpImg := range Images {
		// two images' uuid maybe the same.
		if tmpImg.RepoTag == img.RepoTag {
			Images = append(Images[:idx], Images[idx+1:]...)
			break
		}
	}

	if err := Dump(); err != nil {
		return err
	}

	return os.RemoveAll(img.RootDir())
}

func GetImageByNameOrUuid(identifier string) (*Image, error) {
	if err := Load(); err != nil {
		return nil, err
	}

	if identifier == "" {
		return nil, fmt.Errorf("missing image's name or uuid")
	}

	for _, img := range Images {
		if img.Uuid == identifier || img.RepoTag == identifier ||
			img.RepoTag == identifier+":latest" {
			return img, nil
		}
	}

	return nil, fmt.Errorf("no such image: %s", identifier)
}

func Dump() error {
	if err := util.EnSureFileExists(ImagesConfigFile); err != nil {
		return err
	}

	flags := os.O_WRONLY | os.O_TRUNC | os.O_CREATE
	imagesConfigFile, err := os.OpenFile(ImagesConfigFile, int(flags), 0644)
	defer imagesConfigFile.Close()
	if err != nil {
		return err
	}

	imagesBytes, err := json.Marshal(Images)
	if err != nil {
		return fmt.Errorf("failed to encode images object using json: %v", err)
	}

	_, err = imagesConfigFile.Write(imagesBytes)
	return err
}

func Load() error {
	if err := util.EnSureFileExists(ImagesConfigFile); err != nil {
		return err
	}

	flags := os.O_RDONLY | os.O_CREATE
	configFile, err := os.OpenFile(ImagesConfigFile, int(flags), 0644)
	defer configFile.Close()
	if err != nil {
		return err
	}

	jsonBytes := make([]byte, MaxBytes)
	n, err := configFile.Read(jsonBytes)
	if n == 0 {
		return nil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonBytes[:n], &Images)
}

func ChangeCounts(repoTag, action string) error {
	if err := Load(); err != nil {
		return err
	}

	for _, img := range Images {
		if img.RepoTag == repoTag {
			switch action {
			case "create":
				img.Counts++
			case "delete":
				img.Counts--
			}
			return Dump()
		}
	}

	return nil
}
