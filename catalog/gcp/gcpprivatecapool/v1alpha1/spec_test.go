package gcpprivatecapoolv1alpha1

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	gcpprivatecacertificatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecacertificate/v1alpha1"
	gcpprivatecacertificateauthorityv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecacertificateauthority/v1alpha1"
	gcpprivatecacertificatetemplatev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprivatecacertificatetemplate/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpPrivateCaPoolSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpPrivateCaPoolSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	base := func() *GcpPrivateCaPool {
		return &GcpPrivateCaPool{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpPrivateCaPool",
			Metadata:   &shared.CloudResourceMetadata{Name: "internal-tls"},
			Spec: &GcpPrivateCaPoolSpec{
				Location: "us-central1",
				Tier:     "DEVOPS",
			},
		}
	}

	ginkgo.It("should accept a minimal pool", func() {
		gomega.Expect(validator.Validate(base())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an Enterprise pool with every lever", func() {
		r := base()
		r.Spec.ProjectId = litRef("p")
		r.Spec.CaPoolId = "internal_tls-1"
		r.Spec.Tier = "ENTERPRISE"
		r.Spec.KmsKeyName = litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k")
		r.Spec.Labels = map[string]string{"team": "platform"}
		r.Spec.DeletionPolicy = "PREVENT"
		r.Spec.PublishingOptions = &GcpPrivateCaPoolPublishingOptions{PublishCaCert: true, PublishCrl: true, EncodingFormat: "DER"}
		r.Spec.IssuancePolicy = &GcpPrivateCaPoolIssuancePolicy{
			AllowedKeyTypes: []*GcpPrivateCaPoolAllowedKeyType{
				{Rsa: &GcpPrivateCaPoolRsaKeyType{MinModulusSize: 2048, MaxModulusSize: 4096}},
				{EllipticCurveSignatureAlgorithm: "ECDSA_P256"},
				{Rsa: &GcpPrivateCaPoolRsaKeyType{}},
			},
			MaximumLifetime:      "7776000s",
			BackdateDuration:     "172800s",
			AllowedIssuanceModes: &GcpPrivateCaPoolIssuanceModes{AllowCsrBasedIssuance: true},
			IdentityConstraints: &GcpPrivateCaPoolIdentityConstraints{
				AllowSubjectAltNamesPassthrough: true,
				CelExpression:                   &GcpPrivateCaPoolCelExpression{Expression: "subject_alt_names.all(san, san.type == DNS)", Title: "dns only"},
			},
			BaselineValues: &GcpPrivateCaPoolX509Parameters{
				KeyUsage: &GcpPrivateCaPoolKeyUsage{
					BaseKeyUsage:             &GcpPrivateCaPoolBaseKeyUsage{DigitalSignature: true, KeyEncipherment: true},
					ExtendedKeyUsage:         &GcpPrivateCaPoolExtendedKeyUsage{ServerAuth: true, ClientAuth: true},
					UnknownExtendedKeyUsages: []*GcpPrivateCaPoolObjectId{{ObjectIdPath: []int32{1, 3, 6, 1, 5, 5, 7, 3, 17}}},
				},
				CaOptions:            &GcpPrivateCaPoolCaOptions{IsCa: proto.Bool(false), MaxIssuerPathLength: proto.Int32(0)},
				PolicyIds:            []*GcpPrivateCaPoolObjectId{{ObjectIdPath: []int32{2, 23, 140, 1, 2, 1}}},
				AiaOcspServers:       []string{"http://ocsp.example.com"},
				AdditionalExtensions: []*GcpPrivateCaPoolX509Extension{{ObjectId: &GcpPrivateCaPoolObjectId{ObjectIdPath: []int32{1, 2, 3}}, Value: "AAEC", Critical: true}},
				NameConstraints:      &GcpPrivateCaPoolNameConstraints{Critical: true, PermittedDnsNames: []string{"internal.example.com"}},
			},
		}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a missing or unknown tier", func() {
		for _, tier := range []string{"", "STANDARD", "devops"} {
			r := base()
			r.Spec.Tier = tier
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), tier)
		}
	})

	ginkgo.It("should reject a missing or malformed location", func() {
		for _, location := range []string{"", "us", "us-central1-a"} {
			r := base()
			r.Spec.Location = location
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), location)
		}
	})

	ginkgo.It("should reject a pool id Google refuses", func() {
		for _, id := range []string{"has space", "dots.are.out", string(make([]byte, 64))} {
			r := base()
			r.Spec.CaPoolId = id
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), id)
		}
	})

	ginkgo.It("should take exactly one form per allowed key type", func() {
		r := base()
		r.Spec.IssuancePolicy = &GcpPrivateCaPoolIssuancePolicy{AllowedKeyTypes: []*GcpPrivateCaPoolAllowedKeyType{{}}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "neither")
		r.Spec.IssuancePolicy.AllowedKeyTypes = []*GcpPrivateCaPoolAllowedKeyType{{Rsa: &GcpPrivateCaPoolRsaKeyType{}, EllipticCurveSignatureAlgorithm: "ECDSA_P384"}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "both")
		r.Spec.IssuancePolicy.AllowedKeyTypes = []*GcpPrivateCaPoolAllowedKeyType{{EllipticCurveSignatureAlgorithm: "ED448"}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "unknown curve")
	})

	ginkgo.It("should reject an RSA range upside down or negative", func() {
		r := base()
		r.Spec.IssuancePolicy = &GcpPrivateCaPoolIssuancePolicy{AllowedKeyTypes: []*GcpPrivateCaPoolAllowedKeyType{{Rsa: &GcpPrivateCaPoolRsaKeyType{MinModulusSize: 4096, MaxModulusSize: 2048}}}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
		r.Spec.IssuancePolicy.AllowedKeyTypes[0].Rsa = &GcpPrivateCaPoolRsaKeyType{MinModulusSize: -1}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
		r.Spec.IssuancePolicy.AllowedKeyTypes[0].Rsa = &GcpPrivateCaPoolRsaKeyType{MinModulusSize: 3072}
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed(), "a floor with no ceiling")
	})

	ginkgo.It("should hold durations to Google's shape and the 48-hour backdate limit", func() {
		r := base()
		r.Spec.IssuancePolicy = &GcpPrivateCaPoolIssuancePolicy{BackdateDuration: "172801s"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "over 48h")
		r.Spec.IssuancePolicy.BackdateDuration = "1h"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "not seconds")
		r.Spec.IssuancePolicy.BackdateDuration = "3600.5s"
		gomega.Expect(validator.Validate(r)).To(gomega.Succeed())
		r.Spec.IssuancePolicy.MaximumLifetime = "90d"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), "days")
	})

	ginkgo.It("should require the CEL expression text in the pool's identity constraints", func() {
		r := base()
		r.Spec.IssuancePolicy = &GcpPrivateCaPoolIssuancePolicy{IdentityConstraints: &GcpPrivateCaPoolIdentityConstraints{CelExpression: &GcpPrivateCaPoolCelExpression{Title: "t"}}}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject malformed X.509 baseline values", func() {
		cases := map[string]*GcpPrivateCaPoolX509Parameters{
			"empty OID":               {PolicyIds: []*GcpPrivateCaPoolObjectId{{}}},
			"negative arc":            {PolicyIds: []*GcpPrivateCaPoolObjectId{{ObjectIdPath: []int32{2, -1}}}},
			"extension without id":    {AdditionalExtensions: []*GcpPrivateCaPoolX509Extension{{Value: "AA=="}}},
			"extension without value": {AdditionalExtensions: []*GcpPrivateCaPoolX509Extension{{ObjectId: &GcpPrivateCaPoolObjectId{ObjectIdPath: []int32{1}}}}},
			"negative path length":    {CaOptions: &GcpPrivateCaPoolCaOptions{MaxIssuerPathLength: proto.Int32(-1)}},
		}
		for name, values := range cases {
			r := base()
			r.Spec.IssuancePolicy = &GcpPrivateCaPoolIssuancePolicy{BaselineValues: values}
			gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed(), name)
		}
	})

	ginkgo.It("should reject an unknown encoding format or deletion policy", func() {
		r := base()
		r.Spec.PublishingOptions = &GcpPrivateCaPoolPublishingOptions{EncodingFormat: "BER"}
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
		r = base()
		r.Spec.DeletionPolicy = "FORCE"
		gomega.Expect(validator.Validate(r)).ToNot(gomega.Succeed())
	})
})

