package crkreflect

import (
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

// IdPrefix returns the id prefix for a cloud resource kind
func IdPrefix(kind cloudresourcekind.CloudResourceKind) string {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		// intentionally suppressing the error to make it easy for caller
		return ""
	}
	return kindMeta.IdPrefix
}
