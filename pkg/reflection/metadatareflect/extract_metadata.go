package metadatareflect

import (
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"google.golang.org/protobuf/proto"
)

func ExtractMetadata(msg proto.Message) *shared.CloudResourceMetadata {
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
	var metadata shared.CloudResourceMetadata
	err = proto.Unmarshal(bytes, &metadata)
	if err != nil {
		return nil
	}

	return &metadata
}

// ExtractLabels extracts labels from a manifest's metadata
// Returns nil if no metadata or labels are found
func ExtractLabels(msg proto.Message) map[string]string {
	metadata := ExtractMetadata(msg)
	if metadata == nil {
		return nil
	}
	return metadata.Labels
}
