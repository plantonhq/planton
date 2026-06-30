package crkreflect

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"google.golang.org/protobuf/proto"
)

var NoKindMetaError = errors.Errorf("no kind meta found")

func KindMeta(kind cloudresourcekind.CloudResourceKind) (*cloudresourcekind.CloudResourceKindMeta, error) {
	// Get the descriptor for the enum value (CloudResourceKind)
	enumValueDescriptor := kind.Descriptor().Values().ByNumber(kind.Number())
	if enumValueDescriptor == nil {
		return nil, errors.Errorf("no descriptor found for kind: %v", kind)
	}

	// Get the options from the enum value descriptor
	options := enumValueDescriptor.Options()
	if options == nil {
		return nil, errors.Errorf("no options found for kind: %v", kind)
	}

	// Extract the meta field from the options
	meta, ok := proto.GetExtension(options, cloudresourcekind.E_KindMeta).(*cloudresourcekind.CloudResourceKindMeta)
	if !ok || meta == nil {
		return nil, NoKindMetaError
	}

	// Return the meta information
	return meta, nil
}
