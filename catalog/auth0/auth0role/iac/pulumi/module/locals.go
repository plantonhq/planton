package module

import (
	auth0rolev1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0role/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals contains computed values for the Auth0 Role deployment.
type Locals struct {
	Auth0Role *auth0rolev1alpha1.Auth0Role

	// ResourceName is the stable Pulumi logical resource name (from metadata.name).
	ResourceName string
	// RoleName is the human-readable Auth0 role name (spec.name, or metadata.name when blank).
	RoleName    string
	Description string

	Permissions []*auth0rolev1alpha1.Auth0RolePermission
}

// initializeLocals creates and populates the Locals struct from IaC input.
func initializeLocals(ctx *pulumi.Context, iacInput *auth0rolev1alpha1.Auth0RoleIacInput) *Locals {
	locals := &Locals{}

	locals.Auth0Role = iacInput.Target

	spec := iacInput.Target.Spec
	metadata := iacInput.Target.Metadata

	locals.ResourceName = metadata.Name

	// Role name defaults to metadata.name when spec.name is not provided.
	if spec.Name != "" {
		locals.RoleName = spec.Name
	} else {
		locals.RoleName = metadata.Name
	}

	locals.Description = spec.Description
	locals.Permissions = spec.Permissions

	return locals
}
