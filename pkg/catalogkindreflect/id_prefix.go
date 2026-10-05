package catalogkindreflect

import (
	"github.com/plantonhq/planton/shared/catalogkind"
)

// IdPrefix returns the id prefix for a catalog kind
func IdPrefix(kind catalogkind.CatalogKind) string {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		// intentionally suppressing the error to make it easy for caller
		return ""
	}
	return kindMeta.IdPrefix
}
