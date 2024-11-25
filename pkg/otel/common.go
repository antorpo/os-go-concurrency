package otel

import (
	"fmt"
)

const (
	_defaultAgentHost = "localhost"
	_defaultAgentPort = "4317"
)

// ShutdownFunc for shutting down the tracer provider and its components.
type ShutdownFunc func() error

func getEndpoint() string {
	return fmt.Sprintf("%s:%s", _defaultAgentHost, _defaultAgentPort)
}
