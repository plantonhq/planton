package module

import (
	"strings"

	azureuserassignedidentityv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureuserassignedidentity/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureUserAssignedIdentity *azureuserassignedidentityv1.AzureUserAssignedIdentity
	ResourceGroupName         string
	ResolvedScopes            []string
	AzureTags                 map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureuserassignedidentityv1.AzureUserAssignedIdentityStackInput) *Locals {
	locals := &Locals{}

	locals.AzureUserAssignedIdentity = stackInput.Target
	target := stackInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware resolves
	// valueFrom references before IaC modules run, so .GetValue() always returns the
	// resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Resolve all role assignment scopes.
	// Each scope is a StringValueOrRef that may reference another resource's output.
	// By the time IaC modules run, all valueFrom references are already resolved.
	locals.ResolvedScopes = make([]string, len(target.Spec.RoleAssignments))
	for i, ra := range target.Spec.RoleAssignments {
		locals.ResolvedScopes[i] = ra.Scope.GetValue()
	}

	// Create Azure tags for resource tagging
	locals.AzureTags = map[string]string{
		"resource":      "true",
		"resource_name": target.Metadata.Name,
		"resource_kind": strings.ToLower(cloudresourcekind.CloudResourceKind_AzureUserAssignedIdentity.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags["resource_id"] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags["organization"] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags["environment"] = target.Metadata.Env
	}

	return locals
}
