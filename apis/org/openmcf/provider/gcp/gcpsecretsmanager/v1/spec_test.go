package gcpsecretsmanagerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestGcpSecretsManagerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpSecretsManagerSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpSecretsManagerSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_secrets_manager", func() {

			ginkgo.It("should not return a validation error for minimal valid fields with literal value", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "GcpSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-secrets-manager",
					},
					Spec: &GcpSecretsManagerSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-project-123",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple secrets", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "GcpSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "app-secrets",
						Org:  "acme-corp",
						Env:  "production",
					},
					Spec: &GcpSecretsManagerSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-gcp-project",
							},
						},
						SecretNames: []string{
							"database-password",
							"api-key",
							"oauth-client-secret",
							"jwt-signing-key",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with empty secret names", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "GcpSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "empty-secrets",
					},
					Spec: &GcpSecretsManagerSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-project-456",
							},
						},
						SecretNames: []string{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with environment metadata", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "GcpSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-secrets",
						Env:  "prod",
						Org:  "engineering",
					},
					Spec: &GcpSecretsManagerSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-project-123",
							},
						},
						SecretNames: []string{
							"stripe-secret-key",
							"sendgrid-api-key",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with value_from reference", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "GcpSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ref-secrets",
						Env:  "prod",
					},
					Spec: &GcpSecretsManagerSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Kind: cloudresourcekind.CloudResourceKind_GcpProject,
									Name: "main-project",
								},
							},
						},
						SecretNames: []string{
							"api-key",
							"database-password",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("gcp_secrets_manager", func() {

			ginkgo.It("should return a validation error when project_id is missing", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "GcpSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-secrets-manager",
					},
					Spec: &GcpSecretsManagerSpec{
						ProjectId: nil,
						SecretNames: []string{
							"api-key",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("project_id"))
			})

			// HISTORY: This test originally asserted BeNil() — documenting that an empty
			// StringValueOrRef passed proto validation because `required = true` only
			// checked message presence, not content. A message-level CEL rule was added
			// to StringValueOrRef (id: "string_value_or_ref.non_empty") to fix this.
			// The assertion is now inverted to reflect the corrected behavior.
			//
			// Comprehensive boundary tests for the CEL rule live on the permanent test
			// resource at _test/testcloudresourceone/v1/spec_test.go, not here.
			ginkgo.It("should return a validation error with empty StringValueOrRef (CEL rule rejects empty content)", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "GcpSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-secrets-manager",
					},
					Spec: &GcpSecretsManagerSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is nil", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "GcpSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-secrets-manager",
					},
					Spec: nil,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("spec"))
			})

			ginkgo.It("should return a validation error when metadata is nil", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "GcpSecretsManager",
					Metadata:   nil,
					Spec: &GcpSecretsManagerSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-project-123",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("metadata"))
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &GcpSecretsManager{
					ApiVersion: "invalid.api.version/v1",
					Kind:       "GcpSecretsManager",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-secrets-manager",
					},
					Spec: &GcpSecretsManagerSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-project-123",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("api_version"))
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &GcpSecretsManager{
					ApiVersion: "gcp.openmcf.org/v1",
					Kind:       "InvalidKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-secrets-manager",
					},
					Spec: &GcpSecretsManagerSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-project-123",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("kind"))
			})
		})
	})
})
