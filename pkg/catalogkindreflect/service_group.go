package catalogkindreflect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/shared/catalogkind"
	"google.golang.org/protobuf/proto"
)

// ServiceGroup returns the provider-console service group one kind is browsed
// under (kind_meta.service_group) — the coarse UX taxonomy, not the enum's
// family sub-bands. Kinds of providers without a service taxonomy (auth0,
// openfga, _test) return the unspecified value: an honest absence, not an
// error — the registry tests own the required-vs-prohibited contract.
func ServiceGroup(kind catalogkind.CatalogKind) (catalogkind.CatalogProviderServiceGroup, error) {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		return catalogkind.CatalogProviderServiceGroup_catalog_provider_service_group_unspecified,
			errors.Wrap(err, "while getting catalog kind meta")
	}
	return kindMeta.ServiceGroup, nil
}

// ServiceGroupMetaOf returns the service_group_meta extension (owning
// provider, display name) of one service-group enum value.
func ServiceGroupMetaOf(group catalogkind.CatalogProviderServiceGroup) (*catalogkind.CatalogProviderServiceGroupMeta, error) {
	enumValueDescriptor := group.Descriptor().Values().ByNumber(group.Number())
	if enumValueDescriptor == nil {
		return nil, errors.Errorf("no descriptor found for service group: %v", group)
	}

	options := enumValueDescriptor.Options()
	if options == nil {
		return nil, errors.Errorf("no options found for service group: %v", group)
	}

	meta, ok := proto.GetExtension(options, catalogkind.E_ServiceGroupMeta).(*catalogkind.CatalogProviderServiceGroupMeta)
	if !ok || meta == nil {
		return nil, errors.Errorf("no meta information found for service group: %v", group)
	}
	return meta, nil
}
