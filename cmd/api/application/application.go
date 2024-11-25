package application

import (
	"fmt"

	"github.com/antorpo/os-go-concurrency/cmd/api/application/controller"
	"github.com/antorpo/os-go-concurrency/internal/application/usecase"
	"github.com/antorpo/os-go-concurrency/internal/infrastructure/config"
	"github.com/antorpo/os-go-concurrency/pkg/log"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const _appName = "os-go-concurrency"

type Application struct {
	AppName           string
	Router            *gin.Engine
	Logger            log.Logger
	Tracer            trace.Tracer
	Meter             metric.Meter
	Config            config.IConfiguration
	ProductUseCase    usecase.IProductUseCase
	ProductController controller.IProductController
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
	err := app.registerConfiguration()
	if err != nil {
		return nil, err
	}

	// Use Case
	app.registerProductUseCase()

	// Controllers
	app.registerProductController()

	return app, nil
}

func (app *Application) registerConfiguration() error {
	cfg := config.NewConfiguration()
	if cfg == nil {
		return fmt.Errorf("unable to load configuration")
	}

	if err := cfg.LoadConfig(); err != nil {
		return err
	}

	app.Config = cfg
	return nil
}

func (app *Application) registerProductUseCase() {
	app.ProductUseCase = usecase.NewProductUseCase(app.Logger, app.Meter, app.Config)
}

func (app *Application) registerProductController() {
	app.ProductController = controller.NewProductController(app.Logger, app.Meter, app.ProductUseCase)
}
