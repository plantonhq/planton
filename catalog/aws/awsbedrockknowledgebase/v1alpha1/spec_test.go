package awsbedrockknowledgebasev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsBedrockKnowledgeBaseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsBedrockKnowledgeBaseSpec Validation Suite")
}

func svr(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// minimalManagedKb is the smallest valid manifest: region, role, and the
// MANAGED type (AWS manages the vector store).
func minimalManagedKb() *AwsBedrockKnowledgeBaseSpec {
	return &AwsBedrockKnowledgeBaseSpec{
		Region:            "us-west-2",
		RoleArn:           svr("arn:aws:iam::123456789012:role/bedrock-kb"),
		KnowledgeBaseType: &AwsBedrockKnowledgeBaseSpec_Managed{Managed: &AwsBedrockKnowledgeBaseManagedConfig{}},
	}
}

// vectorS3VectorsKb is a VECTOR knowledge base on S3 Vectors.
func vectorS3VectorsKb() *AwsBedrockKnowledgeBaseSpec {
	return &AwsBedrockKnowledgeBaseSpec{
		Region:  "us-west-2",
		RoleArn: svr("arn:aws:iam::123456789012:role/bedrock-kb"),
		KnowledgeBaseType: &AwsBedrockKnowledgeBaseSpec_Vector{Vector: &AwsBedrockKnowledgeBaseVectorConfig{
			EmbeddingModelArn: "arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-text-v2:0",
		}},
		Storage: &AwsBedrockKnowledgeBaseStorage{
			S3Vectors: &AwsBedrockKnowledgeBaseS3VectorsStorage{
				IndexArn: "arn:aws:s3vectors:us-west-2:123456789012:bucket/kb-vectors/index/kb-index",
			},
		},
	}
}

var _ = ginkgo.Describe("AwsBedrockKnowledgeBaseSpec validations", func() {

	// -----------------------------------------------------------------
	// Valid inputs
	// -----------------------------------------------------------------
	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("with the minimal MANAGED shape", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(minimalManagedKb())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a VECTOR type on S3 Vectors", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(vectorS3VectorsKb())
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a VECTOR type on OpenSearch Serverless plus an S3 data source", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := vectorS3VectorsKb()
				spec.Storage = &AwsBedrockKnowledgeBaseStorage{
					OpensearchServerless: &AwsBedrockKnowledgeBaseOpenSearchServerlessStorage{
						CollectionArn:   svr("arn:aws:aoss:us-west-2:123456789012:collection/abc"),
						VectorIndexName: "bedrock-knowledge-base-default-index",
						FieldMapping: &AwsBedrockKnowledgeBaseFieldMapping{
							VectorField:   "bedrock-knowledge-base-default-vector",
							TextField:     "AMAZON_BEDROCK_TEXT_CHUNK",
							MetadataField: "AMAZON_BEDROCK_METADATA",
						},
					},
				}
				spec.GetVector().EmbeddingModel = &AwsBedrockKnowledgeBaseEmbeddingModelConfig{
					Dimensions:        1024,
					EmbeddingDataType: "FLOAT32",
				}
				spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
					Name:               "docs",
					Description:        "product documentation",
					DataDeletionPolicy: "DELETE",
					Connector: &AwsBedrockKnowledgeBaseDataSource_S3{S3: &AwsBedrockKnowledgeBaseS3DataSource{
						BucketArn:       svr("arn:aws:s3:::kb-docs"),
						InclusionPrefix: "manuals/",
					}},
					VectorIngestion: &AwsBedrockKnowledgeBaseVectorIngestion{
						Chunking: &AwsBedrockKnowledgeBaseChunking{
							Strategy: "FIXED_SIZE",
							FixedSize: &AwsBedrockKnowledgeBaseFixedSizeChunking{
								MaxTokens:         300,
								OverlapPercentage: 20,
							},
						},
					},
				}}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a SQL type on Redshift Serverless", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsBedrockKnowledgeBaseSpec{
					Region:  "us-west-2",
					RoleArn: svr("arn:aws:iam::123456789012:role/bedrock-kb"),
					KnowledgeBaseType: &AwsBedrockKnowledgeBaseSpec_Sql{Sql: &AwsBedrockKnowledgeBaseSqlConfig{
						Serverless: &AwsBedrockKnowledgeBaseRedshiftServerless{
							WorkgroupArn: svr("arn:aws:redshift-serverless:us-west-2:123456789012:workgroup/abc"),
							Auth: &AwsBedrockKnowledgeBaseRedshiftServerlessAuth{
								Type: "IAM",
							},
						},
						Warehouse: &AwsBedrockKnowledgeBaseRedshiftStorage{
							Redshift: &AwsBedrockKnowledgeBaseRedshiftDatabaseStorage{
								DatabaseName: "analytics",
							},
						},
						QueryGeneration: &AwsBedrockKnowledgeBaseQueryGeneration{
							ExecutionTimeoutSeconds: 60,
							CuratedQueries: []*AwsBedrockKnowledgeBaseCuratedQuery{
								{NaturalLanguage: "How many orders last month?", Sql: "SELECT count(*) FROM orders"},
							},
							Tables: []*AwsBedrockKnowledgeBaseQueryTable{
								{
									Name:        "analytics.public.orders",
									Description: "one row per order",
									Columns: []*AwsBedrockKnowledgeBaseQueryColumn{
										{Name: "internal_notes", Inclusion: "EXCLUDE"},
									},
								},
							},
						},
					}},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a KENDRA type", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsBedrockKnowledgeBaseSpec{
					Region:  "us-west-2",
					RoleArn: svr("arn:aws:iam::123456789012:role/bedrock-kb"),
					KnowledgeBaseType: &AwsBedrockKnowledgeBaseSpec_Kendra{Kendra: &AwsBedrockKnowledgeBaseKendraConfig{
						KendraIndexArn: "arn:aws:kendra:us-west-2:123456789012:index/abc-def",
					}},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a web crawler data source", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := minimalManagedKb()
				spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
					Name: "site",
					Connector: &AwsBedrockKnowledgeBaseDataSource_Web{Web: &AwsBedrockKnowledgeBaseWebDataSource{
						SeedUrls:  []string{"https://docs.example.com"},
						Scope:     "HOST_ONLY",
						MaxPages:  500,
						RateLimit: 60,
					}},
				}}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	// -----------------------------------------------------------------
	// Type union
	// -----------------------------------------------------------------
	ginkgo.Describe("Knowledge-base type union", func() {

		ginkgo.It("should reject a spec with no type arm", func() {
			spec := minimalManagedKb()
			spec.KnowledgeBaseType = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
		})

		ginkgo.It("carries exactly one type arm by construction (setting a second arm replaces the first)", func() {
			spec := vectorS3VectorsKb()
			spec.KnowledgeBaseType = &AwsBedrockKnowledgeBaseSpec_Managed{Managed: &AwsBedrockKnowledgeBaseManagedConfig{}}
			gomega.Expect(spec.GetVector()).To(gomega.BeNil())
			// The vector store that came with the vector arm is now forbidden.
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
			spec.Storage = nil
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should reject a vector type without storage", func() {
			spec := vectorS3VectorsKb()
			spec.Storage = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("storage"))
		})

		ginkgo.It("should reject a managed type with storage", func() {
			spec := minimalManagedKb()
			spec.Storage = &AwsBedrockKnowledgeBaseStorage{
				S3Vectors: &AwsBedrockKnowledgeBaseS3VectorsStorage{IndexArn: "arn:aws:s3vectors:us-west-2:123456789012:bucket/b/index/i"},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -----------------------------------------------------------------
	// Storage backends
	// -----------------------------------------------------------------
	ginkgo.Describe("Storage backends", func() {

		ginkgo.It("should reject two storage backends", func() {
			spec := vectorS3VectorsKb()
			spec.Storage.Pinecone = &AwsBedrockKnowledgeBasePineconeStorage{
				ConnectionString:     "https://idx.pinecone.io",
				CredentialsSecretArn: svr("arn:aws:secretsmanager:us-west-2:123456789012:secret:pc"),
				FieldMapping: &AwsBedrockKnowledgeBaseTextMetadataFieldMapping{
					TextField:     "text",
					MetadataField: "metadata",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject S3 Vectors with both index_arn and bucket addressing", func() {
			spec := vectorS3VectorsKb()
			spec.Storage.S3Vectors.VectorBucketArn = svr("arn:aws:s3vectors:us-west-2:123456789012:bucket/b")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject S3 Vectors with a bucket but no index name", func() {
			spec := vectorS3VectorsKb()
			spec.Storage.S3Vectors = &AwsBedrockKnowledgeBaseS3VectorsStorage{
				VectorBucketArn: svr("arn:aws:s3vectors:us-west-2:123456789012:bucket/b"),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept S3 Vectors addressed by bucket plus index name", func() {
			spec := vectorS3VectorsKb()
			spec.Storage.S3Vectors = &AwsBedrockKnowledgeBaseS3VectorsStorage{
				VectorBucketArn: svr("arn:aws:s3vectors:us-west-2:123456789012:bucket/b"),
				IndexName:       "kb-index",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -----------------------------------------------------------------
	// SQL shapes
	// -----------------------------------------------------------------
	ginkgo.Describe("SQL shapes", func() {

		ginkgo.It("should reject a provisioned USERNAME auth without database_user", func() {
			spec := &AwsBedrockKnowledgeBaseSpec{
				Region:  "us-west-2",
				RoleArn: svr("arn:aws:iam::123456789012:role/r"),
				KnowledgeBaseType: &AwsBedrockKnowledgeBaseSpec_Sql{Sql: &AwsBedrockKnowledgeBaseSqlConfig{
					Provisioned: &AwsBedrockKnowledgeBaseRedshiftProvisioned{
						ClusterIdentifier: svr("analytics"),
						Auth:              &AwsBedrockKnowledgeBaseRedshiftProvisionedAuth{Type: "USERNAME"},
					},
					Warehouse: &AwsBedrockKnowledgeBaseRedshiftStorage{
						Redshift: &AwsBedrockKnowledgeBaseRedshiftDatabaseStorage{DatabaseName: "analytics"},
					},
				}},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a serverless USERNAME_PASSWORD auth without the secret", func() {
			spec := &AwsBedrockKnowledgeBaseSpec{
				Region:  "us-west-2",
				RoleArn: svr("arn:aws:iam::123456789012:role/r"),
				KnowledgeBaseType: &AwsBedrockKnowledgeBaseSpec_Sql{Sql: &AwsBedrockKnowledgeBaseSqlConfig{
					Serverless: &AwsBedrockKnowledgeBaseRedshiftServerless{
						WorkgroupArn: svr("arn:aws:redshift-serverless:us-west-2:123456789012:workgroup/w"),
						Auth:         &AwsBedrockKnowledgeBaseRedshiftServerlessAuth{Type: "USERNAME_PASSWORD"},
					},
					Warehouse: &AwsBedrockKnowledgeBaseRedshiftStorage{
						DataCatalog: &AwsBedrockKnowledgeBaseDataCatalogStorage{TableNames: []string{"t"}},
					},
				}},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a warehouse with both metadata sources", func() {
			spec := &AwsBedrockKnowledgeBaseSpec{
				Region:  "us-west-2",
				RoleArn: svr("arn:aws:iam::123456789012:role/r"),
				KnowledgeBaseType: &AwsBedrockKnowledgeBaseSpec_Sql{Sql: &AwsBedrockKnowledgeBaseSqlConfig{
					Serverless: &AwsBedrockKnowledgeBaseRedshiftServerless{
						WorkgroupArn: svr("arn:aws:redshift-serverless:us-west-2:123456789012:workgroup/w"),
						Auth:         &AwsBedrockKnowledgeBaseRedshiftServerlessAuth{Type: "IAM"},
					},
					Warehouse: &AwsBedrockKnowledgeBaseRedshiftStorage{
						DataCatalog: &AwsBedrockKnowledgeBaseDataCatalogStorage{TableNames: []string{"t"}},
						Redshift:    &AwsBedrockKnowledgeBaseRedshiftDatabaseStorage{DatabaseName: "d"},
					},
				}},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -----------------------------------------------------------------
	// Data sources
	// -----------------------------------------------------------------
	ginkgo.Describe("Data sources", func() {

		ginkgo.It("should reject duplicate data source names", func() {
			spec := minimalManagedKb()
			ds := func() *AwsBedrockKnowledgeBaseDataSource {
				return &AwsBedrockKnowledgeBaseDataSource{
					Name:      "docs",
					Connector: &AwsBedrockKnowledgeBaseDataSource_S3{S3: &AwsBedrockKnowledgeBaseS3DataSource{BucketArn: svr("arn:aws:s3:::b")}},
				}
			}
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{ds(), ds()}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a data source with no connector", func() {
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{Name: "docs"}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("carries exactly one connector by construction (setting a second arm replaces the first)", func() {
			ds := &AwsBedrockKnowledgeBaseDataSource{
				Name:      "docs",
				Connector: &AwsBedrockKnowledgeBaseDataSource_S3{S3: &AwsBedrockKnowledgeBaseS3DataSource{BucketArn: svr("arn:aws:s3:::b")}},
			}
			ds.Connector = &AwsBedrockKnowledgeBaseDataSource_Web{Web: &AwsBedrockKnowledgeBaseWebDataSource{SeedUrls: []string{"https://x.com"}}}
			gomega.Expect(ds.GetS3()).To(gomega.BeNil())
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{ds}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("should reject a seed URL without a scheme", func() {
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
				Name:      "site",
				Connector: &AwsBedrockKnowledgeBaseDataSource_Web{Web: &AwsBedrockKnowledgeBaseWebDataSource{SeedUrls: []string{"docs.example.com"}}},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a SharePoint tenant id that is not a UUID", func() {
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
				Name: "sp",
				Connector: &AwsBedrockKnowledgeBaseDataSource_Sharepoint{Sharepoint: &AwsBedrockKnowledgeBaseSharePointDataSource{
					SiteUrls:             []string{"https://x.sharepoint.com/sites/docs"},
					Domain:               "x",
					TenantId:             "not-a-uuid",
					AuthType:             "OAUTH2_CLIENT_CREDENTIALS",
					CredentialsSecretArn: svr("arn:aws:secretsmanager:us-west-2:123456789012:secret:sp"),
				}},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -----------------------------------------------------------------
	// Chunking and parsing shapes
	// -----------------------------------------------------------------
	ginkgo.Describe("Chunking and parsing", func() {

		ginkgo.It("should reject a FIXED_SIZE strategy without its block", func() {
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
				Name:      "docs",
				Connector: &AwsBedrockKnowledgeBaseDataSource_S3{S3: &AwsBedrockKnowledgeBaseS3DataSource{BucketArn: svr("arn:aws:s3:::b")}},
				VectorIngestion: &AwsBedrockKnowledgeBaseVectorIngestion{
					Chunking: &AwsBedrockKnowledgeBaseChunking{Strategy: "FIXED_SIZE"},
				},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a NONE strategy carrying a block", func() {
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
				Name:      "docs",
				Connector: &AwsBedrockKnowledgeBaseDataSource_S3{S3: &AwsBedrockKnowledgeBaseS3DataSource{BucketArn: svr("arn:aws:s3:::b")}},
				VectorIngestion: &AwsBedrockKnowledgeBaseVectorIngestion{
					Chunking: &AwsBedrockKnowledgeBaseChunking{
						Strategy:  "NONE",
						FixedSize: &AwsBedrockKnowledgeBaseFixedSizeChunking{MaxTokens: 300, OverlapPercentage: 20},
					},
				},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should require exactly two hierarchical levels", func() {
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
				Name:      "docs",
				Connector: &AwsBedrockKnowledgeBaseDataSource_S3{S3: &AwsBedrockKnowledgeBaseS3DataSource{BucketArn: svr("arn:aws:s3:::b")}},
				VectorIngestion: &AwsBedrockKnowledgeBaseVectorIngestion{
					Chunking: &AwsBedrockKnowledgeBaseChunking{
						Strategy: "HIERARCHICAL",
						Hierarchical: &AwsBedrockKnowledgeBaseHierarchicalChunking{
							OverlapTokens: 60,
							Levels: []*AwsBedrockKnowledgeBaseChunkingLevel{
								{MaxTokens: 1500},
							},
						},
					},
				},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject BEDROCK_FOUNDATION_MODEL parsing without its block", func() {
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
				Name:      "docs",
				Connector: &AwsBedrockKnowledgeBaseDataSource_S3{S3: &AwsBedrockKnowledgeBaseS3DataSource{BucketArn: svr("arn:aws:s3:::b")}},
				VectorIngestion: &AwsBedrockKnowledgeBaseVectorIngestion{
					Parsing: &AwsBedrockKnowledgeBaseParsing{Strategy: "BEDROCK_FOUNDATION_MODEL"},
				},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject multimodal on a non-BDA strategy", func() {
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
				Name:      "docs",
				Connector: &AwsBedrockKnowledgeBaseDataSource_S3{S3: &AwsBedrockKnowledgeBaseS3DataSource{BucketArn: svr("arn:aws:s3:::b")}},
				VectorIngestion: &AwsBedrockKnowledgeBaseVectorIngestion{
					Parsing: &AwsBedrockKnowledgeBaseParsing{Strategy: "SMART_PARSING", Multimodal: true},
				},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept SMART_PARSING with no block", func() {
			spec := minimalManagedKb()
			spec.DataSources = []*AwsBedrockKnowledgeBaseDataSource{{
				Name:      "docs",
				Connector: &AwsBedrockKnowledgeBaseDataSource_S3{S3: &AwsBedrockKnowledgeBaseS3DataSource{BucketArn: svr("arn:aws:s3:::b")}},
				VectorIngestion: &AwsBedrockKnowledgeBaseVectorIngestion{
					Parsing: &AwsBedrockKnowledgeBaseParsing{Strategy: "SMART_PARSING"},
				},
			}}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
