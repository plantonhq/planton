package cloudflarepagesprojectv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func ref(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validProject() *CloudflarePagesProject {
	return &CloudflarePagesProject{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflarePagesProject",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-site"},
		Spec: &CloudflarePagesProjectSpec{
			AccountId:        validAccountID,
			Name:             "test-site",
			ProductionBranch: "main",
		},
	}
}

func TestCloudflarePagesProjectSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflarePagesProjectSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflarePagesProjectSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal direct-upload project", func() {
			gomega.Expect(protovalidate.Validate(validProject())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a git-connected project with build config", func() {
			in := validProject()
			in.Spec.BuildConfig = &CloudflarePagesBuildConfig{BuildCommand: "npm run build", DestinationDir: "dist"}
			in.Spec.Source = &CloudflarePagesSource{
				Type: "github",
				Config: &CloudflarePagesSourceConfig{
					Owner:                    "acme",
					RepoName:                 "site",
					ProductionBranch:         "main",
					PrCommentsEnabled:        true,
					PreviewDeploymentSetting: "custom",
					PreviewBranchIncludes:    []string{"dev"},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts deployment configs with bindings, vars and secrets", func() {
			in := validProject()
			in.Spec.DeploymentConfigs = &CloudflarePagesDeploymentConfigs{
				Production: &CloudflarePagesDeploymentConfig{
					CompatibilityDate: "2025-01-15",
					UsageModel:        "standard",
					Vars:              map[string]string{"LOG_LEVEL": "info"},
					Secrets:           []*CloudflarePagesSecret{{Name: "API_KEY", Value: ref("secret")}},
					KvNamespaces:      []*CloudflarePagesKvBinding{{Name: "CONFIG", NamespaceId: ref("0f1e2d3c4b5a69788796a5b4c3d2e1f0")}},
					D1Databases:       []*CloudflarePagesD1Binding{{Name: "DB", DatabaseId: ref("9a1b2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d")}},
					R2Buckets:         []*CloudflarePagesR2Binding{{Name: "MEDIA", BucketName: ref("media"), Jurisdiction: "eu"}},
					QueueProducers:    []*CloudflarePagesQueueProducerBinding{{Name: "JOBS", QueueName: ref("jobs-queue")}},
					Services:          []*CloudflarePagesServiceBinding{{Name: "AUTH", Service: ref("auth-worker")}},
					Placement:         &CloudflarePagesPlacement{Mode: "smart"},
					Limits:            &CloudflarePagesLimits{CpuMs: 50},
				},
			}
			in.Spec.Domains = []string{"www.example.com"}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			in := validProject()
			in.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an uppercase project name", func() {
			in := validProject()
			in.Spec.Name = "Test-Site"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing production_branch", func() {
			in := validProject()
			in.Spec.ProductionBranch = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid source type", func() {
			in := validProject()
			in.Spec.Source = &CloudflarePagesSource{Type: "bitbucket", Config: &CloudflarePagesSourceConfig{}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a source with no config", func() {
			in := validProject()
			in.Spec.Source = &CloudflarePagesSource{Type: "github"}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid preview_deployment_setting", func() {
			in := validProject()
			in.Spec.Source = &CloudflarePagesSource{Type: "github", Config: &CloudflarePagesSourceConfig{PreviewDeploymentSetting: "sometimes"}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid usage_model", func() {
			in := validProject()
			in.Spec.DeploymentConfigs = &CloudflarePagesDeploymentConfigs{
				Production: &CloudflarePagesDeploymentConfig{UsageModel: "premium"},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a secret binding with no value", func() {
			in := validProject()
			in.Spec.DeploymentConfigs = &CloudflarePagesDeploymentConfigs{
				Production: &CloudflarePagesDeploymentConfig{Secrets: []*CloudflarePagesSecret{{Name: "API_KEY"}}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a kv binding with an empty namespace_id", func() {
			in := validProject()
			in.Spec.DeploymentConfigs = &CloudflarePagesDeploymentConfigs{
				Production: &CloudflarePagesDeploymentConfig{KvNamespaces: []*CloudflarePagesKvBinding{{Name: "CONFIG", NamespaceId: ref("")}}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
