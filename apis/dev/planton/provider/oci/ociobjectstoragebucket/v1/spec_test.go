package ociobjectstoragebucketv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciObjectStorageBucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciObjectStorageBucketSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidBucket() *OciObjectStorageBucket {
	return &OciObjectStorageBucket{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciObjectStorageBucket",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-bucket",
		},
		Spec: &OciObjectStorageBucketSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			Namespace:     "axe1234abc",
			Name:          "my-test-bucket",
		},
	}
}

var _ = ginkgo.Describe("OciObjectStorageBucketSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_objectstorage_bucket", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidBucket()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with access_type set", func() {
				input := minimalValidBucket()
				input.Spec.AccessType = OciObjectStorageBucketSpec_object_read
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with storage_tier archive", func() {
				input := minimalValidBucket()
				input.Spec.StorageTier = OciObjectStorageBucketSpec_archive
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with versioning enabled", func() {
				input := minimalValidBucket()
				input.Spec.Versioning = OciObjectStorageBucketSpec_enabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with auto_tiering infrequent_access", func() {
				input := minimalValidBucket()
				input.Spec.AutoTiering = OciObjectStorageBucketSpec_infrequent_access
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with object_events_enabled", func() {
				input := minimalValidBucket()
				input.Spec.ObjectEventsEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with kms_key_id", func() {
				input := minimalValidBucket()
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with metadata", func() {
				input := minimalValidBucket()
				input.Spec.Metadata = map[string]string{"team": "platform", "env": "prod"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidBucket()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a simple retention rule", func() {
				input := minimalValidBucket()
				input.Spec.RetentionRules = []*OciObjectStorageBucketSpec_RetentionRule{
					{DisplayName: "compliance-30d"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with retention rule with duration", func() {
				input := minimalValidBucket()
				input.Spec.RetentionRules = []*OciObjectStorageBucketSpec_RetentionRule{
					{
						DisplayName: "compliance-1y",
						Duration: &OciObjectStorageBucketSpec_Duration{
							TimeAmount: 1,
							TimeUnit:   OciObjectStorageBucketSpec_years,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with retention rule with time_rule_locked", func() {
				input := minimalValidBucket()
				input.Spec.RetentionRules = []*OciObjectStorageBucketSpec_RetentionRule{
					{
						DisplayName: "locked-rule",
						Duration: &OciObjectStorageBucketSpec_Duration{
							TimeAmount: 365,
							TimeUnit:   OciObjectStorageBucketSpec_days,
						},
						TimeRuleLocked: "2027-01-01T00:00:00Z",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a simple lifecycle rule", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "archive-old-objects",
						Action:     OciObjectStorageBucketSpec_lifecycle_archive,
						IsEnabled:  true,
						TimeAmount: 90,
						TimeUnit:   OciObjectStorageBucketSpec_days,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with lifecycle rule with object_name_filter", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "delete-logs",
						Action:     OciObjectStorageBucketSpec_lifecycle_delete,
						IsEnabled:  true,
						TimeAmount: 30,
						TimeUnit:   OciObjectStorageBucketSpec_days,
						Target:     "objects",
						ObjectNameFilter: &OciObjectStorageBucketSpec_ObjectNameFilter{
							InclusionPatterns: []string{"logs/*"},
							ExclusionPatterns: []string{"logs/audit/*"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with abort lifecycle rule targeting multipart-uploads", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "abort-incomplete-uploads",
						Action:     OciObjectStorageBucketSpec_lifecycle_abort,
						IsEnabled:  true,
						TimeAmount: 7,
						TimeUnit:   OciObjectStorageBucketSpec_days,
						Target:     "multipart-uploads",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with lifecycle rule targeting previous-object-versions", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "delete-old-versions",
						Action:     OciObjectStorageBucketSpec_lifecycle_delete,
						IsEnabled:  true,
						TimeAmount: 90,
						TimeUnit:   OciObjectStorageBucketSpec_days,
						Target:     "previous-object-versions",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a replication policy", func() {
				input := minimalValidBucket()
				input.Spec.ReplicationPolicies = []*OciObjectStorageBucketSpec_ReplicationPolicy{
					{
						Name:                  "replicate-to-ashburn",
						DestinationBucketName: "my-test-bucket-replica",
						DestinationRegionName: "us-ashburn-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidBucket()
				input.Spec.AccessType = OciObjectStorageBucketSpec_no_public_access
				input.Spec.StorageTier = OciObjectStorageBucketSpec_standard
				input.Spec.Versioning = OciObjectStorageBucketSpec_enabled
				input.Spec.AutoTiering = OciObjectStorageBucketSpec_infrequent_access
				input.Spec.ObjectEventsEnabled = true
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				input.Spec.Metadata = map[string]string{"team": "platform"}
				input.Spec.RetentionRules = []*OciObjectStorageBucketSpec_RetentionRule{
					{
						DisplayName: "compliance-1y",
						Duration: &OciObjectStorageBucketSpec_Duration{
							TimeAmount: 1,
							TimeUnit:   OciObjectStorageBucketSpec_years,
						},
					},
				}
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "archive-old",
						Action:     OciObjectStorageBucketSpec_lifecycle_archive,
						IsEnabled:  true,
						TimeAmount: 90,
						TimeUnit:   OciObjectStorageBucketSpec_days,
					},
					{
						Name:       "abort-uploads",
						Action:     OciObjectStorageBucketSpec_lifecycle_abort,
						IsEnabled:  true,
						TimeAmount: 7,
						TimeUnit:   OciObjectStorageBucketSpec_days,
						Target:     "multipart-uploads",
					},
				}
				input.Spec.ReplicationPolicies = []*OciObjectStorageBucketSpec_ReplicationPolicy{
					{
						Name:                  "dr-ashburn",
						DestinationBucketName: "my-bucket-dr",
						DestinationRegionName: "us-ashburn-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_objectstorage_bucket", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidBucket()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidBucket()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidBucket()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciObjectStorageBucket{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciObjectStorageBucket",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-bucket"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidBucket()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when namespace is empty", func() {
				input := minimalValidBucket()
				input.Spec.Namespace = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is empty", func() {
				input := minimalValidBucket()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when retention rule display_name is empty", func() {
				input := minimalValidBucket()
				input.Spec.RetentionRules = []*OciObjectStorageBucketSpec_RetentionRule{
					{DisplayName: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when retention rule duration time_amount is zero", func() {
				input := minimalValidBucket()
				input.Spec.RetentionRules = []*OciObjectStorageBucketSpec_RetentionRule{
					{
						DisplayName: "bad-duration",
						Duration: &OciObjectStorageBucketSpec_Duration{
							TimeAmount: 0,
							TimeUnit:   OciObjectStorageBucketSpec_days,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when retention rule duration time_unit is unspecified", func() {
				input := minimalValidBucket()
				input.Spec.RetentionRules = []*OciObjectStorageBucketSpec_RetentionRule{
					{
						DisplayName: "bad-unit",
						Duration: &OciObjectStorageBucketSpec_Duration{
							TimeAmount: 30,
							TimeUnit:   OciObjectStorageBucketSpec_time_unit_unspecified,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when lifecycle rule name is empty", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "",
						Action:     OciObjectStorageBucketSpec_lifecycle_delete,
						IsEnabled:  true,
						TimeAmount: 30,
						TimeUnit:   OciObjectStorageBucketSpec_days,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when lifecycle rule action is unspecified", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "bad-action",
						Action:     OciObjectStorageBucketSpec_lifecycle_action_unspecified,
						IsEnabled:  true,
						TimeAmount: 30,
						TimeUnit:   OciObjectStorageBucketSpec_days,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when lifecycle rule time_amount is zero", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "bad-time",
						Action:     OciObjectStorageBucketSpec_lifecycle_delete,
						IsEnabled:  true,
						TimeAmount: 0,
						TimeUnit:   OciObjectStorageBucketSpec_days,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when lifecycle rule time_unit is unspecified", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "bad-unit",
						Action:     OciObjectStorageBucketSpec_lifecycle_delete,
						IsEnabled:  true,
						TimeAmount: 30,
						TimeUnit:   OciObjectStorageBucketSpec_time_unit_unspecified,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when lifecycle rule has invalid target", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "bad-target",
						Action:     OciObjectStorageBucketSpec_lifecycle_delete,
						IsEnabled:  true,
						TimeAmount: 30,
						TimeUnit:   OciObjectStorageBucketSpec_days,
						Target:     "invalid-target",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when abort action does not target multipart-uploads", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "abort-wrong-target",
						Action:     OciObjectStorageBucketSpec_lifecycle_abort,
						IsEnabled:  true,
						TimeAmount: 7,
						TimeUnit:   OciObjectStorageBucketSpec_days,
						Target:     "objects",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when multipart-uploads target has object_name_filter", func() {
				input := minimalValidBucket()
				input.Spec.LifecycleRules = []*OciObjectStorageBucketSpec_LifecycleRule{
					{
						Name:       "abort-with-filter",
						Action:     OciObjectStorageBucketSpec_lifecycle_abort,
						IsEnabled:  true,
						TimeAmount: 7,
						TimeUnit:   OciObjectStorageBucketSpec_days,
						Target:     "multipart-uploads",
						ObjectNameFilter: &OciObjectStorageBucketSpec_ObjectNameFilter{
							InclusionPatterns: []string{"*.tmp"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when replication policy name is empty", func() {
				input := minimalValidBucket()
				input.Spec.ReplicationPolicies = []*OciObjectStorageBucketSpec_ReplicationPolicy{
					{
						Name:                  "",
						DestinationBucketName: "dest-bucket",
						DestinationRegionName: "us-ashburn-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when replication policy destination_bucket_name is empty", func() {
				input := minimalValidBucket()
				input.Spec.ReplicationPolicies = []*OciObjectStorageBucketSpec_ReplicationPolicy{
					{
						Name:                  "replicate",
						DestinationBucketName: "",
						DestinationRegionName: "us-ashburn-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when replication policy destination_region_name is empty", func() {
				input := minimalValidBucket()
				input.Spec.ReplicationPolicies = []*OciObjectStorageBucketSpec_ReplicationPolicy{
					{
						Name:                  "replicate",
						DestinationBucketName: "dest-bucket",
						DestinationRegionName: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
