package openstackapplicationcredentialv1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOpenStackApplicationCredentialSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenStackApplicationCredentialSpec Validation Suite")
}

func boolPtr(b bool) *bool {
	return &b
}

var _ = Describe("OpenStackApplicationCredentialSpec validations", func() {

	Context("positive cases", func() {

		It("should accept a minimal valid spec (all defaults)", func() {
			spec := &OpenStackApplicationCredentialSpec{}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with description", func() {
			spec := &OpenStackApplicationCredentialSpec{
				Description: "CI/CD pipeline credential for staging",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with unrestricted set to true", func() {
			spec := &OpenStackApplicationCredentialSpec{
				Unrestricted: boolPtr(true),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with unrestricted set to false", func() {
			spec := &OpenStackApplicationCredentialSpec{
				Unrestricted: boolPtr(false),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with user-provided secret", func() {
			spec := &OpenStackApplicationCredentialSpec{
				Secret: "my-secure-secret-value",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with roles", func() {
			spec := &OpenStackApplicationCredentialSpec{
				Roles: []string{"member", "reader"},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a spec with access rules", func() {
			spec := &OpenStackApplicationCredentialSpec{
				AccessRules: []*AccessRule{
					{
						Path:    "/v2.1/servers",
						Method:  "GET",
						Service: "compute",
					},
					{
						Path:    "/v3/volumes",
						Method:  "POST",
						Service: "block-storage",
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept all valid HTTP methods in access rules", func() {
			methods := []string{"POST", "GET", "HEAD", "PATCH", "PUT", "DELETE"}
			for _, method := range methods {
				spec := &OpenStackApplicationCredentialSpec{
					AccessRules: []*AccessRule{
						{
							Path:    "/v3/test",
							Method:  method,
							Service: "identity",
						},
					},
				}
				Expect(protovalidate.Validate(spec)).To(BeNil())
			}
		})

		It("should accept a spec with expires_at", func() {
			spec := &OpenStackApplicationCredentialSpec{
				ExpiresAt: "2027-01-01T00:00:00Z",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a fully populated spec", func() {
			spec := &OpenStackApplicationCredentialSpec{
				Description:  "Production deployment credential",
				Unrestricted: boolPtr(false),
				Secret:       "super-secret",
				Roles:        []string{"member"},
				AccessRules: []*AccessRule{
					{
						Path:    "/v2.1/servers/*",
						Method:  "GET",
						Service: "compute",
					},
				},
				ExpiresAt: "2027-12-31T23:59:59Z",
				Region:    "RegionOne",
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("negative cases", func() {

		It("should reject access rule with empty path", func() {
			spec := &OpenStackApplicationCredentialSpec{
				AccessRules: []*AccessRule{
					{
						Path:    "",
						Method:  "GET",
						Service: "compute",
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject access rule with empty method", func() {
			spec := &OpenStackApplicationCredentialSpec{
				AccessRules: []*AccessRule{
					{
						Path:    "/v2.1/servers",
						Method:  "",
						Service: "compute",
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject access rule with invalid method", func() {
			spec := &OpenStackApplicationCredentialSpec{
				AccessRules: []*AccessRule{
					{
						Path:    "/v2.1/servers",
						Method:  "INVALID",
						Service: "compute",
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject access rule with lowercase method", func() {
			spec := &OpenStackApplicationCredentialSpec{
				AccessRules: []*AccessRule{
					{
						Path:    "/v2.1/servers",
						Method:  "get",
						Service: "compute",
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject access rule with empty service", func() {
			spec := &OpenStackApplicationCredentialSpec{
				AccessRules: []*AccessRule{
					{
						Path:    "/v2.1/servers",
						Method:  "GET",
						Service: "",
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})
})
