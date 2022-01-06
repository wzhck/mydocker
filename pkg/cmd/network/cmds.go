package network

import (
	"github.com/urfave/cli"
	"weike.sh/mydocker/pkg/network"
)

var Command = cli.Command{
	Name:  "network",
	Usage: "Manage container networks",
	Subcommands: []cli.Command{
		Create,
		Remove,
		List,
		Connect,
		DisConnect,
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

var Remove = cli.Command{
	Name:  "rm",
	Usage: "Remove one or more networks",
	Action: func(ctx *cli.Context) error {
		return operateNetworks(ctx, "delete")
	},
}

var RemoveNetworks = cli.Command{
	Name:  "rmn",
	Usage: "Remove one or more networks",
	Action: func(ctx *cli.Context) error {
		return operateNetworks(ctx, "delete")
	},
}

var List = cli.Command{
	Name:  "list",
	Usage: "List networks on the host",
	Action: func(ctx *cli.Context) error {
		return listNetworks(ctx)
	},
}

var ListNetworks = cli.Command{
	Name:  "networks",
	Usage: "List networks on the host",
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
