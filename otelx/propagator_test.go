package otelx

import (
	"slices"
	"testing"
)

func TestBuildTextMapPropagator(t *testing.T) {
	t.Parallel()

	fields := buildTextMapPropagator().Fields()
	for _, want := range []string{"traceparent", "tracestate", "baggage"} {
		if !slices.Contains(fields, want) {
			t.Errorf("Fields() missing %q: %v", want, fields)
		}
	}
	if slices.Contains(fields, "X-Cloud-Trace-Context") {
		t.Errorf("Fields() includes deprecated Cloud Trace header: %v", fields)
	}
}
