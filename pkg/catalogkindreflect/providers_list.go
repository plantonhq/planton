package catalogkindreflect

import "github.com/plantonhq/planton/shared/catalogkind"

func ProvidersList() []catalogkind.CatalogProvider {
	resp := make([]catalogkind.CatalogProvider, 0)
	// Iterate over all the enum values in ApiResourceKind
	for _, enumValue := range catalogkind.CatalogProvider_value {
		resp = append(resp, catalogkind.CatalogProvider(enumValue))
	}
	return resp
}
