package cgroups

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"path"
)

const (
	blkio                        = "blkio"
	blkioWeight                  = "blkio.weight"
	blkioLeafWeight              = "blkio.leaf_weight"
	blkioWeightDevice            = "blkio.weight_device"
	blkioLeafWeightDevice        = "blkio.leaf_weight_device"
	blkioThrottleReadBpsDevice   = "blkio.throttle.read_bps_device"
	blkioThrottleWriteBpsDevice  = "blkio.throttle.write_bps_device"
	blkioThrottleReadIOPSDevice  = "blkio.throttle.read_iops_device"
	blkioThrottleWriteIOPSDevice = "blkio.throttle.write_iops_device"
)

type blockIODevice struct {
	// device's major number.
	Major uint64 `json:"Major"`
	// device's minor number.
	Minor uint64 `json:"Minor"`
}

// `major:minor weight` or `major:minor leaf_weight`
type WeightDevice struct {
	blockIODevice
	// Bandwidth rate for the device, range is from 10 to 1000
	Weight uint64 `json:"Weight"`
	// Bandwidth rate for the device while competing with the cgroup's
	// child cgroups, range is from 10 to 1000, cfq scheduler only
	LeafWeight uint64 `json:"LeafWeight"`
}

// `major:minor rate_per_second`
type ThrottleDevice struct {
	blockIODevice
	// io rate limit per cgroup per device.
	Rates uint64 `json:"Rates"`
}

type BlkioSubsystem struct{}

func (_ *BlkioSubsystem) Name() string {
	return blkio
}

func (_ *BlkioSubsystem) RootName() string {
	return blkio
}

func (_ *BlkioSubsystem) Apply(cgPath string, pid int) error {
	return apply(blkio, cgPath, pid)
}

func (_ *BlkioSubsystem) Remove(cgPath string) error {
	return remove(blkio, cgPath)
}

func (_ *BlkioSubsystem) Set(cgPath string, r *Resources) error {
	blkioPath, err := getSubsystemPath(blkio, cgPath)
	if err != nil {
		return err
	}

	if r.BlkioWeight >= 10 && r.BlkioWeight <= 1000 {
		confFile := path.Join(blkioPath, blkioWeight)
		confValue := []byte(fmt.Sprintf("%d", r.BlkioWeight))
		log.Debugf("set %s => %s", blkioWeight, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	if r.BlkioLeafWeight >= 10 && r.BlkioLeafWeight <= 1000 {
		confFile := path.Join(blkioPath, blkioLeafWeight)
		confValue := []byte(fmt.Sprintf("%d", r.BlkioLeafWeight))
		log.Debugf("set %s => %s", blkioLeafWeight, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	for _, device := range r.BlkioWeightDevice {
		confFile := path.Join(blkioPath, blkioWeightDevice)
		confValue := []byte(fmt.Sprintf("%d:%d %d", device.Major, device.Minor, device.Weight))
		log.Debugf("set %s => %s", blkioWeightDevice, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	for _, device := range r.BlkioLeafWeightDevice {
		confFile := path.Join(blkioPath, blkioLeafWeightDevice)
		confValue := []byte(fmt.Sprintf("%d:%d %d", device.Major, device.Minor, device.LeafWeight))
		log.Debugf("set %s => %s", blkioLeafWeightDevice, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	for _, device := range r.BlkioThrottleReadBpsDevice {
		confFile := path.Join(blkioPath, blkioThrottleReadBpsDevice)
		confValue := []byte(fmt.Sprintf("%d:%d %d", device.Major, device.Minor, device.Rates))
		log.Debugf("set %s => %s", blkioThrottleReadBpsDevice, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	for _, device := range r.BlkioThrottleWriteBpsDevice {
		confFile := path.Join(blkioPath, blkioThrottleWriteBpsDevice)
		confValue := []byte(fmt.Sprintf("%d:%d %d", device.Major, device.Minor, device.Rates))
		log.Debugf("set %s => %s", blkioThrottleWriteBpsDevice, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	for _, device := range r.BlkioThrottleReadIOPSDevice {
		confFile := path.Join(blkioPath, blkioThrottleReadIOPSDevice)
		confValue := []byte(fmt.Sprintf("%d:%d %d", device.Major, device.Minor, device.Rates))
		log.Debugf("set %s => %s", blkioThrottleReadIOPSDevice, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	for _, device := range r.BlkioThrottleWriteIOPSDevice {
		confFile := path.Join(blkioPath, blkioThrottleWriteIOPSDevice)
		confValue := []byte(fmt.Sprintf("%d:%d %d", device.Major, device.Minor, device.Rates))
		log.Debugf("set %s => %s", blkioThrottleWriteIOPSDevice, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}

func blockDeviceExists(major, minor int) bool {
	arg := fmt.Sprintf("lsblk -o MAJ:MIN | grep -q %d:%d", major, minor)
	cmd := exec.Command("bash", "-c", arg)
	return cmd.Run() == nil
}
