package docker

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/waypoint/sdk/component"
	"github.com/hashicorp/waypoint/sdk/terminal"
)

// Builder uses `docker build` to build a Docker iamge.
type Builder struct{}
}

// BuildFunc implements component.Builder
func (b *Builder) BuildFunc() interface{} {
	return b.Build
}

// Build
func (b *Builder) Build(
	ctx context.Context,
	ui terminal.UI,
	src *component.Source,
) (*Image, error) {
	stdout, stderr, err := ui.OutputWriters()
	if err != nil {
		return nil, err
	}

	result := &Image{
		Image: fmt.Sprintf("waypoint.local/%s", src.App),
		Tag:   "latest",
	}

	// Build the image with Docker build
	cmd := exec.CommandContext(ctx, "docker", "build", ".", "-t", result.Name())
	cmd.Dir = src.Path
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	if !b.DisableCEB {
		tmpdir, err := ioutil.TempDir("", "waypoint")
		if err != nil {
			return nil, err
		}

		defer os.RemoveAll(tmpdir)

		err = assets.RestoreAsset(tmpdir, "ceb/ceb")
		if err != nil {
			return nil, err
		}

		err = epinject.AlterEntrypoint(ctx, result.Name(), func(cur []string) (*epinject.NewEntrypoint, error) {
			ep := &epinject.NewEntrypoint{
				Entrypoint: append([]string{"/bin/wpceb"}, cur...),
				InjectFiles: map[string]string{
					filepath.Join(tmpdir, "ceb/ceb"): "/bin/wpceb",
				},
			}

			return ep, nil
		})

		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
