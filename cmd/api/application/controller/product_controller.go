package controller

import (
	"fmt"
	"net/http"

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
	logger         log.Logger
	meter          metric.Meter
	productUseCase usecase.IProductUseCase
	productCount   metric.Int64Counter
}

func NewProductController(logger log.Logger, meter metric.Meter, productUseCase usecase.IProductUseCase) IProductController {
	productCount, err := meter.Int64Counter("products_processed_count")
	if err != nil {
		logger.Error("failed to create metric counter", log.Err(err))
	}

	return &productController{
		logger:         logger,
		meter:          meter,
		productUseCase: productUseCase,
		productCount:   productCount,
	}
}

func (c *productController) ProcessProducts(ctx *gin.Context) {
	var request entities.RequestProducts
	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mode := ctx.DefaultQuery("mode", "sequential")

	// Metric
	productLen := len(request.Products)
	attributes := []attribute.KeyValue{
		attribute.String("processing_mode", mode),
	}
	c.productCount.Add(ctx.Request.Context(), int64(productLen), metric.WithAttributes(attributes...))
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

	ctx.JSON(http.StatusOK, resp)
}
