package alicloudpolardbclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudPolardbClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudPolardbClusterSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudPolardbClusterSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-polardb",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-hangzhou",
					DbType:      "MySQL",
					DbVersion:   "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prod-polardb",
					Org:  "acme-corp",
					Env:  "production",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-shanghai",
					DbType:      "MySQL",
					DbVersion:   "8.0",
					DbNodeClass: "polar.mysql.x4.xlarge",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-xyz789"},
					},
					DbNodeCount:                            proto.Int32(4),
					Description:                            "Production PolarDB cluster",
					PayType:                                proto.String("PostPaid"),
					ZoneId:                                 "cn-shanghai-a",
					SecurityIps:                            []string{"10.0.0.0/8"},
					SecurityGroupIds:                       []string{"sg-abc123"},
					MaintainTime:                           "02:00Z-03:00Z",
					ResourceGroupId:                        "rg-abc123",
					Tags:                                   map[string]string{"team": "platform"},
					CreationCategory:                       proto.String("Normal"),
					SubCategory:                            proto.String("Exclusive"),
					StorageType:                            proto.String("PSL5"),
					TdeStatus:                              proto.String("Disabled"),
					DeletionLock:                           proto.Int32(1),
					CollectorStatus:                        proto.String("Enable"),
					BackupRetentionPolicyOnClusterDeletion: proto.String("LATEST"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PostgreSQL engine", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "pg-cluster",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-hangzhou",
					DbType:      "PostgreSQL",
					DbVersion:   "14",
					DbNodeClass: "polar.pg.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-pg123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Oracle engine", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "oracle-cluster",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-hangzhou",
					DbType:      "Oracle",
					DbVersion:   "11",
					DbNodeClass: "polar.o.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-oracle123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with databases and accounts", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "db-with-accounts",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-hangzhou",
					DbType:      "MySQL",
					DbVersion:   "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					Databases: []*AliCloudPolardbDatabase{
						{DbName: "appdb", CharacterSetName: "utf8mb4"},
						{DbName: "analytics", DbDescription: "Analytics database"},
					},
					Accounts: []*AliCloudPolardbAccount{
						{
							AccountName:     "app_user",
							AccountPassword: "SecureP@ss123",
							AccountType:     proto.String("Normal"),
							Privileges: []*AliCloudPolardbAccountPrivilege{
								{
									DbNames:          []string{"appdb"},
									AccountPrivilege: proto.String("ReadWrite"),
								},
								{
									DbNames:          []string{"analytics"},
									AccountPrivilege: proto.String("ReadOnly"),
								},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PostgreSQL databases using collate and ctype", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "pg-with-collation",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-hangzhou",
					DbType:      "PostgreSQL",
					DbVersion:   "14",
					DbNodeClass: "polar.pg.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-pg123"},
					},
					Databases: []*AliCloudPolardbDatabase{
						{
							DbName:           "mydb",
							CharacterSetName: "UTF8",
							Collate:          "C",
							Ctype:            "C",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PrePaid billing and subscription settings", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prepaid-polardb",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-hangzhou",
					DbType:      "MySQL",
					DbVersion:   "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					PayType:         proto.String("PrePaid"),
					Period:          proto.Int32(12),
					RenewalStatus:   proto.String("AutoRenewal"),
					AutoRenewPeriod: proto.Int32(3),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Standard Edition storage configuration", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "standard-edition",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-hangzhou",
					DbType:      "MySQL",
					DbVersion:   "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					CreationCategory: proto.String("SENormal"),
					StorageType:      proto.String("ESSDPL1"),
					StorageSpace:     proto.Int32(100),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with parameters", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "param-polardb",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-hangzhou",
					DbType:      "MySQL",
					DbVersion:   "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					Parameters: []*AliCloudPolardbParameter{
						{Name: "loose_innodb_buffer_pool_size", Value: "1073741824"},
						{Name: "max_connections", Value: "500"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with vswitch_id as valueFrom reference", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata: &shared.CloudResourceMetadata{
					Name: "ref-polardb",
				},
				Spec: &AliCloudPolardbClusterSpec{
					Region:      "cn-hangzhou",
					DbType:      "MySQL",
					DbVersion:   "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
							ValueFrom: &fkv1.ValueFromRef{
								Name: "my-vswitch",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when db_type is invalid", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "SQLServer", DbVersion: "2019",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_id is missing", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when db_node_class is empty", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when pay_type is invalid", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					PayType: proto.String("PayAsYouGo"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when creation_category is invalid", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					CreationCategory: proto.String("SuperHA"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when storage_type is invalid", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					StorageType: proto.String("cloud_ssd"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when storage_space is below minimum", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					StorageSpace: proto.Int32(5),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when deletion_lock is not 0 or 1", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					DeletionLock: proto.Int32(2),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when account_password is too short", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					Accounts: []*AliCloudPolardbAccount{
						{AccountName: "user1", AccountPassword: "short"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when privilege db_names is empty", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					Accounts: []*AliCloudPolardbAccount{
						{
							AccountName:     "user1",
							AccountPassword: "SecureP@ss123",
							Privileges: []*AliCloudPolardbAccountPrivilege{
								{DbNames: []string{}, AccountPrivilege: proto.String("ReadOnly")},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when account_privilege value is invalid", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					Accounts: []*AliCloudPolardbAccount{
						{
							AccountName:     "user1",
							AccountPassword: "SecureP@ss123",
							Privileges: []*AliCloudPolardbAccountPrivilege{
								{DbNames: []string{"mydb"}, AccountPrivilege: proto.String("DBOwner")},
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when description is too short", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					Description: "x",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when db_node_count exceeds maximum", func() {
			input := &AliCloudPolardbCluster{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AliCloudPolardbCluster",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudPolardbClusterSpec{
					Region: "cn-hangzhou", DbType: "MySQL", DbVersion: "8.0",
					DbNodeClass: "polar.mysql.x4.large",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					DbNodeCount: proto.Int32(20),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
