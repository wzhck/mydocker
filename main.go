package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/cmd"
	"github.com/weikeit/mydocker/cmd/container"
	"github.com/weikeit/mydocker/cmd/image"
	"github.com/weikeit/mydocker/cmd/network"
	_ "github.com/weikeit/mydocker/pkg/init"
	"os"
)

const usage = `mydocker is a simple container runtime implementation.
The purpose of this project is to learn how docker works and how to
write a docker-like container runtime by ourselves, enjoy it!`

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage

	app.Commands = []cli.Command{
		container.Init,
		container.Run,
		container.List,
		container.Logs,
		container.Exec,
		container.Stop,
		container.Start,
		container.Restart,
		container.Remove,
		network.RemoveNetworks,
		image.RemoveImages,
		image.Pull,
		cmd.Inspect,
		network.ListNetworks,
		image.ListImages,
		network.Command,
		image.Command,
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "print mydocker debug logs",
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
