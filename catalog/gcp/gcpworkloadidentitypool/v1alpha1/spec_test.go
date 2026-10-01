package gcpworkloadidentitypoolv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpWorkloadIdentityPoolSpec Suite")
}

func ptr(s string) *string {
	return &s
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpWorkloadIdentityPoolSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpWorkloadIdentityPool.
	minimal := func() *GcpWorkloadIdentityPool {
		return &GcpWorkloadIdentityPool{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpWorkloadIdentityPool",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-pool",
			},
			Spec: &GcpWorkloadIdentityPoolSpec{
				WorkloadIdentityPoolId: "github-actions",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully populated federation spec", func() {
		msg := minimal()
		msg.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-gcp-project"},
		}
		msg.Spec.DisplayName = "GitHub Actions"
		msg.Spec.Description = "Keyless federation for the engineering org's CI"
		msg.Spec.Disabled = true
		msg.Spec.Mode = ptr("FEDERATION_ONLY")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 4-character pool id (lower bound)", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolId = "ab-1"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 32-character pool id (upper bound)", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolId = strings.Repeat("a", 32)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every creatable mode value", func() {
		for _, mode := range []string{"FEDERATION_ONLY", "TRUST_DOMAIN"} {
			msg := minimal()
			msg.Spec.Mode = ptr(mode)
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "mode %s should be valid", mode)
		}
	})

	ginkgo.It("should accept certificate issuance with own CA pools", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{
			CaPools: map[string]*foreignkeyv1.StringValueOrRef{
				"us-central1": litRef("projects/my-project/locations/us-central1/caPools/my-pool"),
			},
			KeyAlgorithm:             ptr("ECDSA_P256"),
			Lifetime:                 ptr("86400s"),
			RotationWindowPercentage: int32Ptr(60),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a trust config with anchors", func() {
		msg := minimal()
		msg.Spec.InlineTrustConfig = &GcpWorkloadIdentityPoolTrustConfig{
			AdditionalTrustBundles: []*GcpWorkloadIdentityPoolTrustBundle{{
				TrustDomain: "example.com",
				TrustAnchors: []*GcpWorkloadIdentityPoolTrustAnchor{{
					PemCertificate: "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
				}},
			}},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing pool id", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a pool id with the reserved gcp- prefix", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolId = "gcp-my-pool"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a pool id with uppercase characters", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolId = "GitHub-Actions"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a pool id shorter than 4 characters", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolId = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a pool id longer than 32 characters", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolId = strings.Repeat("a", 33)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a display_name longer than 32 characters", func() {
		msg := minimal()
		msg.Spec.DisplayName = strings.Repeat("d", 33)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a description longer than 256 characters", func() {
		msg := minimal()
		msg.Spec.Description = strings.Repeat("d", 257)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid mode value", func() {
		msg := minimal()
		msg.Spec.Mode = ptr("HYBRID")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject the Google-managed SYSTEM_TRUST_DOMAIN mode", func() {
		msg := minimal()
		msg.Spec.Mode = ptr("SYSTEM_TRUST_DOMAIN")
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject certificate issuance without CA pools", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept GcpPrivateCaPool references as ca_pools values", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{
			CaPools: map[string]*foreignkeyv1.StringValueOrRef{
				"us-central1":  nameRef("workload-ca-us-central1"),
				"europe-west1": litRef("projects/p/locations/europe-west1/caPools/workload-ca"),
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed literal ca_pools value", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{
			CaPools: map[string]*foreignkeyv1.StringValueOrRef{
				"us-central1":  nameRef("workload-ca-us-central1"),
				"europe-west1": litRef("workload-ca"),
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("a literal ca_pools value must be"))
	})

	ginkgo.It("should reject an invalid key_algorithm", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{
			CaPools:      map[string]*foreignkeyv1.StringValueOrRef{"us-central1": litRef("projects/p/locations/us-central1/caPools/c")},
			KeyAlgorithm: ptr("ED25519"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a lifetime without the seconds suffix", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{
			CaPools:  map[string]*foreignkeyv1.StringValueOrRef{"us-central1": litRef("projects/p/locations/us-central1/caPools/c")},
			Lifetime: ptr("24h"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a rotation window outside 50-80", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{
			CaPools:                  map[string]*foreignkeyv1.StringValueOrRef{"us-central1": litRef("projects/p/locations/us-central1/caPools/c")},
			RotationWindowPercentage: int32Ptr(90),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a trust config with no bundles", func() {
		msg := minimal()
		msg.Spec.InlineTrustConfig = &GcpWorkloadIdentityPoolTrustConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a trust bundle without anchors", func() {
		msg := minimal()
		msg.Spec.InlineTrustConfig = &GcpWorkloadIdentityPoolTrustConfig{
			AdditionalTrustBundles: []*GcpWorkloadIdentityPoolTrustBundle{{
				TrustDomain: "example.com",
			}},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a trust bundle that also trusts the default shared CA", func() {
		msg := minimal()
		msg.Spec.InlineTrustConfig = &GcpWorkloadIdentityPoolTrustConfig{
			AdditionalTrustBundles: []*GcpWorkloadIdentityPoolTrustBundle{{
				TrustDomain: "example.com",
				TrustAnchors: []*GcpWorkloadIdentityPoolTrustAnchor{{
					PemCertificate: "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
				}},
				TrustDefaultSharedCa: true,
			}},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept certificate issuance from the default shared CA", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{
			UseDefaultSharedCa: true,
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject certificate issuance with both CA sources", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{
			CaPools:            map[string]*foreignkeyv1.StringValueOrRef{"us-central1": litRef("projects/p/locations/us-central1/caPools/pool")},
			UseDefaultSharedCa: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject certificate issuance with no CA source", func() {
		msg := minimal()
		msg.Spec.InlineCertificateIssuanceConfig = &GcpWorkloadIdentityPoolCertificateIssuance{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept attestation rules", func() {
		msg := minimal()
		msg.Spec.AttestationRules = []*GcpWorkloadIdentityPoolAttestationRule{{
			GoogleCloudResource: "//run.googleapis.com/projects/123/type/Service/*",
		}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject an attestation rule without a resource", func() {
		msg := minimal()
		msg.Spec.AttestationRules = []*GcpWorkloadIdentityPoolAttestationRule{{}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON", ""} {
			msg := minimal()
			msg.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "FORCE"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})

func int32Ptr(v int32) *int32 {
	return &v
}
