package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/cmd"
	"github.com/weikeit/mydocker/cmd/container"
	"github.com/weikeit/mydocker/cmd/image"
	"github.com/weikeit/mydocker/cmd/network"
	netpkg "github.com/weikeit/mydocker/pkg/network"
	"github.com/x-cray/logrus-prefixed-formatter"
	"math/rand"
	"os"
	"time"
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
