package azuremssqlfailovergroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMssqlFailoverGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMssqlFailoverGroupSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func validResource() *AzureMssqlFailoverGroup {
	return &AzureMssqlFailoverGroup{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMssqlFailoverGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-fog",
		},
		Spec: &AzureMssqlFailoverGroupSpec{
			Name:     "test-fog",
			ServerId: ref("primary-server"),
			PartnerServers: []*AzureMssqlFailoverGroupPartnerServer{
				{ServerId: ref("partner-server")},
			},
			ReadWriteEndpointFailoverPolicy: &AzureMssqlFailoverGroupReadWritePolicy{
				Mode:         AzureMssqlFailoverGroupFailoverMode_AUTOMATIC,
				GraceMinutes: 60,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMssqlFailoverGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_mssql_failover_group", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept manual failover with no grace minutes", func() {
				input := validResource()
				input.Spec.ReadWriteEndpointFailoverPolicy = &AzureMssqlFailoverGroupReadWritePolicy{
					Mode: AzureMssqlFailoverGroupFailoverMode_MANUAL,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept databases and a literal server id", func() {
				input := validResource()
				input.Spec.ServerId = literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Sql/servers/primary")
				input.Spec.DatabaseIds = []*foreignkeyv1.StringValueOrRef{ref("app-db")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the read-only failover policy toggle", func() {
				input := validResource()
				enabled := false
				input.Spec.ReadonlyEndpointFailoverPolicyEnabled = &enabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_mssql_failover_group", func() {

			ginkgo.It("should return a validation error when the server id is missing", func() {
				input := validResource()
				input.Spec.ServerId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when there are no partner servers", func() {
				input := validResource()
				input.Spec.PartnerServers = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the read-write policy is missing", func() {
				input := validResource()
				input.Spec.ReadWriteEndpointFailoverPolicy = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when automatic grace_minutes is below 60", func() {
				input := validResource()
				input.Spec.ReadWriteEndpointFailoverPolicy = &AzureMssqlFailoverGroupReadWritePolicy{
					Mode:         AzureMssqlFailoverGroupFailoverMode_AUTOMATIC,
					GraceMinutes: 30,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when manual mode sets grace_minutes", func() {
				input := validResource()
				input.Spec.ReadWriteEndpointFailoverPolicy = &AzureMssqlFailoverGroupReadWritePolicy{
					Mode:         AzureMssqlFailoverGroupFailoverMode_MANUAL,
					GraceMinutes: 60,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the failover mode is unspecified", func() {
				input := validResource()
				input.Spec.ReadWriteEndpointFailoverPolicy = &AzureMssqlFailoverGroupReadWritePolicy{
					Mode: AzureMssqlFailoverGroupFailoverMode_azure_mssql_failover_group_failover_mode_unspecified,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name has uppercase letters", func() {
				input := validResource()
				input.Spec.Name = "Test-FOG"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
