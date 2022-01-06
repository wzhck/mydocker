package image

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"
	"weike.sh/mydocker/pkg/image"
)

func pull(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("misssing image's repo and tag")
	}

	for _, imageName := range ctx.Args() {
		if !strings.Contains(imageName, ":") {
			imageName = imageName + ":latest"
		}
		if image.Exist(imageName) {
			continue
		}
		if err := image.Pull(imageName); err != nil {
			return err
		}
	}

	return nil
}

func remove(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("misssing image's repo and tag")
	}

	for _, identifier := range ctx.Args() {
		if err := image.Delete(identifier); err != nil {
			return err
		}
	}

	return nil
}

func list(_ *cli.Context) error {
	if err := image.Load(); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 8, 1, 3, ' ', 0)
	fmt.Fprint(w, "IMAGE ID\tREPO\tTAG\tCOUNTS\tCREATED\tSIZE\n")
	for _, img := range image.Images {
		repoTags := strings.Split(img.RepoTag, ":")
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			img.Uuid,
			repoTags[0],
			repoTags[1],
			img.Counts,
			img.CreateTime,
			img.Size,
		)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush %v", err)
	}
	return nil
}
