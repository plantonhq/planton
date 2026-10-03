package module

import (
	azureroleassignmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureroleassignment/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureRoleAssignment *azureroleassignmentv1alpha1.AzureRoleAssignment

	// Scope and PrincipalId are StringValueOrRef fields; the platform
	// middleware resolves valueFrom references before IaC modules run, so
	// GetValue() always returns the resolved literal.
	Scope       string
	PrincipalId string

	// PrincipalType is the ARM string for the spec's enum ("User", "Group",
	// "ServicePrincipal"), or empty when unspecified so Azure infers the type
	// from the directory object.
	PrincipalType string
}

// Note: role assignments carry no tags -- Microsoft.Authorization resources do
// not support ARM tags, so the usual metadata-derived tag map is intentionally
// absent from these locals.
func initializeLocals(ctx *pulumi.Context, iacInput *azureroleassignmentv1alpha1.AzureRoleAssignmentIacInput) *Locals {
	locals := &Locals{}

	locals.AzureRoleAssignment = iacInput.Target
	spec := iacInput.Target.Spec

	locals.Scope = spec.Scope.GetValue()
	locals.PrincipalId = spec.PrincipalId.GetValue()
	locals.PrincipalType = principalTypeToArm(spec.PrincipalType)

	return locals
}

// principalTypeToArm maps the spec enum onto ARM's PrincipalType strings.
// The unspecified value maps to "" so the field is omitted from the request,
// letting Azure infer the type -- passing an explicit (wrong) type is what
// breaks assignments under ABAC-constrained creators, so we never guess.
func principalTypeToArm(t azureroleassignmentv1alpha1.AzureRoleAssignmentPrincipalType) string {
	switch t {
	case azureroleassignmentv1alpha1.AzureRoleAssignmentPrincipalType_SERVICE_PRINCIPAL:
		return "ServicePrincipal"
	case azureroleassignmentv1alpha1.AzureRoleAssignmentPrincipalType_USER:
		return "User"
	case azureroleassignmentv1alpha1.AzureRoleAssignmentPrincipalType_GROUP:
		return "Group"
	default:
		return ""
	}
}
