package ocifilesystemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciFileSystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciFileSystemSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidFileSystem() *OciFileSystem {
	return &OciFileSystem{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciFileSystem",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-fs",
		},
		Spec: &OciFileSystemSpec{
			CompartmentId:      newStringValueOrRef("ocid1.compartment.oc1..example"),
			AvailabilityDomain: "Uocm:US-ASHBURN-AD-1",
			MountTarget: &OciFileSystemSpec_MountTarget{
				SubnetId: newStringValueOrRef("ocid1.subnet.oc1..example"),
			},
			Exports: []*OciFileSystemSpec_Export{
				{Path: "/shared-data"},
			},
		},
	}
}

var _ = ginkgo.Describe("OciFileSystemSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_file_storage_file_system", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidFileSystem()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display_name set", func() {
				input := minimalValidFileSystem()
				input.Spec.DisplayName = "my-nfs-share"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with kms_key_id for encryption", func() {
				input := minimalValidFileSystem()
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with filesystem_snapshot_policy_id", func() {
				input := minimalValidFileSystem()
				input.Spec.FilesystemSnapshotPolicyId = newStringValueOrRef("ocid1.filesystemsnapshotpolicy.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with mount target display_name", func() {
				input := minimalValidFileSystem()
				input.Spec.MountTarget.DisplayName = "my-mount-target"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with mount target hostname_label and ip_address", func() {
				input := minimalValidFileSystem()
				input.Spec.MountTarget.HostnameLabel = "nfs-server"
				input.Spec.MountTarget.IpAddress = "10.0.1.50"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with mount target nsg_ids", func() {
				input := minimalValidFileSystem()
				input.Spec.MountTarget.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.networksecuritygroup.oc1..example"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with mount target requested_throughput", func() {
				input := minimalValidFileSystem()
				input.Spec.MountTarget.RequestedThroughput = 1024
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with mount target export set limits", func() {
				input := minimalValidFileSystem()
				input.Spec.MountTarget.MaxFsStatBytes = 1099511627776
				input.Spec.MountTarget.MaxFsStatFiles = 1000000
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple exports", func() {
				input := minimalValidFileSystem()
				input.Spec.Exports = []*OciFileSystemSpec_Export{
					{Path: "/shared-data"},
					{Path: "/app-config"},
					{Path: "/logs"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with export options", func() {
				input := minimalValidFileSystem()
				input.Spec.Exports = []*OciFileSystemSpec_Export{
					{
						Path: "/shared-data",
						ExportOptions: []*OciFileSystemSpec_ExportOption{
							{
								Source:                      "10.0.0.0/16",
								Access:                      OciFileSystemSpec_read_write,
								IdentitySquash:              OciFileSystemSpec_root_squash,
								RequirePrivilegedSourcePort: true,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple export options per export", func() {
				input := minimalValidFileSystem()
				input.Spec.Exports = []*OciFileSystemSpec_Export{
					{
						Path: "/shared-data",
						ExportOptions: []*OciFileSystemSpec_ExportOption{
							{
								Source:         "10.0.1.0/24",
								Access:         OciFileSystemSpec_read_write,
								IdentitySquash: OciFileSystemSpec_no_squash,
							},
							{
								Source:         "10.0.2.0/24",
								Access:         OciFileSystemSpec_read_only,
								IdentitySquash: OciFileSystemSpec_root_squash,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with export option anonymous uid/gid", func() {
				input := minimalValidFileSystem()
				input.Spec.Exports = []*OciFileSystemSpec_Export{
					{
						Path: "/shared-data",
						ExportOptions: []*OciFileSystemSpec_ExportOption{
							{
								Source:                   "0.0.0.0/0",
								Access:                   OciFileSystemSpec_read_only,
								IdentitySquash:           OciFileSystemSpec_all_squash,
								IsAnonymousAccessAllowed: true,
								AnonymousUid:             65534,
								AnonymousGid:             65534,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidFileSystem()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidFileSystem()
				input.Spec.DisplayName = "production-nfs"
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				input.Spec.FilesystemSnapshotPolicyId = newStringValueOrRef("ocid1.fsspolicy.oc1..example")
				input.Spec.MountTarget.DisplayName = "prod-mount-target"
				input.Spec.MountTarget.HostnameLabel = "nfs-prod"
				input.Spec.MountTarget.IpAddress = "10.0.1.100"
				input.Spec.MountTarget.NsgIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.nsg.oc1..example"),
				}
				input.Spec.MountTarget.RequestedThroughput = 2048
				input.Spec.MountTarget.MaxFsStatBytes = 10995116277760
				input.Spec.MountTarget.MaxFsStatFiles = 10000000
				input.Spec.Exports = []*OciFileSystemSpec_Export{
					{
						Path: "/shared-data",
						ExportOptions: []*OciFileSystemSpec_ExportOption{
							{
								Source:                      "10.0.0.0/16",
								Access:                      OciFileSystemSpec_read_write,
								IdentitySquash:              OciFileSystemSpec_root_squash,
								RequirePrivilegedSourcePort: true,
								AnonymousUid:                65534,
								AnonymousGid:                65534,
							},
						},
					},
					{
						Path: "/read-only-archive",
						ExportOptions: []*OciFileSystemSpec_ExportOption{
							{
								Source:         "0.0.0.0/0",
								Access:         OciFileSystemSpec_read_only,
								IdentitySquash: OciFileSystemSpec_all_squash,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_file_storage_file_system", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidFileSystem()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidFileSystem()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidFileSystem()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciFileSystem{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciFileSystem",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-fs"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidFileSystem()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when availability_domain is empty", func() {
				input := minimalValidFileSystem()
				input.Spec.AvailabilityDomain = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when mount_target is missing", func() {
				input := minimalValidFileSystem()
				input.Spec.MountTarget = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when mount_target subnet_id is missing", func() {
				input := minimalValidFileSystem()
				input.Spec.MountTarget.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when exports is empty", func() {
				input := minimalValidFileSystem()
				input.Spec.Exports = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when exports list is explicitly empty", func() {
				input := minimalValidFileSystem()
				input.Spec.Exports = []*OciFileSystemSpec_Export{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when export path is empty", func() {
				input := minimalValidFileSystem()
				input.Spec.Exports = []*OciFileSystemSpec_Export{
					{Path: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when export path does not start with slash", func() {
				input := minimalValidFileSystem()
				input.Spec.Exports = []*OciFileSystemSpec_Export{
					{Path: "shared-data"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when export option source is empty", func() {
				input := minimalValidFileSystem()
				input.Spec.Exports = []*OciFileSystemSpec_Export{
					{
						Path: "/shared-data",
						ExportOptions: []*OciFileSystemSpec_ExportOption{
							{Source: ""},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
