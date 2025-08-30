package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	vegeta "github.com/tsenart/vegeta/lib"
)

var sImagePath string
var lbImagePath string
var promImagePath string
var grafImagePath string
var promConfigPaths string
var grafConfigPaths string

type DemoContainers struct {
	backends     [10]testcontainers.Container
	loadBalander testcontainers.Container
	prom         testcontainers.Container
	graf         testcontainers.Container
}

type logConsumer struct {
	name string
}

func (c logConsumer) Accept(l testcontainers.Log) {
	log.Printf("%s: %s: %s", c.name, l.LogType, l.Content)
}

func stopContainerOnDone(ctx context.Context, c testcontainers.Container) {
	go func() {
		<-ctx.Done()
		testcontainers.TerminateContainer(c)
	}()
}

func SetupContainer(ctx context.Context, name, imageName string, nw *testcontainers.DockerNetwork, port string, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	//l := logConsumer{name: name}
	opts = append(opts,
		testcontainers.WithExposedPorts(port),
		network.WithNetwork([]string{name}, nw),
		//	testcontainers.WithLogConsumers(l),
	)
	c, err := testcontainers.Run(ctx, imageName, opts...)
	if err != nil {
		return nil, fmt.Errorf("testcontainers.Run(%v): %w", name, err)
	}

	stopContainerOnDone(ctx, c)
	return c, nil
}

func SetupContainers(ctx context.Context) (*DemoContainers, error) {
	client, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		return nil, err
	}

	err = loadImage(ctx, client, sImagePath, "hiserver")
	if err != nil {
		return nil, fmt.Errorf("loadImage(%v): %w", sImagePath, err)
	}

	err = loadImage(ctx, client, lbImagePath, "lb")
	if err != nil {
		return nil, err
	}

	err = loadImage(ctx, client, promImagePath, "prom")
	if err != nil {
		return nil, err
	}

	err = loadImage(ctx, client, grafImagePath, "graf")
	if err != nil {
		return nil, err
	}

	d := &DemoContainers{}

	demoNetwork, err := network.New(ctx)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		demoNetwork.Remove(context.Background())
	}()

	var backendEndpoints []string

	for i := range d.backends {
		name := fmt.Sprintf("server%d", i)
		c, err := SetupContainer(ctx, name, "hiserver", demoNetwork, "8000/tcp")
		if err != nil {
			return nil, err
		}
		d.backends[i] = c

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

	lb, err := SetupContainer(ctx, "lb", "lb", demoNetwork, "9001/tcp", testcontainers.WithCmdArgs("--backends", strings.Join(backendEndpoints, ",")))
	if err != nil {
		return nil, err
	}
	d.loadBalander = lb

	ep, err := lb.PortEndpoint(ctx, "9001", "http")
	if err != nil {
		return nil, err
	}
	fmt.Printf("Started lb on %s\n", ep)

	promloc, err := runfiles.Rlocation("_main/prom/prometheus.yml")
	if err != nil {
		return nil, fmt.Errorf("runfiles.Rlocation(): %w", err)
	}
	root := filepath.Dir(promloc)
	var containerFiles []testcontainers.ContainerFile
	for _, path := range strings.Split(promConfigPaths, " ") {
		config, err := runfiles.Rlocation(path)
		if err != nil {
			return nil, fmt.Errorf("runfiles.Rlocation(%v): %w", path, err)
		}
		rel, err := filepath.Rel(root, config)
		if err != nil {
			return nil, fmt.Errorf("filepath.Rel(%v, %v): %w", root, config, err)
		}
		containerFilePath := filepath.Join("/etc/prometheus", rel)
		containerFiles = append(containerFiles,
			testcontainers.ContainerFile{
				HostFilePath:      config,
				ContainerFilePath: containerFilePath,
				FileMode:          0o644,
			})
	}
	prom, err := SetupContainer(ctx, "prom", "prom", demoNetwork, "9090/tcp",
		testcontainers.WithFiles(containerFiles...),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/")),
	)
	if err != nil {
		return nil, err
	}

	d.prom = prom

	pep, err := prom.PortEndpoint(ctx, "9090", "http")
	if err != nil {
		return nil, err
	}
	fmt.Printf("Started prom on %s\n", pep)

	grafloc, err := runfiles.Rlocation("blts/graf/grafana.ini")
	if err != nil {
		return nil, fmt.Errorf("runfiles.Rlocation(): %w", err)
	}
	root = filepath.Dir(grafloc)
	for _, path := range strings.Split(grafConfigPaths, " ") {
		config, err := runfiles.Rlocation(path)
		if err != nil {
			return nil, fmt.Errorf("runfiles.Rlocation(%v): %w", path, err)
		}
		rel, err := filepath.Rel(root, config)
		if err != nil {
			return nil, fmt.Errorf("filepath.Rel(%v, %v): %w", root, config, err)
		}
		containerFilePath := filepath.Join("/etc/grafana", rel)

		containerFiles = append(containerFiles, testcontainers.ContainerFile{
			HostFilePath:      config,
			ContainerFilePath: containerFilePath,
			FileMode:          0o644,
		})
	}

	graf, err := SetupContainer(ctx, "graf", "graf", demoNetwork, "3000/tcp",
		testcontainers.WithFiles(containerFiles...),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/")),
	)
	if err != nil {
		return nil, err
	}
	d.graf = graf
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
	}

	fmt.Println("done, press enter again")
	fmt.Scanln()
}
