package cloudflarer2bucketv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestCloudflareR2BucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareR2BucketSpec Custom Validation Tests")
}

// newBucket builds a minimal valid bucket the individual tests mutate.
func newBucket(name string, spec *CloudflareR2BucketSpec) *CloudflareR2Bucket {
	return &CloudflareR2Bucket{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareR2Bucket",
		Metadata:   &shared.CloudResourceMetadata{Name: name},
		Spec:       spec,
	}
}

func literalRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

const testAccountID = "00000000000000000000000000000000"

var _ = ginkgo.Describe("CloudflareR2BucketSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("cloudflare_r2_bucket core", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := newBucket("test-r2-bucket", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Location:   CloudflareR2Location_weur,
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when location is omitted (auto)", func() {
				input := newBucket("test-r2-bucket-auto", &CloudflareR2BucketSpec{
					BucketName: "test-auto-bucket",
					AccountId:  testAccountID,
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with public access enabled", func() {
				input := newBucket("test-r2-bucket-public", &CloudflareR2BucketSpec{
					BucketName:   "test-public-bucket",
					AccountId:    testAccountID,
					PublicAccess: true,
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("event notifications", func() {

			ginkgo.It("should accept an event notification targeting a queue", func() {
				input := newBucket("test-r2-events", &CloudflareR2BucketSpec{
					BucketName: "test-events-bucket",
					AccountId:  testAccountID,
					EventNotifications: []*CloudflareR2BucketEventNotification{
						{
							Queue: literalRef("0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"),
							Rules: []*CloudflareR2BucketEventNotificationRule{
								{Actions: []string{"PutObject", "DeleteObject"}, Prefix: "uploads/"},
							},
						},
					},
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should reject an event notification with no rules", func() {
				input := newBucket("test-r2-events-norules", &CloudflareR2BucketSpec{
					BucketName: "test-events-bucket",
					AccountId:  testAccountID,
					EventNotifications: []*CloudflareR2BucketEventNotification{
						{Queue: literalRef("0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d")},
					},
				})
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("jurisdiction and storage_class", func() {

			ginkgo.It("should accept each valid jurisdiction value", func() {
				for _, j := range []string{"", "default", "eu", "fedramp"} {
					input := newBucket("test-jurisdiction", &CloudflareR2BucketSpec{
						BucketName:   "test-bucket",
						AccountId:    testAccountID,
						Jurisdiction: j,
					})
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "jurisdiction=%q", j)
				}
			})

			ginkgo.It("should accept each storage class", func() {
				for _, sc := range []CloudflareR2StorageClass{
					CloudflareR2StorageClass_Standard,
					CloudflareR2StorageClass_InfrequentAccess,
				} {
					input := newBucket("test-storage-class", &CloudflareR2BucketSpec{
						BucketName:   "test-bucket",
						AccountId:    testAccountID,
						StorageClass: sc,
					})
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})
		})

		ginkgo.Context("custom_domains", func() {

			ginkgo.It("should accept multiple enabled custom domains with min_tls and ciphers", func() {
				input := newBucket("test-custom-domains", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					CustomDomains: []*CloudflareR2BucketCustomDomainConfig{
						{Enabled: true, ZoneId: literalRef(testAccountID), Domain: "media.example.com", MinTls: "1.2"},
						{Enabled: true, ZoneId: literalRef(testAccountID), Domain: "cdn.example.com", Ciphers: []string{"ECDHE-RSA-AES128-GCM-SHA256"}},
					},
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a disabled custom domain with no zone_id/domain", func() {
				input := newBucket("test-custom-domain-disabled", &CloudflareR2BucketSpec{
					BucketName:    "test-bucket",
					AccountId:     testAccountID,
					CustomDomains: []*CloudflareR2BucketCustomDomainConfig{{Enabled: false}},
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("cors / lifecycle / lock", func() {

			ginkgo.It("should accept a valid CORS configuration", func() {
				input := newBucket("test-cors", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Cors: &CloudflareR2BucketCorsConfig{
						Rules: []*CloudflareR2BucketCorsRule{{
							Allowed: &CloudflareR2BucketCorsAllowed{
								Methods: []CloudflareR2CorsAllowedMethod{CloudflareR2CorsAllowedMethod_GET, CloudflareR2CorsAllowedMethod_PUT},
								Origins: []string{"https://app.example.com"},
							},
							MaxAgeSeconds: 3600,
						}},
					},
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a valid lifecycle configuration", func() {
				input := newBucket("test-lifecycle", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Lifecycle: &CloudflareR2BucketLifecycleConfig{
						Rules: []*CloudflareR2BucketLifecycleRule{{
							Id:                              "expire-logs",
							Conditions:                      &CloudflareR2BucketLifecycleConditions{Prefix: "logs/"},
							Enabled:                         true,
							AbortMultipartUploadsTransition: &CloudflareR2BucketLifecycleAbortMultipartUploadsTransition{MaxAgeSeconds: 604800},
							DeleteObjectsTransition: &CloudflareR2BucketLifecycleDeleteObjectsTransition{
								Condition: &CloudflareR2BucketLifecycleTransitionCondition{Type: CloudflareR2ConditionType_Age, MaxAgeSeconds: 2592000},
							},
							StorageClassTransitions: []*CloudflareR2BucketLifecycleStorageClassTransition{{
								Condition: &CloudflareR2BucketLifecycleTransitionCondition{Type: CloudflareR2ConditionType_Date, Date: "2027-01-01T00:00:00Z"},
							}},
						}},
					},
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept lock rules for age, date, and indefinite", func() {
				input := newBucket("test-lock", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Lock: &CloudflareR2BucketLockConfig{
						Rules: []*CloudflareR2BucketLockRule{
							{Id: "age", Enabled: true, Condition: &CloudflareR2BucketLockRuleCondition{Type: CloudflareR2ConditionType_Age, MaxAgeSeconds: 86400}},
							{Id: "date", Enabled: true, Condition: &CloudflareR2BucketLockRuleCondition{Type: CloudflareR2ConditionType_Date, Date: "2030-01-01T00:00:00Z"}},
							{Id: "indef", Enabled: true, Condition: &CloudflareR2BucketLockRuleCondition{Type: CloudflareR2ConditionType_Indefinite}},
						},
					},
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("account_id validation", func() {
			ginkgo.It("should return error if account_id is missing", func() {
				input := newBucket("test-no-account", &CloudflareR2BucketSpec{BucketName: "test-bucket"})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if account_id is not 32 characters", func() {
				input := newBucket("test-short-account", &CloudflareR2BucketSpec{BucketName: "test-bucket", AccountId: "123"})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if account_id contains non-hex characters", func() {
				input := newBucket("test-invalid-hex", &CloudflareR2BucketSpec{BucketName: "test-bucket", AccountId: "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("bucket_name validation", func() {
			ginkgo.It("should return error if bucket_name is missing", func() {
				input := newBucket("test-no-bucket-name", &CloudflareR2BucketSpec{AccountId: testAccountID})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if bucket_name is too short", func() {
				input := newBucket("test-short-bucket", &CloudflareR2BucketSpec{BucketName: "ab", AccountId: testAccountID})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if bucket_name contains invalid characters", func() {
				input := newBucket("test-invalid-bucket", &CloudflareR2BucketSpec{BucketName: "Test_Bucket", AccountId: testAccountID})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("jurisdiction and min_tls validation", func() {
			ginkgo.It("should return error for an unknown jurisdiction", func() {
				input := newBucket("test-bad-jurisdiction", &CloudflareR2BucketSpec{BucketName: "test-bucket", AccountId: testAccountID, Jurisdiction: "mars"})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error for an unknown min_tls value", func() {
				input := newBucket("test-bad-min-tls", &CloudflareR2BucketSpec{
					BucketName:    "test-bucket",
					AccountId:     testAccountID,
					CustomDomains: []*CloudflareR2BucketCustomDomainConfig{{Enabled: true, ZoneId: literalRef(testAccountID), Domain: "media.example.com", MinTls: "1.4"}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("custom_domain validation", func() {
			ginkgo.It("should return error if custom domain enabled but zone_id is missing", func() {
				input := newBucket("test-custom-domain-no-zone", &CloudflareR2BucketSpec{
					BucketName:    "test-bucket",
					AccountId:     testAccountID,
					CustomDomains: []*CloudflareR2BucketCustomDomainConfig{{Enabled: true, Domain: "media.example.com"}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if custom domain enabled but domain is missing", func() {
				input := newBucket("test-custom-domain-no-domain", &CloudflareR2BucketSpec{
					BucketName:    "test-bucket",
					AccountId:     testAccountID,
					CustomDomains: []*CloudflareR2BucketCustomDomainConfig{{Enabled: true, ZoneId: literalRef(testAccountID)}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("cors validation", func() {
			ginkgo.It("should return error if a CORS rule has no allowed methods", func() {
				input := newBucket("test-cors-no-methods", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Cors: &CloudflareR2BucketCorsConfig{Rules: []*CloudflareR2BucketCorsRule{{
						Allowed: &CloudflareR2BucketCorsAllowed{Origins: []string{"https://app.example.com"}},
					}}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if max_age_seconds exceeds 86400", func() {
				input := newBucket("test-cors-bad-maxage", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Cors: &CloudflareR2BucketCorsConfig{Rules: []*CloudflareR2BucketCorsRule{{
						Allowed:       &CloudflareR2BucketCorsAllowed{Methods: []CloudflareR2CorsAllowedMethod{CloudflareR2CorsAllowedMethod_GET}, Origins: []string{"*"}},
						MaxAgeSeconds: 100000,
					}}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("lifecycle validation", func() {
			ginkgo.It("should return error if a lifecycle rule has no id", func() {
				input := newBucket("test-lifecycle-no-id", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Lifecycle: &CloudflareR2BucketLifecycleConfig{Rules: []*CloudflareR2BucketLifecycleRule{{
						Conditions: &CloudflareR2BucketLifecycleConditions{Prefix: ""},
						Enabled:    true,
					}}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if an Age condition has no max_age_seconds", func() {
				input := newBucket("test-lifecycle-age-no-maxage", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Lifecycle: &CloudflareR2BucketLifecycleConfig{Rules: []*CloudflareR2BucketLifecycleRule{{
						Id:         "r",
						Conditions: &CloudflareR2BucketLifecycleConditions{},
						Enabled:    true,
						DeleteObjectsTransition: &CloudflareR2BucketLifecycleDeleteObjectsTransition{
							Condition: &CloudflareR2BucketLifecycleTransitionCondition{Type: CloudflareR2ConditionType_Age},
						},
					}}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error if a lifecycle condition uses Indefinite", func() {
				input := newBucket("test-lifecycle-indefinite", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Lifecycle: &CloudflareR2BucketLifecycleConfig{Rules: []*CloudflareR2BucketLifecycleRule{{
						Id:         "r",
						Conditions: &CloudflareR2BucketLifecycleConditions{},
						Enabled:    true,
						DeleteObjectsTransition: &CloudflareR2BucketLifecycleDeleteObjectsTransition{
							Condition: &CloudflareR2BucketLifecycleTransitionCondition{Type: CloudflareR2ConditionType_Indefinite},
						},
					}}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("lock validation", func() {
			ginkgo.It("should return error if a Date lock condition has no date", func() {
				input := newBucket("test-lock-date-no-date", &CloudflareR2BucketSpec{
					BucketName: "test-bucket",
					AccountId:  testAccountID,
					Lock: &CloudflareR2BucketLockConfig{Rules: []*CloudflareR2BucketLockRule{{
						Id:        "r",
						Enabled:   true,
						Condition: &CloudflareR2BucketLockRuleCondition{Type: CloudflareR2ConditionType_Date},
					}}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
