package image

type Image struct {
	Uuid       string   `json:"Uuid"`
	Size       string   `json:"Size"`
	Counts     int      `json:"Counts"`
	RepoTag    string   `json:"RepoTag"`
	WorkingDir string   `json:"WorkingDir"`
	CreateTime string   `json:"CreateTime"`
	Entrypoint []string `json:"Entrypoint"`
	Command    []string `json:"Command"`
	Envs       []string `json:"Envs"`
}
