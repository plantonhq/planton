package crkreflect

import (
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/proto"
)

// KindToKindMetaMap builds a map of CloudResourceKind -> CloudResourceKindMeta
func KindToKindMetaMap() map[cloudresourcekind.CloudResourceKind]*cloudresourcekind.CloudResourceKindMeta {
	result := make(map[cloudresourcekind.CloudResourceKind]*cloudresourcekind.CloudResourceKindMeta)

	for _, kind := range KindsList() {
		val := kind.Descriptor().Values().ByNumber(kind.Number())
		if val == nil {
			continue
		}
		ext := proto.GetExtension(val.Options(), cloudresourcekind.E_KindMeta)
		if meta, ok := ext.(*cloudresourcekind.CloudResourceKindMeta); ok {
			if meta == nil {
				continue
			}
			result[kind] = meta
		}
	}

	return result
}
