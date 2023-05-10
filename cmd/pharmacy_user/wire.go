//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/hoquangnam45/pharmacy-auth/internal/biz"
	"github.com/hoquangnam45/pharmacy-auth/internal/conf"
	"github.com/hoquangnam45/pharmacy-auth/internal/data"
	"github.com/hoquangnam45/pharmacy-auth/internal/server"
	"github.com/hoquangnam45/pharmacy-auth/internal/service"
	"github.com/hoquangnam45/pharmacy-common-go/util/log"

	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Service, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
