package module

import (
	"strings"

	azuremonitormetricalertv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitormetricalert/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMonitorMetricAlert *azuremonitormetricalertv1alpha1.AzureMonitorMetricAlert
	ResourceGroupName       string
	AzureTags               map[string]string
}

// aggregationStrings maps the aggregation enum to ARM's wire values.
var aggregationStrings = map[azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertAggregation]string{
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertAggregation_AVERAGE: "Average",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertAggregation_COUNT:   "Count",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertAggregation_MINIMUM: "Minimum",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertAggregation_MAXIMUM: "Maximum",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertAggregation_TOTAL:   "Total",
}

// operatorStrings is one shared vocabulary for both criteria families (the
// spec CELs keep each family to its legal subset).
var operatorStrings = map[azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertOperator]string{
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertOperator_EQUALS:                "Equals",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertOperator_GREATER_THAN:          "GreaterThan",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertOperator_GREATER_THAN_OR_EQUAL: "GreaterThanOrEqual",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertOperator_LESS_THAN:             "LessThan",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertOperator_LESS_THAN_OR_EQUAL:    "LessThanOrEqual",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertOperator_GREATER_OR_LESS_THAN:  "GreaterOrLessThan",
}

// dimensionOperatorStrings maps the dimension-filter enum to ARM's values.
var dimensionOperatorStrings = map[azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertDimensionOperator]string{
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertDimensionOperator_INCLUDE:     "Include",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertDimensionOperator_EXCLUDE:     "Exclude",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertDimensionOperator_STARTS_WITH: "StartsWith",
}

// sensitivityStrings maps the dynamic-threshold sensitivity enum to ARM's
// values.
var sensitivityStrings = map[azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertSensitivity]string{
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertSensitivity_LOW:    "Low",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertSensitivity_MEDIUM: "Medium",
	azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertSensitivity_HIGH:   "High",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertIacInput) *Locals {
	locals := &Locals{}

	locals.AzureMonitorMetricAlert = iacInput.Target
	target := iacInput.Target

	// The resource_group field is a StringValueOrRef. The platform middleware
	// resolves valueFrom references before IaC modules run, so .GetValue()
	// always returns the resolved literal string.
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Identity tags derived from metadata; user tags merge OVER these (the
	// governance surface belongs to the user).
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureMonitorMetricAlert.String()),
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

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
