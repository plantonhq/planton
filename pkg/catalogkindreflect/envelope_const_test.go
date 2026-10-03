package catalogkindreflect

import (
	"testing"

	validatepb "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// The envelope contract: every kind's api.proto declares hand-written
// buf.validate constants on `api_version` and `kind`, and every runtime
// surface derives the same values from the registry (provider_meta.group +
// kind_meta.version, and the effective kind name). Nothing links the two at
// authoring time, so this gate is what makes drift impossible to ship: a new
// or edited kind whose hand-written constant disagrees with its registry
// metadata fails CI the moment it is introduced.

// envelopeConstOf reads the buf.validate string const declared on a top-level
// field of the kind's message descriptor — the authoritative hand-written
// value, with no proto-text parsing (immune to quote-style variation).
func envelopeConstOf(msg proto.Message, fieldName protoreflect.Name) string {
	fd := msg.ProtoReflect().Descriptor().Fields().ByName(fieldName)
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

func TestEnvelopeConstsAgreeWithKindMeta(t *testing.T) {
	checked := 0
	for _, kind := range KindsList() {
		instance, err := NewInstance(kind)
		if err != nil {
			// Coverage is part of the contract: a registered kind without a
			// message mapping means the codegen skipped it — surface it here
			// rather than silently checking a subset.
			t.Errorf("%s: no message registered in ToMessageMap: %v", kind, err)
			continue
		}

		expectedApiVersion := GroupVersion(kind)
		if expectedApiVersion == "" {
			t.Errorf("%s: GroupVersion() returned empty — kind_meta or provider_meta is incomplete", kind)
			continue
		}
		if gotConst := envelopeConstOf(instance, "api_version"); gotConst != expectedApiVersion {
			t.Errorf("%s: api.proto api_version const %q disagrees with registry-derived %q",
				kind, gotConst, expectedApiVersion)
		}

		expectedKindName := ExtractKindNameByKind(kind)
		if expectedKindName == "" {
			t.Errorf("%s: ExtractKindNameByKind() returned empty", kind)
			continue
		}
		if gotConst := envelopeConstOf(instance, "kind"); gotConst != expectedKindName {
			t.Errorf("%s: api.proto kind const %q disagrees with registry kind name %q",
				kind, gotConst, expectedKindName)
		}

		checked++
	}
	if checked == 0 {
		t.Fatal("no kinds were checked — the registry walk is broken")
	}
	t.Logf("envelope consts verified against kind_meta for %d kinds", checked)
}
