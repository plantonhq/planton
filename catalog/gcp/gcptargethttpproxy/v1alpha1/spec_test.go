package gcptargethttpproxyv1alpha1

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
	ginkgo.RunSpecs(t, "GcpTargetHttpProxySpec Suite")
}

func urlMapSelfLink() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: "https://www.googleapis.com/compute/v1/projects/p/global/urlMaps/web-routing",
		},
	}
}

var _ = ginkgo.Describe("GcpTargetHttpProxySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpTargetHttpProxy {
		return &GcpTargetHttpProxy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpTargetHttpProxy",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-http-proxy",
			},
			Spec: &GcpTargetHttpProxySpec{
				UrlMap: urlMapSelfLink(),
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec with only url_map", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit RFC1035 proxy_name", func() {
		target := minimal()
		target.Spec.ProxyName = "web-http-frontend"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a description", func() {
		target := minimal()
		target.Spec.Description = "Port-80 redirect frontend for www.example.com"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id reference", func() {
		target := minimal()
		target.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-project-123"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept http_keep_alive_timeout_sec at the lower bound", func() {
		target := minimal()
		target.Spec.HttpKeepAliveTimeoutSec = 5
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept http_keep_alive_timeout_sec at the upper bound", func() {
		target := minimal()
		target.Spec.HttpKeepAliveTimeoutSec = 1200
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept zero http_keep_alive_timeout_sec (GCP default)", func() {
		target := minimal()
		target.Spec.HttpKeepAliveTimeoutSec = 0
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept proxy_bind for Traffic Director", func() {
		target := minimal()
		target.Spec.ProxyBind = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a url_map by plain name", func() {
		target := minimal()
		target.Spec.UrlMap = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "web-routing"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing url_map", func() {
		target := minimal()
		target.Spec.UrlMap = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an uppercase proxy_name", func() {
		target := minimal()
		target.Spec.ProxyName = "Web-Frontend"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "RFC1035")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a proxy_name starting with a digit", func() {
		target := minimal()
		target.Spec.ProxyName = "1-frontend"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a proxy_name ending with a hyphen", func() {
		target := minimal()
		target.Spec.ProxyName = "frontend-"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a proxy_name longer than 63 characters", func() {
		target := minimal()
		target.Spec.ProxyName = "a" + strings.Repeat("b", 63)
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a description longer than 2048 characters", func() {
		target := minimal()
		target.Spec.Description = strings.Repeat("d", 2049)
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject http_keep_alive_timeout_sec below the minimum", func() {
		target := minimal()
		target.Spec.HttpKeepAliveTimeoutSec = 4
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "5 and 1200")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject http_keep_alive_timeout_sec above the maximum", func() {
		target := minimal()
		target.Spec.HttpKeepAliveTimeoutSec = 1201
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		target := minimal()
		target.Kind = "GcpTargetHttpsProxy"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a wrong api_version literal", func() {
		target := minimal()
		target.ApiVersion = "gcp.planton.dev/v2"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject missing metadata", func() {
		target := minimal()
		target.Metadata = nil
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

	ginkgo.It("should accept a regional proxy pointing at a regional URL map", func() {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.UrlMap = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/urlMaps/web-routing",
			},
		}
		target.Spec.HttpKeepAliveTimeoutSec = 610
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed region", func() {
		target := minimal()
		target.Spec.Region = "US_Central"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "region must be")).To(gomega.BeTrue())
	})

	ginkgo.It("should accept proxy_bind on a global proxy", func() {
		target := minimal()
		target.Spec.ProxyBind = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject proxy_bind on a regional proxy", func() {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.ProxyBind = true
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "proxy_bind is a global-proxy")).To(gomega.BeTrue())
	})
})
