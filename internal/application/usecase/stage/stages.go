package stage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/antorpo/os-go-concurrency/internal/domain/entities"
)

type MergerHolder struct {
	Availability string
	Price        float64
}

func Source(_ context.Context, input interface{}) (interface{}, error) {
	data, ok := input.(*entities.RequestProducts)
	if !ok {
		er := errors.New("invalid input type")
		return nil, er
	}

	return data, nil
}

func ProductSplitter(_ context.Context, input interface{}) ([]interface{}, error) {
	data, ok := input.(*entities.RequestProducts)
	if !ok {
		er := errors.New("invalid input type")
		return nil, er
	}

	products := make([]interface{}, 0, len(data.Products))
	for _, product := range data.Products {
		products = append(products, product)
	}

	return products, nil
}

func ProductTagger(_ context.Context, input interface{}) string {
	data, ok := input.(entities.Product)
	if !ok {
		return "unknown-product"
	}

	return data.ProductID
}

func CheckAvailability(_ context.Context, _ interface{}) (interface{}, error) {
	return MockCheckAvailability()
}

func GetPricing(_ context.Context, _ interface{}) (interface{}, error) {
	return MockGetPricing()
}

func Merger(_ context.Context, input []interface{}) (interface{}, error) {
	var holder MergerHolder

	for _, member := range input {
		switch v := member.(type) {
		case string:
			holder.Availability = v
		case float64:
			holder.Price = v
		default:
			return nil, fmt.Errorf("unexpected type %T in input slice, expected string or float64", v)
		}
	}

	return &holder, nil
}

func Joiner(_ context.Context, input interface{}, results []interface{}) (interface{}, error) {
	data, ok := input.(*entities.RequestProducts)
	if !ok {
		er := errors.New("invalid input type")
		return nil, er
	}

	enrichedProducts := make([]entities.EnrichedProduct, len(results))

	for i, product := range data.Products {
		holder, _ := results[i].(*MergerHolder)
		enrichedProducts[i] = CalculateEnrichment(holder.Availability, holder.Price, product)
	}

	return &entities.ResponseProducts{Products: enrichedProducts}, nil
}

func Sink(_ context.Context, input interface{}) (interface{}, error) {
	return input, nil
}

func MockCheckAvailability() (string, error) {
	// TODO: Mock external service that injects latency
	time.Sleep(200 * time.Millisecond)
	return "In Stock", nil
}

func MockGetPricing() (float64, error) {
	// TODO: Mock external service that injects latency
	time.Sleep(300 * time.Millisecond)
	return 25.99, nil
}

func CalculateEnrichment(availability string, price float64, product entities.Product) entities.EnrichedProduct {
	time.Sleep(100 * time.Millisecond)

	quantity := 10
	costTotal := price * float64(quantity)
	discount := 0.0
	if price > 20 {
		discount = costTotal * 0.10
	}
	totalCost := costTotal - discount

	return entities.EnrichedProduct{
		ProductID:    product.ProductID,
		Name:         product.Name,
		Availability: availability,
		Price:        price,
		TotalCost:    totalCost,
	}
}
