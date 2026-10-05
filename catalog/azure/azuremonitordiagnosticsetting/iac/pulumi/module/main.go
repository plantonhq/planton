package module

import (
	"fmt"

	"github.com/pkg/errors"
	azuremonitordiagnosticsettingv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitordiagnosticsetting/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/monitoring"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremonitordiagnosticsettingv1alpha1.AzureMonitorDiagnosticSettingIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMonitorDiagnosticSetting.Spec
	targetResourceId := spec.TargetResourceId.GetValue()

	// The diagnostic setting is an extension resource on the target: it
	// selects which platform logs and metrics the target emits and routes
	// them to the configured destinations. The spec enforces Azure's real
	// contracts up front (at least one category, at least one destination,
	// category XOR category_group per log entry).
	settingArgs := &monitoring.DiagnosticSettingArgs{
		Name:             pulumi.String(spec.SettingName),
		TargetResourceId: pulumi.String(targetResourceId),
	}

	// Log selections: exactly one of category or category_group per entry
	// (spec-enforced XOR); every selected category is enabled.
	if len(spec.EnabledLogs) > 0 {
		enabledLogs := monitoring.DiagnosticSettingEnabledLogArray{}
		for _, enabledLog := range spec.EnabledLogs {
			logArgs := monitoring.DiagnosticSettingEnabledLogArgs{}
			if enabledLog.Category != "" {
				logArgs.Category = pulumi.String(enabledLog.Category)
			}
			if enabledLog.CategoryGroup != "" {
				logArgs.CategoryGroup = pulumi.String(enabledLog.CategoryGroup)
			}
			enabledLogs = append(enabledLogs, logArgs)
		}
		settingArgs.EnabledLogs = enabledLogs
	}

	// Metric selections.
	if len(spec.EnabledMetrics) > 0 {
		enabledMetrics := monitoring.DiagnosticSettingEnabledMetricArray{}
		for _, enabledMetric := range spec.EnabledMetrics {
			enabledMetrics = append(enabledMetrics, monitoring.DiagnosticSettingEnabledMetricArgs{
				Category: pulumi.String(enabledMetric.Category),
			})
		}
		settingArgs.EnabledMetrics = enabledMetrics
	}

	// Destinations -- at least one is set (spec-enforced). The eventhub
	// name only rides along with its authorization rule.
	if spec.LogAnalyticsWorkspaceId.GetValue() != "" {
		settingArgs.LogAnalyticsWorkspaceId = pulumi.String(spec.LogAnalyticsWorkspaceId.GetValue())
	}
	if spec.LogAnalyticsDestinationType != azuremonitordiagnosticsettingv1alpha1.AzureMonitorDiagnosticSettingLogAnalyticsDestinationType_azure_monitor_diagnostic_setting_log_analytics_destination_type_unspecified {
		settingArgs.LogAnalyticsDestinationType = pulumi.String(destinationTypeStrings[spec.LogAnalyticsDestinationType])
	}
	if spec.StorageAccountId.GetValue() != "" {
		settingArgs.StorageAccountId = pulumi.String(spec.StorageAccountId.GetValue())
	}
	if spec.EventhubAuthorizationRuleId.GetValue() != "" {
		settingArgs.EventhubAuthorizationRuleId = pulumi.String(spec.EventhubAuthorizationRuleId.GetValue())
	}
	if spec.EventhubName.GetValue() != "" {
		settingArgs.EventhubName = pulumi.String(spec.EventhubName.GetValue())
	}
	if spec.PartnerSolutionId != "" {
		settingArgs.PartnerSolutionId = pulumi.String(spec.PartnerSolutionId)
	}

	createdSetting, err := monitoring.NewDiagnosticSetting(ctx,
		spec.SettingName,
		settingArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create diagnostic setting %s", spec.SettingName)
	}

	// Export outputs. The provider's own state id is a
	// "{target}|{name}" composite, not an ARM id -- the exported id
	// constructs the real ARM extension-resource id so downstream consumers
	// and verification address the setting the way the API does.
	ctx.Export(OpDiagnosticSettingId, pulumi.String(fmt.Sprintf(
		"%s/providers/Microsoft.Insights/diagnosticSettings/%s", targetResourceId, spec.SettingName)))
	ctx.Export(OpDiagnosticSettingName, createdSetting.Name)
	ctx.Export(OpTargetResourceId, createdSetting.TargetResourceId)

	return nil
}
