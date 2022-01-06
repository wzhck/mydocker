package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"weike.sh/mydocker/util"
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
	for idx, repoTag := range tags {
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

		if idx == 0 {
			// call MakeRootfs only once for images with same uuid.
			if err := img.MakeRootfs(); err != nil {
				return fmt.Errorf("failed to make rootfs for image %s: %v",
					imageName, err)
			}
		}

		Images = append(Images, img)
	}

	return Dump()
}

func Delete(identifier string) error {
	thisImg, err := GetImageByNameOrUuid(identifier)
	if err != nil {
		return err
	}

	if thisImg.Counts > 0 {
		return fmt.Errorf("there still exist %d containers using the image %s",
			thisImg.Counts, thisImg.RepoTag)
	}

	// see: https://github.com/go101/go101/wiki/How-to-perfectly-clone-a-slice%3F
	tmpImages := append(Images[:0:0], Images...)
	Images = Images[:0]
	for _, img := range tmpImages {
		if img.Uuid != thisImg.Uuid {
			Images = append(Images, img)
		}
	}

	if err := Dump(); err != nil {
		return err
	}

	return os.RemoveAll(thisImg.RootDir())
}

func GetImageByNameOrUuid(identifier string) (*Image, error) {
	if err := Load(); err != nil {
		return nil, err
	}

	if identifier == "" {
		return nil, fmt.Errorf("missing image's name or uuid")
	}

	for _, img := range Images {
		if img.RepoTag == identifier ||
			img.RepoTag == identifier+":latest" ||
			img.Uuid == identifier {
			return img, nil
		}
	}

	return nil, fmt.Errorf("no such image: %s", identifier)
}

func ChangeCounts(identifier, action string) error {
	thisImg, err := GetImageByNameOrUuid(identifier)
	if err != nil {
		return err
	}

	// change all images with same uuid.
	for _, img := range Images {
		if img.Uuid == thisImg.Uuid {
			switch action {
			case "create":
				img.Counts++
			case "delete":
				img.Counts--
			}
		}
	}

	return Dump()
}

func Dump() error {
	if err := util.EnSureFileExists(ImagesConfigFile); err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(Images)
	if err != nil {
		return fmt.Errorf("failed to json-encode images: %v", err)
	}

	// WriteFile will create the file if it doesn't exist,
	// otherwise WriteFile will truncate it before writing
	if err := ioutil.WriteFile(ImagesConfigFile, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write images configs to file %s: %v",
			ImagesConfigFile, err)
	}

	return nil
}

func Load() error {
	if err := util.EnSureFileExists(ImagesConfigFile); err != nil {
		return err
	}

	jsonBytes, err := ioutil.ReadFile(ImagesConfigFile)
	if len(jsonBytes) == 0 {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read images configFile %s: %v",
			ImagesConfigFile, err)
	}

	if err := json.Unmarshal(jsonBytes, &Images); err != nil {
		return fmt.Errorf("failed to json-decode images: %v", err)
	}

	return nil
}
