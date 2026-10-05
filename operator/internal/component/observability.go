package component

import (
	v1 "github.com/plantonhq/planton/operator/api/v1"
)

// tracesEndpoint is where the install's traces go: spec.observability's
// otlpHttpEndpoint, or empty for an install that traces nothing. The control
// plane and the console read the same value, so a browser span and the API
// call it made land in one trace store.
func tracesEndpoint(planton *v1.PlantonPlatform) string {
	if planton.Spec.Observability == nil {
		return ""
	}
	return planton.Spec.Observability.OtlpHttpEndpoint
}
