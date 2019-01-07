package image

import "path"

const (
	MyDockerDir = "/var/lib/mydocker"
)

var (
	ImagesDir        = path.Join(MyDockerDir, "images")
	ImagesConfigFile = path.Join(ImagesDir, "repositories.json")
)

var Images []*Image
