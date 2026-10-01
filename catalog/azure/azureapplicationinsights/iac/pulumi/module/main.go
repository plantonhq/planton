package module

import (
	"github.com/pkg/errors"
	azureapplicationinsightsv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureapplicationinsights/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appinsights"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureapplicationinsightsv1alpha1.AzureApplicationInsightsStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureApplicationInsights.Spec

	// Workspace-based Application Insights: telemetry lands in the
	// referenced Log Analytics Workspace (classic mode was retired by Azure
	// in February 2024). The workspace binding can be repointed but never
	// removed once set.
	//
	// PARITY-EXCEPTION: pulumi-azure v6.38 bridges only the provider's
	// DEPRECATED negative-form toggles (disableIpMasking,
	// localAuthenticationDisabled, dailyDataCapNotificationsDisabled) while
	// the Terraform module uses the v5-era positive forms. The wire
	// property is identical for each pair -- this module inverts the spec's
	// positive booleans; behavior and outputs match exactly. Re-align when
	// the bridge ships the positive forms.
	insightsArgs := &appinsights.InsightsArgs{
		Name:              pulumi.String(spec.ApplicationInsightsName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		ApplicationType:   pulumi.String(applicationTypeStrings[spec.ApplicationType]),
		WorkspaceId:       pulumi.String(spec.WorkspaceId.GetValue()),
		// Presence-guarded to the proto defaults: stack inputs built from a
		// manifest materialize defaults, but direct stack-input paths do not.
		RetentionInDays:                   pulumi.Int(int(presenceGuardedInt32(spec.RetentionInDays, 90))),
		DailyDataCapInGb:                  pulumi.Float64(presenceGuardedFloat64(spec.DailyDataCapInGb, 100)),
		SamplingPercentage:                pulumi.Float64(presenceGuardedFloat64(spec.SamplingPercentage, 100)),
		DailyDataCapNotificationsDisabled: pulumi.Bool(!presenceGuardedBool(spec.DailyDataCapNotificationsEnabled, true)),
		DisableIpMasking:                  pulumi.Bool(!presenceGuardedBool(spec.IpMaskingEnabled, true)),
		LocalAuthenticationDisabled:       pulumi.Bool(!presenceGuardedBool(spec.LocalAuthenticationEnabled, true)),
		InternetIngestionEnabled:          pulumi.Bool(presenceGuardedBool(spec.InternetIngestionEnabled, true)),
		InternetQueryEnabled:              pulumi.Bool(presenceGuardedBool(spec.InternetQueryEnabled, true)),
		ForceCustomerStorageForProfiler:   pulumi.Bool(spec.ForceCustomerStorageForProfiler),
		Tags:                              pulumi.ToStringMap(locals.AzureTags),
	}

	createdInsights, err := appinsights.NewInsights(ctx,
		spec.ApplicationInsightsName,
		insightsArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Application Insights %s", spec.ApplicationInsightsName)
	}

	// Export stack outputs. connection_string is the composition seam the
	// app-hosting kinds reference.
	ctx.Export(OpApplicationInsightsId, createdInsights.ID())
	ctx.Export(OpApplicationInsightsName, createdInsights.Name)
	ctx.Export(OpInstrumentationKey, pulumi.Unsecret(createdInsights.InstrumentationKey))
	ctx.Export(OpConnectionString, pulumi.Unsecret(createdInsights.ConnectionString))
	ctx.Export(OpAppId, createdInsights.AppId)

	return nil
}

// presenceGuardedBool returns the field's value when set and the proto
// default otherwise -- default materialization is middleware behavior, not a
// wire guarantee.
func presenceGuardedBool(field *bool, defaultValue bool) bool {
	if field == nil {
		return defaultValue
	}
	return *field
}

func presenceGuardedInt32(field *int32, defaultValue int32) int32 {
	if field == nil {
		return defaultValue
	}
	return *field
}

func presenceGuardedFloat64(field *float64, defaultValue float64) float64 {
	if field == nil {
		return defaultValue
	}
	return *field
}
