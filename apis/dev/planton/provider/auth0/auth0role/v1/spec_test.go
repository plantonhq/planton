package auth0rolev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAuth0Role(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0Role Suite")
}

var _ = ginkgo.Describe("Auth0Role Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("auth0_role with minimal configuration (name defaults to metadata.name)", func() {
			var input *Auth0Role

			ginkgo.BeforeEach(func() {
				input = &Auth0Role{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Role",
					Metadata: &shared.CloudResourceMetadata{
						Name: "viewer",
					},
					Spec: &Auth0RoleSpec{},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_role with name and description only", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Role{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Role",
					Metadata: &shared.CloudResourceMetadata{
						Name: "editor",
					},
					Spec: &Auth0RoleSpec{
						Name:        "Editor",
						Description: "Read and write access to the orders API",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_role with a single permission", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Role{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Role",
					Metadata: &shared.CloudResourceMetadata{
						Name: "viewer",
					},
					Spec: &Auth0RoleSpec{
						Name:        "Viewer",
						Description: "Read-only access",
						Permissions: []*Auth0RolePermission{
							{
								Name:                     "read:orders",
								ResourceServerIdentifier: "https://api.example.com/",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_role with permissions spanning multiple resource servers", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Role{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Role",
					Metadata: &shared.CloudResourceMetadata{
						Name: "administrator",
					},
					Spec: &Auth0RoleSpec{
						Name:        "Administrator",
						Description: "Full administrative access across orders and billing APIs",
						Permissions: []*Auth0RolePermission{
							{
								Name:                     "read:orders",
								ResourceServerIdentifier: "https://api.example.com/orders",
							},
							{
								Name:                     "write:orders",
								ResourceServerIdentifier: "https://api.example.com/orders",
							},
							{
								Name:                     "delete:orders",
								ResourceServerIdentifier: "https://api.example.com/orders",
							},
							{
								Name:                     "manage:billing",
								ResourceServerIdentifier: "https://api.example.com/billing",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Role{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Role",
					Metadata:   nil,
					Spec: &Auth0RoleSpec{
						Name: "Viewer",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required spec", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Role{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Role",
					Metadata: &shared.CloudResourceMetadata{
						Name: "viewer",
					},
					Spec: nil,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Role{
					ApiVersion: "wrong.api.version/v1",
					Kind:       "Auth0Role",
					Metadata: &shared.CloudResourceMetadata{
						Name: "viewer",
					},
					Spec: &Auth0RoleSpec{
						Name: "Viewer",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect kind", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Role{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "viewer",
					},
					Spec: &Auth0RoleSpec{
						Name: "Viewer",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("permission with missing name", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Role{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Role",
					Metadata: &shared.CloudResourceMetadata{
						Name: "viewer",
					},
					Spec: &Auth0RoleSpec{
						Name: "Viewer",
						Permissions: []*Auth0RolePermission{
							{
								Name:                     "",
								ResourceServerIdentifier: "https://api.example.com/",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("permission with missing resource_server_identifier", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Role{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Role",
					Metadata: &shared.CloudResourceMetadata{
						Name: "viewer",
					},
					Spec: &Auth0RoleSpec{
						Name: "Viewer",
						Permissions: []*Auth0RolePermission{
							{
								Name:                     "read:orders",
								ResourceServerIdentifier: "",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
