// Package deferrules sets aside the rule violations that belong to a value a
// host resolves later.
//
// A manifest may carry, in place of a literal, a token its host replaces
// before anything deploys -- the Planton Platform's `$secret/...` and
// `$var/...` references, resolved on its runner. A kind's rules are written
// about the literal: a base64 pattern, a CIDR format, a length. Judged against
// the token, they refuse a manifest whose resolved value would pass them. So
// a host that resolves tokens defers those violations here, and applies the
// same rules to the value the token resolves to, where it resolves.
//
// The package knows no token grammar: the host passes the classifier. Only
// violations located on a spec value are deferred -- a scalar string, a map
// value or list item, or a StringValueOrRef whose literal arm holds the token.
// Violations on metadata (a name, a slug) and message-level rules are always
// kept: those are never resolved later.
package deferrules

import (
	"errors"

	"buf.build/go/protovalidate"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// stringValueOrRefFullName is the reference wrapper whose literal arm a token
// may occupy.
const stringValueOrRefFullName protoreflect.FullName = "dev.planton.shared.foreignkey.v1.StringValueOrRef"

// Split partitions a validation error's violations. Those located on a spec
// value the host's isToken classifies are returned as deferred; the rest are
// returned as kept, still a *protovalidate.ValidationError, or nil when none
// remain. An error that is not a ValidationError (a rule that failed to
// compile or evaluate) is returned as kept, unchanged: it says nothing about
// the document's values.
func Split(err error, isToken func(string) bool) (kept error, deferred []*protovalidate.Violation) {
	var validationErr *protovalidate.ValidationError
	if err == nil || isToken == nil || !errors.As(err, &validationErr) {
		return err, nil
	}
	var remaining []*protovalidate.Violation
	for _, v := range validationErr.Violations {
		if onSpec(v) && holdsToken(v, isToken) {
			deferred = append(deferred, v)
			continue
		}
		remaining = append(remaining, v)
	}
	if len(remaining) == 0 {
		return nil, deferred
	}
	return &protovalidate.ValidationError{Violations: remaining}, deferred
}

// onSpec reports whether the violation is located under the document's spec.
func onSpec(v *protovalidate.Violation) bool {
	elements := v.Proto.GetField().GetElements()
	return len(elements) > 0 && elements[0].GetFieldName() == "spec"
}

// holdsToken reports whether the value the violation is about is a token: a
// string, or a StringValueOrRef whose literal arm is one. A violation on a map
// key is about the key, which is never a token.
func holdsToken(v *protovalidate.Violation, isToken func(string) bool) bool {
	if v.Proto.GetForKey() || !v.FieldValue.IsValid() {
		return false
	}
	switch value := v.FieldValue.Interface().(type) {
	case string:
		return isToken(value)
	case protoreflect.Message:
		if value.Descriptor().FullName() != stringValueOrRefFullName {
			return false
		}
		ref, ok := value.Interface().(*foreignkeyv1.StringValueOrRef)
		return ok && ref.GetValueFrom() == nil && isToken(ref.GetValue())
	}
	return false
}
