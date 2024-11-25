package usecase

import (
	"context"

	"github.com/antorpo/os-go-concurrency/internal/application/usecase/stage"
	"github.com/antorpo/os-go-concurrency/internal/domain/entities"
	"github.com/antorpo/os-go-concurrency/internal/infrastructure/config"
	"github.com/antorpo/os-go-concurrency/pkg/log"
	"github.com/antorpo/os-go-concurrency/pkg/pipeline"
	"go.opentelemetry.io/otel/metric"
)

type productUseCase struct {
	logger log.Logger
	meter  metric.Meter
	config config.IConfiguration
}

type IProductUseCase interface {
	ProcessSequential(context.Context, *entities.RequestProducts) (*entities.ResponseProducts, error)
	ProcessConcurrent(context.Context, *entities.RequestProducts) (*entities.ResponseProducts, error)
}

func NewProductUseCase(logger log.Logger, meter metric.Meter, config config.IConfiguration) IProductUseCase {
	return &productUseCase{
		logger: logger,
		meter:  meter,
		config: config,
	}
}

func (p *productUseCase) ProcessSequential(_ context.Context, products *entities.RequestProducts) (*entities.ResponseProducts, error) {
	enrichedProducts := make([]entities.EnrichedProduct, len(products.Products))

	for i, product := range products.Products {
		availability, err := stage.MockCheckAvailability()
		if err != nil {
			return nil, err
		}

		price, err := stage.MockGetPricing()
		if err != nil {
			return nil, err
		}

		enrichedProducts[i] = stage.CalculateEnrichment(availability, price, product)
	}

	return &entities.ResponseProducts{Products: enrichedProducts}, nil
}

func (p *productUseCase) ProcessConcurrent(ctx context.Context, products *entities.RequestProducts) (*entities.ResponseProducts, error) {
	// Number of workers in worker pool
	workers := p.config.GetConfig().App.Workers

	pipeline.EncryptedMode = false
	productPipeline := &pipeline.Pipeline{
		Name:   "Product pipeline",
		Source: stage.Source,
		Flow: pipeline.Flow{
			&pipeline.Iterator{
				Name:     "Concurrent processing using fan-in/fan-out",
				Splitter: stage.ProductSplitter,
				MaxP:     &workers,
				Stream: pipeline.Flow{
					&pipeline.Broadcast{
						Name: "External data-sources concurrent",
						Streams: []pipeline.Flow{
							{
								&pipeline.SimplePipe{Resolver: stage.CheckAvailability},
							},
							{
								&pipeline.SimplePipe{Resolver: stage.GetPricing},
							},
						},
						Merger: stage.Merger,
					},
				},
				Joiner: stage.Joiner,
				Tagger: stage.ProductTagger,
			},
		},
		Sink: stage.Sink,
	}

	enrichedProducts, err := pipeline.Run(ctx, products, productPipeline, false)
	if err != nil {
		return nil, err
	}

	return enrichedProducts.(*entities.ResponseProducts), nil
}
