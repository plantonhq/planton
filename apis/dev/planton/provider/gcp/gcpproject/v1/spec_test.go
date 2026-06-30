package gcpprojectv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestGcpProjectSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpProjectSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpProjectSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_project with minimal required fields", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gcp-project",
					},
					Spec: &GcpProjectSpec{
						ProjectId:        "my-proj", // Required field
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH", // Valid billing account format
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("gcp_project with add_suffix enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				addSuffix := true
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-gcp-project",
					},
					Spec: &GcpProjectSpec{
						ProjectId:        "test-project",
						AddSuffix:        &addSuffix,
						ParentType:       GcpProjectParentType_folder,
						ParentId:         "345678901234",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("gcp_project with all optional fields", func() {
			ginkgo.It("should not return a validation error", func() {
				disableNetwork := true
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-test-project",
					},
					Spec: &GcpProjectSpec{
						ProjectId:             "full-test-123",
						ParentType:            GcpProjectParentType_organization,
						ParentId:              "987654321098",
						BillingAccountId:      "ABCDEF-123456-ABCDEF",
						Labels:                map[string]string{"env": "dev", "team": "platform"},
						DisableDefaultNetwork: &disableNetwork,
						EnabledApis:           []string{"compute.googleapis.com", "storage.googleapis.com"},
						OwnerMember:           "admin@example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("gcp_project with valid project_id formats", func() {
			ginkgo.It("should accept 6 character project_id", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "proj01",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept 30 character project_id", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "my-very-long-project-name-12",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept project_id with hyphens", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "my-test-project-123",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept project_id starting with letter ending with digit", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "project123",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("project_id validation", func() {
			ginkgo.It("should return error when project_id is missing", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						// ProjectId is missing
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error when project_id is empty", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error when project_id is too short (< 6 chars)", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "proj1",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error when project_id is too long (> 30 chars)", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "my-very-long-project-name-that-exceeds-thirty-chars",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error when project_id starts with digit", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "123project",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error when project_id starts with hyphen", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "-myproject",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error when project_id ends with hyphen", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "myproject-",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error when project_id contains uppercase letters", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "myProject",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error when project_id contains underscores", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "my_project",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return error when project_id contains special characters", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "my-project!",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("billing_account_id validation", func() {
			ginkgo.It("should return error for invalid billing_account_id format", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "myproject",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "invalid-billing-id",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("enabled_apis validation", func() {
			ginkgo.It("should return error when API doesn't end with .googleapis.com", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "myproject",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
						EnabledApis:      []string{"compute.google.com"}, // Invalid - should be googleapis.com
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("owner_member validation", func() {
			ginkgo.It("should return error for invalid email format", func() {
				input := &GcpProject{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpProject",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
					Spec: &GcpProjectSpec{
						ProjectId:        "myproject",
						ParentType:       GcpProjectParentType_organization,
						ParentId:         "123456789012",
						BillingAccountId: "0123AB-4567CD-89EFGH",
						OwnerMember:      "not-an-email",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
