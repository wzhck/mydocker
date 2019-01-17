package container

import (
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/urfave/cli"
	"github.com/vishvananda/netlink"
	"github.com/weikeit/mydocker/pkg/cgroups"
	"github.com/weikeit/mydocker/pkg/image"
	"github.com/weikeit/mydocker/pkg/network"
	"github.com/weikeit/mydocker/util"
	"net"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func NewContainer(ctx *cli.Context) (*Container, error) {
	detach := ctx.Bool("detach")

	name := ctx.String("name")
	if name == "" {
		// generate a random name if necessary.
		name = strings.ToLower(randomdata.SillyName())
	}

	if c, _ := GetContainerByNameOrUuid(name); c != nil {
		return nil, fmt.Errorf("the container name %s already exist", name)
	}

	uuid := util.Sha256Sum(name)[:12]

	dns := ctx.StringSlice("dns")

	imgNameOrUuid := ctx.String("image")
	if imgNameOrUuid == "" {
		return nil, fmt.Errorf("the image name is required")
	}

	img, err := image.GetImageByNameOrUuid(imgNameOrUuid)
	if err != nil {
		return nil, err
	}

	var commands []string
	if len(img.Entrypoint) > 0 {
		commands = append(commands, img.Entrypoint...)
	}
	if len(ctx.Args()) > 0 {
		commands = append(commands, ctx.Args()...)
	} else if len(img.Command) > 0 {
		commands = append(commands, img.Command...)
	}

	if len(commands) == 0 {
		return nil, fmt.Errorf("missing container commands")
	}

	storageDriver := ctx.String("storage-driver")
	driverConfig, ok := DriverConfigs[storageDriver]
	if !ok {
		return nil, fmt.Errorf("storage driver only support: %s",
			reflect.ValueOf(DriverConfigs).MapKeys())
	}
	if !Drivers[storageDriver].Allowed() {
		return nil, fmt.Errorf("the driver %s is NOT allowed! "+
			"Note: aufs needs ubuntu release; overlay2 needs "+
			"kernel-4.0.0+", storageDriver)
	}

	rootfs := &Rootfs{
		ContainerDir: path.Join(ContainersDir, uuid),
		ImageDir:     img.RootDir(),
		WriteDir:     path.Join(ContainersDir, uuid, driverConfig["writeDir"]),
		MergeDir:     path.Join(ContainersDir, uuid, driverConfig["mergeDir"]),
	}

	resources, err := cgroups.NewResources(ctx)
	if err != nil {
		return nil, err
	}

	volumes := make(map[string]string)
	for _, volumeArg := range ctx.StringSlice("volume") {
		volumePeers := strings.Split(volumeArg, ":")
		if len(volumePeers) == 2 && volumePeers[0] != "" && volumePeers[1] != "" {
			source := strings.TrimRight(volumePeers[0], "/")
			volumes[source] = path.Join(rootfs.MergeDir, volumePeers[1])
		} else {
			return nil, fmt.Errorf("the argument of -v should be '-v /src:/dst'")
		}
	}

	envs := make(map[string]string)
	// note: should put img.Envs before ctx's envs.
	for _, envArg := range append(img.Envs, ctx.StringSlice("env")...) {
		envPeers := strings.Split(envArg, "=")
		if len(envPeers) >= 2 && envPeers[0] != "" {
			// note: the value maybe containe the character =
			envs[envPeers[0]] = strings.Join(envPeers[1:], "=")
		} else {
			return nil, fmt.Errorf("the argument of -e should be '-e key=value'")
		}
	}

	nwNames := ctx.StringSlice("network")
	if len(nwNames) == 0 {
		nwNames = append(nwNames, network.DefaultNetwork)
	}

	ports, err := parsePortMaps(ctx)
	if err != nil {
		return nil, err
	}

	nwNames = util.Uniq(nwNames)
	var endpoints []*network.Endpoint

	// if context contains `--net none`
	// don't allocate ip for container.
	if !util.Contains(nwNames, "none") {
		endpoints, err = CreateEndpoints(name, nwNames, ports)
		if err != nil {
			return nil, err
		}
	}

	if err := image.ChangeCounts(img.RepoTag, "create"); err != nil {
		return nil, err
	}

	return &Container{
		Detach:        detach,
		Uuid:          uuid,
		Name:          name,
		Dns:           dns,
		Image:         imgNameOrUuid,
		Commands:      commands,
		Rootfs:        rootfs,
		Volumes:       volumes,
		Envs:          envs,
		Ports:         ports,
		Endpoints:     endpoints,
		Status:        Creating,
		CreateTime:    time.Now().Format("2006-01-02 15:04:05"),
		StorageDriver: storageDriver,
		Cgroups: &cgroups.Cgroups{
			Path:      MyDocker + "/" + uuid,
			Resources: resources,
		},
	}, nil
}

func parsePortMaps(ctx *cli.Context) (map[string]string, error) {
	ports := make(map[string]string)

	allContainers, err := GetAllContainers()
	if err != nil {
		return nil, err
	}
	for _, portArg := range ctx.StringSlice("publish") {
		portPeers := strings.Split(portArg, ":")
		if len(portPeers) == 2 && portPeers[0] != "" && portPeers[1] != "" {
			for _, portStr := range portPeers {
				if portNum, err := strconv.Atoi(portStr); err != nil {
					return nil, fmt.Errorf("the port %s is not integer", portStr)
				} else if portNum < 0 || portNum > 65535 {
					return nil, fmt.Errorf("the port %s is out of [0, 65535]", portStr)
				}
			}

			outPort, _ := strconv.Atoi(portPeers[0])
			inPort, _ := strconv.Atoi(portPeers[1])
			ports[strconv.Itoa(outPort)] = strconv.Itoa(inPort)

			if server, err := net.Listen("tcp", ":"+strconv.Itoa(outPort)); err != nil {
				return nil, fmt.Errorf("the host port %d is already in use", outPort)
			} else {
				server.Close()
			}

			for _, c := range allContainers {
				for out := range c.Ports {
					if out == strconv.Itoa(outPort) {
						return nil, fmt.Errorf("the host port %d is already in use", outPort)
					}
				}
			}
		} else {
			return nil, fmt.Errorf("the argument of -p should be '-p out:in'")
		}
	}

	return ports, nil
}

func CreateEndpoints(cName string, nwNames []string, ports map[string]string) ([]*network.Endpoint, error) {
	var endpoints []*network.Endpoint

	for _, nwName := range nwNames {
		nw, ok := network.Networks[nwName]
		if !ok {
			return nil, fmt.Errorf("no such network %s, please create it first", nwName)
		}

		ipaddr, err := network.IPAllocator.Allocate(nw)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate new ip from the network %s: %v",
				nwName, err)
		}

		br, err := netlink.LinkByName(nw.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get the bridge of the network %s: %v",
				nw.Name, err)
		}

		// the sha256sum has 64 decimal digits.
		hashed := util.Sha256Sum(nwName + "/" + cName)

		la := netlink.NewLinkAttrs()
		la.Name = "veth-" + hashed[:8]
		// bind this veth onto the bridge
		la.MasterIndex = br.Attrs().Index

		endpoints = append(endpoints, &network.Endpoint{
			Uuid:    hashed[52:],
			IPAddr:  ipaddr,
			Network: nw,
			Ports:   ports,
			Device: &netlink.Veth{
				LinkAttrs: la,
				PeerName:  "ceth-" + hashed[:8],
			},
		})
	}

	return endpoints, nil
}
