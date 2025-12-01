package app

import (
	"context"
	"log/slog"

	"github.com/s1lentmol/q-flow-backend/services/queue/config"
	grpcapp "github.com/s1lentmol/q-flow-backend/services/queue/internal/app/grpc"
	notifyclient "github.com/s1lentmol/q-flow-backend/services/queue/internal/clients/notification"
	"github.com/s1lentmol/q-flow-backend/services/queue/internal/services/queue"
	"github.com/s1lentmol/q-flow-backend/services/queue/internal/storage/postgres"
	migrator "github.com/s1lentmol/q-flow-backend/services/queue/migrations"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	GRPCSrv *grpcapp.App
	storage *postgres.Storage
	notif   *grpc.ClientConn
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	if err := migrator.Migrate(cfg.GetDSN()); err != nil {
		return nil, err
	}

	store, err := postgres.New(ctx, cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	conn, err := grpc.DialContext(ctx, cfg.Notify.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	notif := notifyclient.New(conn)

	queueService := queue.New(log, store, notif)

	grpcApp := grpcapp.New(log, queueService, cfg.GRPC.Port)

	return &App{
		GRPCSrv: grpcApp,
		storage: store,
		notif:   conn,
	}, nil
}

func (a *App) Stop() {
	a.GRPCSrv.Stop()
	a.storage.Close()
	if a.notif != nil {
		_ = a.notif.Close()
	}
}
