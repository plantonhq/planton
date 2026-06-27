package cloudflarehyperdriveconfigv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func secretValue(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validHyperdrive() *CloudflareHyperdriveConfig {
	return &CloudflareHyperdriveConfig{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareHyperdriveConfig",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-hyperdrive"},
		Spec: &CloudflareHyperdriveConfigSpec{
			AccountId: validAccountID,
			Name:      "app-prod-pg",
			Origin: &CloudflareHyperdriveOrigin{
				Database: "app_production",
				Scheme:   CloudflareHyperdriveScheme_postgres,
				User:     "app_user",
				Host:     "db.example.com",
				Port:     5432,
				Password: secretValue("s3cr3t"),
			},
		},
	}
}

func TestCloudflareHyperdriveConfigSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareHyperdriveConfigSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareHyperdriveConfigSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal valid config", func() {
			gomega.Expect(protovalidate.Validate(validHyperdrive())).To(gomega.BeNil())
		})

		ginkgo.It("accepts caching, mtls, and a valid connection limit", func() {
			in := validHyperdrive()
			in.Spec.Caching = &CloudflareHyperdriveCaching{Disabled: false, MaxAge: 60, StaleWhileRevalidate: 15}
			in.Spec.Mtls = &CloudflareHyperdriveMtls{Sslmode: "verify-full"}
			in.Spec.OriginConnectionLimit = 25
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts each scheme", func() {
			for _, s := range []CloudflareHyperdriveScheme{
				CloudflareHyperdriveScheme_postgres,
				CloudflareHyperdriveScheme_postgresql,
				CloudflareHyperdriveScheme_mysql,
			} {
				in := validHyperdrive()
				in.Spec.Origin.Scheme = s
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})

		ginkgo.It("accepts access-client credentials", func() {
			in := validHyperdrive()
			in.Spec.Origin.AccessClientId = "client-id"
			in.Spec.Origin.AccessClientSecret = secretValue("client-secret")
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a VPC Service origin (service_id without mtls)", func() {
			in := validHyperdrive()
			in.Spec.Origin.ServiceId = "vpc-service-123"
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			in := validHyperdrive()
			in.Spec.AccountId = "not-a-valid-account-id-string!!!"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing name", func() {
			in := validHyperdrive()
			in.Spec.Name = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing origin", func() {
			in := validHyperdrive()
			in.Spec.Origin = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an origin missing the database", func() {
			in := validHyperdrive()
			in.Spec.Origin.Database = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unspecified scheme", func() {
			in := validHyperdrive()
			in.Spec.Origin.Scheme = CloudflareHyperdriveScheme_scheme_unspecified
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing password", func() {
			in := validHyperdrive()
			in.Spec.Origin.Password = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid sslmode", func() {
			in := validHyperdrive()
			in.Spec.Mtls = &CloudflareHyperdriveMtls{Sslmode: "insecure"}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an origin_connection_limit below the minimum of 5", func() {
			in := validHyperdrive()
			in.Spec.OriginConnectionLimit = 3
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a port above the valid range", func() {
			in := validHyperdrive()
			in.Spec.Origin.Port = 70000
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects mtls combined with a VPC Service origin", func() {
			in := validHyperdrive()
			in.Spec.Origin.ServiceId = "vpc-service-123"
			in.Spec.Mtls = &CloudflareHyperdriveMtls{Sslmode: "verify-full"}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
