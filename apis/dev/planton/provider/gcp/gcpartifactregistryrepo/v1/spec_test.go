package gcpartifactregistryrepov1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestGcpArtifactRegistryRepo(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpArtifactRegistryRepo Suite")
}

var _ = ginkgo.Describe("GcpArtifactRegistryRepo Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_artifact_registry_repo", func() {

			ginkgo.It("should not return a validation error for minimal valid fields with literal value", func() {
				input := &GcpArtifactRegistryRepo{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpArtifactRegistryRepo",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-artifact-registry",
					},
					Spec: &GcpArtifactRegistryRepoSpec{
						RepoFormat: GcpArtifactRegistryRepoFormat_DOCKER,
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-project-123",
							},
						},
						Region: "us-central1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with public access enabled", func() {
				input := &GcpArtifactRegistryRepo{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpArtifactRegistryRepo",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-repo",
					},
					Spec: &GcpArtifactRegistryRepoSpec{
						RepoFormat: GcpArtifactRegistryRepoFormat_DOCKER,
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "some-project-id",
							},
						},
						Region:             "us-west2",
						EnablePublicAccess: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with value_from reference", func() {
				input := &GcpArtifactRegistryRepo{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpArtifactRegistryRepo",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-repo-with-ref",
					},
					Spec: &GcpArtifactRegistryRepoSpec{
						RepoFormat: GcpArtifactRegistryRepoFormat_DOCKER,
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "main-project",
								},
							},
						},
						Region: "us-central1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
