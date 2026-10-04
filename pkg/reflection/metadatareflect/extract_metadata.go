package metadatareflect

import (
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func ExtractMetadata(msg proto.Message) *shared.CatalogObjectMetadata {
	msgReflect := msg.ProtoReflect()

	// Check if the "status" field exists
	metadataField := msgReflect.Descriptor().Fields().ByName("metadata")
	if metadataField == nil || !msgReflect.Has(metadataField) {
		return nil
	}
	// Get the metadata field message
	metadataReflect := msgReflect.Get(metadataField).Message()

	// Marshal the message to bytes
	bytes, err := proto.Marshal(metadataReflect.Interface())
	if err != nil {
		return nil
	}

	// Unmarshal the bytes into a ResourceAudit
	var metadata shared.CatalogObjectMetadata
	err = proto.Unmarshal(bytes, &metadata)
	if err != nil {
		return nil
	}

	return &metadata
}

// ExtractLabels extracts labels from a manifest's metadata.
// Returns nil if no metadata or labels are found.
//
// Labels are user-facing organizational metadata that planton IaC modules derive
// into cloud-provider tags. Platform-behavior signals (provisioner, backend
// location, kube context, ...) live in annotations — use ExtractAnnotations.
func ExtractLabels(msg proto.Message) map[string]string {
	metadata := ExtractMetadata(msg)
	if metadata == nil {
		return nil
	}
	return metadata.Labels
}

// ExtractAnnotations extracts annotations from a manifest's metadata.
// Returns nil if no metadata or annotations are found.
//
// Annotations carry platform-behavior signals (planton.dev/provisioner, backend
// location keys, kube context, ...). Unlike labels, they are never derived into
// cloud-provider tags, so platform-internal detail cannot leak onto the user's
// real provider resources.
func ExtractAnnotations(msg proto.Message) map[string]string {
	metadata := ExtractMetadata(msg)
	if metadata == nil {
		return nil
	}
	return metadata.Annotations
}
