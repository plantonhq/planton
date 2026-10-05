package ui

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/outputs"
	"github.com/stretchr/testify/assert"
)

// TestFormatOutputLines_SensitiveValuesNeverRender pins the masking law:
// a sensitive output's VALUE must never appear in rendered output — apply
// runs in CI as much as on laptops, and CI logs are persistent and shared.
func TestFormatOutputLines_SensitiveValuesNeverRender(t *testing.T) {
	result := &outputs.CaptureResult{
		Flat: map[string]string{
			"host":         "10.0.0.5",
			"password":     "sup3r-s3cr3t",
			"conn.uri":     "redis://:sup3r-s3cr3t@10.0.0.5",
			"conn.port":    "6379",
			"cluster_name": "cache-prod",
		},
		Secrets: map[string]bool{
			"host":         false,
			"password":     true,
			"conn":         true,
			"cluster_name": false,
		},
	}

	lines := FormatOutputLines(result)
	rendered := strings.Join(lines, "\n")

	assert.NotContains(t, rendered, "sup3r-s3cr3t")
	assert.Contains(t, rendered, "10.0.0.5")
	assert.Contains(t, rendered, "cache-prod")
	assert.Contains(t, rendered, "(sensitive)")

	// Dotted keys inherit their root output's sensitivity — conn.uri and
	// conn.port both mask because conn is secret.
	for _, line := range lines {
		if strings.Contains(line, "conn.") {
			assert.Contains(t, line, "(sensitive)")
		}
	}
}

func TestFormatOutputLines_EmptyAndNilRenderNothing(t *testing.T) {
	assert.Empty(t, FormatOutputLines(nil))
	assert.Empty(t, FormatOutputLines(&outputs.CaptureResult{}))
}

func TestFormatOutputLines_AnUndeclaredOutputNeverRenders(t *testing.T) {
	// A customized module's extra output is not in the kind's schema, so
	// nothing says it is safe to print.
	result := &outputs.CaptureResult{
		Flat:    map[string]string{"host": "10.0.0.5", "extra_token": "sup3r-s3cr3t"},
		Secrets: map[string]bool{"host": false},
	}
	rendered := strings.Join(FormatOutputLines(result), "\n")
	assert.NotContains(t, rendered, "sup3r-s3cr3t")
	assert.Contains(t, rendered, "10.0.0.5")
}
