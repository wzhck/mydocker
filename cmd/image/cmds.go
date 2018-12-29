package image

import (
	"github.com/urfave/cli"
)

var Command = cli.Command{
	Name:  "image",
	Usage: "Manage container images",
	Subcommands: []cli.Command{
		Pull,
		Delete,
		List,
	},
}

var Pull = cli.Command{
	Name:  "pull",
	Usage: "Pull an image from a registry",
	Action: func(ctx *cli.Context) error {
		return pull(ctx)
	},
}

var Delete = cli.Command{
	Name:  "delete",
	Usage: "Remove one or more images",
	Action: func(ctx *cli.Context) error {
		return remove(ctx)
	},
}

var List = cli.Command{
	Name:  "list",
	Usage: "List images on the host",
	Action: func(ctx *cli.Context) error {
		return list(ctx)
	},
}
