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

const prometheusImageName = "docker.io/prom/prometheus"
const grafanaImageName = "docker.io/grafana/grafana"

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

	demoNetwork, err := network.New(ctx)
	if err != nil {
		return nil, err
	}
	d.shutdowns = append(d.shutdowns, func() { demoNetwork.Remove(context.Background()) })

	var backendEndpoints []string

	for i := range d.backends {
		c, err := testcontainers.Run(ctx, "hiserver",
			testcontainers.WithExposedPorts("8000/tcp"),
			network.WithNetwork([]string{fmt.Sprintf("server%d", i)}, demoNetwork),
		)
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

	lb, err := testcontainers.Run(ctx, "lb",
		testcontainers.WithExposedPorts("9001/tcp"),
		network.WithNetwork([]string{"lb"}, demoNetwork),
		testcontainers.WithCmdArgs("--backends", strings.Join(backendEndpoints, ",")),
	)
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

	prom, err := testcontainers.Run(ctx, prometheusImageName,
		network.WithNetwork([]string{"prometheus"}, demoNetwork),
		testcontainers.WithExposedPorts("9090/tcp"),
		testcontainers.WithMounts(
			testcontainers.BindMount(
				filepath.Dir(promconfig),
				testcontainers.ContainerMountTarget("/etc/prometheus"),
			),
		),
	)
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

	graf, err := testcontainers.Run(ctx, grafanaImageName,
		testcontainers.WithExposedPorts("3000/tcp"),
		network.WithNetwork([]string{"graf"}, demoNetwork),
		testcontainers.WithMounts(
			testcontainers.BindMount(
			filepath.Dir(grafconfig),
				testcontainers.ContainerMountTarget("/etc/grafana"),
			),
		),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/")),
	)
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
