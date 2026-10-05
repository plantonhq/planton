//go:build !codegen
// +build !codegen

package secretcoverage

import (
	"strings"
	"testing"

	"github.com/plantonhq/planton/shared/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

// TestSecretOutputShape is the CI guardrail for secret outputs: every `sensitive` mark in a
// kind's outputs sits on a top-level field, carries no value rule, and no output carries
// an exemption reason.
func TestSecretOutputShape(t *testing.T) {
	for _, v := range OutputShapeViolations() {
		t.Errorf("%s:%s -- %s", v.Kind, v.Path, v.Reason)
	}
}

// The shape rules fire on a hand-built outputs message: a mark below the top level, and an
// exemption reason, are each refused; a top-level mark is not.
func TestCollectOutputShapeViolations_FiresOnEachBrokenShape(t *testing.T) {
	sensitive := func() *descriptorpb.FieldOptions {
		opts := &descriptorpb.FieldOptions{}
		proto.SetExtension(opts, options.E_Sensitive, true)
		return opts
	}
	exempt := &descriptorpb.FieldOptions{}
	proto.SetExtension(exempt, options.E_SensitiveExemptReason, "a public value")
	str := descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()
	msg := descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum()
	optional := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum()

	file, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name:       proto.String("fixture/outputs.proto"),
		Package:    proto.String("fixture"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"shared/options/options.proto"},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("FixtureOutputs"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("client_secret"), Number: proto.Int32(1), Type: str, Label: optional, Options: sensitive()},
					{Name: proto.String("site_token"), Number: proto.Int32(2), Type: str, Label: optional, Options: exempt},
					{Name: proto.String("connection"), Number: proto.Int32(3), Type: msg, Label: optional, TypeName: proto.String(".fixture.Connection")},
				},
			},
			{
				Name: proto.String("Connection"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{Name: proto.String("password"), Number: proto.Int32(1), Type: str, Label: optional, Options: sensitive()},
				},
			},
		},
	}, protoregistry.GlobalFiles)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, v := range collectOutputShapeViolations(file.Messages().ByName("FixtureOutputs"), "Fixture") {
		got[v.Path] = v.Reason
	}
	if _, ok := got["status.outputs.client_secret"]; ok {
		t.Errorf("a top-level secret output is the right shape, got %q", got["status.outputs.client_secret"])
	}
	if reason := got["status.outputs.connection.password"]; !strings.Contains(reason, "top-level") {
		t.Errorf("a nested secret output must be refused, got %q", reason)
	}
	if reason := got["status.outputs.site_token"]; !strings.Contains(reason, "sensitive_exempt_reason") {
		t.Errorf("an exemption reason on an output must be refused, got %q", reason)
	}
}
