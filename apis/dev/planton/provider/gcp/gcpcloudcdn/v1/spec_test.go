package gcpcloudcdnv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestGcpCloudCdnSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudCdnSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpCloudCdnSpec Validation Tests", func() {

	ginkgo.Describe("Valid Input Tests", func() {
		ginkgo.Context("GCS Backend (basic)", func() {
			ginkgo.It("should not return validation error for GCS backend with minimal config", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cloud-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "my-test-bucket",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return validation error for GCS backend with full config", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cloud-cdn-full",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName:          "my-static-site-bucket",
									EnableUniformAccess: ptrBool(true),
								},
							},
						},
						CacheMode:             ptrCacheMode(CacheMode_CACHE_ALL_STATIC),
						DefaultTtlSeconds:     ptrInt32(3600),
						MaxTtlSeconds:         ptrInt32(86400),
						ClientTtlSeconds:      ptrInt32(1800),
						EnableNegativeCaching: ptrBool(true),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("Cloud Run Backend", func() {
			ginkgo.It("should not return validation error for Cloud Run backend", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "cloud-run-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_CloudRunService{
								CloudRunService: &CloudRunBackendConfig{
									ServiceName: "my-api-service",
									Region:      "us-central1",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("Compute Engine Backend", func() {
			ginkgo.It("should not return validation error for Compute Engine backend", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "compute-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_ComputeService{
								ComputeService: &ComputeBackendConfig{
									InstanceGroupName: "webapp-mig-us-central1",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return validation error for Compute backend with health check", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "compute-cdn-healthcheck",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_ComputeService{
								ComputeService: &ComputeBackendConfig{
									InstanceGroupName: "webapp-mig",
									HealthCheck: &HealthCheckConfig{
										Path:                 ptrString("/healthz"),
										Port:                 ptrInt32(8080),
										CheckIntervalSeconds: ptrInt32(10),
										TimeoutSeconds:       ptrInt32(5),
										HealthyThreshold:     ptrInt32(2),
										UnhealthyThreshold:   ptrInt32(3),
									},
									Protocol: ptrBackendProtocol(BackendProtocol_HTTP),
									Port:     ptrInt32(8080),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("External Origin Backend", func() {
			ginkgo.It("should not return validation error for external origin", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "external-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_ExternalOrigin{
								ExternalOrigin: &ExternalBackendConfig{
									Hostname: "assets.example.com",
									Port:     ptrInt32(443),
									Protocol: ptrBackendProtocol(BackendProtocol_HTTPS),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("Advanced Configuration", func() {
			ginkgo.It("should not return validation error with cache key policy", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "advanced-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "my-bucket",
								},
							},
						},
						AdvancedConfig: &GcpCloudCdnAdvancedConfig{
							CacheKeyPolicy: &CacheKeyPolicy{
								IncludeQueryString:   ptrBool(true),
								QueryStringWhitelist: []string{"version", "lang"},
								IncludeProtocol:      ptrBool(true),
								IncludeHost:          ptrBool(true),
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return validation error with negative caching policies", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "negative-caching-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "my-bucket",
								},
							},
						},
						EnableNegativeCaching: ptrBool(true),
						AdvancedConfig: &GcpCloudCdnAdvancedConfig{
							NegativeCachingPolicies: []*NegativeCachingPolicy{
								{Code: 404, TtlSeconds: 600},
								{Code: 500, TtlSeconds: 10},
								{Code: 503, TtlSeconds: 30},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return validation error with signed URLs", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "signed-url-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "private-content-bucket",
								},
							},
						},
						AdvancedConfig: &GcpCloudCdnAdvancedConfig{
							SignedUrlConfig: &SignedUrlConfig{
								Enabled: true,
								Keys: []*SignedUrlKey{
									{
										KeyName:  "primary-key-2024",
										KeyValue: "aGVsbG93b3JsZHRoaXNpc2Fwcm9kdWN0aW9ua2V5",
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("Frontend Configuration", func() {
			ginkgo.It("should not return validation error with Google-managed SSL", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ssl-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "my-bucket",
								},
							},
						},
						FrontendConfig: &GcpCloudCdnFrontendConfig{
							CustomDomains: []string{"www.example.com", "example.com"},
							SslCertificate: &SslCertificateConfig{
								CertificateType: &SslCertificateConfig_GoogleManaged{
									GoogleManaged: &GoogleManagedCertificateConfig{
										Domains: []string{"www.example.com", "example.com"},
									},
								},
							},
							EnableHttpsRedirect: ptrBool(true),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Invalid Input Tests", func() {
		ginkgo.Context("Missing Required Fields", func() {
			ginkgo.It("should return validation error when gcp_project_id is missing", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "my-bucket",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return validation error when backend is missing", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return validation error when GCS bucket_name is missing", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									// Missing bucket_name
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("Invalid Field Values", func() {
			// Note: Pattern validation for gcp_project_id is no longer enforced at proto level
			// since it uses StringValueOrRef. Validation is handled at the IaC layer.
			ginkgo.It("should accept any value in StringValueOrRef for gcp_project_id", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "any-valid-string"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "my-bucket",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return validation error for invalid bucket name pattern", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "Invalid_Bucket_Name", // Uppercase not allowed
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return validation error for TTL values out of range", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "my-bucket",
								},
							},
						},
						DefaultTtlSeconds: ptrInt32(999999999), // Too large (>31536000)
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return validation error for invalid negative caching status code", func() {
				input := &GcpCloudCdn{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpCloudCdn",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cdn",
					},
					Spec: &GcpCloudCdnSpec{
						GcpProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						Backend: &GcpCloudCdnBackend{
							BackendType: &GcpCloudCdnBackend_GcsBucket{
								GcsBucket: &GcsBackendConfig{
									BucketName: "my-bucket",
								},
							},
						},
						AdvancedConfig: &GcpCloudCdnAdvancedConfig{
							NegativeCachingPolicies: []*NegativeCachingPolicy{
								{Code: 200, TtlSeconds: 600}, // 200 is not an error code (must be 400-599)
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

// Helper functions to create pointers
func ptrString(s string) *string {
	return &s
}

func ptrInt32(i int32) *int32 {
	return &i
}

func ptrBool(b bool) *bool {
	return &b
}

func ptrCacheMode(cm CacheMode) *CacheMode {
	return &cm
}

func ptrBackendProtocol(bp BackendProtocol) *BackendProtocol {
	return &bp
}
