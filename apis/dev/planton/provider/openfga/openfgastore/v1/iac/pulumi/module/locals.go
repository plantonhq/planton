package module

import (
	openfgastorev1 "github.com/plantonhq/planton/apis/dev/planton/provider/openfga/openfgastore/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals contains computed values derived from the stack input.
// This is a placeholder for consistency - OpenFGA has no Pulumi provider.
type Locals struct {
	// StoreName is the name of the OpenFGA store from spec.
	StoreName string
}

// initializeLocals creates and returns the computed local values from stack input.
// This is a placeholder for consistency - OpenFGA has no Pulumi provider.
func initializeLocals(_ *pulumi.Context, stackInput *openfgastorev1.OpenFgaStoreStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	return &Locals{
		StoreName: spec.Name,
	}
}
