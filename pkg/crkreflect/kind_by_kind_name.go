package crkreflect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

func KindByKindName(kindName string) (cloudresourcekind.CloudResourceKind, error) {
	// Iterate over the ApiResourceKind enum values
	for _, enumValue := range cloudresourcekind.CloudResourceKind_value {
		kind := cloudresourcekind.CloudResourceKind(enumValue)
		kindMeta, err := KindMeta(kind)
		if err != nil {
			continue
		}
		// Compare the kind_name in the metadata with the message kind name
		// If kindMeta.Name is empty, fall back to comparing with enum value name
		metaName := kindMeta.Name
		if metaName == "" {
			metaName = kind.String()
		}
		if metaName == kindName {
			// If it matches, return the corresponding ApiResourceKind
			return kind, nil
		}
	}
	return cloudresourcekind.CloudResourceKind_unspecified,
		errors.Errorf("no matching CloudResourceKind found for kind: %s", kindName)
}
