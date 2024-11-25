package application

import (
	"fmt"

	"github.com/antorpo/os-go-concurrency/cmd/api/application/controller"
	"github.com/antorpo/os-go-concurrency/internal/infrastructure/config"
	"github.com/antorpo/os-go-concurrency/pkg/log"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const _appName = "os-go-concurrency"

type Application struct {
	AppName        string
	Router         *gin.Engine
	Logger         log.Logger
	Tracer         trace.Tracer
	Meter          metric.Meter
	TestController controller.ITestController
}

func BuildApplication() (*Application, error) {
	app := &Application{}
	app.AppName = _appName

	// Router
	gin.SetMode(gin.DebugMode)
	app.Router = gin.New()

	// Logger
	lvl := log.NewAtomicLevelAt(log.InfoLevel)
	app.Logger = log.NewProductionLogger(&lvl)

	// OTel
	app.Tracer = otel.GetTracerProvider().Tracer(_appName)
	app.Meter = otel.GetMeterProvider().Meter(_appName)

	// Configuration
	_, err := app.registerConfiguration()
	if err != nil {
		return nil, err
	}

	// Controllers
	app.TestController = controller.NewTestController()

	return app, nil
}

func (app *Application) registerConfiguration() (config.IConfiguration, error) {
	cfg := config.NewConfiguration()
	if cfg == nil {
		return nil, fmt.Errorf("unable to load configuration")
	}

	if err := cfg.LoadConfig(); err != nil {
		return nil, err
	}

	return cfg, nil
}
