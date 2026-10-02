package module

import (
	"github.com/pkg/errors"
	gcpdeploypolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdeploypolicy/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/clouddeploy"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// deployPolicy enables the Cloud Deploy API and creates the policy with its
// rules and selectors.
func deployPolicy(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDeployPolicy.Spec
	resourceName := locals.GcpDeployPolicy.Metadata.Name
	project := spec.GetProjectId().GetValue()

	// The Cloud Deploy API. DisableOnDestroy stays false: tearing down one
	// policy must never disable the API for every pipeline in the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("clouddeploy.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdApi, err := projects.NewService(ctx, "gcpcdpol-clouddeploy.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable clouddeploy.googleapis.com")
	}

	// The policy ID defaults to metadata.name -- identical to the Terraform
	// module.
	deployPolicyId := spec.DeployPolicyId
	if deployPolicyId == "" {
		deployPolicyId = resourceName
	}

	args := &clouddeploy.DeployPolicyArgs{
		Project:        optionalString(project),
		Location:       pulumi.String(spec.Location),
		Name:           pulumi.String(deployPolicyId),
		Description:    optionalString(spec.Description),
		Labels:         pulumi.ToStringMap(locals.GcpLabels),
		DeletionPolicy: optionalString(spec.DeletionPolicy),
		Rules:          rules(spec.Rules),
		Selectors:      selectors(spec.Selectors),
	}
	if len(spec.Annotations) > 0 {
		args.Annotations = pulumi.ToStringMap(spec.Annotations)
	}
	if spec.Suspended {
		args.Suspended = pulumi.BoolPtr(true)
	}

	createdPolicy, err := clouddeploy.NewDeployPolicy(ctx, resourceName, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create deploy policy")
	}

	ctx.Export(OpName, createdPolicy.ID())
	ctx.Export(OpDeployPolicyId, createdPolicy.Name)
	ctx.Export(OpUid, createdPolicy.Uid)
	return nil
}

func rules(specRules []*gcpdeploypolicyv1alpha1.GcpDeployPolicyRule) clouddeploy.DeployPolicyRuleArray {
	result := clouddeploy.DeployPolicyRuleArray{}
	for _, rule := range specRules {
		ruleArgs := &clouddeploy.DeployPolicyRuleArgs{}
		if r := rule.RolloutRestriction; r != nil {
			ruleArgs.RolloutRestriction = &clouddeploy.DeployPolicyRuleRolloutRestrictionArgs{
				Id:          pulumi.String(r.Id),
				Actions:     optionalStrings(r.Actions),
				Invokers:    optionalStrings(r.Invokers),
				TimeWindows: timeWindows(r.TimeWindows),
			}
		}
		result = append(result, ruleArgs)
	}
	return result
}

