package module

import (
	"strings"

	azuremonitorscheduledqueryalertv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitorscheduledqueryalert/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMonitorScheduledQueryAlert *azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlert
	ResourceGroupName               string
	AzureTags                       map[string]string
}

// timeAggregationStrings maps the aggregation enum to ARM's wire values.
var timeAggregationStrings = map[azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertTimeAggregation]string{
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertTimeAggregation_COUNT:   "Count",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertTimeAggregation_AVERAGE: "Average",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertTimeAggregation_MINIMUM: "Minimum",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertTimeAggregation_MAXIMUM: "Maximum",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertTimeAggregation_TOTAL:   "Total",
}

// operatorStrings maps the comparison enum to ARM's wire values. Note this
// API's equality operator is "Equal" (not the metric-alert API's "Equals").
var operatorStrings = map[azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertOperator]string{
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertOperator_EQUAL:                 "Equal",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertOperator_GREATER_THAN:          "GreaterThan",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertOperator_GREATER_THAN_OR_EQUAL: "GreaterThanOrEqual",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertOperator_LESS_THAN:             "LessThan",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertOperator_LESS_THAN_OR_EQUAL:    "LessThanOrEqual",
}

// dimensionOperatorStrings maps the dimension-split enum to ARM's values.
var dimensionOperatorStrings = map[azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertDimensionOperator]string{
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertDimensionOperator_INCLUDE: "Include",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertDimensionOperator_EXCLUDE: "Exclude",
}

// identityTypeStrings maps the identity-model enum to ARM's values.
var identityTypeStrings = map[azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertIdentityType]string{
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertIdentityType_SYSTEM_ASSIGNED: "SystemAssigned",
	azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertIdentityType_USER_ASSIGNED:   "UserAssigned",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertIacInput) *Locals {
	locals := &Locals{}

	locals.AzureMonitorScheduledQueryAlert = iacInput.Target
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
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureMonitorScheduledQueryAlert.String()),
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
