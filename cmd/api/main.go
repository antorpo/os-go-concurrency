package main

import (
	"context"
	"fmt"
	systemlog "log"

	"github.com/antorpo/os-go-concurrency/cmd/api/application"
	"github.com/antorpo/os-go-concurrency/pkg/log"
	"github.com/antorpo/os-go-concurrency/pkg/otel"
)

const _defaultPort = ":8080"

// @title		OS Go concurrency
// @version		1.0
// @description	Concurrent processing of transactions
func main() {
	app, err := application.StartApp()
	if err != nil {
		systemlog.Fatal(err.Error())
	}

	ctx := context.Background()
	oTelShutdown, _ := otel.Start(ctx, app.AppName)
	defer func() {
		if shutdownErr := oTelShutdown(); shutdownErr != nil {
			app.Logger.Error("failed to shutdown OpenTelemetry: %v", log.Err(shutdownErr))
		}
	}()

	if app != nil {
		app.Logger.Info(fmt.Sprintf("running %s", app.AppName))
		_ = app.Router.Run(_defaultPort)
	}
}
