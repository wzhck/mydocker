package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/cmd/container"
	"github.com/weikeit/mydocker/cmd/network"
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
		container.Delete,
		network.Command,
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Print mydocker debug logs",
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		log.SetFormatter(&prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
