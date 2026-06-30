package module

import (
	openfgastorev1 "github.com/plantonhq/planton/apis/dev/planton/provider/openfga/openfgastore/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is a pass-through placeholder for OpenFGA Store.
//
// IMPORTANT: OpenFGA does not have a Pulumi provider. This module exists only
// to maintain consistency with the Planton deployment component structure.
// It does not create any resources.
//
// To deploy OpenFGA resources, use Terraform/Tofu as the provisioner:
//
//	planton apply --manifest openfga-store.yaml --provisioner tofu
//
// Reference: https://github.com/openfga/terraform-provider-openfga
func Resources(ctx *pulumi.Context, stackInput *openfgastorev1.OpenFgaStoreStackInput) error {
	// Log that this is a pass-through module
	ctx.Log.Warn("OpenFGA does not have a Pulumi provider. This module is a pass-through placeholder.", nil)
	ctx.Log.Warn("Use Terraform/Tofu as the provisioner to deploy OpenFGA resources.", nil)
	ctx.Log.Info("To deploy: planton apply --manifest <manifest.yaml> --provisioner tofu", nil)

	// Export empty outputs to indicate no resources were created
	ctx.Export("notice", pulumi.String("OpenFGA Store was not created. No Pulumi provider available. Use Terraform/Tofu provisioner."))

	return nil
}
