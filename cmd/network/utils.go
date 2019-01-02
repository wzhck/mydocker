package network

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/pkg/network"
	"os"
	"text/tabwriter"
)

func listNetworks(_ *cli.Context) error {
	w := tabwriter.NewWriter(os.Stdout, 8, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIPNETS\tGATEWAY\tCOUNTS\tDRIVER\tCREATED\n")
	for _, nw := range network.Networks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			nw.Name,
			nw.IPNet.String(),
			nw.Gateway.IP.String(),
			nw.Counts,
			nw.Driver,
			nw.CreateTime,
		)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush %v", err)
	}
	return nil
}

func operateNetworks(ctx *cli.Context, action string) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("missing network's name")
	}

	var err error
	unknownErr := fmt.Errorf("unknown action: %s", action)
	for _, arg := range ctx.Args() {
		nw, ok := network.Networks[arg]
		if !ok {
			return fmt.Errorf("no such network: %s", arg)
		}

		switch action {
		case network.Delete:
			err = nw.Delete()
		default:
			err = unknownErr
		}

		if err != nil {
			return err
		}
	}

	return nil
}
