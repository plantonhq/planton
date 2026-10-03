package module

import (
	"github.com/pkg/errors"
	azuremonitormetricalertv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremonitormetricalert/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/monitoring"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremonitormetricalertv1alpha1.AzureMonitorMetricAlertIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder,
	// which resolves the right credential mechanism (static client secret,
	// keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMonitorMetricAlert.Spec

	// The metric alert evaluates platform metrics on a rolling window and
	// fires the referenced action groups. Exactly one condition family is
	// configured (spec-enforced).
	scopes := pulumi.StringArray{}
	for _, scope := range spec.Scopes {
		scopes = append(scopes, pulumi.String(scope.GetValue()))
	}

	alertArgs := &monitoring.MetricAlertArgs{
		Name:              pulumi.String(spec.AlertName),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Scopes:            scopes,
		// Presence-guarded to the proto defaults: IaC inputs built from a
		// manifest materialize defaults, but direct iac-input paths do not.
		Enabled:      pulumi.Bool(presenceGuardedBool(spec.Enabled, true)),
		AutoMitigate: pulumi.Bool(presenceGuardedBool(spec.AutoMitigate, true)),
		Severity:     pulumi.Int(int(presenceGuardedInt32(spec.Severity, 3))),
		Frequency:    pulumi.String(presenceGuardedString(spec.Frequency, "PT1M")),
		WindowSize:   pulumi.String(presenceGuardedString(spec.WindowSize, "PT5M")),
		Tags:         pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.Description != "" {
		alertArgs.Description = pulumi.String(spec.Description)
	}
	// Required for multi-resource / group / subscription scopes so Azure can
	// resolve the metric definition.
	if spec.TargetResourceType != "" {
		alertArgs.TargetResourceType = pulumi.String(spec.TargetResourceType)
	}
	if spec.TargetResourceLocation != "" {
		alertArgs.TargetResourceLocation = pulumi.String(spec.TargetResourceLocation)
	}

	// Static thresholds -- multiple criteria AND together.
	if len(spec.StaticCriteria) > 0 {
		criterias := monitoring.MetricAlertCriteriaArray{}
		for _, criteria := range spec.StaticCriteria {
			criteriaArgs := monitoring.MetricAlertCriteriaArgs{
				MetricNamespace:      pulumi.String(criteria.MetricNamespace),
				MetricName:           pulumi.String(criteria.MetricName),
				Aggregation:          pulumi.String(aggregationStrings[criteria.Aggregation]),
				Operator:             pulumi.String(operatorStrings[criteria.Operator]),
				Threshold:            pulumi.Float64(criteria.Threshold),
				SkipMetricValidation: pulumi.Bool(criteria.SkipMetricValidation),
			}
			if len(criteria.Dimensions) > 0 {
				dimensions := monitoring.MetricAlertCriteriaDimensionArray{}
				for _, dimension := range criteria.Dimensions {
					dimensions = append(dimensions, monitoring.MetricAlertCriteriaDimensionArgs{
						Name:     pulumi.String(dimension.Name),
						Operator: pulumi.String(dimensionOperatorStrings[dimension.Operator]),
						Values:   pulumi.ToStringArray(dimension.Values),
					})
				}
				criteriaArgs.Dimensions = dimensions
			}
			criterias = append(criterias, criteriaArgs)
		}
		alertArgs.Criterias = criterias
	}

	// Dynamic machine-learning threshold -- Azure learns the metric's normal
	// band; sensitivity controls how tightly the band hugs it.
	if spec.DynamicCriteria != nil {
		dynamicArgs := &monitoring.MetricAlertDynamicCriteriaArgs{
			MetricNamespace:  pulumi.String(spec.DynamicCriteria.MetricNamespace),
			MetricName:       pulumi.String(spec.DynamicCriteria.MetricName),
			Aggregation:      pulumi.String(aggregationStrings[spec.DynamicCriteria.Aggregation]),
			Operator:         pulumi.String(operatorStrings[spec.DynamicCriteria.Operator]),
			AlertSensitivity: pulumi.String(sensitivityStrings[spec.DynamicCriteria.AlertSensitivity]),
			// Presence-guarded to the proto defaults (4/4 -- the provider's
			// own defaults, sent explicitly so both engines match).
			EvaluationTotalCount:   pulumi.Int(int(presenceGuardedInt32(spec.DynamicCriteria.EvaluationTotalCount, 4))),
			EvaluationFailureCount: pulumi.Int(int(presenceGuardedInt32(spec.DynamicCriteria.EvaluationFailureCount, 4))),
			SkipMetricValidation:   pulumi.Bool(spec.DynamicCriteria.SkipMetricValidation),
		}
		if spec.DynamicCriteria.IgnoreDataBefore != "" {
			dynamicArgs.IgnoreDataBefore = pulumi.String(spec.DynamicCriteria.IgnoreDataBefore)
		}
		if len(spec.DynamicCriteria.Dimensions) > 0 {
			dimensions := monitoring.MetricAlertDynamicCriteriaDimensionArray{}
			for _, dimension := range spec.DynamicCriteria.Dimensions {
				dimensions = append(dimensions, monitoring.MetricAlertDynamicCriteriaDimensionArgs{
					Name:     pulumi.String(dimension.Name),
					Operator: pulumi.String(dimensionOperatorStrings[dimension.Operator]),
					Values:   pulumi.ToStringArray(dimension.Values),
				})
			}
			dynamicArgs.Dimensions = dimensions
		}
		alertArgs.DynamicCriteria = dynamicArgs
	}

	// Web-test availability -- fires when the referenced Application
	// Insights availability test fails from N locations.
	if spec.WebTestAvailabilityCriteria != nil {
		alertArgs.ApplicationInsightsWebTestLocationAvailabilityCriteria = &monitoring.MetricAlertApplicationInsightsWebTestLocationAvailabilityCriteriaArgs{
			WebTestId:           pulumi.String(spec.WebTestAvailabilityCriteria.WebTestId.GetValue()),
			ComponentId:         pulumi.String(spec.WebTestAvailabilityCriteria.ComponentId.GetValue()),
			FailedLocationCount: pulumi.Int(int(spec.WebTestAvailabilityCriteria.FailedLocationCount)),
		}
	}

	if len(spec.Actions) > 0 {
		actions := monitoring.MetricAlertActionArray{}
		for _, action := range spec.Actions {
			actionArgs := monitoring.MetricAlertActionArgs{
				ActionGroupId: pulumi.String(action.ActionGroupId.GetValue()),
			}
			if len(action.WebhookProperties) > 0 {
				actionArgs.WebhookProperties = pulumi.ToStringMap(action.WebhookProperties)
			}
			actions = append(actions, actionArgs)
		}
		alertArgs.Actions = actions
	}

	createdAlert, err := monitoring.NewMetricAlert(ctx,
		spec.AlertName,
		alertArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create metric alert %s", spec.AlertName)
	}

	// Export outputs.
	ctx.Export(OpMetricAlertId, createdAlert.ID())
	ctx.Export(OpMetricAlertName, createdAlert.Name)

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
