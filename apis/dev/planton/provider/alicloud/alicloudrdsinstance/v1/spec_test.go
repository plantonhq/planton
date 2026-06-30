package alicloudrdsinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	fkv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudRdsInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudRdsInstanceSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudRdsInstanceSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-rds",
				},
				Spec: &AliCloudRdsInstanceSpec{
					Region:          "cn-hangzhou",
					Engine:          "MySQL",
					EngineVersion:   "8.0",
					InstanceType:    "rds.mysql.s2.large",
					InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prod-mysql",
					Org:  "acme-corp",
					Env:  "production",
				},
				Spec: &AliCloudRdsInstanceSpec{
					Region:          "cn-shanghai",
					Engine:          "MySQL",
					EngineVersion:   "8.0",
					InstanceType:    "rds.mysql.s2.large",
					InstanceStorage: 200,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-xyz789"},
					},
					InstanceName:          "prod-mysql-primary",
					InstanceChargeType:    proto.String("Postpaid"),
					Category:              proto.String("HighAvailability"),
					DbInstanceStorageType: proto.String("cloud_essd"),
					ZoneId:                "cn-shanghai-a",
					ZoneIdSlaveA:          "cn-shanghai-b",
					SecurityIps:           []string{"10.0.0.0/8"},
					SecurityGroupIds:      []string{"sg-abc123"},
					MonitoringPeriod:      proto.Int32(60),
					MaintainTime:          "02:00Z-06:00Z",
					DeletionProtection:    proto.Bool(true),
					SslAction:             proto.String("Open"),
					EncryptionKey:         "kms-key-123",
					ResourceGroupId:       "rg-abc123",
					Tags:                  map[string]string{"team": "platform"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PostgreSQL engine", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "pg-test",
				},
				Spec: &AliCloudRdsInstanceSpec{
					Region:          "cn-hangzhou",
					Engine:          "PostgreSQL",
					EngineVersion:   "16.0",
					InstanceType:    "rds.pg.s2.large",
					InstanceStorage: 100,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-pg123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with databases and accounts", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "db-with-accounts",
				},
				Spec: &AliCloudRdsInstanceSpec{
					Region:          "cn-hangzhou",
					Engine:          "MySQL",
					EngineVersion:   "8.0",
					InstanceType:    "rds.mysql.s2.large",
					InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					Databases: []*AliCloudRdsDatabase{
						{Name: "appdb", CharacterSet: "utf8mb4"},
						{Name: "analytics", Description: "Analytics database"},
					},
					Accounts: []*AliCloudRdsAccount{
						{
							AccountName:     "app_user",
							AccountPassword: "SecureP@ss123",
							AccountType:     proto.String("Normal"),
							Privileges: []*AliCloudRdsAccountPrivilege{
								{
									DatabaseNames: []string{"appdb"},
									Privilege:     proto.String("ReadWrite"),
								},
								{
									DatabaseNames: []string{"analytics"},
									Privilege:     proto.String("ReadOnly"),
								},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Prepaid billing and period", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prepaid-rds",
				},
				Spec: &AliCloudRdsInstanceSpec{
					Region:          "cn-hangzhou",
					Engine:          "MySQL",
					EngineVersion:   "8.0",
					InstanceType:    "rds.mysql.s2.large",
					InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					InstanceChargeType: proto.String("Prepaid"),
					Period:             proto.Int32(12),
					AutoRenew:          proto.Bool(true),
					AutoRenewPeriod:    proto.Int32(3),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with parameters", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "param-rds",
				},
				Spec: &AliCloudRdsInstanceSpec{
					Region:          "cn-hangzhou",
					Engine:          "MySQL",
					EngineVersion:   "8.0",
					InstanceType:    "rds.mysql.s2.large",
					InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					Parameters: []*AliCloudRdsParameter{
						{Name: "innodb_buffer_pool_size", Value: "1073741824"},
						{Name: "max_connections", Value: "500"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "WrongKind",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when engine is invalid", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MongoDB", EngineVersion: "6.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_id is missing", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_storage is zero", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 0,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when category is invalid", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					Category: proto.String("SuperHA"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ssl_action is invalid", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					SslAction: proto.String("Enable"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when account_password is too short", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					Accounts: []*AliCloudRdsAccount{
						{AccountName: "user1", AccountPassword: "short"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when privilege database_names is empty", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					Accounts: []*AliCloudRdsAccount{
						{
							AccountName:     "user1",
							AccountPassword: "SecureP@ss123",
							Privileges: []*AliCloudRdsAccountPrivilege{
								{DatabaseNames: []string{}, Privilege: proto.String("ReadOnly")},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when privilege value is invalid", func() {
			input := &AliCloudRdsInstance{
				ApiVersion: "alicloud.planton.dev/v1",
				Kind:       "AliCloudRdsInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudRdsInstanceSpec{
					Region: "cn-hangzhou", Engine: "MySQL", EngineVersion: "8.0",
					InstanceType: "rds.mysql.s2.large", InstanceStorage: 50,
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					Accounts: []*AliCloudRdsAccount{
						{
							AccountName:     "user1",
							AccountPassword: "SecureP@ss123",
							Privileges: []*AliCloudRdsAccountPrivilege{
								{DatabaseNames: []string{"mydb"}, Privilege: proto.String("Admin")},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
