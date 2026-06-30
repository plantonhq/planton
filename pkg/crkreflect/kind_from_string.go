package crkreflect

import (
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"strings"
)

var AliasMap = map[cloudresourcekind.CloudResourceKind][]string{}

func KindFromString(cloudResourceKindString string) cloudresourcekind.CloudResourceKind {
	// Check aliases first (exact match)
	for kind, aliases := range AliasMap {
		for _, alias := range aliases {
			if alias == cloudResourceKindString {
				return kind
			}
		}
	}

	// Normalize the input string for comparison
	normalizedInput := strings.ReplaceAll(cloudResourceKindString, "-", "")
	normalizedInput = strings.ReplaceAll(normalizedInput, "_", "")
	normalizedInput = strings.ToLower(normalizedInput)

	for _, k := range KindsList() {
		// Normalize the enum value for comparison
		normalizedEnum := strings.ReplaceAll(k.String(), "-", "")
		normalizedEnum = strings.ReplaceAll(normalizedEnum, "_", "")
		normalizedEnum = strings.ToLower(normalizedEnum)

		// Compare normalized values
		if normalizedInput == normalizedEnum {
			return k
		}
	}

	return cloudresourcekind.CloudResourceKind_unspecified
}
