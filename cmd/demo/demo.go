package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	vegeta "github.com/tsenart/vegeta/lib"
	dockerclient "github.com/docker/docker/client"
)

type DemoContainers struct {
	backends     [10]testcontainers.Container
	loadBalander testcontainers.Container
	prom         testcontainers.Container
	graf         testcontainers.Container

	shutdowns []func()
}

func (d *DemoContainers) Shutdown() {
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

		resp, err := client.ImageLoad(ctx, pr, dockerclient.ImageLoadWithQuiet(false))
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

func Wait(w wait.Strategy) containerStartOption {
	return func(r *testcontainers.ContainerRequest) {
		r.WaitingFor = w
	}
}

func Dir(source, target string) containerStartOption {
	return func(r *testcontainers.ContainerRequest) {
		r.Mounts = append(r.Mounts, testcontainers.BindMount(source, testcontainers.ContainerMountTarget(target)))
	}
}

func Env(key, value string) containerStartOption {
	return func(r *testcontainers.ContainerRequest) {
		if r.Env == nil {
			r.Env = make(map[string]string)
		}
		r.Env[key] = value
	}
}

func Alias(network, name string) containerStartOption {
	return func(r *testcontainers.ContainerRequest) {
		if r.NetworkAliases == nil {
			r.NetworkAliases = make(map[string][]string)
		}
		r.NetworkAliases[network] = append(r.NetworkAliases[network], name)
	}
}

func startContainer(ctx context.Context, client *testcontainers.DockerClient, imageName string, opts ...containerStartOption) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image: imageName,
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
	d.shutdowns = append(d.shutdowns, func() { network.Remove(context.Background()) })

	var backendEndpoints []string

	for i := range d.backends {
		c, err := startContainer(ctx, client, "hiserver", Port("8000"), Net(network.Name), Alias(network.Name, fmt.Sprintf("server%d", i)))
		if err != nil {
			return nil, err
		}
		d.backends[i] = c
		d.shutdowns = append(d.shutdowns, func() { testcontainers.TerminateContainer(c) })

		ep, err := c.PortEndpoint(ctx, "8000", "http")
		if err != nil {
			return nil, err
		}
		fmt.Printf("Started hiserver on %s\n", ep)
		ip, err := c.ContainerIP(ctx)
		if err != nil {
			return nil, err
		}
		backendEndpoints = append(backendEndpoints, fmt.Sprintf("%s:8000", ip))
	}

	lb, err := startContainer(ctx, client, "lb", Port("9001"), Net(network.Name), Cmd("--backends"), Cmd(strings.Join(backendEndpoints, ",")), Alias(network.Name, "lb"))
	if err != nil {
		return nil, err
	}
	d.loadBalander = lb
	d.shutdowns = append(d.shutdowns, func() { testcontainers.TerminateContainer(lb) })

	ep, err := lb.PortEndpoint(ctx, "9001", "http")
	if err != nil {
		return nil, err
	}
	fmt.Printf("Started lb on %s\n", ep)

	promconfig, err := runfiles.Rlocation("blts/prom/prometheus.yml")
	if err != nil {
		return nil, err
	}

	prom, err := startContainer(ctx, client, "docker.io/prom/prometheus", Net(network.Name), Port("9090"), Dir(filepath.Dir(promconfig), "/etc/prometheus"), Alias(network.Name, "prometheus"))
	if err != nil {
		return nil, err
	}

	d.prom = prom
	d.shutdowns = append(d.shutdowns, func() { testcontainers.TerminateContainer(prom) })

	pep, err := prom.PortEndpoint(ctx, "9090", "http")
	if err != nil {
		return nil, err
	}
	fmt.Printf("Started prom on %s\n", pep)

	grafconfig, err := runfiles.Rlocation("blts/graf/grafana.ini")
	if err != nil {
		return nil, err
	}

	graf, err := startContainer(ctx, client, "docker.io/grafana/grafana", Net(network.Name), Port("3000"), Dir(filepath.Dir(grafconfig), "/etc/grafana"), Wait(wait.ForHTTP("/")))
	if err != nil {
		return nil, err
	}
	d.graf = graf
	d.shutdowns = append(d.shutdowns, func() { testcontainers.TerminateContainer(graf) })
	graf.Exec(ctx, []string{"ls", "-al", "/etc/grafana"})
	graf.Exec(ctx, []string{"cat", "/etc/grafana/grafana.ini"})

	gep, err := graf.PortEndpoint(ctx, "3000", "http")
	if err != nil {
		return nil, err
	}
	fmt.Printf("Started graf on %s, opening in browser\n", gep)
	openURL(gep)

	return d, nil
}

func main() {
	d, err := SetupContainers(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("press enter to start nice load")
	fmt.Scanln()

	lbep, err := d.loadBalander.PortEndpoint(context.Background(), "9001", "http")
	if err != nil {
		log.Fatal(err)
	}
	target := vegeta.Target{
		Method: "GET",
		URL:    lbep,
	}
	tr := vegeta.NewStaticTargeter(target)
	a := vegeta.NewAttacker()

	r := a.Attack(tr, vegeta.ConstantPacer{Freq: 100, Per: time.Second}, time.Minute, "nice")
	for range r {
		fmt.Printf("r %v\n", r)
	}

	fmt.Println("done, press enter again")
	fmt.Scanln()

	d.Shutdown()
}
