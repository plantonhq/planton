package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	cd "github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/clouddeploy"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// automations creates the pipeline's automations, each keyed by its own
// declared ID so adding or removing one never renames or recreates its
// siblings. Every automation takes the pipeline's project, region, and
// deletion policy, and the created pipeline's ID.
func automations(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, pipeline *cd.DeliveryPipeline) error {
	spec := locals.GcpDeliveryPipeline.Spec
	resourceName := locals.GcpDeliveryPipeline.Metadata.Name

	for _, automation := range spec.Automations {
		targets := cd.AutomationSelectorTargetArray{}
		for _, target := range automation.GetSelector().GetTargets() {
			targets = append(targets, cd.AutomationSelectorTargetArgs{
				Id:     optionalString(target.GetId().GetValue()),
				Labels: stringMap(target.Labels),
			})
		}

		rules := cd.AutomationRuleArray{}
		for _, rule := range automation.Rules {
			ruleArgs := cd.AutomationRuleArgs{}
			if r := rule.AdvanceRolloutRule; r != nil {
				ruleArgs.AdvanceRolloutRule = &cd.AutomationRuleAdvanceRolloutRuleArgs{
					Id:           pulumi.String(r.Id),
					SourcePhases: stringArray(r.SourcePhases),
					Wait:         optionalString(r.Wait),
				}
			}
			if r := rule.PromoteReleaseRule; r != nil {
				ruleArgs.PromoteReleaseRule = &cd.AutomationRulePromoteReleaseRuleArgs{
					Id:                  pulumi.String(r.Id),
					Wait:                optionalString(r.Wait),
					DestinationTargetId: optionalString(r.GetDestinationTargetId().GetValue()),
					DestinationPhase:    optionalString(r.DestinationPhase),
				}
			}
			if r := rule.RepairRolloutRule; r != nil {
				phases := cd.AutomationRuleRepairRolloutRuleRepairPhaseArray{}
				for _, phase := range r.RepairPhases {
					phaseArgs := cd.AutomationRuleRepairRolloutRuleRepairPhaseArgs{}
					if retry := phase.Retry; retry != nil {
						phaseArgs.Retry = &cd.AutomationRuleRepairRolloutRuleRepairPhaseRetryArgs{
							Attempts:    pulumi.String(retry.Attempts),
							Wait:        optionalString(retry.Wait),
							BackoffMode: optionalString(retry.BackoffMode),
						}
					}
					if rollback := phase.Rollback; rollback != nil {
						phaseArgs.Rollback = &cd.AutomationRuleRepairRolloutRuleRepairPhaseRollbackArgs{
							DestinationPhase:                optionalString(rollback.DestinationPhase),
							DisableRollbackIfRolloutPending: optionalTrue(rollback.DisableRollbackIfRolloutPending),
						}
					}
					phases = append(phases, phaseArgs)
				}
				ruleArgs.RepairRolloutRule = &cd.AutomationRuleRepairRolloutRuleArgs{
					Id:           pulumi.String(r.Id),
					Phases:       stringArray(r.Phases),
					Jobs:         stringArray(r.Jobs),
					RepairPhases: phases,
				}
			}
			if r := rule.TimedPromoteReleaseRule; r != nil {
				ruleArgs.TimedPromoteReleaseRule = &cd.AutomationRuleTimedPromoteReleaseRuleArgs{
					Id:                  pulumi.String(r.Id),
					Schedule:            pulumi.String(r.Schedule),
					TimeZone:            pulumi.String(r.TimeZone),
					DestinationTargetId: optionalString(r.GetDestinationTargetId().GetValue()),
					DestinationPhase:    optionalString(r.DestinationPhase),
				}
			}
			rules = append(rules, ruleArgs)
		}

		args := &cd.AutomationArgs{
			Project:          optionalString(spec.GetProjectId().GetValue()),
			Location:         pulumi.String(spec.Location),
			DeliveryPipeline: pipeline.Name,
			Name:             pulumi.String(automation.AutomationId),
			Description:      optionalString(automation.Description),
			Labels:           pulumi.ToStringMap(mergeLabels(automation.Labels, locals.AttributionLabels)),
			Annotations:      stringMap(automation.Annotations),
			Suspended:        optionalTrue(automation.Suspended),
			ServiceAccount:   pulumi.String(automation.GetServiceAccount().GetValue()),
			Selector:         &cd.AutomationSelectorArgs{Targets: targets},
			Rules:            rules,
			DeletionPolicy:   optionalString(spec.DeletionPolicy),
		}
		if _, err := cd.NewAutomation(ctx, fmt.Sprintf("%s-automation-%s", resourceName, automation.AutomationId),
			args, pulumi.Provider(gcpProvider)); err != nil {
			return errors.Wrapf(err, "failed to create automation %s", automation.AutomationId)
		}
	}
	return nil
}
