package main

import (
	"context"
	"io"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/moby/moby/client"
	"github.com/testcontainers/testcontainers-go"
)

func loadImage(ctx context.Context, dockerClient *testcontainers.DockerClient, imagePath, imageName string) error {
	fullPath, err := runfiles.Rlocation(imagePath)
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

		resp, err := dockerClient.ImageLoad(ctx, pr, client.ImageLoadWithQuiet(false))
		if err != nil {
			return err
		}

		_, err = io.ReadAll(resp)
		if err != nil {
			return err
		}

		resp.Close()
	}
	return nil
}
