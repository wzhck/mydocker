package container

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"weike.sh/mydocker/pkg/cgroups"
	"weike.sh/mydocker/pkg/container"
)

var Init = cli.Command{
	Name:   "init",
	Usage:  "Run user's process in container. Do not call it outside!",
	Hidden: true,
	Action: func(ctx *cli.Context) error {
		log.Debugf("auto-calling initCommand...")
		return container.RunContainerInitProcess()
	},
}

var runFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "detach,d",
		Usage: "Run the container in background",
	},
	cli.StringFlag{
		Name:  "name,n",
		Usage: "Assign a name to the container",
	},
	cli.StringFlag{
		Name:  "hostname",
		Usage: "Set hostname in the container",
	},
	cli.StringSliceFlag{
		Name:  "dns",
		Usage: "Set DNS servers in the container",
		Value: &cli.StringSlice{"8.8.8.8", "8.8.4.4"},
	},
	cli.StringFlag{
		Name:  "image,i",
		Usage: "The image to be used (name or id)",
	},
	cli.StringSliceFlag{
		Name:  "env,e",
		Usage: "Set environment variables, e.g. -e key=value",
	},
	cli.StringSliceFlag{
		Name:  "volume,v",
		Usage: "Bind a local directory/file, e.g. -v /src:/dst",
	},
	cli.StringSliceFlag{
		Name:  "network,net",
		Usage: "Connect the container to a network (none to disable)",
	},
	cli.StringSliceFlag{
		Name:  "publish,p",
		Usage: "Publish the container's port(s) to the host",
	},
	cli.StringFlag{
		Name:  "storage-driver,s",
		Usage: "Storage driver to be used",
		Value: "overlay2",
	},
}

var Run = cli.Command{
	Name:  "run",
	Usage: "Create a new mydocker container",
	Flags: append(runFlags, cgroups.Flags...),
	Action: func(ctx *cli.Context) error {
		c, err := container.NewContainer(ctx)
		if err != nil {
			return err
		}
		return c.Run()
	},
}

var List = cli.Command{
	Name:  "ps",
	Usage: "List all containers on the host",
	Action: func(ctx *cli.Context) error {
		return listContainers(ctx)
	},
}

var Logs = cli.Command{
	Name:  "logs",
	Usage: "Show all the logs of a container",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "follow,f",
			Usage: "Follow the log's output",
		},
	},
	Action: func(ctx *cli.Context) error {
		c, err := getContainerFromArg(ctx)
		if err != nil {
			return err
		}
		return c.Logs(ctx)
	},
}

var Exec = cli.Command{
	Name:  "exec",
	Usage: "Run a command in a running container",
	Action: func(ctx *cli.Context) error {
		c, cmdArray, err := parseExecArgs(ctx)
		if err != nil {
			return err
		}
		if c == nil {
			return nil
		}
		return c.Exec(cmdArray)
	},
}

var Stop = cli.Command{
	Name:  container.Stop,
	Usage: "Stop one or more containers",
	Action: func(ctx *cli.Context) error {
		return operateContainers(ctx, container.Stop)
	},
}

var Start = cli.Command{
	Name:  container.Start,
	Usage: "Start one or more containers",
	Action: func(ctx *cli.Context) error {
		return operateContainers(ctx, container.Start)
	},
}

var Restart = cli.Command{
	Name:  container.Restart,
	Usage: "Restart one or more containers",
	Action: func(ctx *cli.Context) error {
		return operateContainers(ctx, container.Restart)
	},
}

var Remove = cli.Command{
	Name:  "rm",
	Usage: "Remove one or more containers",
	Action: func(ctx *cli.Context) error {
		return operateContainers(ctx, container.Delete)
	},
}
