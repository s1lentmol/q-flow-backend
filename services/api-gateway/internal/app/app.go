package app

import (
	"context"
	"log/slog"
	"net"
	"strconv"

	"github.com/s1lentmol/q-flow-backend/services/api-gateway/config"
	authclient "github.com/s1lentmol/q-flow-backend/services/api-gateway/internal/clients/auth"
	notifyclient "github.com/s1lentmol/q-flow-backend/services/api-gateway/internal/clients/notification"
	queueclient "github.com/s1lentmol/q-flow-backend/services/api-gateway/internal/clients/queue"
	"github.com/s1lentmol/q-flow-backend/services/api-gateway/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	Server *server.Server
	conns  []*grpc.ClientConn
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	dialCtx, cancel := context.WithTimeout(ctx, cfg.GRPC.Timeout)
	defer cancel()

	authConn, err := grpc.DialContext(dialCtx, cfg.GRPC.AuthAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	queueConn, err := grpc.DialContext(dialCtx, cfg.GRPC.QueueAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	notifConn, err := grpc.DialContext(dialCtx, cfg.GRPC.NotificationAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	srv := server.New(
		log,
		authclient.New(authConn),
		queueclient.New(queueConn),
		notifyclient.New(notifConn),
		cfg.App.Secret,
		cfg.App.ID,
	)

	return &App{
		Server: srv,
		conns:  []*grpc.ClientConn{authConn, queueConn, notifConn},
	}, nil
}

func (a *App) Run(ctx context.Context, addr string) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Server.Run(addr)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return a.Server.Shutdown(context.Background())
	}
}

func (a *App) Close() {
	for _, conn := range a.conns {
		_ = conn.Close()
	}
}

func ListenAddr(port int) string {
	return net.JoinHostPort("", strconv.Itoa(port))
}
