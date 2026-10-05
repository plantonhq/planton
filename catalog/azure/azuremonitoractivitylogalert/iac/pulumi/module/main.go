package module

import (
	"github.com/pkg/errors"
	azuremonitoractivitylogalertv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitoractivitylogalert/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/monitoring"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremonitoractivitylogalertv1alpha1.AzureMonitorActivityLogAlertIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMonitorActivityLogAlert.Spec
	c := spec.Criteria

	// Build the criteria block. The category is required; every other field
	// narrows within it and is sent only when set (the spec's exclusivity
	// CELs guarantee valid combinations). Plural list fields carry the
	// single case as a one-element list.
	criteriaArgs := monitoring.ActivityLogAlertCriteriaArgs{
		Category: pulumi.String(categoryString(c.Category)),
	}
	if c.OperationName != "" {
		criteriaArgs.OperationName = pulumi.String(c.OperationName)
	}
	if c.Caller != "" {
		criteriaArgs.Caller = pulumi.String(c.Caller)
	}
	if len(c.Levels) > 0 {
		criteriaArgs.Levels = pulumi.ToStringArray(levelStrings(c.Levels))
	}
	if len(c.ResourceProviders) > 0 {
		criteriaArgs.ResourceProviders = pulumi.ToStringArray(c.ResourceProviders)
	}
	if len(c.ResourceTypes) > 0 {
		criteriaArgs.ResourceTypes = pulumi.ToStringArray(c.ResourceTypes)
	}
	if len(c.ResourceGroups) > 0 {
		criteriaArgs.ResourceGroups = pulumi.ToStringArray(c.ResourceGroups)
	}
	if len(c.ResourceIds) > 0 {
		criteriaArgs.ResourceIds = pulumi.ToStringArray(c.ResourceIds)
	}
	if len(c.Statuses) > 0 {
		criteriaArgs.Statuses = pulumi.ToStringArray(c.Statuses)
	}
	if len(c.SubStatuses) > 0 {
		criteriaArgs.SubStatuses = pulumi.ToStringArray(c.SubStatuses)
	}
	if s := recommendationCategoryString(c.RecommendationCategory); s != "" {
		criteriaArgs.RecommendationCategory = pulumi.String(s)
	}
	if s := recommendationImpactString(c.RecommendationImpact); s != "" {
		criteriaArgs.RecommendationImpact = pulumi.String(s)
	}
	if c.RecommendationType != "" {
		criteriaArgs.RecommendationType = pulumi.String(c.RecommendationType)
	}
	if c.ResourceHealth != nil {
		rh := &monitoring.ActivityLogAlertCriteriaResourceHealthArgs{}
		if len(c.ResourceHealth.Current) > 0 {
			rh.Currents = pulumi.ToStringArray(healthStatusStrings(c.ResourceHealth.Current))
		}
		if len(c.ResourceHealth.Previous) > 0 {
			rh.Previouses = pulumi.ToStringArray(healthStatusStrings(c.ResourceHealth.Previous))
		}
		if len(c.ResourceHealth.Reason) > 0 {
			rh.Reasons = pulumi.ToStringArray(healthReasonStrings(c.ResourceHealth.Reason))
		}
		criteriaArgs.ResourceHealth = rh
	}
	if c.ServiceHealth != nil {
		sh := &monitoring.ActivityLogAlertCriteriaServiceHealthArgs{}
		if len(c.ServiceHealth.Events) > 0 {
			sh.Events = pulumi.ToStringArray(serviceHealthEventStrings(c.ServiceHealth.Events))
		}
		if len(c.ServiceHealth.Locations) > 0 {
			sh.Locations = pulumi.ToStringArray(c.ServiceHealth.Locations)
		}
		if len(c.ServiceHealth.Services) > 0 {
			sh.Services = pulumi.ToStringArray(c.ServiceHealth.Services)
		}
		criteriaArgs.ServiceHealth = sh
	}

	// Actions: each notifies an action group with optional webhook props.
	// Action-group references resolve to literal ARM IDs before the module
	// runs.
	actions := monitoring.ActivityLogAlertActionArray{}
	for _, a := range spec.Actions {
		actionArgs := monitoring.ActivityLogAlertActionArgs{
			ActionGroupId: pulumi.String(a.ActionGroupId.GetValue()),
		}
		if len(a.WebhookProperties) > 0 {
			actionArgs.WebhookProperties = pulumi.ToStringMap(a.WebhookProperties)
		}
		actions = append(actions, actionArgs)
	}

	alertArgs := &monitoring.ActivityLogAlertArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Location:          pulumi.String(locals.Location),
		Scopes:            pulumi.ToStringArray(locals.ScopeIds),
		Criteria:          criteriaArgs,
		Actions:           actions,
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}
	if spec.Description != "" {
		alertArgs.Description = pulumi.String(spec.Description)
	}
	// Azure defaults enabled to true; only an explicit choice is sent so an
	// unspecified spec deploys identically on both engines.
	if spec.Enabled != nil {
		alertArgs.Enabled = pulumi.Bool(spec.GetEnabled())
	}

	createdAlert, err := monitoring.NewActivityLogAlert(ctx,
		spec.Name,
		alertArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create activity log alert %s", spec.Name)
	}

	ctx.Export(OpActivityLogAlertId, createdAlert.ID())
	ctx.Export(OpActivityLogAlertName, createdAlert.Name)

	return nil
}
