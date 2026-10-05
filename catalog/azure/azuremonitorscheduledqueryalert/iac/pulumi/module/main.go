package module

import (
	"github.com/pkg/errors"
	azuremonitorscheduledqueryalertv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitorscheduledqueryalert/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/monitoring"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremonitorscheduledqueryalertv1alpha1.AzureMonitorScheduledQueryAlertIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMonitorScheduledQueryAlert.Spec

	// The scheduled query alert runs KQL against the scoped workspace (or
	// Application Insights resource) on a schedule and fires action groups
	// when its condition holds. The rule is regional and must live in the
	// same region as the resource it queries.
	criterias := monitoring.ScheduledQueryRulesAlertV2CriteriaArray{}
	for _, criteria := range spec.Criteria {
		criteriaArgs := monitoring.ScheduledQueryRulesAlertV2CriteriaArgs{
			Query:                 pulumi.String(criteria.Query),
			TimeAggregationMethod: pulumi.String(timeAggregationStrings[criteria.TimeAggregationMethod]),
			Operator:              pulumi.String(operatorStrings[criteria.Operator]),
			Threshold:             pulumi.Float64(criteria.Threshold),
		}
		// Required for non-Count aggregations, forbidden for Count -- Azure's
		// apply-time pairing, documented on the spec field.
		if criteria.MetricMeasureColumn != "" {
			criteriaArgs.MetricMeasureColumn = pulumi.String(criteria.MetricMeasureColumn)
		}
		if criteria.ResourceIdColumn != "" {
			criteriaArgs.ResourceIdColumn = pulumi.String(criteria.ResourceIdColumn)
		}
		if len(criteria.Dimensions) > 0 {
			dimensions := monitoring.ScheduledQueryRulesAlertV2CriteriaDimensionArray{}
			for _, dimension := range criteria.Dimensions {
				dimensions = append(dimensions, monitoring.ScheduledQueryRulesAlertV2CriteriaDimensionArgs{
					Name:     pulumi.String(dimension.Name),
					Operator: pulumi.String(dimensionOperatorStrings[dimension.Operator]),
					Values:   pulumi.ToStringArray(dimension.Values),
				})
			}
			criteriaArgs.Dimensions = dimensions
		}
		// The flap damper: require N of M recent evaluations to breach
		// before firing.
		if criteria.FailingPeriods != nil {
			criteriaArgs.FailingPeriods = &monitoring.ScheduledQueryRulesAlertV2CriteriaFailingPeriodsArgs{
				MinimumFailingPeriodsToTriggerAlert: pulumi.Int(int(criteria.FailingPeriods.MinimumFailingPeriodsToTriggerAlert)),
				NumberOfEvaluationPeriods:           pulumi.Int(int(criteria.FailingPeriods.NumberOfEvaluationPeriods)),
			}
		}
		criterias = append(criterias, criteriaArgs)
	}

	alertArgs := &monitoring.ScheduledQueryRulesAlertV2Args{
		Name:              pulumi.String(spec.AlertName),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Location:          pulumi.String(spec.Region),
		// The provider caps scopes at exactly one, so the pulumi bridge
		// flattens the one-item list to a singular string input; the TF
		// module wraps the same value in a one-item list. Wire-identical.
		Scopes:    pulumi.String(spec.Scope.GetValue()),
		Criterias: criterias,
		// Presence-guarded to the proto defaults: IaC inputs built from a
		// manifest materialize defaults, but direct iac-input paths do not.
		Enabled:             pulumi.Bool(presenceGuardedBool(spec.Enabled, true)),
		Severity:            pulumi.Int(int(presenceGuardedInt32(spec.Severity, 3))),
		EvaluationFrequency: pulumi.String(presenceGuardedString(spec.EvaluationFrequency, "PT5M")),
		WindowDuration:      pulumi.String(presenceGuardedString(spec.WindowDuration, "PT5M")),
		// Mutually exclusive with the mute duration (spec-enforced).
		AutoMitigationEnabled:         pulumi.Bool(spec.AutoMitigationEnabled),
		WorkspaceAlertsStorageEnabled: pulumi.Bool(spec.WorkspaceAlertsStorageEnabled),
		SkipQueryValidation:           pulumi.Bool(spec.SkipQueryValidation),
		Tags:                          pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.DisplayName != "" {
		alertArgs.DisplayName = pulumi.String(spec.DisplayName)
	}
	if spec.Description != "" {
		alertArgs.Description = pulumi.String(spec.Description)
	}
	if spec.QueryTimeRangeOverride != "" {
		alertArgs.QueryTimeRangeOverride = pulumi.String(spec.QueryTimeRangeOverride)
	}
	if spec.MuteActionsAfterAlertDuration != "" {
		alertArgs.MuteActionsAfterAlertDuration = pulumi.String(spec.MuteActionsAfterAlertDuration)
	}
	if len(spec.TargetResourceTypes) > 0 {
		alertArgs.TargetResourceTypes = pulumi.ToStringArray(spec.TargetResourceTypes)
	}

	// The rule's managed identity -- required when the workspace enforces
	// Entra-only query access.
	if spec.Identity != nil {
		identityArgs := &monitoring.ScheduledQueryRulesAlertV2IdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.UserAssignedIdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.UserAssignedIdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		alertArgs.Identity = identityArgs
	}

	if spec.Action != nil {
		actionArgs := &monitoring.ScheduledQueryRulesAlertV2ActionArgs{}
		if len(spec.Action.ActionGroupIds) > 0 {
			actionGroups := pulumi.StringArray{}
			for _, actionGroupId := range spec.Action.ActionGroupIds {
				actionGroups = append(actionGroups, pulumi.String(actionGroupId.GetValue()))
			}
			actionArgs.ActionGroups = actionGroups
		}
		if len(spec.Action.CustomProperties) > 0 {
			actionArgs.CustomProperties = pulumi.ToStringMap(spec.Action.CustomProperties)
		}
		if spec.Action.EmailSubject != "" {
			actionArgs.EmailSubject = pulumi.String(spec.Action.EmailSubject)
		}
		alertArgs.Action = actionArgs
	}

	createdAlert, err := monitoring.NewScheduledQueryRulesAlertV2(ctx,
		spec.AlertName,
		alertArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create scheduled query alert %s", spec.AlertName)
	}

	// Export outputs. Empty principal id unless SYSTEM_ASSIGNED is
	// enabled -- mirrors the TF module's try(identity[0].principal_id, "").
	ctx.Export(OpScheduledQueryAlertId, createdAlert.ID())
	ctx.Export(OpScheduledQueryAlertName, createdAlert.Name)
	ctx.Export(OpIdentityPrincipalId, createdAlert.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput))

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

func presenceGuardedString(field *string, defaultValue string) string {
	if field == nil || *field == "" {
		return defaultValue
	}
	return *field
}
