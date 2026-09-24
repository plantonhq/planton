// The secret_home option's catalog-wide law: every field that says "a secret
// does not go here" names a sibling that exists and that is not itself a
// field a secret does not go in. The option's value is rendered into every
// reference page and read by platforms to phrase their refusal, so a typo'd
// home would send an author -- or an agent -- to a field that is not there.
// What this pins: the walk reaches every kind's spec (a floor on the marked
// fields found), and each home resolves in the marked field's own message.
package explain

import (
	"testing"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// markedFieldsFloor is the number of secret_home fields the catalog carried
// when the law was written; the walk must find at least that many, so a walk
// that silently stops reaching specs fails instead of passing on nothing.
const markedFieldsFloor = 4

func TestEverySecretHomeNamesAnExistingSiblingThatAcceptsSecrets(t *testing.T) {
	visited := map[protoreflect.FullName]bool{}
	marked := 0
	var walk func(protoreflect.MessageDescriptor)
	walk = func(message protoreflect.MessageDescriptor) {
		if visited[message.FullName()] {
			return
		}
		visited[message.FullName()] = true
		fields := message.Fields()
		for i := 0; i < fields.Len(); i++ {
			field := fields.Get(i)
			if home := secretHomeOf(field); home != "" {
				marked++
				sibling := message.Fields().ByName(protoreflect.Name(home))
				switch {
				case sibling == nil:
					t.Errorf("%s names secret home %q, which %s does not have", field.FullName(), home, message.FullName())
				case sibling == field:
					t.Errorf("%s names itself as its secret home", field.FullName())
				case secretHomeOf(sibling) != "":
					t.Errorf("%s names secret home %q, which refuses secrets itself", field.FullName(), home)
				}
			}
			if next := field.Message(); next != nil {
				walk(next)
			}
			if field.IsMap() {
				if value := field.MapValue().Message(); value != nil {
					walk(value)
				}
			}
		}
	}
	for _, message := range crkreflect.ToMessageMap {
		walk(message.ProtoReflect().Descriptor())
	}
	if marked < markedFieldsFloor {
		t.Fatalf("found %d secret_home fields across every kind, fewer than the %d the catalog carries -- the walk is not reaching the specs", marked, markedFieldsFloor)
	}
}

func secretHomeOf(field protoreflect.FieldDescriptor) string {
	opts, ok := field.Options().(proto.Message)
	if !ok || opts == nil || !proto.HasExtension(opts, options.E_SecretHome) {
		return ""
	}
	return proto.GetExtension(opts, options.E_SecretHome).(string)
}
