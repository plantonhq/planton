package gcpcertmanagerissuanceconfigv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCertManagerIssuanceConfigSpec Suite")
}

func valueOf(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

var _ = ginkgo.Describe("GcpCertManagerIssuanceConfigSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCertManagerIssuanceConfig {
		return &GcpCertManagerIssuanceConfig{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCertManagerIssuanceConfig",
			Metadata:   &shared.CloudResourceMetadata{Name: "internal-tls"},
			Spec: &GcpCertManagerIssuanceConfigSpec{
				CaPool:                   valueOf("projects/my-project/locations/us-central1/caPools/internal"),
				KeyAlgorithm:             "ECDSA_P256",
				Lifetime:                 "2592000s",
				RotationWindowPercentage: 66,
			},
		}
	}

	ginkgo.It("accepts a 30-day ECDSA config renewing at 66%", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("accepts the lifetime bounds, fractional seconds, and RSA", func() {
		for _, lifetime := range []string{"1814400s", "2000000.5s", "2592000.000s"} {
			msg := minimal()
			msg.Spec.Lifetime = lifetime
			msg.Spec.RotationWindowPercentage = 50
			msg.Spec.KeyAlgorithm = "RSA_2048"
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), lifetime)
		}
	})

	ginkgo.It("refuses a lifetime outside 21-30 days or not in seconds", func() {
		for _, lifetime := range []string{"1814399s", "2592000.5s", "2592001s", "30d", "720h", ""} {
			msg := minimal()
			msg.Spec.Lifetime = lifetime
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), lifetime)
		}
	})

	ginkgo.It("enforces Google's seven-day renewal margins", func() {
		msg := minimal()
		msg.Spec.Lifetime = "1814400s" // 21 days: usable window 34-66
		for pct, ok := range map[int32]bool{33: false, 34: true, 66: true, 67: false} {
			msg.Spec.RotationWindowPercentage = pct
			if ok {
				gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), "pct %d", pct)
			} else {
				gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "pct %d", pct)
			}
		}
		msg.Spec.Lifetime = "2592000s" // 30 days: usable window 24-76
		msg.Spec.RotationWindowPercentage = 23
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.RotationWindowPercentage = 76
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("refuses an out-of-range percentage, an unknown algorithm, and a malformed pool", func() {
		msg := minimal()
		msg.Spec.RotationWindowPercentage = 0
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.KeyAlgorithm = "ECDSA_P384"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CaPool = valueOf("internal")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CaPool = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
