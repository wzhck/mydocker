package image

import (
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:  "image",
	Usage: "Manage container images",
	Subcommands: []cli.Command{
		Pull,
		Remove,
		List,
	},
}

var (
	Pull = cli.Command{
		Name:   "pull",
		Usage:  "Pull an image from a registry",
		Action: pull,
	}

	Remove = cli.Command{
		Name:   "rm",
		Usage:  "Remove one or more images",
		Action: remove,
	}

	RemoveImages = cli.Command{
		Name:   "rmi",
		Usage:  "Remove one or more images",
		Action: remove,
	}

	List = cli.Command{
		Name:   "list",
		Usage:  "List images on the host",
		Action: list,
	}

	ListImages = cli.Command{
		Name:   "images",
		Usage:  "List images on the host",
		Action: list,
	}
)
