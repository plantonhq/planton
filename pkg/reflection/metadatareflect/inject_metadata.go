package metadatareflect

import (
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// InjectMetadata sets the metadata field of msg to meta and returns the updated message.
// If meta is nil, msg is nil, or the message lacks a metadata field, msg is returned unchanged.
func InjectMetadata(msg proto.Message, meta *shared.CloudResourceMetadata) proto.Message {
	if meta == nil || msg == nil {
		return msg
	}

	msgReflect := msg.ProtoReflect()

	metadataField := msgReflect.Descriptor().Fields().ByName("metadata")
	if metadataField == nil {
		return msg
	}

	// Convert meta to a protoreflect.Message and assign it.
	msgReflect.Set(metadataField, protoreflect.ValueOfMessage(meta.ProtoReflect()))
	return msg
}
