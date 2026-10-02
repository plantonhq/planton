package gcpcertmanagertrustconfigv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCertManagerTrustConfigSpec Suite")
}

const pem = "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n"

var _ = ginkgo.Describe("GcpCertManagerTrustConfigSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCertManagerTrustConfig {
		return &GcpCertManagerTrustConfig{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCertManagerTrustConfig",
			Metadata:   &shared.CloudResourceMetadata{Name: "partner-mtls"},
			Spec: &GcpCertManagerTrustConfigSpec{
				TrustStores: []*GcpCertManagerTrustConfigTrustStore{{
					TrustAnchors:    []string{pem},
					IntermediateCas: []string{pem},
				}},
			},
		}
	}

	ginkgo.It("accepts one trust store with an anchor and an intermediate", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("accepts allowlisted certificates alone, a regional location, and every deletion policy", func() {
		msg := minimal()
		msg.Spec.TrustStores = nil
		msg.Spec.AllowlistedCertificates = []string{pem}
		msg.Spec.Location = "us-central1"
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), policy)
		}
	})

	ginkgo.It("refuses a second trust store (Google allows one)", func() {
		msg := minimal()
		msg.Spec.TrustStores = append(msg.Spec.TrustStores, &GcpCertManagerTrustConfigTrustStore{TrustAnchors: []string{pem}})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("refuses certificates that are not PEM", func() {
		msg := minimal()
		msg.Spec.TrustStores[0].TrustAnchors = []string{"MIIB-base64-only"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.TrustStores[0].IntermediateCas = []string{""}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.AllowlistedCertificates = []string{"not a certificate"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("refuses a malformed name and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.TrustConfigName = "9-starts-with-digit"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
