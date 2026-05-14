package outputs

import "strings"

// preprocessKeys normalizes IaC output map keys to match proto field paths.
//
// IaC engines (Terraform/Pulumi) may emit keys with conventions that differ
// from protobuf field naming. This function applies the same normalization as
// Java's StackOutputsMapKeyPreprocessor.process():
//
//   - ".\[" -> "[" : Remove the dot before square brackets. Terraform uses
//     "subnets.[0].id" but the dot-path walker expects "subnets[0].id".
//   - "-" -> "_" : Hyphens to underscores. Proto field names use snake_case;
//     some IaC outputs use hyphenated names.
//
// The original map is not modified; a new map is returned.
func preprocessKeys(outputs map[string]string) map[string]string {
	result := make(map[string]string, len(outputs))
	for key, val := range outputs {
		normalized := strings.ReplaceAll(key, ".[", "[")
		normalized = strings.ReplaceAll(normalized, "-", "_")
		result[normalized] = val
	}
	return result
}
