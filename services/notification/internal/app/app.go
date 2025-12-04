package app

import (
	"context"
	"log/slog"

	"github.com/s1lentmol/q-flow-backend/services/notification/config"
	grpcapp "github.com/s1lentmol/q-flow-backend/services/notification/internal/app/grpc"
	"github.com/s1lentmol/q-flow-backend/services/notification/internal/services/notification"
	"github.com/s1lentmol/q-flow-backend/services/notification/internal/storage/postgres"
	migrator "github.com/s1lentmol/q-flow-backend/services/notification/migrations"
)

type App struct {
	GRPCSrv *grpcapp.App
	storage *postgres.Storage
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	if err := migrator.Migrate(cfg.DSN()); err != nil {
		return nil, err
	}

	store, err := postgres.New(ctx, cfg.DSN())
	if err != nil {
		return nil, err
	}

	notifService := notification.New(log, store, cfg.Telegram.Token)

	grpcApp := grpcapp.New(log, notifService, cfg.GRPC.Port)

	return &App{
		GRPCSrv: grpcApp,
		storage: store,
	}, nil
}

func (a *App) Stop() {
	a.GRPCSrv.Stop()
	a.storage.Close()
}
