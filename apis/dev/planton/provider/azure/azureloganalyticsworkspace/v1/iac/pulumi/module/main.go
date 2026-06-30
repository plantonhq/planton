package module

import (
	"github.com/pkg/errors"
	azureloganalyticsworkspacev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azureloganalyticsworkspace/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/operationalinsights"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureloganalyticsworkspacev1.AzureLogAnalyticsWorkspaceStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureLogAnalyticsWorkspace.Spec

	// Build workspace arguments
	workspaceArgs := &operationalinsights.AnalyticsWorkspaceArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Sku:               pulumi.String(spec.GetSku()),
		RetentionInDays:   pulumi.Int(int(spec.GetRetentionInDays())),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Set daily quota if specified and not unlimited (-1)
	dailyQuota := spec.GetDailyQuotaGb()
	if dailyQuota >= 0 {
		workspaceArgs.DailyQuotaGb = pulumi.Float64(dailyQuota)
	}

	// Create the Log Analytics Workspace
	workspace, err := operationalinsights.NewAnalyticsWorkspace(ctx,
		spec.Name,
		workspaceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Log Analytics Workspace %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpWorkspaceId, workspace.ID())
	ctx.Export(OpWorkspaceName, workspace.Name)
	ctx.Export(OpPrimarySharedKey, workspace.PrimarySharedKey)
	ctx.Export(OpSecondarySharedKey, workspace.SecondarySharedKey)

	return nil
}
