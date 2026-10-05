package awsgluecatalogdatabasev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
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
	// description bounds + catalog_id + parameters
	// -------------------------------------------------------------------------

	ginkgo.It("fails when description exceeds 2048 characters", func() {
		spec.Description = strings.Repeat("x", 2049)
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a cross-account catalog_id", func() {
		spec.CatalogId = "123456789012"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts catalog metadata parameters", func() {
		spec.Parameters = map[string]string{
			"classification": "parquet",
			"team":           "data-platform",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Lake Formation create-table default permissions
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a grant of ALL to IAM_ALLOWED_PRINCIPALS", func() {
		spec.CreateTableDefaultPermissions = []*AwsGlueCatalogDatabasePrincipalPermissions{
			{Permissions: []string{"ALL"}, Principal: "IAM_ALLOWED_PRINCIPALS"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a scoped grant to an IAM role ARN", func() {
		spec.CreateTableDefaultPermissions = []*AwsGlueCatalogDatabasePrincipalPermissions{
			{
				Permissions: []string{"SELECT", "ALTER", "DROP"},
				Principal:   "arn:aws:iam::123456789012:role/data-lake-admin",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts an empty-permissions entry (disables the default IAM grant)", func() {
		spec.CreateTableDefaultPermissions = []*AwsGlueCatalogDatabasePrincipalPermissions{
			{Permissions: []string{}},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails on an invalid Lake Formation permission value", func() {
		spec.CreateTableDefaultPermissions = []*AwsGlueCatalogDatabasePrincipalPermissions{
			{Permissions: []string{"READ_WRITE"}, Principal: "IAM_ALLOWED_PRINCIPALS"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails on a lowercase permission value", func() {
		spec.CreateTableDefaultPermissions = []*AwsGlueCatalogDatabasePrincipalPermissions{
			{Permissions: []string{"select"}},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Resource link (target_database)
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a same-region resource link", func() {
		spec.TargetDatabase = &AwsGlueCatalogDatabaseTarget{
			CatalogId:    "987654321098",
			DatabaseName: "shared_sales_data",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a cross-region resource link", func() {
		spec.TargetDatabase = &AwsGlueCatalogDatabaseTarget{
			CatalogId:    "987654321098",
			DatabaseName: "shared_sales_data",
			Region:       "us-east-1",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when a resource link omits the owning catalog_id", func() {
		spec.TargetDatabase = &AwsGlueCatalogDatabaseTarget{
			DatabaseName: "shared_sales_data",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when a resource link omits the target database_name", func() {
		spec.TargetDatabase = &AwsGlueCatalogDatabaseTarget{
			CatalogId: "987654321098",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Federated database + CEL: target_xor_federated
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a federated database over a Redshift datashare", func() {
		spec.FederatedDatabase = &AwsGlueCatalogDatabaseFederation{
			Identifier:     "arn:aws:redshift:us-west-2:123456789012:datashare:abc-123/sales_share",
			ConnectionName: "aws:redshift",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when target_database and federated_database are combined", func() {
		spec.TargetDatabase = &AwsGlueCatalogDatabaseTarget{
			CatalogId:    "987654321098",
			DatabaseName: "shared_sales_data",
		}
		spec.FederatedDatabase = &AwsGlueCatalogDatabaseFederation{
			Identifier:     "arn:aws:redshift:us-west-2:123456789012:datashare:abc-123/sales_share",
			ConnectionName: "aws:redshift",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// API envelope validations (from api.proto)
	// -------------------------------------------------------------------------

	ginkgo.It("fails when apiVersion is wrong", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "wrong/v1",
			Kind:       "AwsGlueCatalogDatabase",
			Metadata:   &shared.CatalogObjectMetadata{Name: "test-db"},
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "WrongKind",
			Metadata:   &shared.CatalogObjectMetadata{Name: "test-db"},
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsGlueCatalogDatabase",
			Spec:       spec,
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsGlueCatalogDatabase",
			Metadata:   &shared.CatalogObjectMetadata{Name: "test-db"},
		}
		err := protovalidate.Validate(envelope)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts a valid complete envelope", func() {
		envelope := &AwsGlueCatalogDatabase{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsGlueCatalogDatabase",
			Metadata:   &shared.CatalogObjectMetadata{Name: "analytics-db"},
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
