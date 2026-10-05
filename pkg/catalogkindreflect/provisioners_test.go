package catalogkindreflect

import (
	"testing"

	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
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
	for _, kind := range []catalogkind.CatalogKind{
		catalogkind.CatalogKind_OpenFgaStore,
		catalogkind.CatalogKind_OpenFgaAuthorizationModel,
		catalogkind.CatalogKind_OpenFgaRelationshipTuple,
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

// Stripe publishes no Pulumi provider and its kinds are proven on OpenTofu alone: every Stripe
// kind runs on OpenTofu and on nothing else, not even the Terraform binary.
func TestStripeKindsRunOnOpenTofuOnly(t *testing.T) {
	stripeKinds := 0
	for _, kind := range KindsList() {
		if GetProvider(kind) != catalogkind.CatalogProvider_stripe {
			continue
		}
		stripeKinds++
		declared, err := Provisioners(kind)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if len(declared) != 1 || declared[0] != shared.IacProvisioner_tofu {
			t.Errorf("%s declares %v, want exactly [tofu]", kind, declared)
		}
	}
	if stripeKinds == 0 {
		t.Fatal("no Stripe kinds registered")
	}
}

// A kind that declares nothing runs on every engine: the default must never narrow.
func TestUndeclaredKindRunsOnEveryEngine(t *testing.T) {
	for _, provisioner := range []shared.IacProvisioner{shared.IacProvisioner_tofu, shared.IacProvisioner_terraform, shared.IacProvisioner_pulumi} {
		got, err := RunsOn(catalogkind.CatalogKind_AwsS3Bucket, provisioner)
		if err != nil || !got {
			t.Errorf("RunsOn(AwsS3Bucket, %s) = %v, %v; want true, nil", provisioner, got, err)
		}
	}
}
