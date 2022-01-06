package container

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"
	"weike.sh/mydocker/pkg/container"
)

func listContainers(_ *cli.Context) error {
	allContainers, err := container.GetAllContainers()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 1, 3, ' ', 0)
	fmt.Fprintf(w, "CONTAINER ID\tNAME\tIMAGE\tSTATUS\tDRIVER\tPID\tCOMMAND\tIPS\tPORTS\tCREATED\n")
	for _, c := range allContainers {
		var ipaddrs, ports string

		if len(c.Endpoints) > 0 {
			for _, ep := range c.Endpoints {
				ipaddrs += fmt.Sprintf("%s, ", ep.IPAddr)
			}

			// the ports of all endpoints are the same.
			for out, in := range c.Endpoints[0].Ports {
				ports += fmt.Sprintf("%s->%s, ", out, in)
			}

			ipaddrs = strings.TrimRight(ipaddrs, ", ")
			ports = strings.TrimRight(ports, ", ")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
			c.Uuid,
			c.Name,
			c.Image,
			c.Status,
			c.StorageDriver,
			c.Cgroups.Pid,
			c.Commands,
			ipaddrs,
			ports,
			c.CreateTime)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %v", err)
	}

	return nil
}

func getContainerFromArg(ctx *cli.Context) (*container.Container, error) {
	if len(ctx.Args()) < 1 {
		return nil, fmt.Errorf("missing container's name or uuid")
	}

	identifier := ctx.Args().Get(0)
	c, err := container.GetContainerByNameOrUuid(identifier)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func parseExecArgs(ctx *cli.Context) (*container.Container, []string, error) {
	switch argsLen := len(ctx.Args()); argsLen {
	case 0:
		return nil, nil, fmt.Errorf("missing container's name/uuid and command")
	case 1:
		return nil, nil, fmt.Errorf("missing command to be executed")
	}

	identifier := ctx.Args().Get(0)
	c, err := container.GetContainerByNameOrUuid(identifier)
	if err != nil {
		return nil, nil, err
	}

	if c == nil {
		return nil, nil, fmt.Errorf("invalid container name or uuid: %s", identifier)
	}

	if c.Status != container.Running {
		return nil, nil, fmt.Errorf("the container %s is %s, not running", identifier, c.Status)
	}

	var cmdArray []string
	for _, arg := range ctx.Args().Tail() {
		if arg != "--" {
			// if the arg contains space e.g. -a '1 2 3'
			// we need to treat the whole string "1 2 3"
			// as one argument by adding a pair of single
			// quotes, because Exec() will use a space to
			// seperate all standalone argumenmts.
			if strings.Contains(arg, " ") {
				arg = fmt.Sprintf("'%s'", arg)
			}
			cmdArray = append(cmdArray, arg)
		}
	}

	return c, cmdArray, nil
}

func operateContainers(ctx *cli.Context, action string) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("missing container's name or uuid")
	}

	unknownErr := fmt.Errorf("unknown action: %s", action)
	for _, arg := range ctx.Args() {
		c, err := container.GetContainerByNameOrUuid(arg)
		if err != nil {
			return err
		}

		switch action {
		case container.Stop:
			err = c.Stop()
		case container.Start:
			err = c.Start()
		case container.Restart:
			err = c.Restart()
		case container.Delete:
			err = c.Delete()
		default:
			err = unknownErr
		}

		if err != nil {
			return err
		}
	}

	return nil
}
