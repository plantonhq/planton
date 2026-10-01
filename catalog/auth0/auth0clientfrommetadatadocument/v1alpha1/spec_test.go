package auth0clientfrommetadatadocumentv1alpha1

import (
	"fmt"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestAuth0ClientFromMetadataDocument(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0ClientFromMetadataDocument Suite")
}

const documentURL = "https://mcp-client.example.com/.well-known/oauth-client-metadata"

func registration(spec *Auth0ClientFromMetadataDocumentSpec) *Auth0ClientFromMetadataDocument {
	return &Auth0ClientFromMetadataDocument{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0ClientFromMetadataDocument",
		Metadata:   &shared.CloudResourceMetadata{Name: "mcp-client"},
		Spec:       spec,
	}
}

// withURL is a spec carrying only the document's URL.
func withURL(url string) *Auth0ClientFromMetadataDocumentSpec {
	return &Auth0ClientFromMetadataDocumentSpec{ExternalClientId: url}
}

// rotating is a refresh-token block Auth0 accepts.
func rotating() *Auth0ClientFromMetadataDocumentRefreshToken {
	return &Auth0ClientFromMetadataDocumentRefreshToken{
		RotationType:   proto.String("rotating"),
		ExpirationType: proto.String("expiring"),
	}
}

// pairs builds n distinct client_metadata entries.
func pairs(n int) map[string]string {
	m := make(map[string]string, n)
	for i := 0; i < n; i++ {
		m[fmt.Sprintf("key%d", i)] = "value"
	}
	return m
}

func expectValid(spec *Auth0ClientFromMetadataDocumentSpec) {
	gomega.Expect(protovalidate.Validate(registration(spec))).To(gomega.BeNil())
}

func expectInvalid(spec *Auth0ClientFromMetadataDocumentSpec, substring string) {
	err := protovalidate.Validate(registration(spec))
	gomega.Expect(err).ToNot(gomega.BeNil())
	if substring != "" {
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}
}

var _ = ginkgo.Describe("Auth0ClientFromMetadataDocument Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a registration from the document's URL alone", func() {
			expectValid(withURL(documentURL))
		})

		ginkgo.It("accepts a document at a plain file path", func() {
			expectValid(withURL("https://mcp-client.example.com/client.json"))
		})

		ginkgo.It("accepts a document behind an explicit port and percent-escapes", func() {
			expectValid(withURL("https://mcp-client.example.com:8443/metadata%20v1.json"))
		})

		ginkgo.It("accepts every setting the tenant sets over the document", func() {
			spec := withURL(documentURL)
			spec.ExternalClientIdVersion = proto.Int32(2)
			spec.AppType = proto.String("native")
			spec.GrantTypes = []string{"authorization_code", "refresh_token"}
			spec.Description = proto.String("MCP client for the team's tools")
			spec.AllowedOrigins = []string{"https://mcp-client.example.com"}
			spec.WebOrigins = []string{"https://*.example.com"}
			spec.OidcConformant = proto.Bool(true)
			spec.RequireProofOfPossession = proto.Bool(true)
			spec.SkipNonVerifiableCallbackUriConfirmationPrompt = proto.Bool(false)
			spec.RedirectionPolicy = proto.String("open_redirect_protection")
			spec.OrganizationDiscoveryMethods = []string{"email", "organization_name"}
			spec.DefaultOrganization = &Auth0ClientFromMetadataDocumentDefaultOrganization{
				OrganizationId: "org_abc123",
				Flows:          []string{"client_credentials"},
			}
			spec.ClientMetadata = pairs(10)
			spec.JwtConfiguration = &Auth0ClientFromMetadataDocumentJwtConfiguration{
				Alg:               proto.String("PS256"),
				LifetimeInSeconds: proto.Int32(3600),
			}
			spec.RefreshToken = &Auth0ClientFromMetadataDocumentRefreshToken{
				RotationType:              proto.String("rotating"),
				ExpirationType:            proto.String("expiring"),
				Leeway:                    proto.Int32(0),
				TokenLifetime:             proto.Int32(2592000),
				InfiniteTokenLifetime:     proto.Bool(false),
				IdleTokenLifetime:         proto.Int32(1296000),
				InfiniteIdleTokenLifetime: proto.Bool(false),
			}
			spec.TokenQuota = &Auth0ClientFromMetadataDocumentTokenQuota{
				ClientCredentials: &Auth0ClientFromMetadataDocumentTokenQuotaClientCredentials{
					Enforce: proto.Bool(false),
					PerDay:  proto.Int32(1000),
					PerHour: proto.Int32(100),
				},
			}
			expectValid(spec)
		})

		ginkgo.It("accepts the regular_web and spa application types", func() {
			for _, appType := range []string{"regular_web", "spa"} {
				spec := withURL(documentURL)
				spec.AppType = proto.String(appType)
				expectValid(spec)
			}
		})

		ginkgo.It("accepts a non-rotating refresh token with an idle lifetime equal to its lifetime", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = &Auth0ClientFromMetadataDocumentRefreshToken{
				RotationType:      proto.String("non-rotating"),
				ExpirationType:    proto.String("expiring"),
				TokenLifetime:     proto.Int32(157788000),
				IdleTokenLifetime: proto.Int32(157788000),
			}
			expectValid(spec)
		})

		ginkgo.It("accepts a token quota that only sets its daily limit", func() {
			spec := withURL(documentURL)
			spec.TokenQuota = &Auth0ClientFromMetadataDocumentTokenQuota{
				ClientCredentials: &Auth0ClientFromMetadataDocumentTokenQuotaClientCredentials{PerDay: proto.Int32(1)},
			}
			expectValid(spec)
		})
	})

	ginkgo.Describe("When the document's URL is invalid", func() {
		ginkgo.It("rejects a missing URL", func() {
			expectInvalid(&Auth0ClientFromMetadataDocumentSpec{}, "")
		})

		ginkgo.It("rejects a plain http URL", func() {
			expectInvalid(withURL("http://mcp-client.example.com/client.json"), "Auth0 fetches client metadata only over HTTPS")
		})

		ginkgo.It("rejects a value that is not a URL", func() {
			expectInvalid(withURL("mcp-client.example.com/client.json"), "https:// URL")
		})

		ginkgo.It("rejects a URL longer than 120 characters", func() {
			expectInvalid(withURL("https://mcp-client.example.com/"+strings.Repeat("a", 100)), "")
		})

		ginkgo.It("rejects a document at the site's root", func() {
			expectInvalid(withURL("https://mcp-client.example.com/"), "needs a path after the host")
			expectInvalid(withURL("https://mcp-client.example.com"), "needs a path after the host")
		})

		ginkgo.It("rejects a query string", func() {
			expectInvalid(withURL("https://mcp-client.example.com/client.json?v=2"), "cannot carry a query")
		})

		ginkgo.It("rejects a fragment", func() {
			expectInvalid(withURL("https://mcp-client.example.com/client.json#top"), "cannot carry a query")
		})

		ginkgo.It("rejects a user and password in the URL", func() {
			expectInvalid(withURL("https://user:secret@mcp-client.example.com/client.json"), "cannot carry a query")
		})

		ginkgo.It("rejects port 0", func() {
			expectInvalid(withURL("https://mcp-client.example.com:0/client.json"), "cannot carry a query")
		})

		ginkgo.It("rejects whitespace", func() {
			expectInvalid(withURL("https://mcp-client.example.com/client.json "), "")
		})

		ginkgo.It("rejects loopback hosts", func() {
			for _, url := range []string{
				"https://localhost/client.json",
				"https://LOCALHOST:8443/client.json",
				"https://127.0.0.1/client.json",
				"https://[::1]/client.json",
			} {
				expectInvalid(withURL(url), "cannot point at localhost")
			}
		})

		ginkgo.It("rejects dot segments, plain or escaped", func() {
			for _, url := range []string{
				"https://mcp-client.example.com/./client.json",
				"https://mcp-client.example.com/meta/../client.json",
				"https://mcp-client.example.com/%2E%2e/client.json",
			} {
				expectInvalid(withURL(url), "cannot contain . or .. path segments")
			}
		})

		ginkgo.It("rejects a percent sign that starts no escape", func() {
			expectInvalid(withURL("https://mcp-client.example.com/100%.json"), "two-digit hex escape")
			expectInvalid(withURL("https://mcp-client.example.com/client%4"), "two-digit hex escape")
		})
	})

	ginkgo.Describe("When a setting over the document is invalid", func() {
		ginkgo.It("rejects an application type Auth0 does not register from a document", func() {
			spec := withURL(documentURL)
			spec.AppType = proto.String("non_interactive")
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects a grant type these applications cannot hold", func() {
			spec := withURL(documentURL)
			spec.GrantTypes = []string{"authorization_code", "client_credentials"}
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects a description longer than 140 characters", func() {
			spec := withURL(documentURL)
			spec.Description = proto.String(strings.Repeat("a", 141))
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects turning OpenID Connect conformance off", func() {
			spec := withURL(documentURL)
			spec.OidcConformant = proto.Bool(false)
			expectInvalid(spec, "oidc_conformant can only be true")
		})

		ginkgo.It("rejects an unknown redirection policy", func() {
			spec := withURL(documentURL)
			spec.RedirectionPolicy = proto.String("never")
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects an unknown organization discovery method", func() {
			spec := withURL(documentURL)
			spec.OrganizationDiscoveryMethods = []string{"domain"}
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects more than ten metadata pairs", func() {
			spec := withURL(documentURL)
			spec.ClientMetadata = pairs(11)
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects a metadata key or value longer than 255 characters", func() {
			spec := withURL(documentURL)
			spec.ClientMetadata = map[string]string{strings.Repeat("k", 256): "value"}
			expectInvalid(spec, "")
			spec.ClientMetadata = map[string]string{"owner": strings.Repeat("v", 256)}
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects a symmetric signing algorithm", func() {
			spec := withURL(documentURL)
			spec.JwtConfiguration = &Auth0ClientFromMetadataDocumentJwtConfiguration{Alg: proto.String("HS256")}
			expectInvalid(spec, "")
		})
	})

	ginkgo.Describe("When the default organization is invalid", func() {
		ginkgo.It("rejects a missing organization id", func() {
			spec := withURL(documentURL)
			spec.DefaultOrganization = &Auth0ClientFromMetadataDocumentDefaultOrganization{Flows: []string{"client_credentials"}}
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects an id that is not an organization's", func() {
			spec := withURL(documentURL)
			spec.DefaultOrganization = &Auth0ClientFromMetadataDocumentDefaultOrganization{
				OrganizationId: "acme",
				Flows:          []string{"client_credentials"},
			}
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects a default organization with no flow", func() {
			spec := withURL(documentURL)
			spec.DefaultOrganization = &Auth0ClientFromMetadataDocumentDefaultOrganization{OrganizationId: "org_abc123"}
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects a flow Auth0 does not define", func() {
			spec := withURL(documentURL)
			spec.DefaultOrganization = &Auth0ClientFromMetadataDocumentDefaultOrganization{
				OrganizationId: "org_abc123",
				Flows:          []string{"authorization_code"},
			}
			expectInvalid(spec, "")
		})
	})

	ginkgo.Describe("When the refresh-token settings are invalid", func() {
		ginkgo.It("rejects a block without its rotation type", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = &Auth0ClientFromMetadataDocumentRefreshToken{ExpirationType: proto.String("expiring")}
			expectInvalid(spec, "refresh_token needs both rotation_type")
		})

		ginkgo.It("rejects a block without its expiration type", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = &Auth0ClientFromMetadataDocumentRefreshToken{RotationType: proto.String("rotating")}
			expectInvalid(spec, "refresh_token needs both rotation_type")
		})

		ginkgo.It("rejects non-expiring refresh tokens", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = rotating()
			spec.RefreshToken.ExpirationType = proto.String("non-expiring")
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects an unknown rotation type", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = rotating()
			spec.RefreshToken.RotationType = proto.String("sliding")
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects a negative leeway", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = rotating()
			spec.RefreshToken.Leeway = proto.Int32(-1)
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects a lifetime outside 1 to 157788000 seconds", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = rotating()
			spec.RefreshToken.TokenLifetime = proto.Int32(0)
			expectInvalid(spec, "")
			spec.RefreshToken.TokenLifetime = proto.Int32(157788001)
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects an idle lifetime of zero", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = rotating()
			spec.RefreshToken.IdleTokenLifetime = proto.Int32(0)
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects an idle lifetime longer than the lifetime", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = rotating()
			spec.RefreshToken.TokenLifetime = proto.Int32(86400)
			spec.RefreshToken.IdleTokenLifetime = proto.Int32(86401)
			expectInvalid(spec, "idle_token_lifetime cannot exceed token_lifetime")
		})

		ginkgo.It("rejects an infinite idle lifetime", func() {
			spec := withURL(documentURL)
			spec.RefreshToken = rotating()
			spec.RefreshToken.InfiniteIdleTokenLifetime = proto.Bool(true)
			expectInvalid(spec, "infinite_idle_token_lifetime can only be false")
		})
	})

	ginkgo.Describe("When the token quota is invalid", func() {
		ginkgo.It("rejects a quota without its client-credentials limits", func() {
			spec := withURL(documentURL)
			spec.TokenQuota = &Auth0ClientFromMetadataDocumentTokenQuota{}
			expectInvalid(spec, "")
		})

		ginkgo.It("rejects limits below one token", func() {
			spec := withURL(documentURL)
			spec.TokenQuota = &Auth0ClientFromMetadataDocumentTokenQuota{
				ClientCredentials: &Auth0ClientFromMetadataDocumentTokenQuotaClientCredentials{PerDay: proto.Int32(0)},
			}
			expectInvalid(spec, "")
			spec.TokenQuota.ClientCredentials = &Auth0ClientFromMetadataDocumentTokenQuotaClientCredentials{PerHour: proto.Int32(0)}
			expectInvalid(spec, "")
		})
	})
})
