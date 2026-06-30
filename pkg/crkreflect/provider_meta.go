package crkreflect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/proto"
)

func ProviderMeta(kind cloudresourcekind.CloudResourceKind) (*cloudresourcekind.CloudResourceProviderMeta, error) {
	kindMeta, err := KindMeta(kind)
	if err != nil {
		return nil, errors.Wrap(err, "while getting cloud resource kind meta")
	}
	provider := kindMeta.Provider

	// Get the descriptor for the enum value (CloudResourceProvider)
	enumValueDescriptor := provider.Descriptor().Values().ByNumber(provider.Number())
	if enumValueDescriptor == nil {
		return nil, errors.Errorf("no descriptor found for provider: %v", provider)
	}

	// Get the options from the enum value descriptor
	options := enumValueDescriptor.Options()
	if options == nil {
		return nil, errors.Errorf("no options found for provider: %v", provider)
	}

	// Extract the meta field from the options
	providerMeta, ok := proto.GetExtension(options, cloudresourcekind.E_ProviderMeta).(*cloudresourcekind.CloudResourceProviderMeta)
	if !ok || providerMeta == nil {
		return nil, errors.Errorf("no meta information found for provider: %v", provider)
	}

	// Return the meta information
	return providerMeta, nil
}
