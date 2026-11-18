package server

import (
	stdhttp "net/http"

	appV1 "recon-ai/api/app/v1"
	"recon-ai/app/bff/internal/conf"
	"recon-ai/app/bff/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewHTTPServer)

// NewHTTPServer 装配 BFF 的 HTTP 服务
func NewHTTPServer(c *conf.Bootstrap, svc *service.AppService, logger log.Logger) *khttp.Server {
	var opts = []khttp.ServerOption{
		khttp.Middleware(
			middleware.Chain(
				recovery.Recovery(),
			),
		),
		// 本地前后端分离开发需要 CORS；生产由 nginx 同源反代，可移除
		khttp.Filter(corsFilter),
	}

	if c.Server.Http.Network != "" {
		opts = append(opts, khttp.Network(c.Server.Http.Network))
	}
	if c.Server.Http.Addr != "" {
		opts = append(opts, khttp.Address(c.Server.Http.Addr))
	}
	if c.Server.Http.Timeout != nil {
		opts = append(opts, khttp.Timeout(c.Server.Http.Timeout.AsDuration()))
	}

	srv := khttp.NewServer(opts...)

	// proto 里通过 google.api.http 声明的路由
	appV1.RegisterAppHTTPServer(srv, svc)

	// 文件上传无法用 google.api.http 表达，单独注册
	router := srv.Route("/")
	router.POST("/api/tasks", svc.UploadTask)

	return srv
}

func corsFilter(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == stdhttp.MethodOptions {
			w.WriteHeader(stdhttp.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
