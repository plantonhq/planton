package module

import (
	"fmt"

	openfgarelationshiptuplev1 "github.com/plantonhq/planton/apis/dev/planton/provider/openfga/openfgarelationshiptuple/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals contains computed values derived from the stack input.
// This is a placeholder for consistency - OpenFGA has no Pulumi provider.
type Locals struct {
	// StoreId is the ID of the OpenFGA store this tuple belongs to.
	StoreId string
	// AuthorizationModelId is the optional ID of the authorization model.
	AuthorizationModelId string
	// User is the subject of the relationship tuple (type:id or type:id#relation).
	User string
	// Relation is the relationship type.
	Relation string
	// Object is the resource being accessed (type:id).
	Object string
}

// initializeLocals creates and returns the computed local values from stack input.
// This is a placeholder for consistency - OpenFGA has no Pulumi provider.
func initializeLocals(_ *pulumi.Context, stackInput *openfgarelationshiptuplev1.OpenFgaRelationshipTupleStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec

	// Extract store_id from StringValueOrRef
	storeId := ""
	if spec.StoreId != nil {
		storeId = spec.StoreId.GetValue()
	}

	// Extract authorization_model_id from StringValueOrRef (optional)
	authorizationModelId := ""
	if spec.AuthorizationModelId != nil {
		authorizationModelId = spec.AuthorizationModelId.GetValue()
	}

	// Construct user string from structured User message
	// Format: type:id or type:id#relation (for usersets)
	user := ""
	if spec.User != nil {
		userId := ""
		if spec.User.Id != nil {
			userId = spec.User.Id.GetValue()
		}
		if spec.User.Relation != "" {
			user = fmt.Sprintf("%s:%s#%s", spec.User.Type, userId, spec.User.Relation)
		} else {
			user = fmt.Sprintf("%s:%s", spec.User.Type, userId)
		}
	}

	// Construct object string from structured Object message
	// Format: type:id
	object := ""
	if spec.Object != nil {
		objectId := ""
		if spec.Object.Id != nil {
			objectId = spec.Object.Id.GetValue()
		}
		object = fmt.Sprintf("%s:%s", spec.Object.Type, objectId)
	}

	return &Locals{
		StoreId:              storeId,
		AuthorizationModelId: authorizationModelId,
		User:                 user,
		Relation:             spec.Relation,
		Object:               object,
	}
}
