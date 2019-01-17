package cgroups

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

type Device struct {
	// Type is a (all), c (char), or b (block); all means it
	// applies to all types and all major and minor numbers.
	Type rune `json:"Type"`

	// Path to the device.
	Path string `json:"Path"`

	// Major is the device's major number.
	Major int64 `json:"Major"`

	// Minor is the device's minor number.
	Minor int64 `json:"Minor"`

	// Write to the file devices.allow or devices.deny
	Allow bool `json:"Allow"`

	// Access is a combination of r (read), w (write), and m (mknod)
	Access string `json:"Access"`

	// FileMode permission bits for the device.
	FileMode os.FileMode `json:"FileMode"`

	// Uid of the device.
	Uid uint32 `json:"Uid"`

	// Gid of the device.
	Gid uint32 `json:"Gid"`
}

const (
	devices      = "devices"
	devicesDeny  = "devices.deny"
	devicesAllow = "devices.allow"
)

type DevicesSubsystem struct{}

func (_ *DevicesSubsystem) Name() string {
	return devices
}

func (_ *DevicesSubsystem) RootName() string {
	return devices
}

func (_ *DevicesSubsystem) Apply(cgPath string, pid int) error {
	return apply(devices, cgPath, pid)
}

func (_ *DevicesSubsystem) Remove(cgPath string) error {
	return remove(devices, cgPath)
}

func (_ *DevicesSubsystem) Set(cgPath string, r *Resources) error {
	devicesPath, err := getSubsystemPath(devices, cgPath)
	if err != nil {
		return err
	}

	for _, device := range r.Device {
		deviceFile := devicesDeny
		if device.Allow {
			deviceFile = devicesAllow
		}

		confFile := path.Join(devicesPath, deviceFile)
		confValue := []byte(fmt.Sprintf("%c %d:%d %s", device.Type,
			device.Major, device.Minor, device.Access))
		log.Debugf("set %s => %s", deviceFile, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}
