package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/antorpo/os-go-concurrency/internal/application/usecase"
	"github.com/antorpo/os-go-concurrency/internal/domain/entities"
	"github.com/antorpo/os-go-concurrency/pkg/log"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type IProductController interface {
	ProcessProducts(ctx *gin.Context)
}

type productController struct {
	logger                log.Logger
	meter                 metric.Meter
	productUseCase        usecase.IProductUseCase
	productGauge          metric.Int64Gauge
	responseTimeHistogram metric.Float64Histogram
}

func NewProductController(logger log.Logger, meter metric.Meter, productUseCase usecase.IProductUseCase) IProductController {
	productGauge, err := meter.Int64Gauge("products_processed_gauge")
	if err != nil {
		logger.Error("failed to create metric gauge", log.Err(err))
	}

	responseTimeHistogram, err := meter.Float64Histogram("request_response_time", metric.WithUnit("ms"))
	if err != nil {
		logger.Error("failed to create histogram", log.Err(err))
	}

	return &productController{
		logger:                logger,
		meter:                 meter,
		productUseCase:        productUseCase,
		productGauge:          productGauge,
		responseTimeHistogram: responseTimeHistogram,
	}
}

func (c *productController) ProcessProducts(ctx *gin.Context) {
	startTime := time.Now()

	var request entities.RequestProducts
	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mode := ctx.DefaultQuery("mode", "sequential")

	productLen := len(request.Products)
	attributes := []attribute.KeyValue{
		attribute.String("processing_mode", mode),
	}
	c.productGauge.Record(ctx.Request.Context(), int64(productLen), metric.WithAttributes(attributes...))
	c.logger.Info(fmt.Sprintf("processing %d products", productLen), log.String("processing_mode", mode))

	var resp *entities.ResponseProducts
	var err error

	if mode == "concurrent" {
		resp, err = c.productUseCase.ProcessConcurrent(ctx.Request.Context(), &request)
	} else {
		resp, err = c.productUseCase.ProcessSequential(ctx.Request.Context(), &request)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	duration := time.Since(startTime).Milliseconds()
	c.responseTimeHistogram.Record(ctx.Request.Context(), float64(duration), metric.WithAttributes(attributes...))
	ctx.JSON(http.StatusOK, resp)
}
