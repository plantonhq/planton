package gcpgcsbucketv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestGcpGcsBucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGcsBucketSpec Suite")
}

var _ = ginkgo.Describe("GcpGcsBucketSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	literal := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// Helper to build a minimal valid GcpGcsBucket.
	minimal := func() *GcpGcsBucket {
		return &GcpGcsBucket{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpGcsBucket",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-gcs-bucket",
			},
			Spec: &GcpGcsBucketSpec{
				BucketName: "test-bucket-123",
				Location:   "us-central1",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec (bucket_name + location only)", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a literal project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = literal("my-gcp-project-123")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a project_id reference", func() {
		msg := minimal()
		msg.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{
					Name:      "main-project",
					FieldPath: "status.outputs.project_id",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept multi-region and dual-region locations", func() {
		for _, location := range []string{"US", "EU", "ASIA", "NAM4"} {
			msg := minimal()
			msg.Spec.Location = location
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "location %s should be accepted", location)
		}
	})

	ginkgo.It("should accept bucket names at the length boundaries", func() {
		msg := minimal()
		msg.Spec.BucketName = "abc"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
		msg.Spec.BucketName = strings.Repeat("a", 63)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept dotted bucket names (domain-style)", func() {
		msg := minimal()
		msg.Spec.BucketName = "assets.example.com"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every modern storage class", func() {
		for _, class := range []string{"STANDARD", "NEARLINE", "COLDLINE", "ARCHIVE"} {
			msg := minimal()
			msg.Spec.StorageClass = class
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "storage class %s should be accepted", class)
		}
	})

	ginkgo.It("should accept force_destroy, UBLA, versioning, and requester_pays toggles", func() {
		msg := minimal()
		msg.Spec.ForceDestroy = true
		msg.Spec.UniformBucketLevelAccessEnabled = true
		msg.Spec.VersioningEnabled = true
		msg.Spec.RequesterPays = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept public_access_prevention enforced and inherited", func() {
		for _, v := range []string{"enforced", "inherited"} {
			msg := minimal()
			msg.Spec.PublicAccessPrevention = v
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "value %s should be accepted", v)
		}
	})

	ginkgo.It("should accept autoclass with a terminal storage class", func() {
		msg := minimal()
		msg.Spec.Autoclass = &GcpGcsBucketAutoclass{
			Enabled:              proto.Bool(true),
			TerminalStorageClass: "ARCHIVE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a delete-old-versions lifecycle rule", func() {
		msg := minimal()
		msg.Spec.VersioningEnabled = true
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{Type: "Delete"},
			Condition: &GcpGcsBucketLifecycleCondition{
				NumNewerVersions: proto.Int32(3),
				WithState:        "ARCHIVED",
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a SetStorageClass lifecycle rule with a target class", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{
				Type:         "SetStorageClass",
				StorageClass: "COLDLINE",
			},
			Condition: &GcpGcsBucketLifecycleCondition{
				AgeDays: proto.Int32(90),
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an AbortIncompleteMultipartUpload lifecycle rule", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{Type: "AbortIncompleteMultipartUpload"},
			Condition: &GcpGcsBucketLifecycleCondition{
				AgeDays: proto.Int32(7),
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an explicitly set zero age (matches all objects)", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{Type: "Delete"},
			Condition: &GcpGcsBucketLifecycleCondition{
				AgeDays:       proto.Int32(0),
				MatchesPrefix: []string{"tmp/"},
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept prefix/suffix and date-based lifecycle conditions", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{Type: "Delete"},
			Condition: &GcpGcsBucketLifecycleCondition{
				CreatedBefore:        "2026-01-01",
				MatchesPrefix:        []string{"logs/"},
				MatchesSuffix:        []string{".tmp"},
				MatchesStorageClass:  []string{"STANDARD", "NEARLINE"},
				DaysSinceCustomTime:  proto.Int32(30),
				CustomTimeBefore:     "2026-06-01",
				NoncurrentTimeBefore: "2026-03-01",
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a retention policy", func() {
		msg := minimal()
		msg.Spec.RetentionPolicy = &GcpGcsBucketRetentionPolicy{
			RetentionPeriodSeconds: 2592000,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a locked retention policy", func() {
		msg := minimal()
		msg.Spec.RetentionPolicy = &GcpGcsBucketRetentionPolicy{
			RetentionPeriodSeconds: 94608000,
			IsLocked:               true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept soft delete disabled (0) and a 30-day window", func() {
		msg := minimal()
		msg.Spec.SoftDeletePolicy = &GcpGcsBucketSoftDeletePolicy{
			RetentionDurationSeconds: proto.Int64(0),
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
		msg.Spec.SoftDeletePolicy.RetentionDurationSeconds = proto.Int64(2592000)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a CMEK key reference", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = literal("projects/p/locations/us-central1/keyRings/ring/cryptoKeys/key")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept event-based hold and object retention toggles", func() {
		msg := minimal()
		msg.Spec.DefaultEventBasedHold = true
		msg.Spec.EnableObjectRetention = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept website + CORS configuration", func() {
		msg := minimal()
		msg.Spec.Website = &GcpGcsBucketWebsite{
			MainPageSuffix: "index.html",
			NotFoundPage:   "404.html",
		}
		msg.Spec.CorsRules = []*GcpGcsBucketCorsRule{{
			Origins:         []string{"https://example.com"},
			Methods:         []string{"GET", "HEAD"},
			ResponseHeaders: []string{"Content-Type"},
			MaxAgeSeconds:   3600,
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept logging with a literal and a referenced destination bucket", func() {
		msg := minimal()
		msg.Spec.Logging = &GcpGcsBucketLogging{
			LogBucket:       literal("my-log-bucket"),
			LogObjectPrefix: "access-logs/",
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
		msg.Spec.Logging.LogBucket = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "central-logs"},
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a custom dual-region placement with rpo", func() {
		msg := minimal()
		msg.Spec.Location = "US"
		msg.Spec.CustomPlacementConfig = &GcpGcsBucketCustomPlacementConfig{
			DataLocations: []string{"US-EAST1", "US-WEST1"},
		}
		msg.Spec.Rpo = "ASYNC_TURBO"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept hierarchical namespace with UBLA", func() {
		msg := minimal()
		msg.Spec.HierarchicalNamespaceEnabled = true
		msg.Spec.UniformBucketLevelAccessEnabled = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept labels", func() {
		msg := minimal()
		msg.Spec.Labels = map[string]string{"team": "data", "env": "prod"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept additive iam_members incl. a public reader and a condition", func() {
		msg := minimal()
		msg.Spec.IamMembers = []*GcpGcsBucketIamMember{
			{Role: "roles/storage.objectAdmin", Member: literal("serviceAccount:app@my-project.iam.gserviceaccount.com")},
			{Role: "roles/storage.objectViewer", Member: literal("allUsers")},
			{
				Role:   "roles/storage.objectViewer",
				Member: literal("group:analysts@example.com"),
				Condition: &GcpGcsBucketIamCondition{
					Title:      "reports-prefix-only",
					Expression: `resource.name.startsWith("projects/_/buckets/test-bucket-123/objects/reports/")`,
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a spec without bucket_name", func() {
		msg := minimal()
		msg.Spec.BucketName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("bucket_name"))
	})

	ginkgo.It("should reject a spec without location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("location"))
	})

	ginkgo.It("should reject an uppercase bucket name", func() {
		msg := minimal()
		msg.Spec.BucketName = "INVALID-BUCKET-NAME"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject bucket names outside the length bounds", func() {
		msg := minimal()
		msg.Spec.BucketName = "ab"
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg.Spec.BucketName = strings.Repeat("a", 64)
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unknown public_access_prevention value", func() {
		msg := minimal()
		msg.Spec.PublicAccessPrevention = "disabled"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("public_access_prevention"))
	})

	ginkgo.It("should reject an unknown rpo value", func() {
		msg := minimal()
		msg.Spec.Rpo = "SYNC"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept autoclass switched off", func() {
		msg := minimal()
		msg.Spec.Autoclass = &GcpGcsBucketAutoclass{Enabled: proto.Bool(false)}
		gomega.Expect(protovalidate.Validate(msg)).To(gomega.BeNil())
	})

	ginkgo.It("should reject an autoclass block that does not say on or off", func() {
		msg := minimal()
		msg.Spec.Autoclass = &GcpGcsBucketAutoclass{TerminalStorageClass: "ARCHIVE"}
		err := protovalidate.Validate(msg)
		gomega.Expect(err).ToNot(gomega.BeNil())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("autoclass.enabled"))
	})

	ginkgo.It("should reject autoclass with an invalid terminal storage class", func() {
		msg := minimal()
		msg.Spec.Autoclass = &GcpGcsBucketAutoclass{
			Enabled:              proto.Bool(true),
			TerminalStorageClass: "COLDLINE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject enabled autoclass combined with a SetStorageClass lifecycle rule", func() {
		msg := minimal()
		msg.Spec.Autoclass = &GcpGcsBucketAutoclass{Enabled: proto.Bool(true)}
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{
				Type:         "SetStorageClass",
				StorageClass: "NEARLINE",
			},
			Condition: &GcpGcsBucketLifecycleCondition{AgeDays: proto.Int32(30)},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("autoclass"))
	})

	ginkgo.It("should accept disabled autoclass alongside a SetStorageClass rule", func() {
		msg := minimal()
		msg.Spec.Autoclass = &GcpGcsBucketAutoclass{Enabled: proto.Bool(false)}
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{
				Type:         "SetStorageClass",
				StorageClass: "NEARLINE",
			},
			Condition: &GcpGcsBucketLifecycleCondition{AgeDays: proto.Int32(30)},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a lifecycle rule without an action", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Condition: &GcpGcsBucketLifecycleCondition{AgeDays: proto.Int32(30)},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a lifecycle rule without a condition", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{Type: "Delete"},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unknown lifecycle action type", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action:    &GcpGcsBucketLifecycleAction{Type: "Archive"},
			Condition: &GcpGcsBucketLifecycleCondition{AgeDays: proto.Int32(30)},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject SetStorageClass without a target class", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action:    &GcpGcsBucketLifecycleAction{Type: "SetStorageClass"},
			Condition: &GcpGcsBucketLifecycleCondition{AgeDays: proto.Int32(30)},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("storage_class"))
	})

	ginkgo.It("should reject a Delete action carrying a storage class", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{
				Type:         "Delete",
				StorageClass: "NEARLINE",
			},
			Condition: &GcpGcsBucketLifecycleCondition{AgeDays: proto.Int32(30)},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unknown with_state value", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action:    &GcpGcsBucketLifecycleAction{Type: "Delete"},
			Condition: &GcpGcsBucketLifecycleCondition{WithState: "NONCURRENT"},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a negative lifecycle age", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action:    &GcpGcsBucketLifecycleAction{Type: "Delete"},
			Condition: &GcpGcsBucketLifecycleCondition{AgeDays: proto.Int32(-1)},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a retention policy without a period", func() {
		msg := minimal()
		msg.Spec.RetentionPolicy = &GcpGcsBucketRetentionPolicy{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a retention period beyond the 100-year cap", func() {
		msg := minimal()
		msg.Spec.RetentionPolicy = &GcpGcsBucketRetentionPolicy{
			RetentionPeriodSeconds: 3155760000,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a soft delete window shorter than 7 days (non-zero)", func() {
		msg := minimal()
		msg.Spec.SoftDeletePolicy = &GcpGcsBucketSoftDeletePolicy{
			RetentionDurationSeconds: proto.Int64(3600),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("retention_duration_seconds"))
	})

	ginkgo.It("should reject a soft delete window longer than 90 days", func() {
		msg := minimal()
		msg.Spec.SoftDeletePolicy = &GcpGcsBucketSoftDeletePolicy{
			RetentionDurationSeconds: proto.Int64(31536000),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a CORS rule without origins or methods", func() {
		msg := minimal()
		msg.Spec.CorsRules = []*GcpGcsBucketCorsRule{{
			Methods: []string{"GET"},
		}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg.Spec.CorsRules = []*GcpGcsBucketCorsRule{{
			Origins: []string{"https://example.com"},
		}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject logging without a destination bucket", func() {
		msg := minimal()
		msg.Spec.Logging = &GcpGcsBucketLogging{LogObjectPrefix: "logs/"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a custom placement with fewer or more than two regions", func() {
		msg := minimal()
		msg.Spec.CustomPlacementConfig = &GcpGcsBucketCustomPlacementConfig{
			DataLocations: []string{"US-EAST1"},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg.Spec.CustomPlacementConfig.DataLocations = []string{"US-EAST1", "US-WEST1", "US-CENTRAL1"}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an iam_member without a role", func() {
		msg := minimal()
		msg.Spec.IamMembers = []*GcpGcsBucketIamMember{
			{Member: literal("allUsers")},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an iam_member without a member", func() {
		msg := minimal()
		msg.Spec.IamMembers = []*GcpGcsBucketIamMember{
			{Role: "roles/storage.objectViewer"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an iam condition without a title or expression", func() {
		msg := minimal()
		msg.Spec.IamMembers = []*GcpGcsBucketIamMember{{
			Role:      "roles/storage.objectViewer",
			Member:    literal("allUsers"),
			Condition: &GcpGcsBucketIamCondition{Expression: "true"},
		}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg.Spec.IamMembers[0].Condition = &GcpGcsBucketIamCondition{Title: "t"}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an enabled ip_filter with a public network source", func() {
		msg := minimal()
		msg.Spec.IpFilter = &GcpGcsBucketIpFilter{
			Mode: "Enabled",
			PublicNetworkSource: &GcpGcsBucketIpFilterPublicNetworkSource{
				AllowedIpCidrRanges: []string{"203.0.113.0/24"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an enabled ip_filter with a VPC network source and service-agent exemption", func() {
		msg := minimal()
		msg.Spec.IpFilter = &GcpGcsBucketIpFilter{
			Mode: "Enabled",
			VpcNetworkSources: []*GcpGcsBucketIpFilterVpcNetworkSource{{
				Network:             literal("projects/my-project/global/networks/my-vpc"),
				AllowedIpCidrRanges: []string{"10.0.0.0/8"},
			}},
			AllowAllServiceAgentAccess: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a disabled ip_filter without sources", func() {
		msg := minimal()
		msg.Spec.IpFilter = &GcpGcsBucketIpFilter{Mode: "Disabled"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an enabled ip_filter without any source", func() {
		msg := minimal()
		msg.Spec.IpFilter = &GcpGcsBucketIpFilter{Mode: "Enabled"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("an Enabled ip_filter needs public_network_source and/or vpc_network_sources"))
	})

	ginkgo.It("should reject an ip_filter with an invalid mode", func() {
		msg := minimal()
		msg.Spec.IpFilter = &GcpGcsBucketIpFilter{Mode: "on"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an ip_filter vpc source without a network", func() {
		msg := minimal()
		msg.Spec.IpFilter = &GcpGcsBucketIpFilter{
			Mode: "Enabled",
			VpcNetworkSources: []*GcpGcsBucketIpFilterVpcNetworkSource{{
				AllowedIpCidrRanges: []string{"10.0.0.0/8"},
			}},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an ip_filter source with an empty CIDR list", func() {
		msg := minimal()
		msg.Spec.IpFilter = &GcpGcsBucketIpFilter{
			Mode: "Enabled",
			PublicNetworkSource: &GcpGcsBucketIpFilterPublicNetworkSource{
				AllowedIpCidrRanges: []string{},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		msg := minimal()
		msg.Kind = "GcpStorageBucket"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Size lifecycle conditions ────────────────

	ginkgo.It("should accept a size-banded lifecycle condition (explicit zero legal)", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action: &GcpGcsBucketLifecycleAction{Type: "SetStorageClass", StorageClass: "COLDLINE"},
			Condition: &GcpGcsBucketLifecycleCondition{
				SizeAboveBytes: proto.Int64(0),
				SizeBelowBytes: proto.Int64(1073741824),
			},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a negative size_above_bytes", func() {
		msg := minimal()
		msg.Spec.LifecycleRules = []*GcpGcsBucketLifecycleRule{{
			Action:    &GcpGcsBucketLifecycleAction{Type: "Delete"},
			Condition: &GcpGcsBucketLifecycleCondition{SizeAboveBytes: proto.Int64(-1)},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Encryption enforcement ────────────────

	ginkgo.It("should accept a CMEK-only enforcement posture", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = literal("projects/p/locations/l/keyRings/r/cryptoKeys/k")
		msg.Spec.EncryptionEnforcement = &GcpGcsBucketEncryptionEnforcement{
			GoogleManagedRestrictionMode:    "FullyRestricted",
			CustomerSuppliedRestrictionMode: "FullyRestricted",
			CustomerManagedRestrictionMode:  "NotRestricted",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid restriction mode", func() {
		msg := minimal()
		msg.Spec.EncryptionEnforcement = &GcpGcsBucketEncryptionEnforcement{
			GoogleManagedRestrictionMode: "Blocked",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Deletion policy ────────────────

	ginkgo.It("should accept each valid deletion_policy value", func() {
		for _, policy := range []string{"DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// ──────────────── Folders ────────────────

	ginkgo.It("should accept nested folders on a hierarchical-namespace bucket", func() {
		msg := minimal()
		msg.Spec.HierarchicalNamespaceEnabled = true
		msg.Spec.UniformBucketLevelAccessEnabled = true
		msg.Spec.Folders = []*GcpGcsBucketFolder{
			{Name: "logs/"},
			{Name: "logs/2026/", ForceDestroy: true},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject folders on a bucket without hierarchical namespace", func() {
		msg := minimal()
		msg.Spec.Folders = []*GcpGcsBucketFolder{{Name: "logs/"}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("hierarchical"))
	})

	ginkgo.It("should reject a folder name without the trailing slash", func() {
		msg := minimal()
		msg.Spec.HierarchicalNamespaceEnabled = true
		msg.Spec.UniformBucketLevelAccessEnabled = true
		msg.Spec.Folders = []*GcpGcsBucketFolder{{Name: "logs"}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a folder path nested beyond 5 levels or with empty segments", func() {
		msg := minimal()
		msg.Spec.HierarchicalNamespaceEnabled = true
		msg.Spec.UniformBucketLevelAccessEnabled = true
		msg.Spec.Folders = []*GcpGcsBucketFolder{{Name: "a/b/c/d/e/f/"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg.Spec.Folders = []*GcpGcsBucketFolder{{Name: "a//b/"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
		msg.Spec.Folders = []*GcpGcsBucketFolder{{Name: "/a/"}}
		gomega.Expect(validator.Validate(msg)).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject duplicate folder paths", func() {
		msg := minimal()
		msg.Spec.HierarchicalNamespaceEnabled = true
		msg.Spec.UniformBucketLevelAccessEnabled = true
		msg.Spec.Folders = []*GcpGcsBucketFolder{
			{Name: "logs/"},
			{Name: "logs/"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("folders must not repeat"))
	})

	// ──────────────── Managed folders ────────────────

	ginkgo.It("should accept managed folders on a uniform-access bucket (no HNS needed)", func() {
		msg := minimal()
		msg.Spec.UniformBucketLevelAccessEnabled = true
		msg.Spec.ManagedFolders = []*GcpGcsBucketManagedFolder{
			{Name: "reports/"},
			{Name: "exports/", ForceDestroy: true},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject managed folders without uniform bucket-level access", func() {
		msg := minimal()
		msg.Spec.ManagedFolders = []*GcpGcsBucketManagedFolder{{Name: "reports/"}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("uniform bucket-level access"))
	})

	ginkgo.It("should reject a managed-folder name without the trailing slash", func() {
		msg := minimal()
		msg.Spec.UniformBucketLevelAccessEnabled = true
		msg.Spec.ManagedFolders = []*GcpGcsBucketManagedFolder{{Name: "reports"}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject duplicate managed-folder paths", func() {
		msg := minimal()
		msg.Spec.UniformBucketLevelAccessEnabled = true
		msg.Spec.ManagedFolders = []*GcpGcsBucketManagedFolder{
			{Name: "reports/"},
			{Name: "reports/"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("managed_folders must not repeat"))
	})

	// ──────────────── Notifications ────────────────

	ginkgo.It("should accept a notification with a literal fully-qualified topic", func() {
		msg := minimal()
		msg.Spec.Notifications = []*GcpGcsBucketNotification{{
			Topic:            literal("projects/my-project/topics/uploads"),
			PayloadFormat:    "JSON_API_V1",
			EventTypes:       []string{"OBJECT_FINALIZE", "OBJECT_DELETE"},
			ObjectNamePrefix: "uploads/",
			CustomAttributes: map[string]string{"env": "prod"},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a notification with a referenced topic and empty event_types (all events)", func() {
		msg := minimal()
		msg.Spec.Notifications = []*GcpGcsBucketNotification{{
			Topic: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "uploads-topic"},
				},
			},
			PayloadFormat: "NONE",
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a notification without a topic", func() {
		msg := minimal()
		msg.Spec.Notifications = []*GcpGcsBucketNotification{{
			PayloadFormat: "JSON_API_V1",
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a notification without a payload_format", func() {
		msg := minimal()
		msg.Spec.Notifications = []*GcpGcsBucketNotification{{
			Topic: literal("projects/my-project/topics/uploads"),
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("payload_format"))
	})

	ginkgo.It("should reject an unknown payload_format", func() {
		msg := minimal()
		msg.Spec.Notifications = []*GcpGcsBucketNotification{{
			Topic:         literal("projects/my-project/topics/uploads"),
			PayloadFormat: "PROTOBUF",
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unknown event type", func() {
		msg := minimal()
		msg.Spec.Notifications = []*GcpGcsBucketNotification{{
			Topic:         literal("projects/my-project/topics/uploads"),
			PayloadFormat: "JSON_API_V1",
			EventTypes:    []string{"OBJECT_CREATED"},
		}}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("event_types"))
	})
})
