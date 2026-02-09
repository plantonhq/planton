package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack"
	"github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/identity"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applicationCredential provisions the OpenStack application credential and exports outputs.
func applicationCredential(
	ctx *pulumi.Context,
	locals *Locals,
	openstackProvider *openstack.Provider,
) error {
	spec := locals.OpenStackApplicationCredential.Spec
	resourceName := locals.OpenStackApplicationCredential.Metadata.Name

	appCredArgs := &identity.ApplicationCredentialArgs{
		Name: pulumi.String(resourceName),
	}

	// Set description if provided.
	if spec.Description != "" {
		appCredArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Set unrestricted flag. Middleware guarantees the default (false) is applied.
	appCredArgs.Unrestricted = pulumi.BoolPtr(spec.GetUnrestricted())

	// Set user-provided secret if specified.
	if spec.Secret != "" {
		appCredArgs.Secret = pulumi.StringPtr(spec.Secret)
	}

	// Set roles if provided.
	if len(spec.Roles) > 0 {
		roles := pulumi.StringArray{}
		for _, role := range spec.Roles {
			roles = append(roles, pulumi.String(role))
		}
		appCredArgs.Roles = roles
	}

	// Set access rules if provided.
	if len(spec.AccessRules) > 0 {
		accessRules := identity.ApplicationCredentialAccessRuleArray{}
		for _, rule := range spec.AccessRules {
			accessRules = append(accessRules, &identity.ApplicationCredentialAccessRuleArgs{
				Path:    pulumi.String(rule.Path),
				Method:  pulumi.String(rule.Method),
				Service: pulumi.String(rule.Service),
			})
		}
		appCredArgs.AccessRules = accessRules
	}

	// Set expiration if provided.
	if spec.ExpiresAt != "" {
		appCredArgs.ExpiresAt = pulumi.StringPtr(spec.ExpiresAt)
	}

	// Set region override if provided.
	if spec.Region != "" {
		appCredArgs.Region = pulumi.StringPtr(spec.Region)
	}

	createdAppCred, err := identity.NewApplicationCredential(
		ctx,
		strings.ToLower(resourceName),
		appCredArgs,
		pulumi.Provider(openstackProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create openstack application credential")
	}

	// Export required outputs (matching stack_outputs.proto fields).
	ctx.Export(OpId, createdAppCred.ID())
	ctx.Export(OpName, createdAppCred.Name)
	ctx.Export(OpSecret, createdAppCred.Secret)
	ctx.Export(OpProjectId, createdAppCred.ProjectId)
	ctx.Export(OpRegion, createdAppCred.Region)

	return nil
}
