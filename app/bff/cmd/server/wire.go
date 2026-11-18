//go:build wireinject
// +build wireinject

package main

import (
	"recon-ai/app/bff/internal/biz"
	"recon-ai/app/bff/internal/conf"
	"recon-ai/app/bff/internal/data"
	"recon-ai/app/bff/internal/server"
	"recon-ai/app/bff/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
	))
}
