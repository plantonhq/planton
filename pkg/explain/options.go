package explain

import (
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// OptionInterpreter reads one custom-option family off a field descriptor
// into the report field. The engine runs every registered interpreter on
// every field; each option family (this repo's dev.planton.shared options, a
// host platform's own commons) ships its own interpreter instead of forking
// the walker. Interpreters must only ADD to the field, never reset what an
// earlier interpreter wrote.
type OptionInterpreter func(protoreflect.FieldDescriptor, *Field)

// SharedOptions interprets this repo's dev.planton.shared option family:
// sensitivity, secret homes, recommended defaults, and foreign-key reference
// targets.
func SharedOptions(fd protoreflect.FieldDescriptor, f *Field) {
	opts, ok := fd.Options().(proto.Message)
	if !ok || opts == nil {
		return
	}
	if proto.HasExtension(opts, options.E_Sensitive) {
		f.Sensitive = proto.GetExtension(opts, options.E_Sensitive).(bool)
	}
	if proto.HasExtension(opts, options.E_SecretHome) {
		if home := proto.GetExtension(opts, options.E_SecretHome).(string); home != "" {
			// Rendered in the spelling the report's paths use (the JSON name),
			// so an author can type what the page names.
			f.SecretHome = home
			if sibling := fd.ContainingMessage().Fields().ByName(protoreflect.Name(home)); sibling != nil {
				f.SecretHome = sibling.JSONName()
			}
		}
	}
	if proto.HasExtension(opts, options.E_RecommendedDefault) {
		f.RecommendedDefault = proto.GetExtension(opts, options.E_RecommendedDefault).(string)
	}
	if f.RecommendedDefault == "" && proto.HasExtension(opts, options.E_Default) {
		f.RecommendedDefault = proto.GetExtension(opts, options.E_Default).(string)
	}
	if proto.HasExtension(opts, foreignkeyv1.E_DefaultKind) {
		if kind, ok := proto.GetExtension(opts, foreignkeyv1.E_DefaultKind).(cloudresourcekind.CloudResourceKind); ok &&
			kind != cloudresourcekind.CloudResourceKind_unspecified {
			f.RefKind = kind.String()
		}
	}
	if proto.HasExtension(opts, foreignkeyv1.E_DefaultKindFieldPath) {
		f.RefFieldPath = proto.GetExtension(opts, foreignkeyv1.E_DefaultKindFieldPath).(string)
	}
}
