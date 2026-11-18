package server

import (
	"recon-ai/app/ai/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer 只提供健康检查；AI 服务对内的业务入口是 gRPC
func NewHTTPServer(c *conf.Bootstrap, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			middleware.Chain(
				recovery.Recovery(),
			),
		),
	}

	if c.Server.Http.Network != "" {
		opts = append(opts, http.Network(c.Server.Http.Network))
	}
	if c.Server.Http.Addr != "" {
		opts = append(opts, http.Address(c.Server.Http.Addr))
	}
	if c.Server.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Server.Http.Timeout.AsDuration()))
	}

	srv := http.NewServer(opts...)
	router := srv.Route("/")
	router.GET("/healthz", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{
			"status":  "ok",
			"service": "recon-ai",
		})
	})
	return srv
}
