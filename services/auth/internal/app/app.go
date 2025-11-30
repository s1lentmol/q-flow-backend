package app

import (
	"context"
	"log/slog"

	"github.com/s1lentmol/q-flow-backend/services/auth/config"
	grpcapp "github.com/s1lentmol/q-flow-backend/services/auth/internal/app/grpc"
	"github.com/s1lentmol/q-flow-backend/services/auth/internal/services/auth"
	"github.com/s1lentmol/q-flow-backend/services/auth/internal/storage/postgres"
	migrator "github.com/s1lentmol/q-flow-backend/services/auth/migrations"
)

type App struct {
	GRPCSrv *grpcapp.App
	storage *postgres.Storage
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	if err := migrator.Migrate(cfg.GetDSN()); err != nil {
		return nil, err
	}

	store, err := postgres.New(ctx, cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	authService := auth.New(log, store, store, store, cfg.TokenTTL)

	grpcApp := grpcapp.New(log, authService, cfg.GRPC.Port)

	return &App{
		GRPCSrv: grpcApp,
		storage: store,
	}, nil
}

func (a *App) Stop() {
	a.GRPCSrv.Stop()
	a.storage.Close()
}
