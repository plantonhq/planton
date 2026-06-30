package module

import (
	"github.com/pkg/errors"
	azureapplicationinsightsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azureapplicationinsights/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appinsights"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureapplicationinsightsv1.AzureApplicationInsightsStackInput) error {
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

	spec := locals.AzureApplicationInsights.Spec

	// Create the Application Insights resource
	insights, err := appinsights.NewInsights(ctx,
		spec.Name,
		&appinsights.InsightsArgs{
			Name:               pulumi.String(spec.Name),
			Location:           pulumi.String(spec.Region),
			ResourceGroupName:  pulumi.String(locals.ResourceGroupName),
			ApplicationType:    pulumi.String(spec.GetApplicationType()),
			WorkspaceId:        pulumi.String(locals.WorkspaceId),
			RetentionInDays:    pulumi.Int(int(spec.GetRetentionInDays())),
			DailyDataCapInGb:   pulumi.Float64(spec.GetDailyDataCapInGb()),
			SamplingPercentage: pulumi.Float64(spec.GetSamplingPercentage()),
			Tags:               pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Application Insights %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpAppInsightsId, insights.ID())
	ctx.Export(OpInstrumentationKey, insights.InstrumentationKey)
	ctx.Export(OpConnectionString, insights.ConnectionString)
	ctx.Export(OpAppId, insights.AppId)

	return nil
}
