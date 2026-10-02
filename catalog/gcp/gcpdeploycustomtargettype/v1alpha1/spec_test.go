package gcpdeploycustomtargettypev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDeployCustomTargetTypeSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind cloudresourcekind.CloudResourceKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

var _ = ginkgo.Describe("GcpDeployCustomTargetTypeSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpDeployCustomTargetType {
		return &GcpDeployCustomTargetType{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpDeployCustomTargetType",
			Metadata:   &shared.CloudResourceMetadata{Name: "vendor-deployer"},
			Spec:       &GcpDeployCustomTargetTypeSpec{Location: "us-central1"},
		}
	}

	withTasks := func() *GcpDeployCustomTargetType {
		msg := minimal()
		msg.Spec.ProjectId = reference(cloudresourcekind.CloudResourceKind_GcpProject, "delivery")
		msg.Spec.CustomTargetTypeId = "vendor-deployer"
		msg.Spec.Description = "Deploys releases through the vendor's API"
		msg.Spec.Labels = map[string]string{"team": "platform"}
		msg.Spec.Annotations = map[string]string{"owner": "release-eng"}
		msg.Spec.DeletionPolicy = "PREVENT"
		msg.Spec.Tasks = &GcpDeployCustomTargetTypeTasks{
			Deploy: &GcpDeployCustomTargetTypeTask{Container: &GcpDeployCustomTargetTypeContainer{
				Image:   "us-docker.pkg.dev/acme/deploy/vendor-deployer:1.4",
				Command: []string{"/bin/deploy"},
				Args:    []string{"--verbose"},
				Env:     map[string]string{"VENDOR_REGION": "us"},
			}},
			Render: &GcpDeployCustomTargetTypeTask{Container: &GcpDeployCustomTargetTypeContainer{Image: "us-docker.pkg.dev/acme/deploy/vendor-renderer:1.4"}},
		}
		return msg
	}

	withActions := func() *GcpDeployCustomTargetType {
		msg := minimal()
		msg.Spec.CustomActions = &GcpDeployCustomTargetTypeCustomActions{
			DeployAction: "vendor-deploy",
			RenderAction: "vendor-render",
			IncludeSkaffoldModules: []*GcpDeployCustomTargetTypeSkaffoldModule{
				{Configs: []string{"vendor"}, Git: &GcpDeployCustomTargetTypeGitSource{Repo: "https://github.com/acme/deploy-actions.git", Path: "vendor/skaffold.yaml", Ref: "main"}},
				{GoogleCloudBuildRepo: &GcpDeployCustomTargetTypeCloudBuildRepoSource{Repository: reference(cloudresourcekind.CloudResourceKind_GcpCloudBuildRepository, "deploy-actions")}},
				{GoogleCloudBuildRepo: &GcpDeployCustomTargetTypeCloudBuildRepoSource{Repository: literal("projects/p/locations/us-central1/connections/github/repositories/deploy-actions"), Ref: "v2"}},
				{GoogleCloudStorage: &GcpDeployCustomTargetTypeCloudStorageSource{Source: "gs://acme-deploy/actions/*", Path: "skaffold.yaml"}},
			},
		}
		return msg
	}

	ginkgo.It("should accept a bare type, a task-based type, and an action-based type", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
		gomega.Expect(validator.Validate(withTasks())).To(gomega.Succeed())
		gomega.Expect(validator.Validate(withActions())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a task without a container, as Google does", func() {
		msg := minimal()
		msg.Spec.Tasks = &GcpDeployCustomTargetTypeTasks{Deploy: &GcpDeployCustomTargetTypeTask{}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse custom actions and tasks together", func() {
		msg := withTasks()
		msg.Spec.CustomActions = &GcpDeployCustomTargetTypeCustomActions{DeployAction: "vendor-deploy"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a deploy action, a deploy task, and a container image", func() {
		msg := minimal()
		msg.Spec.CustomActions = &GcpDeployCustomTargetTypeCustomActions{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Tasks = &GcpDeployCustomTargetTypeTasks{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = withTasks()
		msg.Spec.Tasks.Render.Container.Image = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one source per Skaffold module", func() {
		msg := withActions()
		msg.Spec.CustomActions.IncludeSkaffoldModules[0].GoogleCloudStorage = &GcpDeployCustomTargetTypeCloudStorageSource{Source: "gs://acme-deploy/actions/*"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = withActions()
		msg.Spec.CustomActions.IncludeSkaffoldModules[0] = &GcpDeployCustomTargetTypeSkaffoldModule{Configs: []string{"vendor"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require each source's location in Google's form", func() {
		msg := withActions()
		msg.Spec.CustomActions.IncludeSkaffoldModules[0].Git.Repo = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = withActions()
		msg.Spec.CustomActions.IncludeSkaffoldModules[2].GoogleCloudBuildRepo.Repository = literal("deploy-actions")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = withActions()
		msg.Spec.CustomActions.IncludeSkaffoldModules[2].GoogleCloudBuildRepo.Repository = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = withActions()
		msg.Spec.CustomActions.IncludeSkaffoldModules[3].GoogleCloudStorage.Source = "acme-deploy/actions/*"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = withActions()
		msg.Spec.CustomActions.IncludeSkaffoldModules[0].Configs = []string{""}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed type ID, a long description, and an unknown deletion policy", func() {
		for _, id := range []string{"Vendor", "1vendor", "vendor-", "vendor_deployer"} {
			msg := minimal()
			msg.Spec.CustomTargetTypeId = id
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
		msg := minimal()
		msg.Spec.Description = strings.Repeat("a", 256)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "FORCE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
