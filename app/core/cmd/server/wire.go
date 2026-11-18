//go:build wireinject
// +build wireinject

package main

import (
	"recon-ai/app/core/internal/biz"
	"recon-ai/app/core/internal/conf"
	"recon-ai/app/core/internal/data"
	"recon-ai/app/core/internal/server"
	"recon-ai/app/core/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp 由 wire 生成依赖装配代码，见 wire_gen.go
func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
	))
}
