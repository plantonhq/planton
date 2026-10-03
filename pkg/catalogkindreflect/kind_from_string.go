package catalogkindreflect

import (
	"github.com/plantonhq/planton/shared/catalogkind"
	"strings"
)

var AliasMap = map[catalogkind.CatalogKind][]string{}

func KindFromString(catalogKindString string) catalogkind.CatalogKind {
	// Check aliases first (exact match)
	for kind, aliases := range AliasMap {
		for _, alias := range aliases {
			if alias == catalogKindString {
				return kind
			}
		}
	}

	// Normalize the input string for comparison
	normalizedInput := strings.ReplaceAll(catalogKindString, "-", "")
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

	return catalogkind.CatalogKind_unspecified
}
