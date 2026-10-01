package gcpsccnotificationconfigv1alpha1

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	gcpsccbigqueryexportv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccbigqueryexport/v1alpha1"
	gcpsccmuteconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccmuteconfig/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpSccNotificationConfigSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpSccNotificationConfigSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpSccNotificationConfig {
		return &GcpSccNotificationConfig{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSccNotificationConfig",
			Metadata:   &shared.CloudResourceMetadata{Name: "high-findings"},
			Spec: &GcpSccNotificationConfigSpec{
				ConfigId:    "high-findings",
				PubsubTopic: litRef("projects/sec/topics/scc-findings"),
				Filter:      `state = "ACTIVE" AND severity = "HIGH"`,
			},
		}
	}

	ginkgo.It("should accept a project config streaming to a topic", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project config without a topic and with an empty filter", func() {
		msg := minimal()
		msg.Spec.PubsubTopic = nil
		msg.Spec.Filter = ""
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an organization config with every field set", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpSccNotificationConfigScope{OrganizationId: "123456789012"}
		msg.Spec.Description = "High findings to the SOC"
		msg.Spec.Location = "eu"
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a topic on folder and organization configs", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpSccNotificationConfigScope{FolderId: litRef("456")}
		msg.Spec.PubsubTopic = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Scope = &GcpSccNotificationConfigScope{OrganizationId: "123"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject two scope arms and a prefixed organization", func() {
		msg := minimal()
		msg.Spec.Scope = &GcpSccNotificationConfigScope{ProjectId: litRef("p"), FolderId: litRef("456")}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.Scope = &GcpSccNotificationConfigScope{OrganizationId: "organizations/123"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed config id, topic, or location", func() {
		msg := minimal()
		msg.Spec.ConfigId = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ConfigId = "has spaces"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.PubsubTopic = litRef("scc-findings")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Location = "Global"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a description over 1024 characters and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.Description = strings.Repeat("x", 1025)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})

// The three Security Command Center kinds each carry their own copy of the
// scope message (project, folder, or organization). The copies are per kind
// so every kind folder stays self-contained; this test keeps them one
// contract: a field, a reference, or a validation rule changed in one copy
// and not the others fails here, naming the copy.

func scopeShape(message protoreflect.MessageDescriptor) []string {
	lines := []string{"message " + prototext.Format(message.Options())}
	fields := message.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		kind := field.Kind().String()
		if nested := field.Message(); nested != nil {
			kind = "message:" + string(nested.FullName())
		}
		lines = append(lines, fmt.Sprintf("%s=%d %s %s options{%s}",
			field.Name(), field.Number(), kind, field.Cardinality(), prototext.Format(field.Options())))
	}
	sort.Strings(lines)
	return lines
}

func TestSccScopeIsOneContractAcrossTheThreeKinds(t *testing.T) {
	copies := map[string]protoreflect.MessageDescriptor{
		"GcpSccNotificationConfig": (&GcpSccNotificationConfigScope{}).ProtoReflect().Descriptor(),
		"GcpSccMuteConfig":         (&gcpsccmuteconfigv1alpha1.GcpSccMuteConfigScope{}).ProtoReflect().Descriptor(),
		"GcpSccBigQueryExport":     (&gcpsccbigqueryexportv1alpha1.GcpSccBigQueryExportScope{}).ProtoReflect().Descriptor(),
	}
	want := scopeShape(copies["GcpSccNotificationConfig"])
	if len(want) != 4 || !strings.Contains(want[0]+want[1]+want[2]+want[3], "at_most_one_scope") {
		t.Fatalf("the reference scope walked to %d lines without its at-most-one rule:\n  %s", len(want), strings.Join(want, "\n  "))
	}
	for _, name := range []string{"GcpSccMuteConfig", "GcpSccBigQueryExport"} {
		if got := scopeShape(copies[name]); strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s's scope differs from GcpSccNotificationConfig's\nwant:\n  %s\ngot:\n  %s",
				name, strings.Join(want, "\n  "), strings.Join(got, "\n  "))
		}
	}
}
