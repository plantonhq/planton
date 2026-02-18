package alicloudlogprojectv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestAlicloudLogProjectSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudLogProjectSpec Validation Tests")
}

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }

var _ = ginkgo.Describe("AlicloudLogProjectSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudLogProject{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-log-project",
				},
				Spec: &AlicloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-sls-project",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with log stores configured", func() {
			input := &AlicloudLogProject{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-log-project",
				},
				Spec: &AlicloudLogProjectSpec{
					Region:      "cn-shanghai",
					ProjectName: "prod-logging",
					Description: "Production logging project",
					LogStores: []*AlicloudLogStore{
						{
							Name:          "app-logs",
							RetentionDays: int32Ptr(90),
							ShardCount:    int32Ptr(4),
							AutoSplit:     boolPtr(true),
							EnableIndex:   boolPtr(true),
						},
						{
							Name:          "audit-logs",
							RetentionDays: int32Ptr(365),
							ShardCount:    int32Ptr(2),
							EnableIndex:   boolPtr(true),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudLogProject{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudLogProjectSpec{
					Region:          "us-west-1",
					ProjectName:     "full-project-config",
					Description:     "Fully configured project",
					ResourceGroupId: "rg-abc123",
					Tags:            map[string]string{"team": "platform", "cost-center": "eng"},
					LogStores: []*AlicloudLogStore{
						{
							Name:               "request-logs",
							RetentionDays:       int32Ptr(180),
							ShardCount:          int32Ptr(8),
							AutoSplit:           boolPtr(true),
							MaxSplitShardCount:  int32Ptr(128),
							EnableIndex:         boolPtr(true),
							AppendMeta:          boolPtr(true),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AlicloudLogProject{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudLogProjectSpec{
					ProjectName: "my-project",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when project_name is missing", func() {
			input := &AlicloudLogProject{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudLogProjectSpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when project_name is too short", func() {
			input := &AlicloudLogProject{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "ab",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when log store name is too short", func() {
			input := &AlicloudLogProject{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-project",
					LogStores: []*AlicloudLogStore{
						{Name: "ab"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when retention_days is out of range", func() {
			input := &AlicloudLogProject{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-project",
					LogStores: []*AlicloudLogStore{
						{
							Name:          "bad-retention",
							RetentionDays: int32Ptr(5000),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudLogProject{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-project",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudLogProject{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-project",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
