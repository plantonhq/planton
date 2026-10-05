package catalogkindreflect

import (
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
)

// KindToKindMetaMap builds a map of CatalogKind -> CatalogKindMeta
func KindToKindMetaMap() map[catalogkind.CatalogKind]*catalogkind.CatalogKindMeta {
	result := make(map[catalogkind.CatalogKind]*catalogkind.CatalogKindMeta)

	for _, kind := range KindsList() {
		val := kind.Descriptor().Values().ByNumber(kind.Number())
		if val == nil {
			continue
		}
		ext := proto.GetExtension(val.Options(), catalogkind.E_KindMeta)
		if meta, ok := ext.(*catalogkind.CatalogKindMeta); ok {
			if meta == nil {
				continue
			}
			result[kind] = meta
		}
	}

	return result
}
