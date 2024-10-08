package infofx

import (
	"context"

	"github.com/prismedic/scalpel/routerfx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type InfoParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    *zap.SugaredLogger
}

func DisplayInfo(p InfoParams) {
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			p.Logger.Info("Starting application:")
			info, err := GetInfo()
			if err != nil {
				return err
			}
			p.Logger.Info(info.Name)
			p.Logger.Info(info.Platform)
			p.Logger.Info(info.Runtime)
			p.Logger.Info(info.HostName)
			p.Logger.Info(info.BuildCommit)
			p.Logger.Info(info.BuildDate)
			return nil
		},
	})
}

func cleanup(p InfoParams) {
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// common cleanup hooks for all applications
			p.Logger.Info("Shutting down gracefully, press Ctrl+C again to force")
			return nil
		},
	})
}

var Module = fx.Module("info",
	fx.Provide(routerfx.AsControllerRoute(NewHealthController)),
	fx.Invoke(DisplayInfo),
	fx.Invoke(cleanup),
)
