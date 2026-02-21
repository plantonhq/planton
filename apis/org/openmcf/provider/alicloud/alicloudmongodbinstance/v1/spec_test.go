package alicloudmongodbinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAliCloudMongodbInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AliCloudMongodbInstanceSpec Validation Tests")
}

var _ = ginkgo.Describe("AliCloudMongodbInstanceSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-mongodb",
				},
				Spec: &AliCloudMongodbInstanceSpec{
					Region:            "cn-hangzhou",
					EngineVersion:     "7.0",
					DbInstanceClass:   "dds.mongo.mid",
					DbInstanceStorage: 20,
					AccountPassword:   "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prod-mongodb",
					Org:  "acme-corp",
					Env:  "production",
				},
				Spec: &AliCloudMongodbInstanceSpec{
					Region:            "cn-shanghai",
					EngineVersion:     "6.0",
					DbInstanceClass:   "mongo.x8.large",
					DbInstanceStorage: 100,
					AccountPassword:   "ProdP@ssw0rd!",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-xyz789"},
					},
					DbInstanceName:              "prod-mongodb-primary",
					ZoneId:                      "cn-shanghai-a",
					SecondaryZoneId:             "cn-shanghai-b",
					HiddenZoneId:                "cn-shanghai-c",
					ReplicationFactor:           proto.Int32(5),
					ReadonlyReplicas:            proto.Int32(2),
					StorageEngine:               proto.String("WiredTiger"),
					StorageType:                 proto.String("cloud_essd1"),
					ProvisionedIops:             proto.Int32(2000),
					InstanceChargeType:          proto.String("PostPaid"),
					SecurityIpList:              []string{"10.0.0.0/8"},
					SecurityGroupId:             "sg-abc123",
					ResourceGroupId:             "rg-abc123",
					Tags:                        map[string]string{"team": "platform"},
					SslAction:                   proto.String("Open"),
					MaintainStartTime:           "02:00Z",
					MaintainEndTime:             "06:00Z",
					BackupPeriod:                []string{"Monday", "Wednesday", "Friday"},
					BackupTime:                  "02:00Z-03:00Z",
					Parameters:                  map[string]string{"operationProfiling.slowOpThresholdMs": "200"},
					DbInstanceReleaseProtection: proto.Bool(true),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PrePaid billing", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prepaid-mongodb",
				},
				Spec: &AliCloudMongodbInstanceSpec{
					Region:            "cn-hangzhou",
					EngineVersion:     "7.0",
					DbInstanceClass:   "dds.mongo.mid",
					DbInstanceStorage: 50,
					AccountPassword:   "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					InstanceChargeType: proto.String("PrePaid"),
					Period:             proto.Int32(12),
					AutoRenew:          proto.Bool(true),
					AutoRenewDuration:  proto.Int32(3),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with multi-zone HA deployment", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "ha-mongodb",
				},
				Spec: &AliCloudMongodbInstanceSpec{
					Region:            "cn-hangzhou",
					EngineVersion:     "6.0",
					DbInstanceClass:   "mongo.x8.medium",
					DbInstanceStorage: 50,
					AccountPassword:   "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					ZoneId:            "cn-hangzhou-a",
					SecondaryZoneId:   "cn-hangzhou-b",
					HiddenZoneId:      "cn-hangzhou-c",
					ReplicationFactor: proto.Int32(3),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with TDE encryption enabled", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "encrypted-mongodb",
				},
				Spec: &AliCloudMongodbInstanceSpec{
					Region:            "cn-hangzhou",
					EngineVersion:     "7.0",
					DbInstanceClass:   "mongo.x8.large",
					DbInstanceStorage: 100,
					AccountPassword:   "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					TdeStatus:     proto.String("enabled"),
					EncryptionKey: "kms-key-abc123",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with cloud disk encryption", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "disk-encrypted-mongodb",
				},
				Spec: &AliCloudMongodbInstanceSpec{
					Region:            "cn-hangzhou",
					EngineVersion:     "7.0",
					DbInstanceClass:   "dds.mongo.mid",
					DbInstanceStorage: 20,
					AccountPassword:   "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					Encrypted:              proto.Bool(true),
					CloudDiskEncryptionKey: "kms-disk-key-123",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "wrong/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_id is missing", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when engine_version is invalid", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "3.6",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when db_instance_class is empty", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when account_password is too short", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "short",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when replication_factor is invalid", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					ReplicationFactor: proto.Int32(4),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when readonly_replicas exceeds max", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					ReadonlyReplicas: proto.Int32(10),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when storage_engine is invalid", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					StorageEngine: proto.String("InnoDB"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_charge_type is invalid", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					InstanceChargeType: proto.String("Free"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when tde_status is invalid", func() {
			input := &AliCloudMongodbInstance{
				ApiVersion: "ali-cloud.openmcf.org/v1",
				Kind:       "AliCloudMongodbInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AliCloudMongodbInstanceSpec{
					Region: "cn-hangzhou", EngineVersion: "7.0",
					DbInstanceClass: "dds.mongo.mid", DbInstanceStorage: 20,
					AccountPassword: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					TdeStatus: proto.String("disabled"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
