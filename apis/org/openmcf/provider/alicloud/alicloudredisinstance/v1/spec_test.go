package alicloudredisinstancev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAlicloudRedisInstanceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudRedisInstanceSpec Validation Tests")
}

var _ = ginkgo.Describe("AlicloudRedisInstanceSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-redis",
				},
				Spec: &AlicloudRedisInstanceSpec{
					Region:        "cn-hangzhou",
					InstanceClass: "redis.master.small.default",
					Password:      "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prod-redis",
					Org:  "acme-corp",
					Env:  "production",
				},
				Spec: &AlicloudRedisInstanceSpec{
					Region:        "cn-shanghai",
					InstanceClass: "redis.master.large.default",
					Password:      "ProdP@ssw0rd!",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-xyz789"},
					},
					EngineVersion:              proto.String("7.0"),
					InstanceType:               proto.String("Redis"),
					DbInstanceName:             "prod-redis-primary",
					ZoneId:                     "cn-shanghai-a",
					SecondaryZoneId:            "cn-shanghai-b",
					PaymentType:                proto.String("PostPaid"),
					SecurityIps:                []string{"10.0.0.0/8"},
					SecurityGroupId:            "sg-abc123",
					ResourceGroupId:            "rg-abc123",
					Tags:                       map[string]string{"team": "platform"},
					ShardCount:                 proto.Int32(4),
					ReadOnlyCount:              proto.Int32(2),
					SslEnable:                  proto.String("Enable"),
					VpcAuthMode:                proto.String("Open"),
					Config:                     map[string]string{"maxmemory-policy": "allkeys-lru"},
					InstanceReleaseProtection:  proto.Bool(true),
					MaintainStartTime:          "02:00Z",
					MaintainEndTime:            "06:00Z",
					BackupPeriod:               []string{"Monday", "Wednesday", "Friday"},
					BackupTime:                 "02:00Z-03:00Z",
					PrivateConnectionPrefix:    "my-redis",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with PrePaid billing and period", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prepaid-redis",
				},
				Spec: &AlicloudRedisInstanceSpec{
					Region:        "cn-hangzhou",
					InstanceClass: "redis.master.small.default",
					Password:      "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					PaymentType:    proto.String("PrePaid"),
					Period:         proto.String("12"),
					AutoRenew:      proto.Bool(true),
					AutoRenewPeriod: proto.Int32(3),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Memcache instance type", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "memcache-instance",
				},
				Spec: &AlicloudRedisInstanceSpec{
					Region:        "cn-hangzhou",
					InstanceClass: "memcache.master.small.default",
					Password:      "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					InstanceType:  proto.String("Memcache"),
					EngineVersion: proto.String("4.0"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with TDE encryption enabled", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata: &shared.CloudResourceMetadata{
					Name: "encrypted-redis",
				},
				Spec: &AlicloudRedisInstanceSpec{
					Region:        "cn-hangzhou",
					InstanceClass: "redis.master.large.default",
					Password:      "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
					TdeStatus:     proto.String("Enabled"),
					EncryptionKey: "kms-key-abc123",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when region is empty", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_id is missing", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_class is empty", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when password is too short", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "short",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when engine_version is invalid", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					EngineVersion: proto.String("3.5"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when instance_type is invalid", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					InstanceType: proto.String("CouchDB"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when payment_type is invalid", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					PaymentType: proto.String("Free"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when ssl_enable is invalid", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					SslEnable: proto.String("TurnOn"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_auth_mode is invalid", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					VpcAuthMode: proto.String("Partial"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when read_only_count exceeds max", func() {
			input := &AlicloudRedisInstance{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudRedisInstance",
				Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				Spec: &AlicloudRedisInstanceSpec{
					Region: "cn-hangzhou", InstanceClass: "redis.master.small.default",
					Password: "SecureP@ss123",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-123"},
					},
					ReadOnlyCount: proto.Int32(15),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
