package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hoquangnam45/pharmacy-common-go/helper/common"
	"github.com/hoquangnam45/pharmacy-user/internal/conf"

	"github.com/hoquangnam45/pharmacy-common-go/util/log"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z -X main.Name=pharmacy-user"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, server *conf.Server, service *conf.Service) *kratos.App {
	httpPort, err := strconv.ParseInt(strings.SplitN(server.Http.Addr, ":", 2)[1], 10, 64)
	if err != nil {
		panic(err)
	}
	grpcPort, err := strconv.ParseInt(strings.SplitN(server.Grpc.Addr, ":", 2)[1], 10, 64)
	if err != nil {
		panic(err)
	}
	hostAddress, hostPorts, _ := common.InitializeEcsService(logger)

	consulConfig := api.DefaultConfig()
	consulConfig.Address = hostAddress + ":8500"
	client, err := api.NewClient(consulConfig)
	if err != nil {
		panic(err)
	}
	endpoints := []*url.URL{
		{Scheme: "http", Host: fmt.Sprintf("%s:%d", hostAddress, hostPorts[int(httpPort)])},
		{Scheme: "grpc", Host: fmt.Sprintf("%s:%d", hostAddress, hostPorts[int(grpcPort)])},
	}
	return kratos.New(kratos.ID(id),
		kratos.Name(service.Name),
		kratos.Version(service.Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger.(*log.StdLogger).Logger),
		kratos.Server(
			gs,
			hs,
		),
		kratos.Registrar(consul.New(client)),
		kratos.Endpoint(endpoints...),
	)
}

func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			env.NewSource(""),
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Service, log.NewStdLogger(log.Info, id, bc.Service.Name, bc.Service.Version))
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
