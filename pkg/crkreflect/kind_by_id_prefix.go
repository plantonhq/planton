package crkreflect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

// KindByIdPrefix takes an id prefix as input and returns the corresponding CloudResourceKind.
func KindByIdPrefix(idPrefix string) (cloudresourcekind.CloudResourceKind, error) {
	// Iterate over all the enum values in CloudResourceKind
	for _, enumValue := range cloudresourcekind.CloudResourceKind_value {
		kind := cloudresourcekind.CloudResourceKind(enumValue)

		kindMeta, err := KindMeta(kind)
		if err != nil {
			continue
		}
		// Compare the id_prefix in the meta with the input idPrefix
		if kindMeta.IdPrefix == idPrefix {
			// If it matches, return the corresponding CloudResourceKind
			return kind, nil
		}
	}

	// If no match is found, return an error
	return cloudresourcekind.CloudResourceKind_unspecified,
		errors.Errorf("no matching CloudResourceKind found for id prefix: %s", idPrefix)
}