func timeWindows(w *gcpdeploypolicyv1alpha1.GcpDeployPolicyTimeWindows) clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsPtrInput {
	if w == nil {
		return nil
	}
	args := &clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsArgs{
		TimeZone: pulumi.String(w.TimeZone),
	}
	if len(w.OneTimeWindows) > 0 {
		oneTime := clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsOneTimeWindowArray{}
		for _, window := range w.OneTimeWindows {
			oneTime = append(oneTime, &clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsOneTimeWindowArgs{
				StartDate: &clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsOneTimeWindowStartDateArgs{
					Year:  optionalInt(window.GetStartDate().GetYear()),
					Month: optionalInt(window.GetStartDate().GetMonth()),
					Day:   optionalInt(window.GetStartDate().GetDay()),
				},
				StartTime: &clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsOneTimeWindowStartTimeArgs{
					Hours:   optionalInt(window.GetStartTime().GetHours()),
					Minutes: optionalInt(window.GetStartTime().GetMinutes()),
					Seconds: optionalInt(window.GetStartTime().GetSeconds()),
					Nanos:   optionalInt(window.GetStartTime().GetNanos()),
				},
				EndDate: &clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsOneTimeWindowEndDateArgs{
					Year:  optionalInt(window.GetEndDate().GetYear()),
					Month: optionalInt(window.GetEndDate().GetMonth()),
					Day:   optionalInt(window.GetEndDate().GetDay()),
				},
				EndTime: &clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsOneTimeWindowEndTimeArgs{
					Hours:   optionalInt(window.GetEndTime().GetHours()),
					Minutes: optionalInt(window.GetEndTime().GetMinutes()),
					Seconds: optionalInt(window.GetEndTime().GetSeconds()),
					Nanos:   optionalInt(window.GetEndTime().GetNanos()),
				},
			})
		}
		args.OneTimeWindows = oneTime
	}
	if len(w.WeeklyWindows) > 0 {
		weekly := clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsWeeklyWindowArray{}
		for _, window := range w.WeeklyWindows {
			windowArgs := &clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsWeeklyWindowArgs{
				DaysOfWeeks: optionalStrings(window.DaysOfWeek),
			}
			if t := window.StartTime; t != nil {
				windowArgs.StartTime = &clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsWeeklyWindowStartTimeArgs{
					Hours:   optionalInt(t.Hours),
					Minutes: optionalInt(t.Minutes),
					Seconds: optionalInt(t.Seconds),
					Nanos:   optionalInt(t.Nanos),
				}
			}
			if t := window.EndTime; t != nil {
				windowArgs.EndTime = &clouddeploy.DeployPolicyRuleRolloutRestrictionTimeWindowsWeeklyWindowEndTimeArgs{
					Hours:   optionalInt(t.Hours),
					Minutes: optionalInt(t.Minutes),
					Seconds: optionalInt(t.Seconds),
					Nanos:   optionalInt(t.Nanos),
				}
			}
			weekly = append(weekly, windowArgs)
		}
		args.WeeklyWindows = weekly
	}
	return args
}

// selectors builds the policy's selectors. Their labels are match
// criteria, so they carry exactly what the spec declares -- never the
// attribution labels.
func selectors(specSelectors []*gcpdeploypolicyv1alpha1.GcpDeployPolicySelector) clouddeploy.DeployPolicySelectorArray {
	result := clouddeploy.DeployPolicySelectorArray{}
	for _, selector := range specSelectors {
		selectorArgs := &clouddeploy.DeployPolicySelectorArgs{}
		if p := selector.DeliveryPipeline; p != nil {
			pipelineArgs := &clouddeploy.DeployPolicySelectorDeliveryPipelineArgs{
				Id: optionalString(p.GetId().GetValue()),
			}
			if len(p.Labels) > 0 {
				pipelineArgs.Labels = pulumi.ToStringMap(p.Labels)
			}
			selectorArgs.DeliveryPipeline = pipelineArgs
		}
		if t := selector.Target; t != nil {
			targetArgs := &clouddeploy.DeployPolicySelectorTargetArgs{
				Id: optionalString(t.GetId().GetValue()),
			}
			if len(t.Labels) > 0 {
				targetArgs.Labels = pulumi.ToStringMap(t.Labels)
			}
			selectorArgs.Target = targetArgs
		}
		result = append(result, selectorArgs)
	}
	return result
}

// optionalString returns nil for an empty value so the provider default
// applies -- the Terraform module's `!= "" ? value : null`.
func optionalString(value string) pulumi.StringPtrInput {
	if value == "" {
		return nil
	}
	return pulumi.StringPtr(value)
}

// optionalInt returns nil for zero -- the Terraform module's
// `!= 0 ? value : null`.
func optionalInt(value int32) pulumi.IntPtrInput {
	if value == 0 {
		return nil
	}
	return pulumi.IntPtr(int(value))
}

// optionalStrings returns nil for an empty list -- the Terraform module's
// `length(...) > 0 ? value : null`.
func optionalStrings(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}