// The Certificate Authority Service kinds each carry their own copy of
// Google's X509Parameters message (the pool's baseline values, the
// authority's CA certificate, the template's predefined values, the
// certificate's X.509 config), and the authority and certificate each carry
// a copy of the subject messages. The copies are per kind so every kind
// folder stays self-contained; this test is what keeps them one shape: a
// field added, renumbered, or retyped in one copy and not the others fails
// here, naming the copy and the field.

// shape renders a message tree as sorted "path name=number kind cardinality
// presence" lines, following nested messages declared by the same kind (the
// family), so two copies compare equal exactly when they are one shape.
func shape(message protoreflect.MessageDescriptor, prefix string) []string {
	var lines []string
	fields := message.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		kind := field.Kind().String()
		if nested := field.Message(); nested != nil {
			kind = "message:" + strings.TrimPrefix(string(nested.Name()), familyPrefix(message))
		}
		lines = append(lines, fmt.Sprintf("%s%s=%d %s %s presence=%t",
			prefix, field.Name(), field.Number(), kind, field.Cardinality(), field.HasPresence()))
		if nested := field.Message(); nested != nil && nested.ParentFile() == message.ParentFile() {
			lines = append(lines, shape(nested, prefix+string(field.Name())+".")...)
		}
	}
	sort.Strings(lines)
	return lines
}

