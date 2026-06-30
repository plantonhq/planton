//go:build !codegen
// +build !codegen

package outputs

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	fieldNameStatus  = "status"
	fieldNameOutputs = "outputs"
)

// resolveStackOutputsMessage creates a fresh, empty StackOutputs proto message
// for the given CloudResourceKind. It works by navigating the top-level API
// resource message's field path: status -> outputs.
//
// The returned message is the concrete generated Go type (e.g.,
// *auth0resourceserverv1.Auth0ResourceServerStackOutputs), not a dynamicpb
// message. This works because protoreflect.Message.Mutable() on a generated
// message creates the correct concrete sub-message type.
//
// Returns an error if:
//   - The kind is not registered in crkreflect.ToMessageMap
//   - The top-level message does not have a "status" field
//   - The status message does not have an "outputs" field
//   - Either field is not a message type
func resolveStackOutputsMessage(kind cloudresourcekind.CloudResourceKind) (proto.Message, error) {
	topLevel, err := crkreflect.NewInstance(kind)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create instance for kind %s", kind.String())
	}

	ref := topLevel.ProtoReflect()

	statusFd := ref.Descriptor().Fields().ByName(protoreflect.Name(fieldNameStatus))
	if statusFd == nil {
		return nil, errors.Errorf("kind %s: top-level message %s has no %q field",
			kind.String(), ref.Descriptor().FullName(), fieldNameStatus)
	}
	if statusFd.Kind() != protoreflect.MessageKind {
		return nil, errors.Errorf("kind %s: field %q is %s, not a message",
			kind.String(), fieldNameStatus, statusFd.Kind())
	}

	statusMsg := ref.Mutable(statusFd).Message()

	outputsFd := statusMsg.Descriptor().Fields().ByName(protoreflect.Name(fieldNameOutputs))
	if outputsFd == nil {
		return nil, errors.Errorf("kind %s: status message %s has no %q field",
			kind.String(), statusMsg.Descriptor().FullName(), fieldNameOutputs)
	}
	if outputsFd.Kind() != protoreflect.MessageKind {
		return nil, errors.Errorf("kind %s: field %q is %s, not a message",
			kind.String(), fieldNameOutputs, outputsFd.Kind())
	}

	outputsMsg := statusMsg.Mutable(outputsFd).Message()

	return outputsMsg.Interface(), nil
}
