package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"io"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/testcontainers/testcontainers-go"
)

type DemoContainers struct {
	backends     [10]testcontainers.Container
	loadBalander testcontainers.Container
}

func loadImage(ctx context.Context, client *testcontainers.DockerClient, imagePath, imageName string) error {
	fullPath, err := runfiles.Rlocation(filepath.Join("blts", imagePath))
	if err != nil {
		return nil
	}

	path, err := layout.FromPath(fullPath)
	if err != nil {
		return err
	}

	index, err := path.ImageIndex()
	if err != nil {
		return err
	}

	manifest, err := index.IndexManifest()
	if err != nil {
		return err
	}

	for _, manifestDescriptor := range manifest.Manifests {
		hash := manifestDescriptor.Digest

		img, err := path.Image(hash)
		if err != nil {
			return err
		}

		ref, err := name.ParseReference(imageName)
		if err != nil {
			return err
		}
		pr, pw := io.Pipe()
		go func() {
			pw.CloseWithError(tarball.Write(ref, img, pw))
		}()

		notQuiet := false
		resp, err := client.ImageLoad(ctx, pr, notQuiet)
		if err != nil {
			return err
		}

		_, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		resp.Body.Close()
	}
	return nil
}

func SetupContainers(ctx context.Context) (*DemoContainers, error) {
	client, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		return nil, err
	}

	err = loadImage(ctx, client, "cmd/s/image", "hiserver")
	if err != nil {
		return nil, err
	}

	req := testcontainers.ContainerRequest{
		Image:        "hiserver:latest",
		ExposedPorts: []string{"8000/tcp"},
	}
	hiserver, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	// TODO: hiserver.TerminateContainer

	info, err := hiserver.Inspect(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("container info: %s\n", info.Config.Image)
	endpoint, err := hiserver.PortEndpoint(ctx, "8000", "http")
	if err != nil {
		return nil, err
	}

	fmt.Printf("endpoint %s\n", endpoint)

	fmt.Scanln()

	return nil, nil
}

func main() {
	_, err := SetupContainers(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
