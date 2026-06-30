package rules

import (
	"context"
	"fmt"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/check/checkutil"
	options_pb "github.com/plantonhq/planton/apis/dev/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// DefaultRequiresOptionalRule enforces that scalar fields with default values
// must be marked as optional to enable proper field presence tracking.
var DefaultRequiresOptionalRule = &check.RuleSpec{
	ID:      "DEFAULT_REQUIRES_OPTIONAL",
	Purpose: "Checks that scalar fields with (dev.planton.shared.options.default) are marked as optional.",
	Type:    check.RuleTypeLint,
	Handler: checkutil.NewFieldRuleHandler(checkDefaultRequiresOptional),
}

func checkDefaultRequiresOptional(_ context.Context, responseWriter check.ResponseWriter, _ check.Request, field protoreflect.FieldDescriptor) error {
	// Skip non-scalar types (messages are always optional, lists/maps cannot have defaults)
	if field.Kind() == protoreflect.MessageKind || field.IsList() || field.IsMap() {
		return nil
	}

	// Get field options
	opts, ok := field.Options().(*descriptorpb.FieldOptions)
	if !ok {
		return nil
	}

	// Check if field has the default extension
	if !proto.HasExtension(opts, options_pb.E_Default) {
		return nil
	}

	// If field has default but lacks presence tracking, report error
	if !field.HasPresence() {
		message := fmt.Sprintf(
			"Field %q has a default value but is not marked as optional. "+
				"Scalar fields with (dev.planton.shared.options.default) must use the 'optional' keyword "+
				"to enable proper field presence detection.\n\n"+
				"Fix: optional %s %s = %d [(dev.planton.shared.options.default) = \"...\"];",
			field.Name(),
			field.Kind().String(),
			field.Name(),
			field.Number(),
		)
		responseWriter.AddAnnotation(
			check.WithDescriptor(field),
			check.WithMessage(message),
		)
	}

	return nil
}

