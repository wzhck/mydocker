package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"weike.sh/mydocker/pkg/cmd"
	"weike.sh/mydocker/pkg/cmd/container"
	"weike.sh/mydocker/pkg/cmd/image"
	"weike.sh/mydocker/pkg/cmd/network"
	netpkg "weike.sh/mydocker/pkg/network"
)

const usage = `mydocker is a simple container runtime implementation.
The purpose of this project is to learn how docker works and how to
write a docker-like container runtime by ourselves, enjoy it!`

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
		debug := ctx.Bool("debug")
		os.Setenv("debug", fmt.Sprintf("%t", debug))
		if debug {
			log.SetLevel(log.DebugLevel)
		}

		log.SetOutput(os.Stdout)
		log.SetFormatter(&prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})

		// notes: command `mydocker init` is called by
		// `mydocker run` implicitly. so, we can't use
		// `import _ /path/to/init/pkg` to call init()
		if ctx.Args().Get(0) != container.Init.Name {
			return netpkg.Init()
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
