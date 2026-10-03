package gcpbackendbucketv1alpha1

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
	ginkgo.RunSpecs(t, "GcpBackendBucketSpec Suite")
}

var _ = ginkgo.Describe("GcpBackendBucketSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpBackendBucket.
	minimal := func() *GcpBackendBucket {
		return &GcpBackendBucket{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBackendBucket",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-backend-bucket",
			},
			Spec: &GcpBackendBucketSpec{
				BucketName: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-assets-bucket"},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a full CDN configuration", func() {
		target := minimal()
		target.Spec.BackendBucketName = "static-assets"
		target.Spec.Description = "serves fingerprinted static assets"
		target.Spec.EnableCdn = true
		target.Spec.CompressionMode = "AUTOMATIC"
		target.Spec.CustomResponseHeaders = []string{"X-Cache-Status: {cdn_cache_status}"}
		target.Spec.CdnPolicy = &GcpBackendBucketCdnPolicy{
			CacheMode:               "CACHE_ALL_STATIC",
			ClientTtl:               1800,
			DefaultTtl:              3600,
			MaxTtl:                  86400,
			NegativeCaching:         true,
			NegativeCachingPolicy:   []*GcpBackendBucketNegativeCachingPolicy{{Code: 404, Ttl: 60}},
			ServeWhileStale:         600,
			RequestCoalescing:       true,
			SignedUrlCacheMaxAgeSec: 300,
			CacheKeyPolicy: &GcpBackendBucketCacheKeyPolicy{
				QueryStringWhitelist: []string{"v"},
			},
			BypassCacheOnRequestHeaders: []*GcpBackendBucketBypassCacheOnRequestHeader{{HeaderName: "Pragma"}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept USE_ORIGIN_HEADERS with no TTLs", func() {
		target := minimal()
		target.Spec.EnableCdn = true
		target.Spec.CdnPolicy = &GcpBackendBucketCdnPolicy{CacheMode: "USE_ORIGIN_HEADERS"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an INTERNAL_MANAGED scheme without CDN", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = "INTERNAL_MANAGED"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept up to 3 signed-URL keys", func() {
		target := minimal()
		target.Spec.SignedUrlKeys = []*GcpBackendBucketSignedUrlKey{
			{Name: "key-a", KeyValue: "hE1sPOlKZDGmSSJkVbXbBg=="},
			{Name: "key-b", KeyValue: "0123456789_-abcdefghij"},
			{Name: "key-c", KeyValue: "AAAAAAAAAAAAAAAAAAAAAA=="},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an edge security policy reference", func() {
		target := minimal()
		target.Spec.EdgeSecurityPolicy = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "https://www.googleapis.com/compute/v1/projects/p/global/securityPolicies/edge",
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept resource manager tags", func() {
		target := minimal()
		target.Spec.ResourceManagerTags = map[string]string{"tagKeys/123456": "tagValues/789012"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a spec without bucket_name", func() {
		target := minimal()
		target.Spec.BucketName = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid backend_bucket_name", func() {
		target := minimal()
		target.Spec.BackendBucketName = "Invalid_Name"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "RFC1035")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid compression_mode", func() {
		target := minimal()
		target.Spec.CompressionMode = "GZIP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject CDN enabled on an INTERNAL_MANAGED backend bucket", func() {
		target := minimal()
		target.Spec.EnableCdn = true
		target.Spec.LoadBalancingScheme = "INTERNAL_MANAGED"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "CDN only fronts external")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid cache_mode", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendBucketCdnPolicy{CacheMode: "CACHE_EVERYTHING"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject TTLs with USE_ORIGIN_HEADERS", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendBucketCdnPolicy{
			CacheMode:  "USE_ORIGIN_HEADERS",
			DefaultTtl: 3600,
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "USE_ORIGIN_HEADERS")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject max_ttl with FORCE_CACHE_ALL", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendBucketCdnPolicy{
			CacheMode: "FORCE_CACHE_ALL",
			MaxTtl:    86400,
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "FORCE_CACHE_ALL")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an unsupported negative caching code", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendBucketCdnPolicy{
			NegativeCaching:       true,
			NegativeCachingPolicy: []*GcpBackendBucketNegativeCachingPolicy{{Code: 500, Ttl: 60}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "negative-cache")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a negative caching TTL above 1800", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendBucketCdnPolicy{
			NegativeCaching:       true,
			NegativeCachingPolicy: []*GcpBackendBucketNegativeCachingPolicy{{Code: 404, Ttl: 3600}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject more than 3 signed-URL keys", func() {
		target := minimal()
		target.Spec.SignedUrlKeys = []*GcpBackendBucketSignedUrlKey{
			{Name: "key-a", KeyValue: "hE1sPOlKZDGmSSJkVbXbBg=="},
			{Name: "key-b", KeyValue: "0123456789_-abcdefghij"},
			{Name: "key-c", KeyValue: "AAAAAAAAAAAAAAAAAAAAAA=="},
			{Name: "key-d", KeyValue: "BBBBBBBBBBBBBBBBBBBBBB=="},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a custom response header without a colon", func() {
		target := minimal()
		target.Spec.CustomResponseHeaders = []string{"X-Broken-Header"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject more than 5 bypass-cache headers", func() {
		target := minimal()
		headers := []*GcpBackendBucketBypassCacheOnRequestHeader{}
		for _, headerName := range []string{"H1", "H2", "H3", "H4", "H5", "H6"} {
			headers = append(headers, &GcpBackendBucketBypassCacheOnRequestHeader{HeaderName: headerName})
		}
		target.Spec.CdnPolicy = &GcpBackendBucketCdnPolicy{BypassCacheOnRequestHeaders: headers}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		target := minimal()
		target.Kind = "GcpBackendService"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
