package module

import (
	azureroledefinitionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureroledefinition/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureRoleDefinition *azureroledefinitionv1alpha1.AzureRoleDefinition

	// Scope and AssignableScopes are StringValueOrRef fields; the platform
	// middleware resolves valueFrom references before IaC modules run, so
	// GetValue() always returns the resolved literal.
	Scope            string
	AssignableScopes []string
}

// Note: role definitions carry no tags -- Microsoft.Authorization resources do
// not support ARM tags, so the usual metadata-derived tag map is intentionally
// absent from these locals.
func initializeLocals(ctx *pulumi.Context, iacInput *azureroledefinitionv1alpha1.AzureRoleDefinitionIacInput) *Locals {
	locals := &Locals{}

	locals.AzureRoleDefinition = iacInput.Target
	spec := iacInput.Target.Spec

	locals.Scope = spec.Scope.GetValue()

	for _, s := range spec.AssignableScopes {
		locals.AssignableScopes = append(locals.AssignableScopes, s.GetValue())
	}

	return locals
}
