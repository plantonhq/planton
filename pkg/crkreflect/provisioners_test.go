package crkreflect

import (
	"testing"

	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// kind_meta.provisioners is compile-time data every engine-picking surface acts on, so every
// declared value must name a real engine, once.
func TestDeclaredProvisionersAreRealAndDistinct(t *testing.T) {
	for _, kind := range KindsList() {
		meta, err := KindMeta(kind)
		if err != nil {
			continue
		}
		seen := map[string]bool{}
		for _, name := range meta.GetProvisioners() {
			if seen[name] {
				t.Errorf("%s declares provisioner %q twice", kind, name)
			}
			seen[name] = true
		}
		if _, err := Provisioners(kind); err != nil {
			t.Errorf("%v", err)
		}
	}
}

// OpenFGA publishes a Terraform provider and no Pulumi provider: its kinds run on OpenTofu and
// Terraform and on nothing else.
func TestOpenFgaKindsRunOnOpenTofuAndTerraformOnly(t *testing.T) {
	for _, kind := range []cloudresourcekind.CloudResourceKind{
		cloudresourcekind.CloudResourceKind_OpenFgaStore,
		cloudresourcekind.CloudResourceKind_OpenFgaAuthorizationModel,
		cloudresourcekind.CloudResourceKind_OpenFgaRelationshipTuple,
	} {
		for provisioner, want := range map[shared.IacProvisioner]bool{
			shared.IacProvisioner_tofu:      true,
			shared.IacProvisioner_terraform: true,
			shared.IacProvisioner_pulumi:    false,
		} {
			got, err := RunsOn(kind, provisioner)
			if err != nil {
				t.Fatalf("%s: %v", kind, err)
			}
			if got != want {
				t.Errorf("RunsOn(%s, %s) = %v, want %v", kind, provisioner, got, want)
			}
		}
	}
}

// A kind that declares nothing runs on every engine: the default must never narrow.
func TestUndeclaredKindRunsOnEveryEngine(t *testing.T) {
	for _, provisioner := range []shared.IacProvisioner{shared.IacProvisioner_tofu, shared.IacProvisioner_terraform, shared.IacProvisioner_pulumi} {
		got, err := RunsOn(cloudresourcekind.CloudResourceKind_AwsS3Bucket, provisioner)
		if err != nil || !got {
			t.Errorf("RunsOn(AwsS3Bucket, %s) = %v, %v; want true, nil", provisioner, got, err)
		}
	}
}
