package module

import (
	openfgarelationshiptuplev1 "github.com/plantonhq/planton/apis/dev/planton/provider/openfga/openfgarelationshiptuple/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is a pass-through placeholder for OpenFGA Relationship Tuple.
//
// IMPORTANT: OpenFGA does not have a Pulumi provider. This module exists only
// to maintain consistency with the Planton deployment component structure.
// It does not create any resources.
//
// To deploy OpenFGA resources, use Terraform/Tofu as the provisioner:
//
//	planton apply --manifest openfga-relationship-tuple.yaml --provisioner tofu
//
// Reference: https://github.com/openfga/terraform-provider-openfga
func Resources(ctx *pulumi.Context, stackInput *openfgarelationshiptuplev1.OpenFgaRelationshipTupleStackInput) error {
	// Log that this is a pass-through module
	ctx.Log.Warn("OpenFGA does not have a Pulumi provider. This module is a pass-through placeholder.", nil)
	ctx.Log.Warn("Use Terraform/Tofu as the provisioner to deploy OpenFGA resources.", nil)
	ctx.Log.Info("To deploy: planton apply --manifest <manifest.yaml> --provisioner tofu", nil)

	// Export empty outputs to indicate no resources were created
	ctx.Export("notice", pulumi.String("OpenFGA Relationship Tuple was not created. No Pulumi provider available. Use Terraform/Tofu provisioner."))

	return nil
}
