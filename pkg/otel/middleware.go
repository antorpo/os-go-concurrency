package otel

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const (
	_instrumentationName = "github.com/antorpo/os-go-concurrency"
	_durationMetricName  = "http.server.duration"
	_unitKey             = attribute.Key("unit")
)

func Middleware() gin.HandlerFunc {
	traceMiddleware := otelgin.Middleware("")

	meter := otel.GetMeterProvider().Meter(_instrumentationName)
	durationMetric, _ := meter.Int64Histogram(_durationMetricName)

	return func(c *gin.Context) {
		t := time.Now()

		// traces middleware
		traceMiddleware(c)

		// metrics middleware
		attrs := semconv.HTTPServerMetricAttributesFromHTTPRequest("", c.Request)
		attrs = append(attrs, semconv.HTTPRouteKey.String(c.FullPath()), semconv.HTTPStatusCodeKey.Int(c.Writer.Status()))

		// add unit to metrics attributes
		attrs = append(attrs, _unitKey.String("ms"))

		durationMetric.Record(c.Request.Context(), time.Since(t).Milliseconds(), metric.WithAttributes(attrs...))
	}
}
