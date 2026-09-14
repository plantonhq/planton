package module

import (
	"github.com/pkg/errors"
	awsconfigrulev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsconfigrule/v1alpha1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cfg"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// rule creates the Config rule - one of four provider resources
// depending on scope and source - plus the optional remediation, and
// exports outputs.
//
// The spec's CELs guarantee exactly one source arm arrives here and
// that org-only / account-only surfaces never mix, so the renders
// below branch on presence without re-validating.
//
// Lifecycle facts the renders below depend on:
//   - organization rules are DIFFERENT provider resources (one per
//     source kind), deployed from the management or delegated-admin
//     account; member accounts receive them automatically;
//   - a custom-lambda rule needs config.amazonaws.com invoke
//     permission on the function BEFORE create (the consumer's
//     contract on the Lambda's policy);
//   - the remediation configuration attaches by RULE NAME - the
//     parent-child dependency below wires create and teardown order;
//   - AWS caps organization rule names at 64 characters (account
//     rules at 128); metadata.name carries the rule name on both
//     engines.
func rule(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec
	name := locals.Target.Metadata.Name

	ctx.Export(OpRuleName, pulumi.String(name))
	// The organization block's switch (on by default once the block is
	// declared, filled by the platform before this runs) decides the scope.
	if spec.Organization == nil || !spec.Organization.GetEnabled() {
		return accountRule(ctx, locals, provider, name)
	}
	return organizationRule(ctx, locals, provider, name)
}

// accountRule renders the account-scoped aws_config_config_rule with
// its optional remediation configuration.
func accountRule(ctx *pulumi.Context, locals *Locals, provider *aws.Provider, name string) error {
	spec := locals.Spec

	sourceArgs := &cfg.RuleSourceArgs{}
	switch {
	case spec.Managed != nil:
		sourceArgs.Owner = pulumi.String("AWS")
		sourceArgs.SourceIdentifier = pulumi.String(spec.Managed.RuleIdentifier)
	case spec.CustomLambda != nil:
		sourceArgs.Owner = pulumi.String("CUSTOM_LAMBDA")
		sourceArgs.SourceIdentifier = pulumi.String(spec.CustomLambda.FunctionArn.GetValue())
		var details cfg.RuleSourceSourceDetailArray
		for _, d := range spec.CustomLambda.SourceDetails {
			detailArgs := &cfg.RuleSourceSourceDetailArgs{
				MessageType: pulumi.String(d.MessageType),
			}
			if d.MaximumExecutionFrequency != "" {
				detailArgs.MaximumExecutionFrequency = pulumi.String(d.MaximumExecutionFrequency)
			}
			details = append(details, detailArgs)
		}
		if len(details) > 0 {
			sourceArgs.SourceDetails = details
		}
	case spec.CustomPolicy != nil:
		sourceArgs.Owner = pulumi.String("CUSTOM_POLICY")
		sourceArgs.CustomPolicyDetails = &cfg.RuleSourceCustomPolicyDetailsArgs{
			PolicyRuntime:          pulumi.String(spec.CustomPolicy.PolicyRuntime),
			PolicyText:             pulumi.String(spec.CustomPolicy.PolicyText),
			EnableDebugLogDelivery: pulumi.Bool(spec.CustomPolicy.EnableDebugLogDelivery),
		}
		// Guard rules evaluate on configuration changes; the provider
		// requires the trigger detail explicitly on custom-policy
		// sources.
		sourceArgs.SourceDetails = cfg.RuleSourceSourceDetailArray{
			&cfg.RuleSourceSourceDetailArgs{
				MessageType: pulumi.String("ConfigurationItemChangeNotification"),
			},
		}
	}

	args := &cfg.RuleArgs{
		Name:   pulumi.String(name),
		Source: sourceArgs,
		Tags:   pulumi.ToStringMap(locals.AwsTags),
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.InputParameters != "" {
		args.InputParameters = pulumi.String(spec.InputParameters)
	}
	if spec.MaximumExecutionFrequency != "" {
		args.MaximumExecutionFrequency = pulumi.String(spec.MaximumExecutionFrequency)
	}
	if spec.Scope != nil {
		scopeArgs := &cfg.RuleScopeArgs{}
		if spec.Scope.ComplianceResourceId != "" {
			scopeArgs.ComplianceResourceId = pulumi.String(spec.Scope.ComplianceResourceId)
		}
		if len(spec.Scope.ComplianceResourceTypes) > 0 {
			scopeArgs.ComplianceResourceTypes = pulumi.ToStringArray(spec.Scope.ComplianceResourceTypes)
		}
		if spec.Scope.TagKey != "" {
			scopeArgs.TagKey = pulumi.String(spec.Scope.TagKey)
		}
		if spec.Scope.TagValue != "" {
			scopeArgs.TagValue = pulumi.String(spec.Scope.TagValue)
		}
		args.Scope = scopeArgs
	}
	var evaluationModes cfg.RuleEvaluationModeArray
	for _, m := range spec.EvaluationModes {
		evaluationModes = append(evaluationModes, &cfg.RuleEvaluationModeArgs{
			Mode: pulumi.String(m),
		})
	}
	if len(evaluationModes) > 0 {
		args.EvaluationModes = evaluationModes
	}

	createdRule, err := cfg.NewRule(ctx, "rule", args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create config rule")
	}

	if spec.Remediation != nil {
		if err := remediation(ctx, spec.Remediation, createdRule, provider); err != nil {
			return err
		}
	} else {
		ctx.Export(OpRemediationArn, pulumi.String(""))
	}

	ctx.Export(OpRuleArn, createdRule.Arn)
	ctx.Export(OpRuleId, createdRule.RuleId)
	return nil
}

// remediation attaches the SSM-document remediation to the rule.
// Referencing the rule's Name output (not the raw string) wires
// create AND destroy ordering: the remediation goes first on
// teardown.
func remediation(ctx *pulumi.Context, r *awsconfigrulev1alpha1.AwsConfigRuleRemediation, createdRule *cfg.Rule, provider *aws.Provider) error {
	args := &cfg.RemediationConfigurationArgs{
		ConfigRuleName: createdRule.Name,
		TargetType:     pulumi.String("SSM_DOCUMENT"),
		TargetId:       pulumi.String(r.TargetId),
	}
	if r.TargetVersion != "" {
		args.TargetVersion = pulumi.String(r.TargetVersion)
	}
	if r.ResourceType != "" {
		args.ResourceType = pulumi.String(r.ResourceType)
	}
	if r.Automatic {
		args.Automatic = pulumi.Bool(true)
	}
	if r.MaximumAutomaticAttempts > 0 {
		args.MaximumAutomaticAttempts = pulumi.Int(int(r.MaximumAutomaticAttempts))
	}
	if r.RetryAttemptSeconds > 0 {
		args.RetryAttemptSeconds = pulumi.Int(int(r.RetryAttemptSeconds))
	}
	var parameters cfg.RemediationConfigurationParameterArray
	for _, p := range r.Parameters {
		paramArgs := &cfg.RemediationConfigurationParameterArgs{
			Name: pulumi.String(p.Name),
		}
		if p.ResourceValue != "" {
			paramArgs.ResourceValue = pulumi.String(p.ResourceValue)
		}
		if p.StaticValue != "" {
			paramArgs.StaticValue = pulumi.String(p.StaticValue)
		}
		if len(p.StaticValues) > 0 {
			paramArgs.StaticValues = pulumi.ToStringArray(p.StaticValues)
		}
		parameters = append(parameters, paramArgs)
	}
	if len(parameters) > 0 {
		args.Parameters = parameters
	}
	if r.ConcurrentExecutionRatePercentage > 0 || r.ErrorPercentage > 0 {
		ssmControls := &cfg.RemediationConfigurationExecutionControlsSsmControlsArgs{}
		if r.ConcurrentExecutionRatePercentage > 0 {
			ssmControls.ConcurrentExecutionRatePercentage = pulumi.Int(int(r.ConcurrentExecutionRatePercentage))
		}
		if r.ErrorPercentage > 0 {
			ssmControls.ErrorPercentage = pulumi.Int(int(r.ErrorPercentage))
		}
		args.ExecutionControls = &cfg.RemediationConfigurationExecutionControlsArgs{
			SsmControls: ssmControls,
		}
	}

	createdRemediation, err := cfg.NewRemediationConfiguration(ctx, "remediation", args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create remediation configuration")
	}
	ctx.Export(OpRemediationArn, createdRemediation.Arn)
	return nil
}

// organizationRule renders one of the three organization-scoped rule
// resources.
func organizationRule(ctx *pulumi.Context, locals *Locals, provider *aws.Provider, name string) error {
	spec := locals.Spec
	org := spec.Organization

	// Organization rules have no rule_id and carry no remediation.
	ctx.Export(OpRuleId, pulumi.String(""))
	ctx.Export(OpRemediationArn, pulumi.String(""))

	switch {
	case spec.Managed != nil:
		args := &cfg.OrganizationManagedRuleArgs{
			Name:           pulumi.String(name),
			RuleIdentifier: pulumi.String(spec.Managed.RuleIdentifier),
		}
		applyOrgCommon(&orgCommon{
			description:               &args.Description,
			inputParameters:           &args.InputParameters,
			maximumExecutionFrequency: &args.MaximumExecutionFrequency,
			excludedAccounts:          &args.ExcludedAccounts,
			resourceIdScope:           &args.ResourceIdScope,
			resourceTypesScopes:       &args.ResourceTypesScopes,
			tagKeyScope:               &args.TagKeyScope,
			tagValueScope:             &args.TagValueScope,
		}, spec)
		created, err := cfg.NewOrganizationManagedRule(ctx, "org-managed-rule", args, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "create organization managed rule")
		}
		ctx.Export(OpRuleArn, created.Arn)

	case spec.CustomLambda != nil:
		args := &cfg.OrganizationCustomRuleArgs{
			Name:              pulumi.String(name),
			LambdaFunctionArn: pulumi.String(spec.CustomLambda.FunctionArn.GetValue()),
			TriggerTypes:      pulumi.ToStringArray(org.TriggerTypes),
		}
		applyOrgCommon(&orgCommon{
			description:               &args.Description,
			inputParameters:           &args.InputParameters,
			maximumExecutionFrequency: &args.MaximumExecutionFrequency,
			excludedAccounts:          &args.ExcludedAccounts,
			resourceIdScope:           &args.ResourceIdScope,
			resourceTypesScopes:       &args.ResourceTypesScopes,
			tagKeyScope:               &args.TagKeyScope,
			tagValueScope:             &args.TagValueScope,
		}, spec)
		created, err := cfg.NewOrganizationCustomRule(ctx, "org-custom-rule", args, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "create organization custom rule")
		}
		ctx.Export(OpRuleArn, created.Arn)

	case spec.CustomPolicy != nil:
		args := &cfg.OrganizationCustomPolicyRuleArgs{
			Name:          pulumi.String(name),
			PolicyRuntime: pulumi.String(spec.CustomPolicy.PolicyRuntime),
			PolicyText:    pulumi.String(spec.CustomPolicy.PolicyText),
			TriggerTypes:  pulumi.ToStringArray(org.TriggerTypes),
		}
		if len(org.DebugLogDeliveryAccounts) > 0 {
			args.DebugLogDeliveryAccounts = pulumi.ToStringArray(org.DebugLogDeliveryAccounts)
		}
		applyOrgCommon(&orgCommon{
			description:               &args.Description,
			inputParameters:           &args.InputParameters,
			maximumExecutionFrequency: &args.MaximumExecutionFrequency,
			excludedAccounts:          &args.ExcludedAccounts,
			resourceIdScope:           &args.ResourceIdScope,
			resourceTypesScopes:       &args.ResourceTypesScopes,
			tagKeyScope:               &args.TagKeyScope,
			tagValueScope:             &args.TagValueScope,
		}, spec)
		created, err := cfg.NewOrganizationCustomPolicyRule(ctx, "org-custom-policy-rule", args, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "create organization custom policy rule")
		}
		ctx.Export(OpRuleArn, created.Arn)
	}

	return nil
}

