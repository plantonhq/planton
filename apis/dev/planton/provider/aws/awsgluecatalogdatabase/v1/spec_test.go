package awsgluecatalogdatabasev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAwsGlueCatalogDatabaseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsGlueCatalogDatabaseSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsGlueCatalogDatabaseSpec validations", func() {
	var spec *AwsGlueCatalogDatabaseSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsGlueCatalogDatabaseSpec{
			Region: "us-west-2",
		}
	})

	// -------------------------------------------------------------------------
	// Happy path — spec-level validations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal empty spec (all defaults)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with description only", func() {
		spec.Description = "Sales analytics data lake"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with location_uri only", func() {
		spec.LocationUri = "s3://my-data-lake/databases/sales/"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with both description and location_uri", func() {
		spec.Description = "Clickstream events from web and mobile applications"
		spec.LocationUri = "s3://analytics-bucket/clickstream/"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with a long description (near max)", func() {
		spec.Description = "This is a data catalog database for a large enterprise " +
			"data lake containing raw, curated, and aggregated datasets from multiple " +
			"business units including sales, marketing, and engineering."
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a spec with S3 location URI without trailing slash", func() {
		spec.LocationUri = "s3://my-data-lake/databases/events"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready data catalog database", func() {
		spec.Description = "Production data lake — curated datasets for BI and ML pipelines"
		spec.LocationUri = "s3://prod-data-lake-us-east-1/databases/production/"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// API envelope validations (from api.proto)
	// -------------------------------------------------------------------------

	ginkgo.It("fails when apiVersion is wrong", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "wrong/v1",
			Kind:       "AwsGlueCatalogDatabase",
			Metadata:   &shared.CloudResourceMetadata{Name: "test-db"},
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "WrongKind",
			Metadata:   &shared.CloudResourceMetadata{Name: "test-db"},
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsGlueCatalogDatabase",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsGlueCatalogDatabase",
			Metadata:   &shared.CloudResourceMetadata{Name: "test-db"},
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a valid complete envelope", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsGlueCatalogDatabase",
			Metadata:   &shared.CloudResourceMetadata{Name: "analytics-db"},
			Spec: &AwsGlueCatalogDatabaseSpec{
				Region:      "us-west-2",
				Description: "Analytics data catalog",
				LocationUri: "s3://analytics-lake/databases/analytics/",
			},
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).To(gomega.BeNil())
	})
})