// familyPrefix is the kind name each copy's message names start with, so
// "GcpPrivateCaPoolKeyUsage" and "GcpPrivateCaCertificateKeyUsage" compare
// as the same "KeyUsage".
func familyPrefix(message protoreflect.MessageDescriptor) string {
	name := string(message.Name())
	for _, prefix := range []string{
		"GcpPrivateCaCertificateAuthority",
		"GcpPrivateCaCertificateTemplate",
		"GcpPrivateCaCertificate",
		"GcpPrivateCaPool",
	} {
		if strings.HasPrefix(name, prefix) {
			return prefix
		}
	}
	return ""
}

// assertOneShape fails when any copy differs from the first, and when the
// walk finds fewer than floor fields -- so a walk that stops reaching the
// nested messages cannot pass by comparing nothing.
func assertOneShape(t *testing.T, family string, floor int, copies map[string]protoreflect.MessageDescriptor) {
	t.Helper()
	var reference string
	var want []string
	names := make([]string, 0, len(copies))
	for name := range copies {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		got := shape(copies[name], "")
		if want == nil {
			reference, want = name, got
			if len(want) < floor {
				t.Fatalf("%s: walked %d fields of %s, fewer than the %d the family has", family, len(want), name, floor)
			}
			continue
		}
		if strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%s: %s differs from %s\n%s:\n  %s\n%s:\n  %s", family, name, reference,
				reference, strings.Join(want, "\n  "), name, strings.Join(got, "\n  "))
		}
	}
}

func TestX509ParametersIsOneShapeAcrossTheFourKinds(t *testing.T) {
	assertOneShape(t, "X509Parameters", 41, map[string]protoreflect.MessageDescriptor{
		"GcpPrivateCaPool":                 (&GcpPrivateCaPoolX509Parameters{}).ProtoReflect().Descriptor(),
		"GcpPrivateCaCertificateAuthority": (&gcpprivatecacertificateauthorityv1alpha1.GcpPrivateCaCertificateAuthorityX509Parameters{}).ProtoReflect().Descriptor(),
		"GcpPrivateCaCertificateTemplate":  (&gcpprivatecacertificatetemplatev1alpha1.GcpPrivateCaCertificateTemplateX509Parameters{}).ProtoReflect().Descriptor(),
		"GcpPrivateCaCertificate":          (&gcpprivatecacertificatev1alpha1.GcpPrivateCaCertificateX509Parameters{}).ProtoReflect().Descriptor(),
	})
}

func TestSubjectConfigIsOneShapeAcrossTheAuthorityAndTheCertificate(t *testing.T) {
	assertOneShape(t, "SubjectConfig", 14, map[string]protoreflect.MessageDescriptor{
		"GcpPrivateCaCertificateAuthority": (&gcpprivatecacertificateauthorityv1alpha1.GcpPrivateCaCertificateAuthoritySubjectConfig{}).ProtoReflect().Descriptor(),
		"GcpPrivateCaCertificate":          (&gcpprivatecacertificatev1alpha1.GcpPrivateCaCertificateSubjectConfig{}).ProtoReflect().Descriptor(),
	})
}

func TestIdentityConstraintsIsOneShapeAcrossThePoolAndTheTemplate(t *testing.T) {
	assertOneShape(t, "IdentityConstraints", 7, map[string]protoreflect.MessageDescriptor{
		"GcpPrivateCaPool":                (&GcpPrivateCaPoolIdentityConstraints{}).ProtoReflect().Descriptor(),
		"GcpPrivateCaCertificateTemplate": (&gcpprivatecacertificatetemplatev1alpha1.GcpPrivateCaCertificateTemplateIdentityConstraints{}).ProtoReflect().Descriptor(),
	})
}
