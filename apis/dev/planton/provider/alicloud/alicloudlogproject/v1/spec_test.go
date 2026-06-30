package alicloudlogprojectv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAliCloudLogProjectSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudLogProjectSpec Validation Tests")
}

func int32Ptr(i int32) *int32 { return &i }
func boolPtr(b bool) *bool    { return &b }

var _ = ginkgo.Describe("AliCloudLogProjectSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudLogProject{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-log-project",
				},
				Spec: &AliCloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-sls-project",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with log stores configured", func() {
			input := &AliCloudLogProject{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-log-project",
				},
				Spec: &AliCloudLogProjectSpec{
					Region:      "cn-shanghai",
					ProjectName: "prod-logging",
					Description: "Production logging project",
					LogStores: []*AliCloudLogStore{
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
			input := &AliCloudLogProject{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "full-config",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AliCloudLogProjectSpec{
					Region:          "us-west-1",
					ProjectName:     "full-project-config",
					Description:     "Fully configured project",
					ResourceGroupId: "rg-abc123",
					Tags:            map[string]string{"team": "platform", "cost-center": "eng"},
					LogStores: []*AliCloudLogStore{
						{
							Name:               "request-logs",
							RetentionDays:      int32Ptr(180),
							ShardCount:         int32Ptr(8),
							AutoSplit:          boolPtr(true),
							MaxSplitShardCount: int32Ptr(128),
							EnableIndex:        boolPtr(true),
							AppendMeta:         boolPtr(true),
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
			input := &AliCloudLogProject{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudLogProjectSpec{
					ProjectName: "my-project",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when project_name is missing", func() {
			input := &AliCloudLogProject{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudLogProjectSpec{
					Region: "cn-hangzhou",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when project_name is too short", func() {
			input := &AliCloudLogProject{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "ab",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when log store name is too short", func() {
			input := &AliCloudLogProject{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-project",
					LogStores: []*AliCloudLogStore{
						{Name: "ab"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when retention_days is out of range", func() {
			input := &AliCloudLogProject{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-project",
					LogStores: []*AliCloudLogStore{
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
			input := &AliCloudLogProject{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudLogProject",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-project",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudLogProject{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AliCloudLogProjectSpec{
					Region:      "cn-hangzhou",
					ProjectName: "my-project",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
