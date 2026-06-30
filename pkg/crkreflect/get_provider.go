package crkreflect

import (
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

// GetProvider returns the Cloud‑resource **provider** recorded in the
// (provider) enum‑value option of the given CloudResourceKind.
//
// If the kind is unknown or the option is absent, the function returns the
// “unspecified” sentinel.
func GetProvider(
	kind cloudresourcekind.CloudResourceKind,
) cloudresourcekind.CloudResourceProvider {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		return cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified
	}
	return kindMeta.Provider
}
