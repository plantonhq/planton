package module

import (
	"strconv"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// parseIntOrString converts a string value to the appropriate Pulumi input type
// for Kubernetes IntOrString fields.
//
// Kubernetes IntOrString can be either:
//   - An integer (e.g., 1, 5, 10) for absolute values
//   - A percentage string (e.g., "25%", "50%") for relative values
//
// This function determines the correct type:
//   - If the string ends with '%', returns pulumi.String (percentage)
//   - If the string is a valid integer, returns pulumi.Int (absolute number)
//   - Otherwise, returns pulumi.String (fallback, though this may cause validation errors)
//
// Examples:
//   - parseIntOrString("1")    → pulumi.Int(1)
//   - parseIntOrString("0")    → pulumi.Int(0)
//   - parseIntOrString("25%")  → pulumi.String("25%")
//   - parseIntOrString("100%") → pulumi.String("100%")
func parseIntOrString(value string) pulumi.Input {
	if value == "" {
		return nil
	}

	// Check if it's a percentage
	if strings.HasSuffix(value, "%") {
		return pulumi.String(value)
	}

	// Try to parse as integer
	if intValue, err := strconv.Atoi(value); err == nil {
		return pulumi.Int(intValue)
	}

	// Fallback to string (may cause validation error, but preserves original value)
	return pulumi.String(value)
}
