package digitaloceandropletv1alpha1

import (
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/plantonhq/planton/shared"
)

func TestDigitalOceanDropletSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanDropletSpec Custom Validation Tests")
}

func boolPtr(b bool) *bool { return &b }

// validMinimalSpec returns a fresh spec carrying only the required fields —
// region and vpc are deliberately absent, because both are optional (unset
// region lets DigitalOcean choose; unset vpc lands in the region's default
// VPC) — so each test mutates its own copy without cross-test bleed.
func validMinimalSpec() *DigitalOceanDropletSpec {
	return &DigitalOceanDropletSpec{
		DropletName: "test-droplet",
		Size:        "s-1vcpu-1gb",
		Image:       "ubuntu-24-04-x64",
	}
}

func wrap(spec *DigitalOceanDropletSpec) *DigitalOceanDroplet {
	return &DigitalOceanDroplet{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanDroplet",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-droplet",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("DigitalOceanDropletSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("digitalocean_droplet", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (no region, no vpc)", func() {
				err := protovalidate.Validate(wrap(validMinimalSpec()))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full surface with every optional field set", func() {
				spec := validMinimalSpec()
				spec.Region = digitalocean.DigitalOceanRegion_nyc3
				spec.Vpc = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "b5648f9e-a28a-4760-bb87-b2fad07ae295"},
				}
				spec.EnableIpv6 = true
				spec.EnableBackups = true
				spec.BackupPolicy = &DigitalOceanDropletBackupPolicy{
					Plan:    "weekly",
					Weekday: "TUE",
					Hour:    8,
				}
				spec.VolumeIds = []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "93a7a5b4-62ce-11f0-b9db-0a58ac1466b2"}},
				}
				spec.Tags = []string{"env:prod", "team_platform", "web"}
				spec.UserData = "#cloud-config\npackage_update: true\n"
				spec.Monitoring = true
				spec.SshKeys = []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "12345678"}},
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa"}},
				}
				spec.DropletAgent = boolPtr(true)
				spec.GracefulShutdown = true
				spec.ResizeDisk = boolPtr(false)
				spec.PublicNetworking = boolPtr(false)
				spec.GpuPartitionMode = "PARTITION_MODE_DPX_NPS2"
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a hostname-style droplet name with dots and uppercase", func() {
				spec := validMinimalSpec()
				spec.DropletName = "Web-1.example.com"
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a numeric custom-image or snapshot id as the image", func() {
				spec := validMinimalSpec()
				spec.Image = "187625153"
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept backups enabled without a policy (DigitalOcean defaults to daily)", func() {
				spec := validMinimalSpec()
				spec.EnableBackups = true
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a daily backup policy with only the plan set", func() {
				spec := validMinimalSpec()
				spec.EnableBackups = true
				spec.BackupPolicy = &DigitalOceanDropletBackupPolicy{Plan: "daily"}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("droplet_name", func() {

			ginkgo.It("should return a validation error when droplet_name is missing", func() {
				spec := validMinimalSpec()
				spec.DropletName = ""
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a name starting with a hyphen", func() {
				spec := validMinimalSpec()
				spec.DropletName = "-bad-name"
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a name ending with a dot", func() {
				spec := validMinimalSpec()
				spec.DropletName = "bad-name."
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a name longer than 255 characters", func() {
				spec := validMinimalSpec()
				spec.DropletName = strings.Repeat("a", 256)
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("size and image", func() {

			ginkgo.It("should return a validation error when size is missing", func() {
				spec := validMinimalSpec()
				spec.Size = ""
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an uppercase size slug", func() {
				spec := validMinimalSpec()
				spec.Size = "Standard_B2s"
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when image is missing", func() {
				spec := validMinimalSpec()
				spec.Image = ""
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an image slug with spaces", func() {
				spec := validMinimalSpec()
				spec.Image = "Ubuntu 24.04"
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("backup_policy", func() {

			ginkgo.It("should return a validation error for a backup policy without backups enabled", func() {
				spec := validMinimalSpec()
				spec.BackupPolicy = &DigitalOceanDropletBackupPolicy{Plan: "daily"}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an unknown plan", func() {
				spec := validMinimalSpec()
				spec.EnableBackups = true
				spec.BackupPolicy = &DigitalOceanDropletBackupPolicy{Plan: "monthly"}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a lowercase weekday token", func() {
				spec := validMinimalSpec()
				spec.EnableBackups = true
				spec.BackupPolicy = &DigitalOceanDropletBackupPolicy{Plan: "weekly", Weekday: "tue"}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an hour above 20", func() {
				spec := validMinimalSpec()
				spec.EnableBackups = true
				spec.BackupPolicy = &DigitalOceanDropletBackupPolicy{Plan: "weekly", Weekday: "TUE", Hour: 21}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a negative hour", func() {
				spec := validMinimalSpec()
				spec.EnableBackups = true
				spec.BackupPolicy = &DigitalOceanDropletBackupPolicy{Plan: "daily", Hour: -4}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("tags and ssh_keys", func() {

			ginkgo.It("should return a validation error for a tag with spaces", func() {
				spec := validMinimalSpec()
				spec.Tags = []string{"bad tag"}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate tags", func() {
				spec := validMinimalSpec()
				spec.Tags = []string{"web", "web"}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate ssh keys", func() {
				spec := validMinimalSpec()
				spec.SshKeys = []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "12345678"}},
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "12345678"}},
				}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an empty ssh key entry", func() {
				spec := validMinimalSpec()
				spec.SshKeys = []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: ""}},
				}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should accept a literal key beside a reference to a DigitalOceanSshKey", func() {
				spec := validMinimalSpec()
				spec.SshKeys = []*foreignkeyv1.StringValueOrRef{
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa"}},
					{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: "ops-key"}}},
				}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("gpu_partition_mode and user_data", func() {

			ginkgo.It("should return a validation error for an unknown gpu partition mode", func() {
				spec := validMinimalSpec()
				spec.GpuPartitionMode = "full"
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for user_data above 32 KiB", func() {
				spec := validMinimalSpec()
				spec.UserData = "#cloud-config\n" + strings.Repeat("a", 32769)
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
