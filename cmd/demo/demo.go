package main

import (
	"context"
	"log"
	"os"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/testcontainers/testcontainers-go"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

type DemoContainers struct {
	backends     [10]testcontainers.Container
	loadBalander testcontainers.Container
}

func SetupContainers(ctx context.Context) (*DemoContainers, error) {
	imagePath, err := runfiles.Rlocation(filepath.Join("blts", "/cmd/s/image"))
	if err != nil {
		return nil, err
	}
	path, err := layout.FromPath(imagePath)
	if err != nil {
		return nil, err
	}
	img, err := path.ImageIndex()
	if err != nil {
		return nil, err
	}
	manifest, err := img.IndexManifest()
	if err != nil {
		return nil, err
	}
	for _, m := range manifest.Manifests() {
		fmt.Printf("manifgest digest: %v", m.Digest())
	}
	// // client, err := testcontainers.NewDockerClientWithOpts(ctx)
	// // if err != nil {
	// // 	return nil, err
	// // }
	// // reader, err := os.Open(imagePath)
	// // if err != nil {
	// // 	return nil, err
	// // }
	// // resp, err := client.ImageLoad(ctx, reader, false)
	// // if err != nil {
	// // 	return nil, err
	// // }
	// // defer resp.Body.Close()
	return &DemoContainers{}, nil

}

func main() {
	_, err := SetupContainers(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
