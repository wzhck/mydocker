package cmd

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/pkg/container"
	"github.com/weikeit/mydocker/pkg/image"
	"github.com/weikeit/mydocker/pkg/network"
)

var Inspect = cli.Command{
	Name:  "inspect",
	Usage: "Print information of mydocker objects",
	Action: func(ctx *cli.Context) error {
		if len(ctx.Args()) < 1 {
			return fmt.Errorf("missing object's name or uuid")
		}

		for _, arg := range ctx.Args() {
			if c, err := container.GetContainerByNameOrUuid(arg); err == nil {
				showUp(c, "container", arg)
				continue
			}
			if nw, ok := network.Networks[arg]; ok {
				showUp(nw, "network", arg)
				continue
			}
			if img, err := image.GetImageByNameOrUuid(arg); err == nil {
				showUp(img, "image", arg)
				continue
			}
			fmt.Printf("\033[0;31mNo such a mydocker object: %s\033[0m\n", arg)
		}

		return nil
	},
}

func showUp(obj interface{}, cls, arg string) {
	jsonBytes, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		log.Errorf("failed to json-encode %s object %s: %v",
			cls, arg, err)
	}
	fmt.Printf("\033[0;32mShowing %s as a %s:\033[0m\n", arg, cls)
	fmt.Println(string(jsonBytes))
}
