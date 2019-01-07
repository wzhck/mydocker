package network

import (
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/pkg/network"
)

var Command = cli.Command{
	Name:  "network",
	Usage: "Manage container networks",
	Subcommands: []cli.Command{
		Init,
		Create,
		Delete,
		List,
		Connect,
		DisConnect,
	},
}

var Init = cli.Command{
	Name:  "init",
	Usage: "init all the existed networks",
	Action: func(ctx *cli.Context) error {
		return network.Init()
	},
}

var Create = cli.Command{
	Name:  "create",
	Usage: "Create a new container network",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "driver,d",
			Usage: "The network driver to use",
			Value: network.Bridge,
		},
		cli.StringFlag{
			Name:  "subnet,s",
			Usage: "The cidr of the new network, e.g. 10.10.0.0/24",
		},
	},
	Action: func(ctx *cli.Context) error {
		nw, err := network.NewNetwork(ctx)
		if err != nil {
			return err
		}

		return nw.Create()
	},
}

var Delete = cli.Command{
	Name:  "delete",
	Usage: "Delete one or more container networks",
	Action: func(ctx *cli.Context) error {
		return operateNetworks(ctx, "delete")
	},
}

var List = cli.Command{
	Name:  "list",
	Usage: "List all container networks",
	Action: func(ctx *cli.Context) error {
		return listNetworks(ctx)
	},
}

var Connect = cli.Command{
	Name:  "connect",
	Usage: "Connect a container to a network",
	Action: func(ctx *cli.Context) error {
		return handleConnection(ctx, "create")
	},
}

var DisConnect = cli.Command{
	Name:  "disconnect",
	Usage: "Disconnect a container from a network",
	Action: func(ctx *cli.Context) error {
		return handleConnection(ctx, "delete")
	},
}
