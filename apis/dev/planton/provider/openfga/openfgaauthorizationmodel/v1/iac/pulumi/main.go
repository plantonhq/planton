package main

import (
	"github.com/pkg/errors"
	openfgaauthorizationmodelv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openfga/openfgaauthorizationmodel/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/openfga/openfgaauthorizationmodel/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// main is the entry point for the OpenFGA Authorization Model Pulumi module.
//
// IMPORTANT: OpenFGA does not have a Pulumi provider. This module is a
// pass-through placeholder that does not create any resources.
//
// To deploy OpenFGA resources, use Terraform/Tofu as the provisioner:
//
//	planton apply --manifest openfga-authorization-model.yaml --provisioner tofu
func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &openfgaauthorizationmodelv1.OpenFgaAuthorizationModelStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
