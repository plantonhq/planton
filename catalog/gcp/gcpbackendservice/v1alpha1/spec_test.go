package gcpbackendservicev1alpha1

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
	ginkgo.RunSpecs(t, "GcpBackendServiceSpec Suite")
}

func str(v string) *string { return &v }

var _ = ginkgo.Describe("GcpBackendServiceSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpBackendService: a health-checked
	// service with no backends yet — the natural creation order.
	minimal := func() *GcpBackendService {
		return &GcpBackendService{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBackendService",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-backend-service",
			},
			Spec: &GcpBackendServiceSpec{
				HealthCheck: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "https://www.googleapis.com/compute/v1/projects/p/global/healthChecks/hc",
					},
				},
			},
		}
	}

	// Helper to build a valid instance-group backend.
	instanceGroupBackend := func() *GcpBackendServiceBackend {
		return &GcpBackendServiceBackend{
			Group: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "https://www.googleapis.com/compute/v1/projects/p/zones/us-central1-a/instanceGroups/ig",
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a spec without a health check (serverless NEG contract)", func() {
		target := minimal()
		target.Spec.HealthCheck = nil
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a UTILIZATION instance-group backend with dials", func() {
		target := minimal()
		backend := instanceGroupBackend()
		backend.Description = "primary pool"
		backend.MaxUtilization = 0.7
		backend.MaxRate = 1000
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a RATE backend with a per-instance rate", func() {
		target := minimal()
		backend := instanceGroupBackend()
		backend.BalancingMode = str("RATE")
		backend.MaxRatePerInstance = 120.5
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a CONNECTION backend on a TCP service", func() {
		target := minimal()
		target.Spec.Protocol = str("TCP")
		backend := instanceGroupBackend()
		backend.BalancingMode = str("CONNECTION")
		backend.MaxConnections = 500
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a CUSTOM_METRICS backend with metrics", func() {
		target := minimal()
		backend := instanceGroupBackend()
		backend.BalancingMode = str("CUSTOM_METRICS")
		backend.CustomMetrics = []*GcpBackendServiceBackendCustomMetric{
			{Name: "orca.named_metrics.gpu_util", DryRun: true},
		}
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a full CDN configuration", func() {
		target := minimal()
		target.Spec.EnableCdn = true
		target.Spec.CompressionMode = "AUTOMATIC"
		target.Spec.CustomResponseHeaders = []string{"X-Cache-Status: {cdn_cache_status}"}
		target.Spec.CdnPolicy = &GcpBackendServiceCdnPolicy{
			CacheMode:             "CACHE_ALL_STATIC",
			ClientTtl:             1800,
			DefaultTtl:            3600,
			MaxTtl:                86400,
			NegativeCaching:       true,
			NegativeCachingPolicy: []*GcpBackendServiceNegativeCachingPolicy{{Code: 404, Ttl: 60}},
			ServeWhileStale:       600,
			RequestCoalescing:     true,
			CacheKeyPolicy: &GcpBackendServiceCdnCacheKeyPolicy{
				IncludeHost:          true,
				IncludeProtocol:      true,
				IncludeQueryString:   true,
				QueryStringWhitelist: []string{"v"},
				IncludeNamedCookies:  []string{"ab_bucket"},
			},
			BypassCacheOnRequestHeaders: []*GcpBackendServiceBypassCacheOnRequestHeader{{HeaderName: "Pragma"}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a blacklist cache key without a whitelist", func() {
		target := minimal()
		target.Spec.EnableCdn = true
		target.Spec.CdnPolicy = &GcpBackendServiceCdnPolicy{
			CacheKeyPolicy: &GcpBackendServiceCdnCacheKeyPolicy{
				IncludeQueryString:   true,
				QueryStringBlacklist: []string{"utm_source", "utm_campaign"},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept IAP with the Google-managed client", func() {
		target := minimal()
		target.Spec.Iap = &GcpBackendServiceIap{Enabled: true}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept IAP with a paired custom OAuth client", func() {
		target := minimal()
		target.Spec.Iap = &GcpBackendServiceIap{
			Enabled:            true,
			Oauth2ClientId:     "1234567890-abc.apps.googleusercontent.com",
			Oauth2ClientSecret: "GOCSPX-example-secret-value",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept Cloud Armor policy references", func() {
		target := minimal()
		target.Spec.SecurityPolicy = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "https://www.googleapis.com/compute/v1/projects/p/global/securityPolicies/waf",
			},
		}
		target.Spec.EdgeSecurityPolicy = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "https://www.googleapis.com/compute/v1/projects/p/global/securityPolicies/edge",
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept STRONG_COOKIE_AFFINITY with its cookie", func() {
		target := minimal()
		target.Spec.SessionAffinity = str("STRONG_COOKIE_AFFINITY")
		target.Spec.StrongSessionAffinityCookie = &GcpBackendServiceStrongSessionAffinityCookie{
			Name: "route",
			Path: "/app",
			Ttl:  &GcpBackendServiceDuration{Seconds: 3600},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept GENERATED_COOKIE affinity with a cookie TTL", func() {
		target := minimal()
		target.Spec.SessionAffinity = str("GENERATED_COOKIE")
		target.Spec.AffinityCookieTtlSec = 3600
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a Traffic Director mesh configuration", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = str("INTERNAL_SELF_MANAGED")
		target.Spec.LocalityLbPolicy = "MAGLEV"
		target.Spec.ConsistentHash = &GcpBackendServiceConsistentHash{
			HttpHeaderName: "x-session-id",
		}
		target.Spec.CircuitBreakers = &GcpBackendServiceCircuitBreakers{
			MaxRequestsPerConnection: 100,
		}
		target.Spec.OutlierDetection = &GcpBackendServiceOutlierDetection{
			ConsecutiveErrors:    5,
			MaxEjectionPercent:   50,
			BaseEjectionTime:     &GcpBackendServiceDuration{Seconds: 30},
			Interval:             &GcpBackendServiceDuration{Seconds: 1},
			EnforcingSuccessRate: 100,
		}
		target.Spec.MaxStreamDuration = &GcpBackendServiceDuration{Seconds: 300}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept outlier detection on EXTERNAL_MANAGED", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = str("EXTERNAL_MANAGED")
		target.Spec.OutlierDetection = &GcpBackendServiceOutlierDetection{ConsecutiveErrors: 3}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept locality_lb_policies with built-in and custom entries", func() {
		target := minimal()
		target.Spec.LocalityLbPolicies = []*GcpBackendServiceLocalityLbPolicyConfig{
			{Config: &GcpBackendServiceLocalityLbPolicyConfig_CustomPolicy{
				CustomPolicy: &GcpBackendServiceLocalityLbCustomPolicy{
					Name: "example.CustomLoadBalancer",
					Data: `{"threshold": 0.8}`,
				},
			}},
			{Config: &GcpBackendServiceLocalityLbPolicyConfig_Policy{
				Policy: &GcpBackendServiceLocalityLbPolicy{Name: "ROUND_ROBIN"},
			}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept WEIGHTED_ROUND_ROBIN with service-level custom metrics", func() {
		target := minimal()
		target.Spec.LocalityLbPolicy = "WEIGHTED_ROUND_ROBIN"
		target.Spec.CustomMetrics = []*GcpBackendServiceCustomMetric{
			{Name: "orca.cpu_utilization", DryRun: false},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept tls_settings on an HTTPS service", func() {
		target := minimal()
		target.Spec.Protocol = str("HTTPS")
		target.Spec.TlsSettings = &GcpBackendServiceTlsSettings{
			Sni: "origin.example.com",
			SubjectAltNames: []*GcpBackendServiceTlsSubjectAltName{
				{San: &GcpBackendServiceTlsSubjectAltName_DnsName{DnsName: "origin.example.com"}},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the EXTERNAL→EXTERNAL_MANAGED canary states", func() {
		target := minimal()
		target.Spec.ExternalManagedMigrationState = "TEST_BY_PERCENTAGE"
		target.Spec.ExternalManagedMigrationTestingPercentage = 25
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept AWS SigV4 origin authentication", func() {
		target := minimal()
		target.Spec.SecuritySettings = &GcpBackendServiceSecuritySettings{
			AwsV4Authentication: &GcpBackendServiceAwsV4Authentication{
				AccessKeyId:  "AKIAIOSFODNN7EXAMPLE",
				AccessKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				OriginRegion: "us-east-1",
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept up to 3 signed-URL keys", func() {
		target := minimal()
		target.Spec.SignedUrlKeys = []*GcpBackendServiceSignedUrlKey{
			{Name: "key-a", KeyValue: "hE1sPOlKZDGmSSJkVbXbBg=="},
			{Name: "key-b", KeyValue: "0123456789_-abcdefghij"},
			{Name: "key-c", KeyValue: "AAAAAAAAAAAAAAAAAAAAAA=="},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the UNSPECIFIED protocol", func() {
		target := minimal()
		target.Spec.Protocol = str("UNSPECIFIED")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the IN_FLIGHT balancing mode", func() {
		target := minimal()
		backend := instanceGroupBackend()
		backend.BalancingMode = str("IN_FLIGHT")
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
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

	ginkgo.It("should reject an invalid backend_service_name", func() {
		target := minimal()
		target.Spec.BackendServiceName = "Invalid_Name"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "RFC1035")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid protocol", func() {
		target := minimal()
		target.Spec.Protocol = str("SPDY")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid load_balancing_scheme", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = str("REGIONAL")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a backend without a group", func() {
		target := minimal()
		target.Spec.Backends = []*GcpBackendServiceBackend{{}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a RATE backend with no rate target", func() {
		target := minimal()
		backend := instanceGroupBackend()
		backend.BalancingMode = str("RATE")
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "max_rate")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a CONNECTION backend with no connection target", func() {
		target := minimal()
		backend := instanceGroupBackend()
		backend.BalancingMode = str("CONNECTION")
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a CUSTOM_METRICS backend without metrics", func() {
		target := minimal()
		backend := instanceGroupBackend()
		backend.BalancingMode = str("CUSTOM_METRICS")
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a capacity_scaler above 1.0", func() {
		target := minimal()
		backend := instanceGroupBackend()
		scaler := 1.5
		backend.CapacityScaler = &scaler
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject backend preference on the EXTERNAL scheme", func() {
		target := minimal()
		backend := instanceGroupBackend()
		backend.Preference = "PREFERRED"
		target.Spec.Backends = []*GcpBackendServiceBackend{backend}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "preference")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject CDN enabled on an internal scheme", func() {
		target := minimal()
		target.Spec.EnableCdn = true
		target.Spec.LoadBalancingScheme = str("INTERNAL_MANAGED")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "external backend services")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject TTLs with USE_ORIGIN_HEADERS", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendServiceCdnPolicy{
			CacheMode:  "USE_ORIGIN_HEADERS",
			DefaultTtl: 3600,
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "USE_ORIGIN_HEADERS")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject max_ttl with FORCE_CACHE_ALL", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendServiceCdnPolicy{
			CacheMode: "FORCE_CACHE_ALL",
			MaxTtl:    86400,
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject whitelist and blacklist together", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendServiceCdnPolicy{
			CacheKeyPolicy: &GcpBackendServiceCdnCacheKeyPolicy{
				IncludeQueryString:   true,
				QueryStringWhitelist: []string{"v"},
				QueryStringBlacklist: []string{"utm_source"},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "mutually exclusive")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject query filters with the query string excluded", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendServiceCdnPolicy{
			CacheKeyPolicy: &GcpBackendServiceCdnCacheKeyPolicy{
				QueryStringWhitelist: []string{"v"},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unsupported negative caching code", func() {
		target := minimal()
		target.Spec.CdnPolicy = &GcpBackendServiceCdnPolicy{
			NegativeCaching:       true,
			NegativeCachingPolicy: []*GcpBackendServiceNegativeCachingPolicy{{Code: 500, Ttl: 60}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject IAP with only a client id", func() {
		target := minimal()
		target.Spec.Iap = &GcpBackendServiceIap{
			Enabled:        true,
			Oauth2ClientId: "1234567890-abc.apps.googleusercontent.com",
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "set together")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a strong-affinity cookie without the affinity mode", func() {
		target := minimal()
		target.Spec.StrongSessionAffinityCookie = &GcpBackendServiceStrongSessionAffinityCookie{Name: "route"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "STRONG_COOKIE_AFFINITY")).To(gomega.BeTrue())
	})

	ginkgo.It("should accept STRONG_COOKIE_AFFINITY without a cookie block (GCP's default cookie)", func() {
		target := minimal()
		target.Spec.SessionAffinity = str("STRONG_COOKIE_AFFINITY")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject an affinity cookie TTL without GENERATED_COOKIE", func() {
		target := minimal()
		target.Spec.AffinityCookieTtlSec = 3600
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "GENERATED_COOKIE")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject session affinity on a UDP service", func() {
		target := minimal()
		target.Spec.Protocol = str("UDP")
		target.Spec.SessionAffinity = str("CLIENT_IP")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "UDP")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject circuit_breakers outside INTERNAL_SELF_MANAGED", func() {
		target := minimal()
		target.Spec.CircuitBreakers = &GcpBackendServiceCircuitBreakers{MaxRequestsPerConnection: 10}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "INTERNAL_SELF_MANAGED")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject max_stream_duration outside INTERNAL_SELF_MANAGED", func() {
		target := minimal()
		target.Spec.MaxStreamDuration = &GcpBackendServiceDuration{Seconds: 60}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject outlier_detection on the EXTERNAL scheme", func() {
		target := minimal()
		target.Spec.OutlierDetection = &GcpBackendServiceOutlierDetection{ConsecutiveErrors: 3}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject consistent_hash without a hash-based locality policy", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = str("INTERNAL_SELF_MANAGED")
		target.Spec.ConsistentHash = &GcpBackendServiceConsistentHash{HttpHeaderName: "x-id"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "MAGLEV")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a WEIGHTED policy inside locality_lb_policies", func() {
		target := minimal()
		target.Spec.LocalityLbPolicies = []*GcpBackendServiceLocalityLbPolicyConfig{
			{Config: &GcpBackendServiceLocalityLbPolicyConfig_Policy{
				Policy: &GcpBackendServiceLocalityLbPolicy{Name: "WEIGHTED_MAGLEV"},
			}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty locality_lb_policies entry", func() {
		target := minimal()
		target.Spec.LocalityLbPolicies = []*GcpBackendServiceLocalityLbPolicyConfig{{}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject service-level custom_metrics without WEIGHTED_ROUND_ROBIN", func() {
		target := minimal()
		target.Spec.CustomMetrics = []*GcpBackendServiceCustomMetric{{Name: "orca.cpu_utilization"}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "WEIGHTED_ROUND_ROBIN")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject tls_settings on a plain HTTP service", func() {
		target := minimal()
		target.Spec.TlsSettings = &GcpBackendServiceTlsSettings{Sni: "origin.example.com"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "TLS")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a SAN entry with no dns_name or uri", func() {
		target := minimal()
		target.Spec.Protocol = str("HTTPS")
		target.Spec.TlsSettings = &GcpBackendServiceTlsSettings{
			SubjectAltNames: []*GcpBackendServiceTlsSubjectAltName{{}},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a migration testing percentage without TEST_BY_PERCENTAGE", func() {
		target := minimal()
		target.Spec.ExternalManagedMigrationTestingPercentage = 25
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "TEST_BY_PERCENTAGE")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject migration state on a non-EXTERNAL scheme", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = str("EXTERNAL_MANAGED")
		target.Spec.ExternalManagedMigrationState = "PREPARE"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid migration state", func() {
		target := minimal()
		target.Spec.ExternalManagedMigrationState = "ROLLBACK"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject log optional_fields without CUSTOM mode", func() {
		target := minimal()
		target.Spec.LogConfig = &GcpBackendServiceLogConfig{
			Enable:         true,
			OptionalFields: []string{"tls.protocol"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "CUSTOM")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject log optional_mode with logging disabled", func() {
		target := minimal()
		target.Spec.LogConfig = &GcpBackendServiceLogConfig{
			OptionalMode: "INCLUDE_ALL_OPTIONAL",
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject more than 3 signed-URL keys", func() {
		target := minimal()
		target.Spec.SignedUrlKeys = []*GcpBackendServiceSignedUrlKey{
			{Name: "key-a", KeyValue: "hE1sPOlKZDGmSSJkVbXbBg=="},
			{Name: "key-b", KeyValue: "0123456789_-abcdefghij"},
			{Name: "key-c", KeyValue: "AAAAAAAAAAAAAAAAAAAAAA=="},
			{Name: "key-d", KeyValue: "BBBBBBBBBBBBBBBBBBBBBB=="},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a custom request header without a colon", func() {
		target := minimal()
		target.Spec.CustomRequestHeaders = []string{"X-Broken-Header"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a Duration with out-of-range nanos", func() {
		target := minimal()
		target.Spec.SessionAffinity = str("STRONG_COOKIE_AFFINITY")
		target.Spec.StrongSessionAffinityCookie = &GcpBackendServiceStrongSessionAffinityCookie{
			Ttl: &GcpBackendServiceDuration{Seconds: 1, Nanos: 2000000000},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		target := minimal()
		target.Kind = "GcpBackendBucket"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept logged request and response headers with logging enabled", func() {
		target := minimal()
		target.Spec.LogConfig = &GcpBackendServiceLogConfig{
			Enable:          true,
			RequestHeaders:  []string{"X-Request-Id", "User-Agent"},
			ResponseHeaders: []string{"Content-Type"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject logged headers when logging is disabled", func() {
		target := minimal()
		target.Spec.LogConfig = &GcpBackendServiceLogConfig{RequestHeaders: []string{"X-Request-Id"}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("request_headers"))
	})

	ginkgo.It("should reject duplicate logged header names", func() {
		target := minimal()
		target.Spec.LogConfig = &GcpBackendServiceLogConfig{Enable: true, ResponseHeaders: []string{"X-Cache", "X-Cache"}}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	// ──────────────── Regional arm ────────────────

	literal := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
	}
	f64 := func(v float64) *float64 { return &v }
	boolPtr := func(v bool) *bool { return &v }

	// A regional backend service for a regional external ALB: the everyday
	// regional shape, with a regional health check.
	regionalAlb := func() *GcpBackendService {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.LoadBalancingScheme = str("EXTERNAL_MANAGED")
		target.Spec.HealthCheck = literal("https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/healthChecks/hc")
		return target
	}

	// An internal passthrough Network Load Balancer backend service with the
	// passthrough levers.
	internalNlb := func() *GcpBackendService {
		target := regionalAlb()
		target.Spec.LoadBalancingScheme = str("INTERNAL")
		target.Spec.Protocol = str("TCP")
		target.Spec.Network = literal("https://www.googleapis.com/compute/v1/projects/p/global/networks/vpc")
		return target
	}

	ginkgo.It("should accept a regional external managed ALB backend service", func() {
		gomega.Expect(validator.Validate(regionalAlb())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an internal passthrough NLB with failover, connection tracking, and zonal affinity", func() {
		target := internalNlb()
		primary := instanceGroupBackend()
		failover := instanceGroupBackend()
		failover.Failover = true
		target.Spec.Backends = []*GcpBackendServiceBackend{primary, failover}
		target.Spec.FailoverPolicy = &GcpBackendServiceFailoverPolicy{
			DropTrafficIfUnhealthy: boolPtr(true),
			FailoverRatio:          f64(0.5),
		}
		target.Spec.ConnectionTrackingPolicy = &GcpBackendServiceConnectionTrackingPolicy{
			TrackingMode:                            str("PER_SESSION"),
			ConnectionPersistenceOnUnhealthyBackends: str("NEVER_PERSIST"),
		}
		idle := int32(900)
		target.Spec.ConnectionTrackingPolicy.IdleTimeoutSec = &idle
		target.Spec.NetworkPassThroughLbTrafficPolicy = &GcpBackendServiceNetworkPassThroughLbTrafficPolicy{
			ZonalAffinity: &GcpBackendServiceZonalAffinity{
				Spillover:      str("ZONAL_AFFINITY_SPILL_CROSS_ZONE"),
				SpilloverRatio: f64(0.7),
			},
		}
		target.Spec.SessionAffinity = str("CLIENT_IP_NO_DESTINATION")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an HA-policy passthrough NLB without a health check or affinity", func() {
		target := internalNlb()
		target.Spec.HealthCheck = nil
		target.Spec.HaPolicy = &GcpBackendServiceHaPolicy{
			FastIpMove: "GARP_RA",
			Leader: &GcpBackendServiceHaPolicyLeader{
				BackendGroup:    "https://www.googleapis.com/compute/v1/projects/p/zones/us-central1-a/networkEndpointGroups/leader-neg",
				NetworkEndpoint: &GcpBackendServiceHaPolicyLeaderNetworkEndpoint{Instance: "router-a"},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed region", func() {
		target := regionalAlb()
		target.Spec.Region = "US-CENTRAL1"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject the passthrough levers on a global backend service", func() {
		for _, mutate := range []func(*GcpBackendServiceSpec){
			func(s *GcpBackendServiceSpec) {
				s.Network = literal("https://www.googleapis.com/compute/v1/projects/p/global/networks/vpc")
			},
			func(s *GcpBackendServiceSpec) {
				s.FailoverPolicy = &GcpBackendServiceFailoverPolicy{DropTrafficIfUnhealthy: boolPtr(true)}
			},
			func(s *GcpBackendServiceSpec) {
				s.ConnectionTrackingPolicy = &GcpBackendServiceConnectionTrackingPolicy{}
			},
			func(s *GcpBackendServiceSpec) {
				s.NetworkPassThroughLbTrafficPolicy = &GcpBackendServiceNetworkPassThroughLbTrafficPolicy{}
			},
			func(s *GcpBackendServiceSpec) {
				b := instanceGroupBackend()
				b.Failover = true
				s.Backends = []*GcpBackendServiceBackend{b}
			},
		} {
			target := minimal()
			mutate(target.Spec)
			err := validator.Validate(target)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("exist only on a regional backend service"))
		}
	})

	ginkgo.It("should reject the INTERNAL scheme on a global backend service", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = str("INTERNAL")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("INTERNAL scheme"))
	})

	ginkgo.It("should reject CLIENT_IP_NO_DESTINATION affinity on a global backend service", func() {
		target := minimal()
		target.Spec.SessionAffinity = str("CLIENT_IP_NO_DESTINATION")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("CLIENT_IP_NO_DESTINATION exists only"))
	})

	ginkgo.It("should reject the global edge levers on a regional backend service", func() {
		for _, mutate := range []func(*GcpBackendServiceSpec){
			func(s *GcpBackendServiceSpec) { s.CompressionMode = "AUTOMATIC" },
			func(s *GcpBackendServiceSpec) { s.CustomRequestHeaders = []string{"X-Client-Geo: {client_region}"} },
			func(s *GcpBackendServiceSpec) { s.CustomResponseHeaders = []string{"X-Cache: {cdn_cache_status}"} },
			func(s *GcpBackendServiceSpec) {
				s.EdgeSecurityPolicy = literal("https://www.googleapis.com/compute/v1/projects/p/global/securityPolicies/edge")
			},
			func(s *GcpBackendServiceSpec) { s.ServiceLbPolicy = "projects/p/locations/global/serviceLbPolicies/x" },
			func(s *GcpBackendServiceSpec) {
				s.SignedUrlKeys = []*GcpBackendServiceSignedUrlKey{{Name: "k1", KeyValue: "c2lnbmluZ2tleXNlY3JldA=="}}
			},
		} {
			target := regionalAlb()
			mutate(target.Spec)
			err := validator.Validate(target)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("global edge features"))
		}
	})

	ginkgo.It("should reject the Traffic Director levers on a regional backend service", func() {
		target := regionalAlb()
		target.Spec.SecuritySettings = &GcpBackendServiceSecuritySettings{SubjectAltNames: []string{"origin.example.com"}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("Traffic Director / cross-region levers"))
	})

	ginkgo.It("should reject the migration canary and backend preference on a regional backend service", func() {
		canary := regionalAlb()
		canary.Spec.LoadBalancingScheme = str("EXTERNAL")
		canary.Spec.ExternalManagedMigrationState = "PREPARE"
		err := validator.Validate(canary)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("migration canary"))

		preferred := regionalAlb()
		b := instanceGroupBackend()
		b.Preference = "PREFERRED"
		preferred.Spec.Backends = []*GcpBackendServiceBackend{b}
		err = validator.Validate(preferred)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("backends[].preference"))
	})

	ginkgo.It("should reject the CDN knobs the regional resource lacks", func() {
		target := regionalAlb()
		target.Spec.EnableCdn = true
		target.Spec.CdnPolicy = &GcpBackendServiceCdnPolicy{
			CacheMode:         "CACHE_ALL_STATIC",
			RequestCoalescing: true,
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("Cloud CDN's advanced knobs"))

		plain := regionalAlb()
		plain.Spec.EnableCdn = true
		plain.Spec.CdnPolicy = &GcpBackendServiceCdnPolicy{CacheMode: "CACHE_ALL_STATIC", DefaultTtl: 600}
		gomega.Expect(validator.Validate(plain)).To(gomega.Succeed())
	})

	ginkgo.It("should reject Cloud CDN on the INTERNAL scheme", func() {
		target := internalNlb()
		target.Spec.EnableCdn = true
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject ha_policy together with a health check, affinity, failover, or connection tracking", func() {
		target := internalNlb()
		target.Spec.HaPolicy = &GcpBackendServiceHaPolicy{FastIpMove: "DISABLED"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("ha_policy cannot be combined"))

		noHealth := internalNlb()
		noHealth.Spec.HealthCheck = nil
		noHealth.Spec.HaPolicy = &GcpBackendServiceHaPolicy{FastIpMove: "DISABLED"}
		noHealth.Spec.ConnectionTrackingPolicy = &GcpBackendServiceConnectionTrackingPolicy{}
		gomega.Expect(validator.Validate(noHealth)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an empty failover_policy and out-of-range ratios", func() {
		empty := internalNlb()
		empty.Spec.FailoverPolicy = &GcpBackendServiceFailoverPolicy{}
		err := validator.Validate(empty)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("at least one of"))

		ratio := internalNlb()
		ratio.Spec.FailoverPolicy = &GcpBackendServiceFailoverPolicy{FailoverRatio: f64(1.5)}
		gomega.Expect(validator.Validate(ratio)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject invalid connection-tracking, fast-ip-move, and spillover values", func() {
		tracking := internalNlb()
		tracking.Spec.ConnectionTrackingPolicy = &GcpBackendServiceConnectionTrackingPolicy{TrackingMode: str("PER_FLOW")}
		gomega.Expect(validator.Validate(tracking)).ToNot(gomega.Succeed())

		idle := internalNlb()
		short := int32(30)
		idle.Spec.ConnectionTrackingPolicy = &GcpBackendServiceConnectionTrackingPolicy{IdleTimeoutSec: &short}
		gomega.Expect(validator.Validate(idle)).ToNot(gomega.Succeed())

		ha := internalNlb()
		ha.Spec.HealthCheck = nil
		ha.Spec.HaPolicy = &GcpBackendServiceHaPolicy{FastIpMove: "INSTANT"}
		gomega.Expect(validator.Validate(ha)).ToNot(gomega.Succeed())

		spill := internalNlb()
		spill.Spec.NetworkPassThroughLbTrafficPolicy = &GcpBackendServiceNetworkPassThroughLbTrafficPolicy{
			ZonalAffinity: &GcpBackendServiceZonalAffinity{Spillover: str("ZONAL_AFFINITY_ALWAYS")},
		}
		gomega.Expect(validator.Validate(spill)).ToNot(gomega.Succeed())
	})
})
