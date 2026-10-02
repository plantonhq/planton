package gcpbigqueryconnectionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpBigQueryConnectionSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpBigQueryConnectionSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	base := func() *GcpBigQueryConnection {
		return &GcpBigQueryConnection{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBigQueryConnection",
			Metadata:   &shared.CloudResourceMetadata{Name: "lake"},
			Spec:       &GcpBigQueryConnectionSpec{Location: "US"},
		}
	}
	cloudSql := func() *GcpBigQueryConnectionCloudSql {
		return &GcpBigQueryConnectionCloudSql{
			InstanceId: litRef("p:us-central1:orders"),
			Database:   "orders",
			Type:       "POSTGRES",
			Credential: &GcpBigQueryConnectionCloudSqlCredential{Username: "bq", Password: "s3cret"},
		}
	}

	ginkgo.It("should accept each arm on its own", func() {
		arms := []func(*GcpBigQueryConnectionSpec){
			func(s *GcpBigQueryConnectionSpec) { s.CloudResource = true },
			func(s *GcpBigQueryConnectionSpec) {
				s.Aws = &GcpBigQueryConnectionAws{IamRoleId: "arn:aws:iam::123456789012:role/omni"}
			},
			func(s *GcpBigQueryConnectionSpec) {
				s.Azure = &GcpBigQueryConnectionAzure{CustomerTenantId: "tenant", FederatedApplicationClientId: "client"}
			},
			func(s *GcpBigQueryConnectionSpec) {
				s.CloudSpanner = &GcpBigQueryConnectionCloudSpanner{Database: "p/i/d", DatabaseRole: "analyst",
					UseParallelism: true, UseDataBoost: true, MaxParallelism: 4}
			},
			func(s *GcpBigQueryConnectionSpec) { s.CloudSql = cloudSql() },
			func(s *GcpBigQueryConnectionSpec) {
				s.Configuration = &GcpBigQueryConnectionConfiguration{
					ConnectorId:       "google-alloydb",
					Asset:             &GcpBigQueryConnectionConnectorAsset{Database: "orders", GoogleCloudResource: "//alloydb.googleapis.com/projects/p/locations/us-central1/clusters/c/instances/i"},
					UsernamePassword:  &GcpBigQueryConnectionUsernamePassword{Username: "bq", Password: "s3cret"},
					HostPort:          "10.0.0.5:5432",
					NetworkAttachment: "projects/p/regions/us-central1/networkAttachments/bq",
				}
			},
			func(s *GcpBigQueryConnectionSpec) {
				s.Spark = &GcpBigQueryConnectionSpark{
					MetastoreService:             "projects/p/locations/us-central1/services/hive",
					HistoryServerDataprocCluster: "projects/p/regions/us-central1/clusters/history",
				}
			},
		}
		for i, arm := range arms {
			msg := base()
			arm(msg.Spec)
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), "arm %d", i)
		}
	})

	ginkgo.It("should accept the shared fields set", func() {
		msg := base()
		msg.Spec.CloudResource = true
		msg.Spec.ProjectId = litRef("data-project")
		msg.Spec.ConnectionId = "lake"
		msg.Spec.FriendlyName = "Data lake"
		msg.Spec.Description = "BigLake over the raw bucket"
		msg.Spec.KmsKeyName = litRef("projects/p/locations/us/keyRings/r/cryptoKeys/k")
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject no arm and two arms", func() {
		gomega.Expect(validator.Validate(base())).ToNot(gomega.Succeed())
		msg := base()
		msg.Spec.CloudResource = true
		msg.Spec.CloudSql = cloudSql()
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should enforce the Spanner dependencies", func() {
		msg := base()
		msg.Spec.CloudSpanner = &GcpBigQueryConnectionCloudSpanner{Database: "p/i/d", UseDataBoost: true}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.CloudSpanner = &GcpBigQueryConnectionCloudSpanner{Database: "p/i/d", UseParallelism: true, MaxParallelism: 4}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.CloudSpanner = &GcpBigQueryConnectionCloudSpanner{Database: "projects/p/instances/i/databases/d"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.CloudSpanner = &GcpBigQueryConnectionCloudSpanner{Database: "p/i/d", DatabaseRole: "1analyst"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should enforce the Cloud SQL arm's required fields and engine list", func() {
		msg := base()
		msg.Spec.CloudSql = cloudSql()
		msg.Spec.CloudSql.Type = "SQLSERVER"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.CloudSql = cloudSql()
		msg.Spec.CloudSql.Credential.Password = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.CloudSql = cloudSql()
		msg.Spec.CloudSql.InstanceId = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject malformed resource paths and missing required arm fields", func() {
		msg := base()
		msg.Spec.Configuration = &GcpBigQueryConnectionConfiguration{ConnectorId: "google-alloydb",
			Asset: &GcpBigQueryConnectionConnectorAsset{}, NetworkAttachment: "bq"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.Configuration = &GcpBigQueryConnectionConfiguration{ConnectorId: "google-alloydb"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.Spark = &GcpBigQueryConnectionSpark{MetastoreService: "hive"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.Spark = &GcpBigQueryConnectionSpark{HistoryServerDataprocCluster: "projects/p/locations/us-central1/clusters/h"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.Aws = &GcpBigQueryConnectionAws{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = base()
		msg.Spec.Azure = &GcpBigQueryConnectionAzure{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := base()
		msg.Spec.CloudResource = true
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
