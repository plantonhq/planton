package module

import (
	azuremonitordiagnosticsettingv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitordiagnosticsetting/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMonitorDiagnosticSetting *azuremonitordiagnosticsettingv1alpha1.AzureMonitorDiagnosticSetting
}

// destinationTypeStrings maps the workspace-table-layout enum to ARM's wire
// values.
var destinationTypeStrings = map[azuremonitordiagnosticsettingv1alpha1.AzureMonitorDiagnosticSettingLogAnalyticsDestinationType]string{
	azuremonitordiagnosticsettingv1alpha1.AzureMonitorDiagnosticSettingLogAnalyticsDestinationType_DEDICATED:         "Dedicated",
	azuremonitordiagnosticsettingv1alpha1.AzureMonitorDiagnosticSettingLogAnalyticsDestinationType_AZURE_DIAGNOSTICS: "AzureDiagnostics",
}

// The diagnostic setting carries no tags (the ARM extension resource does
// not support them), so locals stay minimal.
func initializeLocals(ctx *pulumi.Context, iacInput *azuremonitordiagnosticsettingv1alpha1.AzureMonitorDiagnosticSettingIacInput) *Locals {
	return &Locals{
		AzureMonitorDiagnosticSetting: iacInput.Target,
	}
}
