package module

import (
	"github.com/pkg/errors"
	azurenetworkwatcherflowlogv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurenetworkwatcherflowlog/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurenetworkwatcherflowlogv1alpha1.AzureNetworkWatcherFlowLogIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureNetworkWatcherFlowLog.Spec

	// The provider requires an explicit value; the platform default is
	// true (a flow log exists to record).
	enabled := true
	if spec.Enabled != nil {
		enabled = *spec.Enabled
	}

	// Schema version 1 is the provider default; 2 adds flow state and
	// byte/packet counters.
	version := 1
	if spec.Version != nil {
		version = int(*spec.Version)
	}

	flowLogArgs := &network.NetworkWatcherFlowLogArgs{
		Name:               pulumi.String(spec.Name),
		NetworkWatcherName: pulumi.String(locals.NetworkWatcherName),
		ResourceGroupName:  pulumi.String(locals.NetworkWatcherResourceGroup),
		Location:           pulumi.String(spec.Region),
		TargetResourceId:   pulumi.String(spec.TargetResourceId.GetValue()),
		StorageAccountId:   pulumi.String(spec.StorageAccountId.GetValue()),
		Enabled:            pulumi.Bool(enabled),
		Version:            pulumi.Int(version),
		RetentionPolicy: &network.NetworkWatcherFlowLogRetentionPolicyArgs{
			Enabled: pulumi.Bool(spec.RetentionPolicy.Enabled),
			Days:    pulumi.Int(int(spec.RetentionPolicy.Days)),
		},
		Tags: pulumi.ToStringMap(locals.AzureTags),
	}

	// Traffic Analytics enrichment into a Log Analytics workspace:
	// workspace_id is the workspace GUID (customer id),
	// workspace_resource_id its ARM id.
	if spec.TrafficAnalytics != nil {
		trafficAnalyticsEnabled := true
		if spec.TrafficAnalytics.Enabled != nil {
			trafficAnalyticsEnabled = *spec.TrafficAnalytics.Enabled
		}
		intervalInMinutes := 60
		if spec.TrafficAnalytics.IntervalInMinutes != nil {
			intervalInMinutes = int(*spec.TrafficAnalytics.IntervalInMinutes)
		}
		flowLogArgs.TrafficAnalytics = &network.NetworkWatcherFlowLogTrafficAnalyticsArgs{
			Enabled:             pulumi.Bool(trafficAnalyticsEnabled),
			WorkspaceId:         pulumi.String(spec.TrafficAnalytics.WorkspaceId.GetValue()),
			WorkspaceRegion:     pulumi.String(spec.TrafficAnalytics.WorkspaceRegion),
			WorkspaceResourceId: pulumi.String(spec.TrafficAnalytics.WorkspaceResourceId.GetValue()),
			IntervalInMinutes:   pulumi.Int(intervalInMinutes),
		}
	}

	createdFlowLog, err := network.NewNetworkWatcherFlowLog(ctx,
		spec.Name,
		flowLogArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create network watcher flow log %s", spec.Name)
	}

	ctx.Export(OpFlowLogId, createdFlowLog.ID())
	ctx.Export(OpFlowLogName, createdFlowLog.Name)
	ctx.Export(OpNetworkWatcherName, createdFlowLog.NetworkWatcherName)

	return nil
}
