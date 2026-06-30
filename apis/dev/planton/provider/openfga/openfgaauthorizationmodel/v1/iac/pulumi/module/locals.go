package module

import (
	openfgaauthorizationmodelv1 "github.com/plantonhq/planton/apis/dev/planton/provider/openfga/openfgaauthorizationmodel/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals contains computed values derived from the stack input.
// This is a placeholder for consistency - OpenFGA has no Pulumi provider.
type Locals struct {
	// StoreId is the ID of the OpenFGA store where the model will be created.
	StoreId string
	// ModelDsl is the authorization model definition in DSL format.
	ModelDsl string
	// ModelJson is the authorization model definition in JSON format.
	ModelJson string
}

// initializeLocals creates and returns the computed local values from stack input.
// This is a placeholder for consistency - OpenFGA has no Pulumi provider.
func initializeLocals(_ *pulumi.Context, stackInput *openfgaauthorizationmodelv1.OpenFgaAuthorizationModelStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	// Extract store_id from StringValueOrRef
	storeId := ""
	if spec.StoreId != nil {
		storeId = spec.StoreId.GetValue()
	}

	return &Locals{
		StoreId:   storeId,
		ModelDsl:  spec.ModelDsl,
		ModelJson: spec.ModelJson,
	}
}