// orgCommon collects the argument slots the three organization rule
// resources share, so the mapping below is written exactly once.
type orgCommon struct {
	description               *pulumi.StringPtrInput
	inputParameters           *pulumi.StringPtrInput
	maximumExecutionFrequency *pulumi.StringPtrInput
	excludedAccounts          *pulumi.StringArrayInput
	resourceIdScope           *pulumi.StringPtrInput
	resourceTypesScopes       *pulumi.StringArrayInput
	tagKeyScope               *pulumi.StringPtrInput
	tagValueScope             *pulumi.StringPtrInput
}

func applyOrgCommon(slots *orgCommon, spec *awsconfigrulev1alpha1.AwsConfigRuleSpec) {
	org := spec.Organization
	if spec.Description != "" {
		*slots.description = pulumi.String(spec.Description)
	}
	if spec.InputParameters != "" {
		*slots.inputParameters = pulumi.String(spec.InputParameters)
	}
	if spec.MaximumExecutionFrequency != "" {
		*slots.maximumExecutionFrequency = pulumi.String(spec.MaximumExecutionFrequency)
	}
	if len(org.ExcludedAccounts) > 0 {
		*slots.excludedAccounts = pulumi.ToStringArray(org.ExcludedAccounts)
	}
	if spec.Scope != nil {
		if spec.Scope.ComplianceResourceId != "" {
			*slots.resourceIdScope = pulumi.String(spec.Scope.ComplianceResourceId)
		}
		if len(spec.Scope.ComplianceResourceTypes) > 0 {
			*slots.resourceTypesScopes = pulumi.ToStringArray(spec.Scope.ComplianceResourceTypes)
		}
		if spec.Scope.TagKey != "" {
			*slots.tagKeyScope = pulumi.String(spec.Scope.TagKey)
		}
		if spec.Scope.TagValue != "" {
			*slots.tagValueScope = pulumi.String(spec.Scope.TagValue)
		}
	}
}
