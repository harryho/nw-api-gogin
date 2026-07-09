package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/harryho/nw-api-gogin/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		panic(err)
	}
	defer application.Shutdown(context.Background())

	addr := app.ServerAddress()
	application.Logger.Info("starting http server", zap.String("address", addr))
	if err := application.Engine.Run(addr); err != nil {
		application.Logger.Error("server shutdown with error", zap.Error(err))
		panic(err)
	}
}
