package awsfsxontapstoragevirtualmachinev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	shared "github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAwsFsxOntapStorageVirtualMachineSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsFsxOntapStorageVirtualMachineSpec Validation Suite")
}

func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func stringPtr(s string) *string {
	return &s
}

var _ = ginkgo.Describe("AwsFsxOntapStorageVirtualMachineSpec validations", func() {
	var spec *AwsFsxOntapStorageVirtualMachineSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsFsxOntapStorageVirtualMachineSpec{
			FileSystemId: strRef("fs-0123456789abcdef0"),
			Name:         "svm_default",
		}
	})

	// -------------------------------------------------------------------------
	// Happy path — valid configurations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec (NFS-only SVM)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts UNIX root_volume_security_style", func() {
		spec.RootVolumeSecurityStyle = stringPtr("UNIX")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts NTFS root_volume_security_style", func() {
		spec.RootVolumeSecurityStyle = stringPtr("NTFS")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts MIXED root_volume_security_style", func() {
		spec.RootVolumeSecurityStyle = stringPtr("MIXED")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts svm_admin_password within valid range", func() {
		spec.SvmAdminPassword = "MyP@ss12"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts svm_admin_password at exactly 8 characters", func() {
		spec.SvmAdminPassword = "Exact8!!"
		gomega.Expect(len(spec.SvmAdminPassword)).To(gomega.Equal(8))
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts svm_admin_password at 50 characters", func() {
		spec.SvmAdminPassword = "Abcdefgh12345678901234567890123456789012345678901!"
		gomega.Expect(len(spec.SvmAdminPassword)).To(gomega.Equal(50))
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a name with underscores", func() {
		spec.Name = "svm_production_01"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a name at 47 characters", func() {
		spec.Name = "svm_abcdefghijklmnopqrstuvwxyz_0123456789_abcde"
		gomega.Expect(len(spec.Name)).To(gomega.Equal(47))
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full SMB configuration with Active Directory", func() {
		spec.RootVolumeSecurityStyle = stringPtr("NTFS")
		spec.SvmAdminPassword = "VsAdmin2024!"
		spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
			NetbiosName:                         "PRODSVM",
			DomainName:                          "corp.example.com",
			DnsIps:                              []string{"10.0.0.1", "10.0.0.2"},
			Username:                            "svc_fsx_join",
			Password:                            "ADJoinP@ssw0rd!",
			OrganizationalUnitDistinguishedName: "OU=FSx,DC=corp,DC=example,DC=com",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts AD configuration with all optional fields", func() {
		spec.RootVolumeSecurityStyle = stringPtr("MIXED")
		spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
			NetbiosName:                         "MYSVM",
			DomainName:                          "ad.internal.com",
			DnsIps:                              []string{"10.0.1.1", "10.0.1.2", "10.0.1.3"},
			Username:                            "admin",
			Password:                            "SecureP@ss1",
			FileSystemAdministratorsGroup:       stringPtr("FSx Admins"),
			OrganizationalUnitDistinguishedName: "OU=Servers,OU=IT,DC=ad,DC=internal,DC=com",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts AD configuration with minimal required fields", func() {
		spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
			DomainName: "corp.example.com",
			DnsIps:     []string{"10.0.0.1"},
			Username:   "admin",
			Password:   "pass123",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts valueFrom reference for file_system_id", func() {
		spec.FileSystemId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{
					Kind:      cloudresourcekind.CloudResourceKind_AwsFsxOntapFileSystem,
					Name:      "my-ontap-fs",
					FieldPath: "status.outputs.file_system_id",
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level validation failures
	// -------------------------------------------------------------------------

	ginkgo.Context("field-level validations", func() {
		ginkgo.It("fails when file_system_id is missing", func() {
			spec.FileSystemId = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when name is empty", func() {
			spec.Name = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when name exceeds 47 characters", func() {
			spec.Name = "svm_abcdefghijklmnopqrstuvwxyz_0123456789_abcdef"
			gomega.Expect(len(spec.Name)).To(gomega.Equal(48))
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when AD domain_name is empty", func() {
			spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
				DomainName: "",
				DnsIps:     []string{"10.0.0.1"},
				Username:   "admin",
				Password:   "pass123",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when AD dns_ips is empty", func() {
			spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
				DomainName: "corp.example.com",
				DnsIps:     []string{},
				Username:   "admin",
				Password:   "pass123",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when AD dns_ips exceeds 3", func() {
			spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
				DomainName: "corp.example.com",
				DnsIps:     []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4"},
				Username:   "admin",
				Password:   "pass123",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when AD username is empty", func() {
			spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
				DomainName: "corp.example.com",
				DnsIps:     []string{"10.0.0.1"},
				Username:   "",
				Password:   "pass123",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when AD password is empty", func() {
			spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
				DomainName: "corp.example.com",
				DnsIps:     []string{"10.0.0.1"},
				Username:   "admin",
				Password:   "",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: security_style_valid
	// -------------------------------------------------------------------------

	ginkgo.Context("security_style_valid", func() {
		ginkgo.It("fails when root_volume_security_style is invalid", func() {
			spec.RootVolumeSecurityStyle = stringPtr("POSIX")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when root_volume_security_style is lowercase", func() {
			spec.RootVolumeSecurityStyle = stringPtr("unix")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when root_volume_security_style is a partial match", func() {
			spec.RootVolumeSecurityStyle = stringPtr("UNI")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: admin_password_length
	// -------------------------------------------------------------------------

	ginkgo.Context("admin_password_length", func() {
		ginkgo.It("fails when svm_admin_password is too short (7 chars)", func() {
			spec.SvmAdminPassword = "Short1!"
			gomega.Expect(len(spec.SvmAdminPassword)).To(gomega.Equal(7))
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when svm_admin_password is too long (51 chars)", func() {
			spec.SvmAdminPassword = "Abcdefgh123456789012345678901234567890123456789012!"
			gomega.Expect(len(spec.SvmAdminPassword)).To(gomega.Equal(51))
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts empty svm_admin_password (optional)", func() {
			spec.SvmAdminPassword = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: name_format
	// -------------------------------------------------------------------------

	ginkgo.Context("name_format", func() {
		ginkgo.It("fails when name contains hyphens", func() {
			spec.Name = "svm-production"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when name contains spaces", func() {
			spec.Name = "svm production"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when name contains dots", func() {
			spec.Name = "svm.production"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when name contains at-sign", func() {
			spec.Name = "svm@prod"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts name with only digits", func() {
			spec.Name = "12345"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts name with mixed case and underscores", func() {
			spec.Name = "SVM_Production_01"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// CEL: netbios_name_length (nested message)
	// -------------------------------------------------------------------------

	ginkgo.Context("netbios_name_length", func() {
		ginkgo.It("fails when netbios_name exceeds 15 characters", func() {
			spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
				NetbiosName: "THISNAMEIS16CHAR",
				DomainName:  "corp.example.com",
				DnsIps:      []string{"10.0.0.1"},
				Username:    "admin",
				Password:    "pass123",
			}
			gomega.Expect(len(spec.ActiveDirectoryConfiguration.NetbiosName)).To(gomega.Equal(16))
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts netbios_name at exactly 15 characters", func() {
			spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
				NetbiosName: "EXACTLY15CHARS!",
				DomainName:  "corp.example.com",
				DnsIps:      []string{"10.0.0.1"},
				Username:    "admin",
				Password:    "pass123",
			}
			gomega.Expect(len(spec.ActiveDirectoryConfiguration.NetbiosName)).To(gomega.Equal(15))
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts empty netbios_name (optional)", func() {
			spec.ActiveDirectoryConfiguration = &AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration{
				NetbiosName: "",
				DomainName:  "corp.example.com",
				DnsIps:      []string{"10.0.0.1"},
				Username:    "admin",
				Password:    "pass123",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	// -------------------------------------------------------------------------
	// API envelope validations
	// -------------------------------------------------------------------------

	ginkgo.Context("API envelope", func() {
		ginkgo.It("validates a complete API resource", func() {
			resource := &AwsFsxOntapStorageVirtualMachine{
				ApiVersion: "aws.openmcf.org/v1",
				Kind:       "AwsFsxOntapStorageVirtualMachine",
				Metadata: &shared.CloudResourceMetadata{
					Name: "my-ontap-svm",
					Id:   "awsfxosvm-test-123",
					Org:  "test-org",
					Env:  "dev",
				},
				Spec: spec,
			}
			err := protovalidate.Validate(resource)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails with wrong api_version", func() {
			resource := &AwsFsxOntapStorageVirtualMachine{
				ApiVersion: "wrong/v1",
				Kind:       "AwsFsxOntapStorageVirtualMachine",
				Metadata: &shared.CloudResourceMetadata{
					Name: "my-ontap-svm",
					Id:   "awsfxosvm-test-123",
					Org:  "test-org",
					Env:  "dev",
				},
				Spec: spec,
			}
			err := protovalidate.Validate(resource)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails with wrong kind", func() {
			resource := &AwsFsxOntapStorageVirtualMachine{
				ApiVersion: "aws.openmcf.org/v1",
				Kind:       "WrongKind",
				Metadata: &shared.CloudResourceMetadata{
					Name: "my-ontap-svm",
					Id:   "awsfxosvm-test-123",
					Org:  "test-org",
					Env:  "dev",
				},
				Spec: spec,
			}
			err := protovalidate.Validate(resource)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
