package data

import (
	"fmt"
	"time"

	aiV1 "recon-ai/api/ai/v1"
	reconV1 "recon-ai/api/recon/v1"
	"recon-ai/app/bff/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewClients)

// Clients 持有下游服务连接
type Clients struct {
	Core reconV1.ReconClient
	AI   aiV1.AiClient

	conns []*grpc.ClientConn
}

func NewClients(c *conf.Bootstrap, logger log.Logger) (*Clients, func(), error) {
	l := log.NewHelper(log.With(logger, "module", "data/clients"))

	timeout := 10 * time.Second
	if c.Client.Timeout != nil {
		timeout = c.Client.Timeout.AsDuration()
	}

	coreConn, err := dial(c.Client.CoreEndpoint, timeout)
	if err != nil {
		return nil, nil, fmt.Errorf("dial core: %w", err)
	}

	aiConn, err := dial(c.Client.AiEndpoint, timeout)
	if err != nil {
		if closeErr := coreConn.Close(); closeErr != nil {
			l.Errorw("msg", "close core conn failed", "err", closeErr)
		}
		return nil, nil, fmt.Errorf("dial ai: %w", err)
	}

	clients := &Clients{
		Core:  reconV1.NewReconClient(coreConn),
		AI:    aiV1.NewAiClient(aiConn),
		conns: []*grpc.ClientConn{coreConn, aiConn},
	}

	cleanup := func() {
		for _, conn := range clients.conns {
			if err := conn.Close(); err != nil {
				l.Errorw("msg", "close conn failed", "err", err)
			}
		}
	}

	return clients, cleanup, nil
}

func dial(endpoint string, timeout time.Duration) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16<<20)),
	)
	if err != nil {
		return nil, err
	}
	_ = timeout // 超时在每次调用时通过 context 控制
	return conn, nil
}
