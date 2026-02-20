package alicloudnasfilesystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	fkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAlicloudNasFileSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AlicloudNasFileSystemSpec Validation Tests")
}

var _ = ginkgo.Describe("AlicloudNasFileSystemSpec Validation Tests", func() {

	ginkgo.Describe("valid input", func() {

		ginkgo.It("should pass with minimal required fields", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test-nas",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-abc123"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-abc123"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with all optional fields populated", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "prod-nas",
					Org:  "my-org",
					Env:  "production",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:         "cn-shanghai",
					FileSystemType: strPtr("standard"),
					ProtocolType:   "NFS",
					StorageType:    "Performance",
					Description:    "Production shared file system for microservices",
					Encryption: &AlicloudNasEncryption{
						EncryptType: 1,
					},
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-prod-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-prod-001"},
					},
					AccessRules: []*AlicloudNasAccessRule{
						{
							SourceCidrIp:   "10.0.0.0/8",
							RwAccessType:   strPtr("RDWR"),
							UserAccessType: strPtr("no_squash"),
							Priority:       int32Ptr(1),
						},
						{
							SourceCidrIp:   "172.16.0.0/12",
							RwAccessType:   strPtr("RDONLY"),
							UserAccessType: strPtr("root_squash"),
							Priority:       int32Ptr(50),
						},
					},
					ResourceGroupId: "rg-prod-123",
					Tags:            map[string]string{"team": "platform", "cost-center": "eng"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with extreme NAS configuration", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "extreme-nas",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:         "cn-hangzhou",
					FileSystemType: strPtr("extreme"),
					ProtocolType:   "NFS",
					StorageType:    "advance",
					Capacity:       500,
					ZoneId:         "cn-hangzhou-a",
					Encryption: &AlicloudNasEncryption{
						EncryptType: 2,
						KmsKeyId:    "kms-key-abc123",
					},
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-hpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-hpc-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with SMB protocol", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "smb-nas",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "SMB",
					StorageType:  "Capacity",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-win-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-win-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with Premium storage type", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "premium-nas",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Premium",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should pass with access rule using all_squash", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "squash-nas",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
					AccessRules: []*AlicloudNasAccessRule{
						{
							SourceCidrIp:   "0.0.0.0/0",
							UserAccessType: strPtr("all_squash"),
							Priority:       int32Ptr(100),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("invalid input", func() {

		ginkgo.It("should fail when region is missing", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when protocol_type is missing", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:      "cn-hangzhou",
					StorageType: "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when protocol_type has invalid value", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "FTP",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when storage_type has invalid value", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "SuperFast",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when file_system_type has invalid value", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:         "cn-hangzhou",
					FileSystemType: strPtr("cpfs"),
					ProtocolType:   "NFS",
					StorageType:    "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vpc_id is missing", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when vswitch_id is missing", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when encrypt_type has invalid value", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					Encryption: &AlicloudNasEncryption{
						EncryptType: 5,
					},
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when access_rule source_cidr_ip is missing", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
					AccessRules: []*AlicloudNasAccessRule{
						{
							RwAccessType: strPtr("RDWR"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when access_rule rw_access_type has invalid value", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
					AccessRules: []*AlicloudNasAccessRule{
						{
							SourceCidrIp: "10.0.0.0/8",
							RwAccessType: strPtr("WRITE_ONLY"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when access_rule user_access_type has invalid value", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
					AccessRules: []*AlicloudNasAccessRule{
						{
							SourceCidrIp:   "10.0.0.0/8",
							UserAccessType: strPtr("nuke_squash"),
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when api_version is wrong", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "wrong/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when kind is wrong", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when metadata is missing", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Spec: &AlicloudNasFileSystemSpec{
					Region:       "cn-hangzhou",
					ProtocolType: "NFS",
					StorageType:  "Performance",
					VpcId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vpc-001"},
					},
					VswitchId: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "vsw-001"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail when spec is missing", func() {
			input := &AlicloudNasFileSystem{
				ApiVersion: "alicloud.openmcf.org/v1",
				Kind:       "AlicloudNasFileSystem",
				Metadata: &shared.CloudResourceMetadata{
					Name: "test",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})

func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
