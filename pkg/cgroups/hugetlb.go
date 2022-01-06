package cgroups

import (
	"fmt"
	"io/ioutil"
	"path"

	log "github.com/sirupsen/logrus"
	"weike.sh/mydocker/util"
)

const (
	hugetlb = "hugetlb"
	// %s means PageSize, e.g., 256MB, 2GB
	hugetlbFile = "hugetlb.%s.limit_in_bytes"
)

type Hugepage struct {
	PageSize string `json:"PageSize"`
	Limit    uint64 `json:"Limit"`
}

type HugetlbSubsystem struct{}

func (_ *HugetlbSubsystem) Name() string {
	return hugetlb
}

func (_ *HugetlbSubsystem) RootName() string {
	return hugetlb
}

func (_ *HugetlbSubsystem) Apply(cgPath string, pid int) error {
	return apply(hugetlb, cgPath, pid)
}

func (_ *HugetlbSubsystem) Remove(cgPath string) error {
	return remove(hugetlb, cgPath)
}

func (_ *HugetlbSubsystem) Set(cgPath string, r *Resources) error {
	hugetlbPath, err := getSubsystemPath(hugetlb, cgPath)
	if err != nil {
		return err
	}

	for _, hugepage := range r.HugepagesLimit {
		hugepageFile := fmt.Sprintf(hugetlbFile, hugepage.PageSize)
		confFile := path.Join(hugetlbPath, hugepageFile)
		if exist, _ := util.FileOrDirExists(hugepageFile); !exist {
			log.Warnf("host doesn't support the hugepage's size %s",
				hugepage.PageSize)
			continue
		}

		confValue := []byte(fmt.Sprintf("%d", hugepage.Limit))
		log.Debugf("set %s => %s", hugepageFile, confValue)
		if err := ioutil.WriteFile(confFile, confValue, 0644); err != nil {
			return err
		}
	}

	return nil
}
