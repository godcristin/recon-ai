package server

import (
	reconV1 "recon-ai/api/recon/v1"
	"recon-ai/app/core/internal/conf"
	"recon-ai/app/core/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer 装配 gRPC 服务：recovery → validate
//
// 这里刻意不挂 logging.Server：它会把请求参数整体打进日志，
// 而 CreateTask 的入参是两份完整 CSV（可达数 MB，且含用户交易信息）。
// 业务日志由 service 层按需打印摘要。
func NewGRPCServer(c *conf.Bootstrap, svc *service.ReconService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			middleware.Chain(
				recovery.Recovery(),
				validate.Validator(),
			),
		),
	}

	if c.Server.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Server.Grpc.Network))
	}
	if c.Server.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Server.Grpc.Addr))
	}
	if c.Server.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Server.Grpc.Timeout.AsDuration()))
	}

	srv := grpc.NewServer(opts...)
	reconV1.RegisterReconServer(srv, svc)
	return srv
}
