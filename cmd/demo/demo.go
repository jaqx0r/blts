package main

import (
	"strings"
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
	 "github.com/testcontainers/testcontainers-go/network"

)

type DemoContainers struct {
	backends     [10]testcontainers.Container
	loadBalander testcontainers.Container

	shutdowns []func()
}

func (d*DemoContainers) Shutdown() {
	for _, f := range d.shutdowns {
		f()
	}
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

type containerStartOption func(*testcontainers.ContainerRequest)

func Port(port string) containerStartOption {
	return func(r *testcontainers.ContainerRequest) {
		r.ExposedPorts = append(r.ExposedPorts, port)
	}
}

func Cmd(cmd string) containerStartOption {
	return func(r *testcontainers.ContainerRequest) {
		r.Cmd = append(r.Cmd, cmd)
	}
}

func Net(name string) containerStartOption {
	return func(r *testcontainers.ContainerRequest) {
		r.Networks = append(r.Networks, name)
	}
}

func startContainer(ctx context.Context, client *testcontainers.DockerClient, imageName string, opts ...containerStartOption) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        imageName,
	}
	for _, o := range opts {
		o(&req)
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return c, nil
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

	err = loadImage(ctx, client, "cmd/lb/image", "lb")
	if err != nil {
		return nil, err
	}

	d := &DemoContainers{}

	network, err := network.New(ctx)
	if err != nil {
		return nil, err
	}
	d.shutdowns = append(d.shutdowns,  func() { network.Remove(context.Background()) })

	var backendEndpoints []string

	for i := range d.backends {
		c, err := startContainer(ctx, client, "hiserver", Port("8000"), Net(network.Name))
		if err != nil {
			return nil, err
		}
		d.backends[i] = c
		d.shutdowns = append(d.shutdowns, func() { testcontainers.TerminateContainer(c) })

		ep, err := c.PortEndpoint(ctx, "8000", "http")
		if err != nil {
			return err
		}
		fmt.Printf("Started hiserver on %s\n", ep)
		ip, err := c.ContainerIP(ctx)
		if err != nil {
			return nil, err
		}
		backendEndpoints = append(backendEndpoints, fmt.Sprintf("%s:8000", ip)
	}

	lb, err := startContainer(ctx, client, "lb", Port("9001"), Net(network.Name), Cmd("--backends"), Cmd(strings.Join(backendEndpoints,  ",")))
	if err != nil {
		return nil, err
	}
	d.loadBalander = lb
	d.shutdowns = append(d.shutdowns, func() { testcontainers.TerminateContainer(lb) })

	ep, err := lb.PortEndpoint(ctx, "9001", "http")
	if err != nil { return nil, err }
	fmt.Printf("Started lb on %s\n", ep)

	return d, nil
}

func main() {
	d, err := SetupContainers(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	defer d.Shutdown()

	fmt.Scanln()
}

