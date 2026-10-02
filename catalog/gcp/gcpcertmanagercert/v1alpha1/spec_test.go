package gcpcertmanagercertv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestGcpCertManagerCertSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCertManagerCertSpec Validation Tests")
}

// literal wraps a string in a StringValueOrRef literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testPemCertificate = "-----BEGIN CERTIFICATE-----\nMIIB...test...\n-----END CERTIFICATE-----"

const testPemPrivateKey = "-----BEGIN PRIVATE KEY-----\nMIIE...test...\n-----END PRIVATE KEY-----"

// managedCert returns a valid managed-arm certificate that cases mutate.
func managedCert() *GcpCertManagerCert {
	return &GcpCertManagerCert{
		ApiVersion: "gcp.planton.dev/v1alpha1",
		Kind:       "GcpCertManagerCert",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-certificate",
		},
		Spec: &GcpCertManagerCertSpec{
			ProjectId: literal("test-project-123"),
			Managed: &GcpCertManagerCertManaged{
				Domains: []string{"example.com"},
				DnsAuthorizations: []*foreignkeyv1.StringValueOrRef{
					literal("projects/test-project-123/locations/global/dnsAuthorizations/example-com-auth"),
				},
			},
		},
	}
}

// selfManagedCert returns a valid self-managed-arm certificate.
func selfManagedCert() *GcpCertManagerCert {
	return &GcpCertManagerCert{
		ApiVersion: "gcp.planton.dev/v1alpha1",
		Kind:       "GcpCertManagerCert",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-uploaded-certificate",
		},
		Spec: &GcpCertManagerCertSpec{
			ProjectId: literal("test-project-123"),
			SelfManaged: &GcpCertManagerCertSelfManaged{
				PemCertificate: testPemCertificate,
				PemPrivateKey:  testPemPrivateKey,
			},
		},
	}
}

var _ = ginkgo.Describe("GcpCertManagerCertSpec Validation Tests", func() {

	ginkgo.Describe("Valid configurations", func() {

		ginkgo.It("should accept a minimal managed certificate with DNS authorization", func() {
			gomega.Expect(protovalidate.Validate(managedCert())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a managed certificate with load-balancer authorization (no auth, no issuance)", func() {
			input := managedCert()
			input.Spec.Managed.DnsAuthorizations = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a wildcard domain with DNS authorization", func() {
			input := managedCert()
			input.Spec.Managed.Domains = []string{"example.com", "*.example.com"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an issuance config instead of DNS authorizations", func() {
			input := managedCert()
			input.Spec.Managed.DnsAuthorizations = nil
			input.Spec.Managed.IssuanceConfig = literal("projects/test-project-123/locations/global/certificateIssuanceConfigs/private-ca")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept ALL_REGIONS scope on a global certificate", func() {
			input := managedCert()
			input.Spec.Scope = "ALL_REGIONS"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a regional certificate", func() {
			input := managedCert()
			input.Spec.Location = "us-central1"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a self-managed certificate", func() {
			gomega.Expect(protovalidate.Validate(selfManagedCert())).To(gomega.BeNil())
		})

		ginkgo.It("should accept an omitted project_id (ambient project)", func() {
			input := managedCert()
			input.Spec.ProjectId = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept labels and an explicit cert name", func() {
			input := managedCert()
			input.Spec.CertName = "prod-web-cert"
			input.Spec.Labels = map[string]string{"team": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept the CLIENT_AUTH scope", func() {
			input := selfManagedCert()
			input.Spec.Scope = "CLIENT_AUTH"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Invalid configurations", func() {

		ginkgo.It("should reject a spec with neither arm", func() {
			input := managedCert()
			input.Spec.Managed = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a spec with both arms", func() {
			input := managedCert()
			input.Spec.SelfManaged = &GcpCertManagerCertSelfManaged{
				PemCertificate: testPemCertificate,
				PemPrivateKey:  testPemPrivateKey,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a managed arm without domains", func() {
			input := managedCert()
			input.Spec.Managed.Domains = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a domain with a trailing dot", func() {
			input := managedCert()
			input.Spec.Managed.Domains = []string{"example.com."}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should accept an issuance config by reference and a wildcard under it", func() {
			input := managedCert()
			input.Spec.Managed.DnsAuthorizations = nil
			input.Spec.Managed.Domains = []string{"*.internal.example.com"}
			input.Spec.Managed.IssuanceConfig = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{
					Name:      "internal-tls",
					FieldPath: "status.outputs.issuance_config_id",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject a literal issuance config that is not a full resource name", func() {
			input := managedCert()
			input.Spec.Managed.DnsAuthorizations = nil
			input.Spec.Managed.IssuanceConfig = literal("private-ca")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject DNS authorizations together with an issuance config", func() {
			input := managedCert()
			input.Spec.Managed.IssuanceConfig = literal("projects/p/locations/global/certificateIssuanceConfigs/ca")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a wildcard domain under load-balancer authorization", func() {
			input := managedCert()
			input.Spec.Managed.Domains = []string{"*.example.com"}
			input.Spec.Managed.DnsAuthorizations = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject ALL_REGIONS scope on a regional certificate", func() {
			input := managedCert()
			input.Spec.Scope = "ALL_REGIONS"
			input.Spec.Location = "us-central1"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid scope", func() {
			input := managedCert()
			input.Spec.Scope = "REGIONAL"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a cert name starting with a digit", func() {
			input := managedCert()
			input.Spec.CertName = "1-cert"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a cert name longer than 64 characters", func() {
			input := managedCert()
			input.Spec.CertName = "a" + strings.Repeat("b", 64)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a self-managed arm missing the private key", func() {
			input := selfManagedCert()
			input.Spec.SelfManaged.PemPrivateKey = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject swapped certificate and key material", func() {
			input := selfManagedCert()
			input.Spec.SelfManaged.PemCertificate = testPemPrivateKey
			input.Spec.SelfManaged.PemPrivateKey = testPemCertificate
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid deletion_policy", func() {
			input := managedCert()
			input.Spec.DeletionPolicy = "KEEP"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("Deletion policy", func() {
		ginkgo.It("should accept each deletion_policy value", func() {
			for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
				input := managedCert()
				input.Spec.DeletionPolicy = v
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})
	})
})
