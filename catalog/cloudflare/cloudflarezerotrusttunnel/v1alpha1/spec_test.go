package cloudflarezerotrusttunnelv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

// validSecret is a base64 string encoding 33 bytes (> 32-byte minimum).
var validSecret = strings.Repeat("A", 44)

func value(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func catchAll() *CloudflareZeroTrustTunnelIngressRule {
	return &CloudflareZeroTrustTunnelIngressRule{Service: "http_status:404"}
}

func validTunnel() *CloudflareZeroTrustTunnel {
	return &CloudflareZeroTrustTunnel{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustTunnel",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-tunnel"},
		Spec: &CloudflareZeroTrustTunnelSpec{
			AccountId: validAccountID,
			Name:      "prod-tunnel",
		},
	}
}

func TestCloudflareZeroTrustTunnelSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustTunnelSpec Validation Suite")
}

var _ = ginkgo.Describe("CloudflareZeroTrustTunnelSpec Validation", func() {
	ginkgo.Describe("Valid inputs", func() {
		ginkgo.It("accepts a minimal tunnel", func() {
			gomega.Expect(protovalidate.Validate(validTunnel())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a tunnel with ingress ending in a catch-all", func() {
			tn := validTunnel()
			tn.Spec.Ingress = []*CloudflareZeroTrustTunnelIngressRule{
				{Hostname: "app.example.com", Service: "http://localhost:8080", Path: "/api/.*"},
				catchAll(),
			}
			gomega.Expect(protovalidate.Validate(tn)).To(gomega.BeNil())
		})

		ginkgo.It("accepts Access-protected ingress referencing an Access application", func() {
			tn := validTunnel()
			tn.Spec.Ingress = []*CloudflareZeroTrustTunnelIngressRule{
				{
					Hostname: "secure.example.com",
					Service:  "http://localhost:9000",
					OriginRequest: &CloudflareZeroTrustTunnelOriginRequest{
						Access: &CloudflareZeroTrustTunnelAccessConfig{
							AudTag:   []*foreignkeyv1.StringValueOrRef{value("aud-tag-abc")},
							TeamName: "acme",
							Required: true,
						},
					},
				},
				catchAll(),
			}
			gomega.Expect(protovalidate.Validate(tn)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a valid base64 tunnel_secret", func() {
			tn := validTunnel()
			tn.Spec.TunnelSecret = validSecret
			gomega.Expect(protovalidate.Validate(tn)).To(gomega.BeNil())
		})

		ginkgo.It("accepts tunnel-level origin_request defaults", func() {
			tn := validTunnel()
			tn.Spec.OriginRequest = &CloudflareZeroTrustTunnelOriginRequest{
				ConnectTimeout: 30,
				ProxyType:      "socks",
				NoTlsVerify:    true,
			}
			gomega.Expect(protovalidate.Validate(tn)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Invalid inputs", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			tn := validTunnel()
			tn.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(tn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing name", func() {
			tn := validTunnel()
			tn.Spec.Name = ""
			gomega.Expect(protovalidate.Validate(tn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects ingress whose last rule has a hostname (no catch-all)", func() {
			tn := validTunnel()
			tn.Spec.Ingress = []*CloudflareZeroTrustTunnelIngressRule{
				{Hostname: "app.example.com", Service: "http://localhost:8080"},
			}
			gomega.Expect(protovalidate.Validate(tn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects ingress on a locally-managed tunnel", func() {
			tn := validTunnel()
			src := CloudflareZeroTrustTunnelConfigSource_local
			tn.Spec.ConfigSrc = &src
			tn.Spec.Ingress = []*CloudflareZeroTrustTunnelIngressRule{catchAll()}
			gomega.Expect(protovalidate.Validate(tn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an ingress rule without a service", func() {
			tn := validTunnel()
			tn.Spec.Ingress = []*CloudflareZeroTrustTunnelIngressRule{
				{Hostname: "app.example.com"},
				catchAll(),
			}
			gomega.Expect(protovalidate.Validate(tn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects Access config without a team_name", func() {
			tn := validTunnel()
			tn.Spec.Ingress = []*CloudflareZeroTrustTunnelIngressRule{
				{
					Hostname: "secure.example.com",
					Service:  "http://localhost:9000",
					OriginRequest: &CloudflareZeroTrustTunnelOriginRequest{
						Access: &CloudflareZeroTrustTunnelAccessConfig{
							AudTag: []*foreignkeyv1.StringValueOrRef{value("aud-tag-abc")},
						},
					},
				},
				catchAll(),
			}
			gomega.Expect(protovalidate.Validate(tn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects Access config with no aud_tag", func() {
			tn := validTunnel()
			tn.Spec.Ingress = []*CloudflareZeroTrustTunnelIngressRule{
				{
					Hostname: "secure.example.com",
					Service:  "http://localhost:9000",
					OriginRequest: &CloudflareZeroTrustTunnelOriginRequest{
						Access: &CloudflareZeroTrustTunnelAccessConfig{TeamName: "acme"},
					},
				},
				catchAll(),
			}
			gomega.Expect(protovalidate.Validate(tn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid proxy_type", func() {
			tn := validTunnel()
			tn.Spec.OriginRequest = &CloudflareZeroTrustTunnelOriginRequest{ProxyType: "http"}
			gomega.Expect(protovalidate.Validate(tn)).ToNot(gomega.BeNil())
		})

		// tunnel_secret carries NO value-content rule by design: the field is
		// sensitive, so consuming platforms store a managed-secret REFERENCE in
		// its place, and a base64 rule written for the raw value could never
		// match the reference (the sensitive-field value-rule class, corrected
		// 2026-08-27 per the secret-coverage gate). Cloudflare validates the
		// decoded secret at apply.
		ginkgo.It("accepts a tunnel_secret in any stored form (reference or raw)", func() {
			tn := validTunnel()
			tn.Spec.TunnelSecret = "${secrets-group/cloudflare/tunnel-secret}"
			gomega.Expect(protovalidate.Validate(tn)).To(gomega.BeNil())
		})
	})
})
