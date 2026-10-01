package manifest

import (
	"fmt"

	validatepb "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// EnvelopeMismatches reports the manifest envelope fields (apiVersion, kind)
// that are PRESENT and conflict with the kind's schema constants, as
// plain-language findings naming the exact fix. A missing value is not a
// mismatch: surfaces that accept partial documents (chart-rendered manifests)
// rely on the platform stamping the authoritative values on write, so only a
// conflicting value is an authoring error everywhere.
func EnvelopeMismatches(manifest proto.Message) []string {
	kindName := string(manifest.ProtoReflect().Descriptor().Name())

	var findings []string
	if got := envelopeValue(manifest, "api_version"); got != "" {
		if expected := envelopeConst(manifest, "api_version"); expected != "" && got != expected {
			findings = append(findings, fmt.Sprintf(
				"apiVersion '%s' does not match kind %s: this kind requires 'apiVersion: %s'",
				got, kindName, expected))
		}
	}
	if got := envelopeValue(manifest, "kind"); got != "" {
		if expected := envelopeConst(manifest, "kind"); expected != "" && got != expected {
			findings = append(findings, fmt.Sprintf(
				"kind '%s' is not the canonical name: write 'kind: %s'", got, expected))
		}
	}
	return findings
}

// envelopeConst reads the buf.validate string const declared on a top-level
// manifest field (api_version or kind). The constants are authored in each
// kind's api.proto and gated against the kind registry by crkreflect's
// registry tests, so the descriptor is the authoritative source of the
// expected value.
func envelopeConst(manifest proto.Message, fieldName protoreflect.Name) string {
	fd := manifest.ProtoReflect().Descriptor().Fields().ByName(fieldName)
	if fd == nil {
		return ""
	}
	opts, ok := fd.Options().(proto.Message)
	if !ok || opts == nil || !proto.HasExtension(opts, validatepb.E_Field) {
		return ""
	}
	rules, ok := proto.GetExtension(opts, validatepb.E_Field).(*validatepb.FieldRules)
	if !ok || rules == nil {
		return ""
	}
	return rules.GetString().GetConst()
}

// envelopeValue reads the current value of a top-level string field.
func envelopeValue(manifest proto.Message, fieldName protoreflect.Name) string {
	fd := manifest.ProtoReflect().Descriptor().Fields().ByName(fieldName)
	if fd == nil {
		return ""
	}
	return manifest.ProtoReflect().Get(fd).String()
}

// StampEnvelope returns a copy of the manifest with a MISSING apiVersion or
// kind set to the kind's schema constant -- what the platform stamps on write
// before it validates the whole document. A present value is left as written
// (EnvelopeMismatches reports a conflicting one), so validating the stamped
// copy checks exactly the document the platform checks.
func StampEnvelope(manifest proto.Message) proto.Message {
	stamped := proto.Clone(manifest)
	for _, fieldName := range []protoreflect.Name{"api_version", "kind"} {
		fd := stamped.ProtoReflect().Descriptor().Fields().ByName(fieldName)
		if fd == nil || envelopeValue(stamped, fieldName) != "" {
			continue
		}
		if expected := envelopeConst(stamped, fieldName); expected != "" {
			stamped.ProtoReflect().Set(fd, protoreflect.ValueOfString(expected))
		}
	}
	return stamped
}
