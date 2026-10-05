package module

import (
	"strings"

	"github.com/pkg/errors"
	azurefederatedidentitycredentialv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefederatedidentitycredential/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// DefaultAudience is the audience Azure AD's token-exchange endpoint expects
// and the value every standard client (azure-sdk, azure/login, the AKS
// workload-identity webhook) requests. Applied when the spec leaves audience
// unset, mirroring the spec's declared default.
const DefaultAudience = "api://AzureADTokenExchange"

type Locals struct {
	AzureFederatedIdentityCredential *azurefederatedidentitycredentialv1alpha1.AzureFederatedIdentityCredential

	// UserAssignedIdentityId is a StringValueOrRef field; the platform
	// middleware resolves valueFrom references before IaC modules run, so
	// GetValue() always returns the resolved literal ARM ID.
	UserAssignedIdentityId string

	// Audience carries the spec value or the token-exchange default.
	Audience string
}

// Note: federated identity credentials carry no tags -- ARM models them as
// untagged child resources of the identity -- so the usual metadata-derived
// tag map is intentionally absent from these locals.
func initializeLocals(ctx *pulumi.Context, iacInput *azurefederatedidentitycredentialv1alpha1.AzureFederatedIdentityCredentialIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFederatedIdentityCredential = iacInput.Target
	spec := iacInput.Target.Spec

	locals.UserAssignedIdentityId = spec.UserAssignedIdentity.GetValue()

	locals.Audience = spec.GetAudience()
	if locals.Audience == "" {
		locals.Audience = DefaultAudience
	}

	return locals
}

// resourceGroupNameFromIdentityId extracts the resource-group name embedded
// in a user-assigned identity's ARM ID
// (/subscriptions/{sub}/resourceGroups/{rg}/providers/...). The provider SDK
// requires the resource group as its own argument even though the parent ID
// already carries it (azurerm derives it the same way internally), so the
// module parses rather than asking the user to restate derivable state.
func resourceGroupNameFromIdentityId(identityId string) (string, error) {
	segments := strings.Split(strings.Trim(identityId, "/"), "/")
	for i := 0; i < len(segments)-1; i++ {
		if strings.EqualFold(segments[i], "resourceGroups") {
			return segments[i+1], nil
		}
	}
	return "", errors.Errorf(
		"user_assigned_identity %q is not a full user-assigned-identity ARM ID "+
			"(expected /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.ManagedIdentity/userAssignedIdentities/{name})",
		identityId)
}
