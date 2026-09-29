package module

import (
	"strings"

	azuremonitoractivitylogalertv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitoractivitylogalert/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMonitorActivityLogAlert *azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlert

	ResourceGroupName string

	// ScopeIds are the resolved ARM IDs the alert watches.
	ScopeIds []string

	// Location is the ARM region string for the alert resource ("global"
	// when unspecified).
	Location string

	AzureTags map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertStackInput) *Locals {
	locals := &Locals{}

	locals.AzureMonitorActivityLogAlert = stackInput.Target
	target := stackInput.Target
	spec := target.Spec

	locals.ResourceGroupName = spec.ResourceGroup.GetValue()

	for _, s := range spec.Scopes {
		locals.ScopeIds = append(locals.ScopeIds, s.GetValue())
	}

	locals.Location = activityLogAlertLocation(spec.Location)

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureMonitorActivityLogAlert.String()),
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
	for k, v := range spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}

// activityLogAlertLocation maps the location enum to its ARM string,
// defaulting to "global" when unspecified.
func activityLogAlertLocation(l azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLocation) string {
	switch l {
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLocation_WEST_EUROPE:
		return "westeurope"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLocation_NORTH_EUROPE:
		return "northeurope"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLocation_EAST_US_2_EUAP:
		return "eastus2euap"
	default:
		return "global"
	}
}

func categoryString(c azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertCategory) string {
	switch c {
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertCategory_ADMINISTRATIVE:
		return "Administrative"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertCategory_AUTOSCALE:
		return "Autoscale"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertCategory_POLICY:
		return "Policy"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertCategory_RECOMMENDATION:
		return "Recommendation"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertCategory_RESOURCE_HEALTH:
		return "ResourceHealth"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertCategory_SECURITY:
		return "Security"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertCategory_SERVICE_HEALTH:
		return "ServiceHealth"
	}
	return ""
}

func levelString(l azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLevel) string {
	switch l {
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLevel_VERBOSE:
		return "Verbose"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLevel_INFORMATIONAL:
		return "Informational"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLevel_WARNING:
		return "Warning"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLevel_ERROR:
		return "Error"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLevel_CRITICAL:
		return "Critical"
	}
	return ""
}

func levelStrings(ls []azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertLevel) []string {
	out := make([]string, 0, len(ls))
	for _, l := range ls {
		if s := levelString(l); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func recommendationCategoryString(c azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationCategory) string {
	switch c {
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationCategory_COST:
		return "Cost"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationCategory_RELIABILITY:
		return "Reliability"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationCategory_OPERATIONAL_EXCELLENCE:
		return "OperationalExcellence"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationCategory_PERFORMANCE:
		return "Performance"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationCategory_HIGH_AVAILABILITY:
		return "HighAvailability"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationCategory_SECURITY_RECOMMENDATION:
		return "Security"
	}
	return ""
}

func recommendationImpactString(i azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationImpact) string {
	switch i {
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationImpact_HIGH:
		return "High"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationImpact_MEDIUM:
		return "Medium"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertRecommendationImpact_LOW:
		return "Low"
	}
	return ""
}

func healthStatusString(h azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthStatus) string {
	switch h {
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthStatus_AVAILABLE:
		return "Available"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthStatus_DEGRADED:
		return "Degraded"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthStatus_UNAVAILABLE:
		return "Unavailable"
	case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthStatus_UNKNOWN:
		return "Unknown"
	}
	return ""
}

func healthStatusStrings(hs []azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthStatus) []string {
	out := make([]string, 0, len(hs))
	for _, h := range hs {
		if s := healthStatusString(h); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func healthReasonStrings(rs []azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthReason) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		switch r {
		case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthReason_PLATFORM_INITIATED:
			out = append(out, "PlatformInitiated")
		case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthReason_USER_INITIATED:
			out = append(out, "UserInitiated")
		case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertHealthReason_REASON_UNKNOWN:
			out = append(out, "Unknown")
		}
	}
	return out
}

func serviceHealthEventStrings(es []azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertServiceHealthEvent) []string {
	out := make([]string, 0, len(es))
	for _, e := range es {
		switch e {
		case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertServiceHealthEvent_INCIDENT:
			out = append(out, "Incident")
		case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertServiceHealthEvent_MAINTENANCE:
			out = append(out, "Maintenance")
		case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertServiceHealthEvent_EVENT_INFORMATIONAL:
			out = append(out, "Informational")
		case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertServiceHealthEvent_ACTION_REQUIRED:
			out = append(out, "ActionRequired")
		case azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertServiceHealthEvent_EVENT_SECURITY:
			out = append(out, "Security")
		}
	}
	return out
}
