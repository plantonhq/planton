package snowflakedatabasev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestSnowflakeDatabaseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "SnowflakeDatabaseSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("SnowflakeDatabaseSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("snowflake_database", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &SnowflakeDatabase{
					ApiVersion: "snowflake.planton.dev/v1",
					Kind:       "SnowflakeDatabase",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-snowflake-database",
					},
					Spec: &SnowflakeDatabaseSpec{
						// No required fields, all are optional
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
