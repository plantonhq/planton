package module

import (
	"github.com/pkg/errors"
	azureserviceplanv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureserviceplan/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureserviceplanv1.AzureServicePlanStackInput) error {
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

	spec := locals.AzureServicePlan.Spec

	// Resolve os_type with default
	osType := spec.GetOsType()
	if osType == "" {
		osType = "Linux"
	}

	// Build Service Plan arguments
	servicePlanArgs := &appservice.ServicePlanArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		OsType:            pulumi.String(osType),
		SkuName:           pulumi.String(spec.SkuName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Set worker count if specified
	if spec.WorkerCount != nil {
		servicePlanArgs.WorkerCount = pulumi.Int(int(spec.GetWorkerCount()))
	}

	// Set zone balancing if specified
	if spec.ZoneBalancingEnabled != nil {
		servicePlanArgs.ZoneBalancingEnabled = pulumi.Bool(spec.GetZoneBalancingEnabled())
	}

	// Set per-site scaling if specified
	if spec.PerSiteScalingEnabled != nil {
		servicePlanArgs.PerSiteScalingEnabled = pulumi.Bool(spec.GetPerSiteScalingEnabled())
	}

	// Set maximum elastic worker count if specified (for EP* SKUs)
	if spec.MaximumElasticWorkerCount != nil {
		servicePlanArgs.MaximumElasticWorkerCount = pulumi.Int(int(spec.GetMaximumElasticWorkerCount()))
	}

	// Create the Service Plan
	servicePlan, err := appservice.NewServicePlan(ctx,
		spec.Name,
		servicePlanArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Service Plan %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpPlanId, servicePlan.ID())
	ctx.Export(OpPlanName, servicePlan.Name)
	ctx.Export(OpOsType, pulumi.String(osType))
	ctx.Export(OpSkuName, pulumi.String(spec.SkuName))

	return nil
}
