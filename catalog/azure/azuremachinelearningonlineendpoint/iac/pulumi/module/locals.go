package module

import (
	"fmt"
	"strings"

	azuremachinelearningonlineendpointv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningonlineendpoint/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMachineLearningOnlineEndpoint *azuremachinelearningonlineendpointv1alpha1.AzureMachineLearningOnlineEndpoint

	// WorkspaceId is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal ARM ID.
	WorkspaceId string

	// ResourceGroupName / WorkspaceName are parsed from WorkspaceId:
	// azure-native addresses ARM children by their ancestor NAMES where
	// the raw ARM layer takes the parent's full ID -- the two engines
	// consume the same spec reference either way.
	ResourceGroupName string
	WorkspaceName     string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

// identityTypeWire maps the spec's identity flavors to ARM's own
// ManagedServiceIdentityType literals (common-types: the dual value has
// NO space) -- the same ARM identity types the Terraform module's azapi
// values normalize to on the wire.
var identityTypeWire = map[azuremachinelearningonlineendpointv1alpha1.AzureMachineLearningOnlineEndpointIdentityType]string{
	azuremachinelearningonlineendpointv1alpha1.AzureMachineLearningOnlineEndpointIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azuremachinelearningonlineendpointv1alpha1.AzureMachineLearningOnlineEndpointIdentityType_USER_ASSIGNED:            "UserAssigned",
	azuremachinelearningonlineendpointv1alpha1.AzureMachineLearningOnlineEndpointIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned,UserAssigned",
}

// parseWorkspaceId splits a workspace ARM ID into its resource-group and
// workspace names. The ID shape is fixed by ARM:
// /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.MachineLearningServices/workspaces/{ws}
func parseWorkspaceId(workspaceId string) (resourceGroupName, workspaceName string, err error) {
	segments := strings.Split(strings.Trim(workspaceId, "/"), "/")
	for i := 0; i+1 < len(segments); i += 2 {
		switch strings.ToLower(segments[i]) {
		case "resourcegroups":
			resourceGroupName = segments[i+1]
		case "workspaces":
			workspaceName = segments[i+1]
		}
	}
	if resourceGroupName == "" || workspaceName == "" {
		return "", "", fmt.Errorf("workspace id %q does not carry resourceGroups and workspaces segments", workspaceId)
	}
	return resourceGroupName, workspaceName, nil
}

func initializeLocals(ctx *pulumi.Context, stackInput *azuremachinelearningonlineendpointv1alpha1.AzureMachineLearningOnlineEndpointStackInput) *Locals {
	locals := &Locals{}

	locals.AzureMachineLearningOnlineEndpoint = stackInput.Target
	target := stackInput.Target

	locals.WorkspaceId = target.Spec.WorkspaceId.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureMachineLearningOnlineEndpoint.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
