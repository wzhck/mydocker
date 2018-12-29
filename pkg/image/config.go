package image

import "path"

const (
	MyDockerDir = "/var/lib/mydocker"
	MaxBytes    = 10000
)

var (
	ImagesDir        = path.Join(MyDockerDir, "images")
	ImagesConfigFile = path.Join(ImagesDir, "repositories.json")
)

var Images []*Image
