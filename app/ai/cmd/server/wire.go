//go:build wireinject
// +build wireinject

package main

import (
	"recon-ai/app/ai/internal/biz"
	"recon-ai/app/ai/internal/conf"
	"recon-ai/app/ai/internal/llm"
	"recon-ai/app/ai/internal/server"
	"recon-ai/app/ai/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp 由 wire 生成依赖装配代码，见 wire_gen.go
func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		llm.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
	))
}
