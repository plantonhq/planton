package auth0userv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAuth0User(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0User Suite")
}

// connectionRef is the reference every valid fixture uses: the user lives in
// an Auth0Connection declared beside it, which is the shape the kind is for.
func connectionRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: "users"},
		},
	}
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func validUser(spec *Auth0UserSpec) *Auth0User {
	return &Auth0User{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0User",
		Metadata:   &shared.CloudResourceMetadata{Name: "staff-root"},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("Auth0User Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("auth0_user with the minimal database shape (connection by reference, email, no password)", func() {
			ginkgo.It("should not return a validation error -- the modules mint the password", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: connectionRef(),
					Email:          "platform-root@example.com",
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_user with a declared password reference and a verified email", func() {
			ginkgo.It("should not return a validation error", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: literal("Username-Password-Authentication"),
					Email:          "ops@example.com",
					EmailVerified:  true,
					Password:       "$secret/ops-initial-password",
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_user with verify_email explicitly false", func() {
			ginkgo.It("should not return a validation error -- the stated false survives presence tracking", func() {
				verify := false
				input := validUser(&Auth0UserSpec{
					ConnectionName: connectionRef(),
					Email:          "seed@example.com",
					VerifyEmail:    &verify,
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				gomega.Expect(input.Spec.VerifyEmail).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_user on a passwordless SMS connection identified by phone number", func() {
			ginkgo.It("should not return a validation error", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: literal("sms"),
					PhoneNumber:    "+14155550123",
					PhoneVerified:  true,
					Passwordless:   true,
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_user identified by username alone", func() {
			ginkgo.It("should not return a validation error", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: connectionRef(),
					Username:       "svc_reports",
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_user with the full profile, metadata, roles, and a direct permission", func() {
			ginkgo.It("should not return a validation error", func() {
				userMetadata, err := structpb.NewStruct(map[string]interface{}{"locale": "en-US"})
				gomega.Expect(err).To(gomega.BeNil())
				appMetadata, err := structpb.NewStruct(map[string]interface{}{"plan": "enterprise"})
				gomega.Expect(err).To(gomega.BeNil())

				input := validUser(&Auth0UserSpec{
					ConnectionName:     connectionRef(),
					Email:              "admin@example.com",
					EmailVerified:      true,
					Name:               "Platform administrator",
					GivenName:          "Platform",
					FamilyName:         "Administrator",
					Nickname:           "admin",
					Picture:            "https://www.example.com/avatar.png",
					UserId:             "platform-admin",
					UserMetadata:       userMetadata,
					AppMetadata:        appMetadata,
					CustomDomainHeader: "login.example.com",
					Roles: []*foreignkeyv1.StringValueOrRef{
						{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{Name: "administrator"},
						}},
						literal("rol_abc123"),
					},
					Permissions: []*Auth0UserPermission{{
						Name:                     "read:reports",
						ResourceServerIdentifier: literal("https://api.example.com/"),
					}},
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0User{
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0User",
					Spec:       &Auth0UserSpec{ConnectionName: connectionRef(), Email: "a@example.com"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required spec", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0User{
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0User",
					Metadata:   &shared.CloudResourceMetadata{Name: "staff-root"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input := validUser(&Auth0UserSpec{ConnectionName: connectionRef(), Email: "a@example.com"})
				input.ApiVersion = "auth0.planton.dev/v1"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect kind", func() {
			ginkgo.It("should return a validation error", func() {
				input := validUser(&Auth0UserSpec{ConnectionName: connectionRef(), Email: "a@example.com"})
				input.Kind = "Auth0Role"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing connection_name", func() {
			ginkgo.It("should return a validation error", func() {
				input := validUser(&Auth0UserSpec{Email: "a@example.com"})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("no sign-in identifier at all", func() {
			ginkgo.It("should return the identifier_required message", func() {
				input := validUser(&Auth0UserSpec{ConnectionName: connectionRef()})
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("at least one sign-in identifier"))
			})
		})

		ginkgo.Context("malformed email", func() {
			ginkgo.It("should return a validation error", func() {
				input := validUser(&Auth0UserSpec{ConnectionName: connectionRef(), Email: "not-an-email"})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("phone number not in E.164 form", func() {
			ginkgo.It("should return a validation error", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: literal("sms"),
					PhoneNumber:    "415-555-0123",
					Passwordless:   true,
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("phone_verified without a phone number", func() {
			ginkgo.It("should return the phone_verified_requires_phone_number message", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: connectionRef(),
					Email:          "a@example.com",
					PhoneVerified:  true,
				})
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("phone_verified can only be set"))
			})
		})

		ginkgo.Context("a password declared on a passwordless connection", func() {
			ginkgo.It("should return the password_and_passwordless_exclusive message", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: literal("email"),
					Email:          "a@example.com",
					Passwordless:   true,
					Password:       "$secret/never-used",
				})
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("has no password to set"))
			})
		})

		ginkgo.Context("picture that is not a URI", func() {
			ginkgo.It("should return a validation error", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: connectionRef(),
					Email:          "a@example.com",
					Picture:        "not a uri",
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("permission with a missing name", func() {
			ginkgo.It("should return a validation error", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: connectionRef(),
					Email:          "a@example.com",
					Permissions: []*Auth0UserPermission{{
						ResourceServerIdentifier: literal("https://api.example.com/"),
					}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("permission with a missing resource server identifier", func() {
			ginkgo.It("should return a validation error", func() {
				input := validUser(&Auth0UserSpec{
					ConnectionName: connectionRef(),
					Email:          "a@example.com",
					Permissions:    []*Auth0UserPermission{{Name: "read:reports"}},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
