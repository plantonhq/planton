package gcptargethttpsproxyv1alpha1

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
	ginkgo.RunSpecs(t, "GcpTargetHttpsProxySpec Suite")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func urlMapSelfLink() *foreignkeyv1.StringValueOrRef {
	return literalRef("https://www.googleapis.com/compute/v1/projects/p/global/urlMaps/web-routing")
}

func certSelfLink() *foreignkeyv1.StringValueOrRef {
	return literalRef("https://www.googleapis.com/compute/v1/projects/p/global/sslCertificates/web-cert")
}

var _ = ginkgo.Describe("GcpTargetHttpsProxySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpTargetHttpsProxy {
		return &GcpTargetHttpsProxy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpTargetHttpsProxy",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-https-proxy",
			},
			Spec: &GcpTargetHttpsProxySpec{
				UrlMap:          urlMapSelfLink(),
				SslCertificates: []*foreignkeyv1.StringValueOrRef{certSelfLink()},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec with url_map and one certificate", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept fifteen ssl_certificates (the GCP limit)", func() {
		target := minimal()
		certs := make([]*foreignkeyv1.StringValueOrRef, 0, 15)
		for i := 0; i < 15; i++ {
			certs = append(certs, certSelfLink())
		}
		target.Spec.SslCertificates = certs
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept certificate_manager_certificates alone", func() {
		target := minimal()
		target.Spec.SslCertificates = nil
		target.Spec.CertificateManagerCertificates = []*foreignkeyv1.StringValueOrRef{
			literalRef("projects/p/locations/global/certificates/cm-cert"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept certificate_map alone", func() {
		target := minimal()
		target.Spec.SslCertificates = nil
		target.Spec.CertificateMap = "//certificatemanager.googleapis.com/projects/p/locations/global/certificateMaps/saas-domains"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a Traffic Director proxy with server_tls_policy and no certificates", func() {
		target := minimal()
		target.Spec.SslCertificates = nil
		target.Spec.ServerTlsPolicy = literalRef("projects/p/locations/global/serverTlsPolicies/mtls-policy")
		target.Spec.ProxyBind = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an ssl_policy reference", func() {
		target := minimal()
		target.Spec.SslPolicy = literalRef("https://www.googleapis.com/compute/v1/projects/p/global/sslPolicies/modern-tls")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each quic_override mode", func() {
		for _, mode := range []string{"NONE", "ENABLE", "DISABLE"} {
			target := minimal()
			target.Spec.QuicOverride = &mode
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept each tls_early_data mode", func() {
		for _, mode := range []string{"STRICT", "PERMISSIVE", "UNRESTRICTED", "DISABLED"} {
			target := minimal()
			target.Spec.TlsEarlyData = mode
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept http_keep_alive_timeout_sec within range", func() {
		target := minimal()
		target.Spec.HttpKeepAliveTimeoutSec = 610
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit RFC1035 proxy_name", func() {
		target := minimal()
		target.Spec.ProxyName = "web-https-frontend"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing url_map", func() {
		target := minimal()
		target.Spec.UrlMap = nil
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject ssl_certificates combined with certificate_manager_certificates", func() {
		target := minimal()
		target.Spec.CertificateManagerCertificates = []*foreignkeyv1.StringValueOrRef{
			literalRef("projects/p/locations/global/certificates/cm-cert"),
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "one certificate mechanism")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject ssl_certificates combined with certificate_map", func() {
		target := minimal()
		target.Spec.CertificateMap = "//certificatemanager.googleapis.com/projects/p/locations/global/certificateMaps/saas-domains"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject certificate_manager_certificates combined with certificate_map", func() {
		target := minimal()
		target.Spec.SslCertificates = nil
		target.Spec.CertificateManagerCertificates = []*foreignkeyv1.StringValueOrRef{
			literalRef("projects/p/locations/global/certificates/cm-cert"),
		}
		target.Spec.CertificateMap = "//certificatemanager.googleapis.com/projects/p/locations/global/certificateMaps/saas-domains"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject sixteen ssl_certificates", func() {
		target := minimal()
		certs := make([]*foreignkeyv1.StringValueOrRef, 0, 16)
		for i := 0; i < 16; i++ {
			certs = append(certs, certSelfLink())
		}
		target.Spec.SslCertificates = certs
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid quic_override", func() {
		target := minimal()
		invalid := "FORCE"
		target.Spec.QuicOverride = &invalid
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "NONE, ENABLE, or DISABLE")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid tls_early_data", func() {
		target := minimal()
		target.Spec.TlsEarlyData = "ALWAYS"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject http_keep_alive_timeout_sec out of range", func() {
		target := minimal()
		target.Spec.HttpKeepAliveTimeoutSec = 2000
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an uppercase proxy_name", func() {
		target := minimal()
		target.Spec.ProxyName = "Web-Frontend"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a proxy_name longer than 63 characters", func() {
		target := minimal()
		target.Spec.ProxyName = "a" + strings.Repeat("b", 63)
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		target := minimal()
		target.Kind = "GcpTargetHttpProxy"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject missing spec", func() {
		target := minimal()
		target.Spec = nil
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	// ──────────────── Regional arm ────────────────

	regional := func() *GcpTargetHttpsProxy {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.UrlMap = literalRef("https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/urlMaps/web-routing")
		target.Spec.SslCertificates = []*foreignkeyv1.StringValueOrRef{
			literalRef("https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/sslCertificates/web-cert"),
		}
		return target
	}

	ginkgo.It("should accept a regional proxy with a regional certificate", func() {
		gomega.Expect(validator.Validate(regional())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional proxy with Certificate Manager certificates and a regional SSL policy", func() {
		target := regional()
		target.Spec.SslCertificates = nil
		target.Spec.CertificateManagerCertificates = []*foreignkeyv1.StringValueOrRef{
			literalRef("projects/p/locations/us-central1/certificates/web"),
		}
		target.Spec.SslPolicy = literalRef("https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/sslPolicies/modern")
		target.Spec.ServerTlsPolicy = literalRef("projects/p/locations/us-central1/serverTlsPolicies/mtls")
		target.Spec.HttpKeepAliveTimeoutSec = 610
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed region", func() {
		target := regional()
		target.Spec.Region = "us central1"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject certificate_map on a regional proxy", func() {
		target := regional()
		target.Spec.SslCertificates = nil
		target.Spec.CertificateMap = "//certificatemanager.googleapis.com/projects/p/locations/global/certificateMaps/saas-domains"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "certificate_map is a global-proxy")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject quic_override on a regional proxy", func() {
		target := regional()
		mode := "ENABLE"
		target.Spec.QuicOverride = &mode
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "quic_override is a global-proxy")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject tls_early_data on a regional proxy", func() {
		target := regional()
		target.Spec.TlsEarlyData = "STRICT"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "tls_early_data is a global-proxy")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject proxy_bind on a regional proxy", func() {
		target := regional()
		target.Spec.ProxyBind = true
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "proxy_bind is a global-proxy")).To(gomega.BeTrue())
	})

	ginkgo.It("should still accept the global-only levers on a global proxy", func() {
		target := minimal()
		mode := "DISABLE"
		target.Spec.QuicOverride = &mode
		target.Spec.TlsEarlyData = "STRICT"
		target.Spec.ProxyBind = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})
})
