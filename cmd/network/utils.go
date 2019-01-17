package network

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/weikeit/mydocker/pkg/container"
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
		case "delete":
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

func handleConnection(ctx *cli.Context, action string) error {
	switch len(ctx.Args()) {
	case 0:
		return fmt.Errorf("missing network's name")
	case 1:
		return fmt.Errorf("missing container's name or uuid")
	}

	nwName := ctx.Args().Get(0)
	nw, ok := network.Networks[nwName]
	if !ok {
		return fmt.Errorf("no such network %s", nwName)
	}

	cName := ctx.Args().Get(1)
	c, err := container.GetContainerByNameOrUuid(cName)
	if err != nil {
		return fmt.Errorf("failed to get the container %s: %v",
			cName, err)
	}

	if c.Status != container.Running {
		return fmt.Errorf("the container %s is not running", c.Uuid)
	}

	nwExist := false
	for _, ep := range c.Endpoints {
		if ep.Network.Name == nwName {
			nwExist = true
			break
		}
	}

	switch action {
	case "create":
		if nwExist {
			// container can't bind multiple endpoints of same network.
			return fmt.Errorf("the container %s has been connected to "+
				"the network %s", c.Uuid, nwName)
		}

		nwNames := []string{nwName}
		eps, err := container.CreateEndpoints(c.Name, nwNames, c.Ports)
		if err != nil {
			return err
		}

		// there's only one endpoint to be added.
		if err := eps[0].Connect(c.Cgroups.Pid); err != nil {
			// note: need to release the ipaddr if failed.
			return network.IPAllocator.Release(nw, &eps[0].IPAddr)
		}

		// update container's Endpoints finally.
		c.Endpoints = append(c.Endpoints, eps[0])

	case "delete":
		if !nwExist {
			return fmt.Errorf("the container %s has no endpoint "+
				"connected to the network %s", c.Uuid, nwName)
		}

		// backup the container's origin Endpoints.
		tmpEndpoints := append(c.Endpoints[:0:0], c.Endpoints...)
		c.Endpoints = c.Endpoints[:0]
		for _, ep := range tmpEndpoints {
			// first, disconnect all the endpoints.
			if err := ep.DisConnect(c.Cgroups.Pid); err != nil {
				return err
			}
			if ep.Network.Name == nwName {
				// note: don't forget to release the ipaddr.
				if err := network.IPAllocator.Release(nw, &ep.IPAddr);
					err != nil {
					return err
				}
			} else {
				// keep all the remaining endpoints.
				c.Endpoints = append(c.Endpoints, ep)
			}
		}

		for _, ep := range c.Endpoints {
			// then, connect all the remaining endpoints.
			if err := ep.Connect(c.Cgroups.Pid); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unknown action %s", action)
	}

	return c.Dump()
}
