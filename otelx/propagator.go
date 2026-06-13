package otelx

import (
	"go.opentelemetry.io/otel/propagation"
	gcppropagator "github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
)

func buildTextMapPropagator(useGCP bool) propagation.TextMapPropagator {
	propagators := []propagation.TextMapPropagator{
		propagation.TraceContext{},
		propagation.Baggage{},
	}
	if useGCP {
		propagators = append(propagators, gcppropagator.CloudTraceFormatPropagator{})
	}
	return propagation.NewCompositeTextMapPropagator(propagators...)
}
