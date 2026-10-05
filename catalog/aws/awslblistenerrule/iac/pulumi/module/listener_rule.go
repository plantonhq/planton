package module

import (
	"strconv"

	"github.com/pkg/errors"
	awslblistenerrulev1alpha1 "github.com/plantonhq/planton/catalog/aws/awslblistenerrule/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/datatypes/stringmaps/convertstringmaps"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// listenerRule provisions the rule. The listener is create-only (moving a
// rule replaces it); priority, conditions, actions, and transforms update in
// place -- priority via a dedicated AWS API, so re-prioritizing rules never
// interrupts traffic.
func listenerRule(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource) error {
	spec := locals.AwsLbListenerRule.Spec
	ruleName := locals.AwsLbListenerRule.Metadata.Name

	actions, err := actionArgs(spec.Actions)
	if err != nil {
		return err
	}

	args := &lb.ListenerRuleArgs{
		ListenerArn: pulumi.String(spec.ListenerArn.GetValue()),
		Conditions:  conditionArgs(spec.Conditions),
		Actions:     actions,
		Tags: convertstringmaps.ConvertGoStringMapToPulumiStringMap(
			stringmaps.AddEntry(locals.AwsTags, "Name", ruleName)),
	}

	// Unset priority lets AWS append after the current highest -- fine for
	// append-only routing; rules that shadow each other set it explicitly.
	if spec.Priority > 0 {
		args.Priority = pulumi.IntPtr(int(spec.Priority))
	}

	if len(spec.Transforms) > 0 {
		args.Transforms = transformArgs(spec.Transforms)
	}

	createdRule, err := lb.NewListenerRule(ctx, ruleName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create listener rule")
	}

	ctx.Export(OpRuleArn, createdRule.Arn)
	// Exported as a string so both engines emit the identical output shape
	// (the output proto field is a string).
	ctx.Export(OpPriority, createdRule.Priority.ApplyT(func(priority int) string {
		return strconv.Itoa(priority)
	}).(pulumi.StringOutput))

	return nil
}

// conditionArgs maps the spec's condition blocks. Exactly one criterion is
// set per block (spec validation guarantees it); the blocks AND together and
// the values inside one block OR together, per AWS semantics.
func conditionArgs(conditions []*awslblistenerrulev1alpha1.AwsLbListenerRuleCondition) lb.ListenerRuleConditionArray {
	result := make(lb.ListenerRuleConditionArray, 0, len(conditions))

	for _, condition := range conditions {
		args := &lb.ListenerRuleConditionArgs{}

		if condition.HostHeader != nil {
			args.HostHeader = &lb.ListenerRuleConditionHostHeaderArgs{
				Values:      pulumi.ToStringArray(condition.HostHeader.Values),
				RegexValues: pulumi.ToStringArray(condition.HostHeader.RegexValues),
			}
		}
		if condition.PathPattern != nil {
			args.PathPattern = &lb.ListenerRuleConditionPathPatternArgs{
				Values:      pulumi.ToStringArray(condition.PathPattern.Values),
				RegexValues: pulumi.ToStringArray(condition.PathPattern.RegexValues),
			}
		}
		if condition.HttpHeader != nil {
			args.HttpHeader = &lb.ListenerRuleConditionHttpHeaderArgs{
				HttpHeaderName: pulumi.String(condition.HttpHeader.HttpHeaderName),
				Values:         pulumi.ToStringArray(condition.HttpHeader.Values),
				RegexValues:    pulumi.ToStringArray(condition.HttpHeader.RegexValues),
			}
		}
		if condition.HttpRequestMethod != nil {
			args.HttpRequestMethod = &lb.ListenerRuleConditionHttpRequestMethodArgs{
				Values: pulumi.ToStringArray(condition.HttpRequestMethod.Values),
			}
		}
		if condition.QueryString != nil {
			pairs := make(lb.ListenerRuleConditionQueryStringArray, 0, len(condition.QueryString.Pairs))
			for _, pair := range condition.QueryString.Pairs {
				pairArgs := &lb.ListenerRuleConditionQueryStringArgs{
					Value: pulumi.String(pair.Value),
				}
				if pair.Key != "" {
					pairArgs.Key = pulumi.StringPtr(pair.Key)
				}
				pairs = append(pairs, pairArgs)
			}
			args.QueryStrings = pairs
		}
		if condition.SourceIp != nil {
			args.SourceIp = &lb.ListenerRuleConditionSourceIpArgs{
				Values: pulumi.ToStringArray(condition.SourceIp.Values),
			}
		}

		result = append(result, args)
	}

	return result
}

// actionArgs maps the spec's action chain onto provider args. The spec
// already guarantees (via CEL) that exactly one configuration message matches
// each action's type, so this is a pure translation.
func actionArgs(actions []*awslblistenerrulev1alpha1.AwsLbListenerRuleAction) (lb.ListenerRuleActionArray, error) {
	result := make(lb.ListenerRuleActionArray, 0, len(actions))

	for i, action := range actions {
		args := &lb.ListenerRuleActionArgs{
			Type: pulumi.String(action.Type),
		}
		if action.Order > 0 {
			args.Order = pulumi.IntPtr(int(action.Order))
		}

		switch action.Type {
		case "forward":
			forward := action.Forward
			// A single unweighted target group uses the simple target_group_arn
			// form; AWS treats the weighted forward block and the simple ARN as
			// different configurations, and the simple form avoids spurious
			// diffs on the common case.
			if len(forward.TargetGroups) == 1 && forward.Stickiness == nil && forward.TargetGroups[0].Weight == 0 {
				args.TargetGroupArn = pulumi.StringPtr(forward.TargetGroups[0].Arn.GetValue())
				break
			}
			forwardArgs := &lb.ListenerRuleActionForwardArgs{}
			targetGroups := make(lb.ListenerRuleActionForwardTargetGroupArray, 0, len(forward.TargetGroups))
			for _, targetGroup := range forward.TargetGroups {
				targetGroupArgs := &lb.ListenerRuleActionForwardTargetGroupArgs{
					Arn: pulumi.String(targetGroup.Arn.GetValue()),
				}
				if targetGroup.Weight > 0 {
					targetGroupArgs.Weight = pulumi.IntPtr(int(targetGroup.Weight))
				}
				targetGroups = append(targetGroups, targetGroupArgs)
			}
			forwardArgs.TargetGroups = targetGroups
			if forward.Stickiness != nil {
				forwardArgs.Stickiness = &lb.ListenerRuleActionForwardStickinessArgs{
					Enabled:  pulumi.BoolPtr(forward.Stickiness.Enabled),
					Duration: pulumi.Int(int(forward.Stickiness.DurationSeconds)),
				}
			}
			args.Forward = forwardArgs

		case "redirect":
			redirect := action.Redirect
			redirectArgs := &lb.ListenerRuleActionRedirectArgs{
				StatusCode: pulumi.String(redirect.StatusCode),
			}
			if redirect.Protocol != "" {
				redirectArgs.Protocol = pulumi.StringPtr(redirect.Protocol)
			}
			if redirect.Port != "" {
				redirectArgs.Port = pulumi.StringPtr(redirect.Port)
			}
			if redirect.Host != "" {
				redirectArgs.Host = pulumi.StringPtr(redirect.Host)
			}
			if redirect.Path != "" {
				redirectArgs.Path = pulumi.StringPtr(redirect.Path)
			}
			if redirect.Query != "" {
				redirectArgs.Query = pulumi.StringPtr(redirect.Query)
			}
			args.Redirect = redirectArgs

		case "fixed-response":
			fixedResponse := action.FixedResponse
			fixedResponseArgs := &lb.ListenerRuleActionFixedResponseArgs{
				ContentType: pulumi.String(fixedResponse.ContentType),
			}
			if fixedResponse.StatusCode != "" {
				fixedResponseArgs.StatusCode = pulumi.StringPtr(fixedResponse.StatusCode)
			}
			if fixedResponse.MessageBody != "" {
				fixedResponseArgs.MessageBody = pulumi.StringPtr(fixedResponse.MessageBody)
			}
			args.FixedResponse = fixedResponseArgs

		case "authenticate-cognito":
			cognito := action.AuthenticateCognito
			cognitoArgs := &lb.ListenerRuleActionAuthenticateCognitoArgs{
				UserPoolArn:      pulumi.String(cognito.UserPoolArn.GetValue()),
				UserPoolClientId: pulumi.String(cognito.UserPoolClientId.GetValue()),
				UserPoolDomain:   pulumi.String(cognito.UserPoolDomain.GetValue()),
			}
			if len(cognito.AuthenticationRequestExtraParams) > 0 {
				cognitoArgs.AuthenticationRequestExtraParams = pulumi.ToStringMap(cognito.AuthenticationRequestExtraParams)
			}
			if cognito.OnUnauthenticatedRequest != "" {
				cognitoArgs.OnUnauthenticatedRequest = pulumi.StringPtr(cognito.OnUnauthenticatedRequest)
			}
			if cognito.Scope != "" {
				cognitoArgs.Scope = pulumi.StringPtr(cognito.Scope)
			}
			if cognito.SessionCookieName != "" {
				cognitoArgs.SessionCookieName = pulumi.StringPtr(cognito.SessionCookieName)
			}
			if cognito.SessionTimeoutSeconds > 0 {
				cognitoArgs.SessionTimeout = pulumi.IntPtr(int(cognito.SessionTimeoutSeconds))
			}
			args.AuthenticateCognito = cognitoArgs

		case "authenticate-oidc":
			oidc := action.AuthenticateOidc
			oidcArgs := &lb.ListenerRuleActionAuthenticateOidcArgs{
				Issuer:                pulumi.String(oidc.Issuer),
				AuthorizationEndpoint: pulumi.String(oidc.AuthorizationEndpoint),
				TokenEndpoint:         pulumi.String(oidc.TokenEndpoint),
				UserInfoEndpoint:      pulumi.String(oidc.UserInfoEndpoint),
				ClientId:              pulumi.String(oidc.ClientId),
				ClientSecret:          pulumi.String(oidc.ClientSecret),
			}
			if len(oidc.AuthenticationRequestExtraParams) > 0 {
				oidcArgs.AuthenticationRequestExtraParams = pulumi.ToStringMap(oidc.AuthenticationRequestExtraParams)
			}
			if oidc.OnUnauthenticatedRequest != "" {
				oidcArgs.OnUnauthenticatedRequest = pulumi.StringPtr(oidc.OnUnauthenticatedRequest)
			}
			if oidc.Scope != "" {
				oidcArgs.Scope = pulumi.StringPtr(oidc.Scope)
			}
			if oidc.SessionCookieName != "" {
				oidcArgs.SessionCookieName = pulumi.StringPtr(oidc.SessionCookieName)
			}
			if oidc.SessionTimeoutSeconds > 0 {
				oidcArgs.SessionTimeout = pulumi.IntPtr(int(oidc.SessionTimeoutSeconds))
			}
			args.AuthenticateOidc = oidcArgs

		case "jwt-validation":
			jwt := action.JwtValidation
			jwtArgs := &lb.ListenerRuleActionJwtValidationArgs{
				Issuer:       pulumi.String(jwt.Issuer),
				JwksEndpoint: pulumi.String(jwt.JwksEndpoint),
			}
			claims := make(lb.ListenerRuleActionJwtValidationAdditionalClaimArray, 0, len(jwt.AdditionalClaims))
			for _, claim := range jwt.AdditionalClaims {
				claims = append(claims, &lb.ListenerRuleActionJwtValidationAdditionalClaimArgs{
					Name:   pulumi.String(claim.Name),
					Format: pulumi.String(claim.Format),
					Values: pulumi.ToStringArray(claim.Values),
				})
			}
			if len(claims) > 0 {
				jwtArgs.AdditionalClaims = claims
			}
			args.JwtValidation = jwtArgs

		default:
			return nil, errors.Errorf("unsupported action type %q at index %d", action.Type, i)
		}

		result = append(result, args)
	}

	return result, nil
}

// transformArgs maps the spec's transforms. Exactly one rewrite config
// matches each transform's type (spec validation guarantees it).
func transformArgs(transforms []*awslblistenerrulev1alpha1.AwsLbListenerRuleTransform) lb.ListenerRuleTransformArray {
	result := make(lb.ListenerRuleTransformArray, 0, len(transforms))

	for _, transform := range transforms {
		args := &lb.ListenerRuleTransformArgs{
			Type: pulumi.String(transform.Type),
		}
		if transform.HostHeaderRewrite != nil {
			args.HostHeaderRewriteConfig = &lb.ListenerRuleTransformHostHeaderRewriteConfigArgs{
				Rewrite: &lb.ListenerRuleTransformHostHeaderRewriteConfigRewriteArgs{
					Regex:   pulumi.String(transform.HostHeaderRewrite.Regex),
					Replace: pulumi.String(transform.HostHeaderRewrite.Replace),
				},
			}
		}
		if transform.UrlRewrite != nil {
			args.UrlRewriteConfig = &lb.ListenerRuleTransformUrlRewriteConfigArgs{
				Rewrite: &lb.ListenerRuleTransformUrlRewriteConfigRewriteArgs{
					Regex:   pulumi.String(transform.UrlRewrite.Regex),
					Replace: pulumi.String(transform.UrlRewrite.Replace),
				},
			}
		}
		result = append(result, args)
	}

	return result
}
