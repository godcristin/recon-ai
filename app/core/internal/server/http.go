package server

import (
	"recon-ai/app/core/internal/conf"
	"recon-ai/app/core/internal/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

func NewHTTPServer(c *conf.Bootstrap, d *data.Data, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			middleware.Chain(
				recovery.Recovery(),
				logging.Server(logger),
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
			"service": "recon-core",
		})
	})

	router.GET("/readyz", func(ctx http.Context) error {
		if err := d.Ping(ctx.Request().Context()); err != nil {
			return ctx.JSON(503, map[string]string{
				"status": "degraded",
				"reason": "database unreachable",
			})
		}
		return ctx.JSON(200, map[string]string{"status": "ready"})
	})

	return srv
}
