package module

import (
	"github.com/pkg/errors"
	azureserviceplanv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureserviceplan/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureserviceplanv1alpha1.AzureServicePlanIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureServicePlan.Spec

	// The App Service Plan: the compute tier every Web App / Function App
	// on it shares. The SKU is deliberately NOT ForceNew -- plans scale
	// up, down, and across tiers in place (Azure rejects the few
	// impossible moves, like Consumption <-> dedicated, at apply time).
	// Name, OS type, region, and resource group ARE ForceNew, and
	// recreating the plan takes every app on it down with it.
	servicePlanArgs := &appservice.ServicePlanArgs{
		Name:              pulumi.String(spec.ServicePlanName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		OsType:            pulumi.String(locals.OsType),
		SkuName:           pulumi.String(locals.SkuName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Placing the plan inside an App Service Environment v3
	// (single-tenant compute). The spec gates this to Isolated SKUs,
	// mirroring Azure's own creation-time rule.
	if spec.AppServiceEnvironmentId != "" {
		servicePlanArgs.AppServiceEnvironmentId = pulumi.String(spec.AppServiceEnvironmentId)
	}

	// Unset lets Azure apply the SKU's default capacity (typically 1);
	// the serverless tiers (Y1/FC1/EP*) manage instance count themselves.
	if spec.WorkerCount != nil {
		servicePlanArgs.WorkerCount = pulumi.Int(int(spec.GetWorkerCount()))
	}

	// Presence-guarded proto defaults: IaC inputs never materialize
	// them, so an unset field must deploy the spec's documented default,
	// not the Go zero value.

	// Premium-plan automatic HTTP-load scaling; the ceiling comes from
	// maximum_elastic_worker_count. Both gates live in the spec,
	// mirroring the provider's own SKU checks.
	premiumAutoScale := false
	if spec.PremiumPlanAutoScaleEnabled != nil {
		premiumAutoScale = spec.GetPremiumPlanAutoScaleEnabled()
	}
	servicePlanArgs.PremiumPlanAutoScaleEnabled = pulumi.BoolPtr(premiumAutoScale)

	if spec.MaximumElasticWorkerCount != nil {
		servicePlanArgs.MaximumElasticWorkerCount = pulumi.Int(int(spec.GetMaximumElasticWorkerCount()))
	}

	// Flipping zone balancing on with fewer than 2 workers forces the
	// plan to be recreated -- keep worker_count a multiple of the
	// region's zone count (typically 3).
	zoneBalancing := false
	if spec.ZoneBalancingEnabled != nil {
		zoneBalancing = spec.GetZoneBalancingEnabled()
	}
	servicePlanArgs.ZoneBalancingEnabled = pulumi.BoolPtr(zoneBalancing)

	perSiteScaling := false
	if spec.PerSiteScalingEnabled != nil {
		perSiteScaling = spec.GetPerSiteScalingEnabled()
	}
	servicePlanArgs.PerSiteScalingEnabled = pulumi.BoolPtr(perSiteScaling)

	servicePlan, err := appservice.NewServicePlan(ctx,
		spec.ServicePlanName,
		servicePlanArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Service Plan %s", spec.ServicePlanName)
	}

	// Export outputs. kind and reserved are Azure-computed
	// attributes read back after creation (the API's own classification
	// of the plan).
	ctx.Export(OpServicePlanId, servicePlan.ID())
	ctx.Export(OpServicePlanName, servicePlan.Name)
	ctx.Export(OpOsType, servicePlan.OsType)
	ctx.Export(OpSkuName, servicePlan.SkuName)
	ctx.Export(OpKind, servicePlan.Kind)
	ctx.Export(OpReserved, servicePlan.Reserved)

	return nil
}
