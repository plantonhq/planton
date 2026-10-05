package module

import (
	"github.com/pkg/errors"
	azurefrontdoorrulesetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorruleset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleSetIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorRuleSet.Spec

	// The rule set itself carries no properties -- it is the named
	// container the rules nest under and routes attach by ARM id.
	createdRuleSet, err := cdn.NewFrontdoorRuleSet(ctx,
		spec.RuleSetName,
		&cdn.FrontdoorRuleSetArgs{
			Name:                  pulumi.String(spec.RuleSetName),
			CdnFrontdoorProfileId: pulumi.String(locals.ProfileId),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door rule set %s", spec.RuleSetName)
	}

	// One provider resource per rule (ARM models rules as children of
	// the set). Evaluation position is the rule's own `order` field, so
	// creation order here carries no meaning; the parent reference gives
	// the dependency edge.
	for _, rule := range spec.Rules {
		ruleArgs := &cdn.FrontdoorRuleArgs{
			Name:                  pulumi.String(rule.Name),
			CdnFrontdoorRuleSetId: createdRuleSet.ID(),
			Order:                 pulumi.Int(int(rule.Order)),
			Actions:               buildRuleActions(rule.Actions),
		}

		// Sent only when chosen: ARM defaults behavior-on-match to
		// Continue (IaC inputs never materialize proto defaults).
		if rule.BehaviorOnMatch != azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleBehaviorOnMatch_azure_front_door_rule_behavior_on_match_unspecified {
			ruleArgs.BehaviorOnMatch = pulumi.String(behaviorOnMatchStrings[rule.BehaviorOnMatch])
		}

		if rule.Conditions != nil {
			ruleArgs.Conditions = buildRuleConditions(rule.Conditions)
		}

		if _, err := cdn.NewFrontdoorRule(ctx,
			rule.Name,
			ruleArgs,
			pulumi.Provider(azureProvider)); err != nil {
			return errors.Wrapf(err, "failed to create front door rule %s", rule.Name)
		}
	}

	// Export outputs. rule_set_id is what AzureFrontDoorRoute's
	// rule_set_ids references; the rules deliberately export no ids
	// (nothing references an individual rule).
	ctx.Export(OpRuleSetId, createdRuleSet.ID())
	ctx.Export(OpRuleSetName, createdRuleSet.Name)

	return nil
}

// buildRuleActions converts the spec's actions message into the
// provider's actions block. The singular redirect/rewrite/override spec
// fields structurally guarantee the provider's only-once contracts.
func buildRuleActions(actions *azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleActions) *cdn.FrontdoorRuleActionsArgs {
	args := &cdn.FrontdoorRuleActionsArgs{}

	if actions.UrlRedirect != nil {
		redirect := &cdn.FrontdoorRuleActionsUrlRedirectActionArgs{
			RedirectType: pulumi.String(redirectTypeStrings[actions.UrlRedirect.RedirectType]),
			// Empty preserves the incoming host -- the provider requires
			// the field to be present, so the empty string is sent as-is.
			DestinationHostname: pulumi.String(actions.UrlRedirect.DestinationHostname),
		}
		// Sent only when chosen: the provider defaults the redirect
		// scheme to MatchRequest. NOTE the redirect dialect: the shared
		// protocol enum maps to Http/Https here but HttpOnly/HttpsOnly on
		// the route-configuration override.
		if actions.UrlRedirect.RedirectProtocol != azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleForwardingProtocol_azure_front_door_rule_forwarding_protocol_unspecified {
			redirect.RedirectProtocol = pulumi.String(redirectProtocolStrings[actions.UrlRedirect.RedirectProtocol])
		}
		if actions.UrlRedirect.DestinationPath != "" {
			redirect.DestinationPath = pulumi.String(actions.UrlRedirect.DestinationPath)
		}
		if actions.UrlRedirect.QueryString != "" {
			redirect.QueryString = pulumi.String(actions.UrlRedirect.QueryString)
		}
		if actions.UrlRedirect.DestinationFragment != "" {
			redirect.DestinationFragment = pulumi.String(actions.UrlRedirect.DestinationFragment)
		}
		args.UrlRedirectAction = redirect
	}

	if actions.UrlRewrite != nil {
		args.UrlRewriteAction = &cdn.FrontdoorRuleActionsUrlRewriteActionArgs{
			SourcePattern:         pulumi.String(actions.UrlRewrite.SourcePattern),
			Destination:           pulumi.String(actions.UrlRewrite.Destination),
			PreserveUnmatchedPath: pulumi.Bool(actions.UrlRewrite.PreserveUnmatchedPath),
		}
	}

	if len(actions.RequestHeaders) > 0 {
		headerActions := make(cdn.FrontdoorRuleActionsRequestHeaderActionArray, 0, len(actions.RequestHeaders))
		for _, header := range actions.RequestHeaders {
			headerAction := &cdn.FrontdoorRuleActionsRequestHeaderActionArgs{
				HeaderAction: pulumi.String(headerActionStrings[header.HeaderAction]),
				HeaderName:   pulumi.String(header.HeaderName),
			}
			// DELETE must carry no value (spec-enforced); the provider
			// rejects an empty value on Append/Overwrite, so it is sent
			// only when present.
			if header.Value != "" {
				headerAction.Value = pulumi.String(header.Value)
			}
			headerActions = append(headerActions, headerAction)
		}
		args.RequestHeaderActions = headerActions
	}

	if len(actions.ResponseHeaders) > 0 {
		headerActions := make(cdn.FrontdoorRuleActionsResponseHeaderActionArray, 0, len(actions.ResponseHeaders))
		for _, header := range actions.ResponseHeaders {
			headerAction := &cdn.FrontdoorRuleActionsResponseHeaderActionArgs{
				HeaderAction: pulumi.String(headerActionStrings[header.HeaderAction]),
				HeaderName:   pulumi.String(header.HeaderName),
			}
			if header.Value != "" {
				headerAction.Value = pulumi.String(header.Value)
			}
			headerActions = append(headerActions, headerAction)
		}
		args.ResponseHeaderActions = headerActions
	}

	if actions.RouteConfigurationOverride != nil {
		override := &cdn.FrontdoorRuleActionsRouteConfigurationOverrideActionArgs{
			// cache_behavior is spec-required: every override makes an
			// explicit cache decision (Disabled turns caching off).
			CacheBehavior: pulumi.String(cacheBehaviorStrings[actions.RouteConfigurationOverride.CacheBehavior]),
		}
		if actions.RouteConfigurationOverride.OriginGroupId != nil {
			override.CdnFrontdoorOriginGroupId = pulumi.String(actions.RouteConfigurationOverride.OriginGroupId.GetValue())
			// The override dialect of the shared protocol enum
			// (HttpOnly/HttpsOnly vs the redirect's Http/Https).
			override.ForwardingProtocol = pulumi.String(overrideForwardingProtocolStrings[actions.RouteConfigurationOverride.ForwardingProtocol])
		}
		if actions.RouteConfigurationOverride.CacheDuration != "" {
			override.CacheDuration = pulumi.String(actions.RouteConfigurationOverride.CacheDuration)
		}
		if actions.RouteConfigurationOverride.QueryStringCachingBehavior != azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleQueryStringCachingBehavior_azure_front_door_rule_query_string_caching_behavior_unspecified {
			override.QueryStringCachingBehavior = pulumi.String(queryStringCachingBehaviorStrings[actions.RouteConfigurationOverride.QueryStringCachingBehavior])
		}
		if len(actions.RouteConfigurationOverride.QueryStringParameters) > 0 {
			override.QueryStringParameters = pulumi.ToStringArray(actions.RouteConfigurationOverride.QueryStringParameters)
		}
		// Sent only when set: the provider treats the flag as part of
		// the cache configuration, so absence lets the cache-disabled
		// path skip it cleanly.
		if actions.RouteConfigurationOverride.CompressionEnabled != nil {
			override.CompressionEnabled = pulumi.Bool(actions.RouteConfigurationOverride.GetCompressionEnabled())
		}
		args.RouteConfigurationOverrideAction = override
	}

	return args
}

// buildRuleConditions converts the spec's conditions message into the
// provider's conditions block. Operator enums translate through the
// shared wire map; the address conditions materialize their documented
// IPMatch default because IaC inputs never carry proto defaults.
func buildRuleConditions(conditions *azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleConditions) *cdn.FrontdoorRuleConditionsArgs {
	args := &cdn.FrontdoorRuleConditionsArgs{}

	if len(conditions.RemoteAddress) > 0 {
		list := make(cdn.FrontdoorRuleConditionsRemoteAddressConditionArray, 0, len(conditions.RemoteAddress))
		for _, condition := range conditions.RemoteAddress {
			list = append(list, &cdn.FrontdoorRuleConditionsRemoteAddressConditionArgs{
				Operator:        pulumi.String(addressOperator(condition.Operator)),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
			})
		}
		args.RemoteAddressConditions = list
	}

	if len(conditions.RequestMethod) > 0 {
		list := make(cdn.FrontdoorRuleConditionsRequestMethodConditionArray, 0, len(conditions.RequestMethod))
		for _, condition := range conditions.RequestMethod {
			list = append(list, &cdn.FrontdoorRuleConditionsRequestMethodConditionArgs{
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
			})
		}
		args.RequestMethodConditions = list
	}

	if len(conditions.QueryString) > 0 {
		list := make(cdn.FrontdoorRuleConditionsQueryStringConditionArray, 0, len(conditions.QueryString))
		for _, condition := range conditions.QueryString {
			list = append(list, &cdn.FrontdoorRuleConditionsQueryStringConditionArgs{
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.QueryStringConditions = list
	}

	if len(conditions.PostArgs) > 0 {
		list := make(cdn.FrontdoorRuleConditionsPostArgsConditionArray, 0, len(conditions.PostArgs))
		for _, condition := range conditions.PostArgs {
			list = append(list, &cdn.FrontdoorRuleConditionsPostArgsConditionArgs{
				PostArgsName:    pulumi.String(condition.PostArgsName),
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.PostArgsConditions = list
	}

	if len(conditions.RequestUri) > 0 {
		list := make(cdn.FrontdoorRuleConditionsRequestUriConditionArray, 0, len(conditions.RequestUri))
		for _, condition := range conditions.RequestUri {
			list = append(list, &cdn.FrontdoorRuleConditionsRequestUriConditionArgs{
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.RequestUriConditions = list
	}

	if len(conditions.RequestHeader) > 0 {
		list := make(cdn.FrontdoorRuleConditionsRequestHeaderConditionArray, 0, len(conditions.RequestHeader))
		for _, condition := range conditions.RequestHeader {
			list = append(list, &cdn.FrontdoorRuleConditionsRequestHeaderConditionArgs{
				HeaderName:      pulumi.String(condition.HeaderName),
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.RequestHeaderConditions = list
	}

	if len(conditions.RequestBody) > 0 {
		list := make(cdn.FrontdoorRuleConditionsRequestBodyConditionArray, 0, len(conditions.RequestBody))
		for _, condition := range conditions.RequestBody {
			list = append(list, &cdn.FrontdoorRuleConditionsRequestBodyConditionArgs{
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.RequestBodyConditions = list
	}

	if len(conditions.RequestScheme) > 0 {
		list := make(cdn.FrontdoorRuleConditionsRequestSchemeConditionArray, 0, len(conditions.RequestScheme))
		for _, condition := range conditions.RequestScheme {
			// The scheme is a single value (the bridge flattens the
			// provider's one-element list); unset deploys HTTP, the
			// documented default.
			scheme := "HTTP"
			if condition.MatchValue != nil {
				scheme = condition.GetMatchValue()
			}
			list = append(list, &cdn.FrontdoorRuleConditionsRequestSchemeConditionArgs{
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.String(scheme),
			})
		}
		args.RequestSchemeConditions = list
	}

	if len(conditions.UrlPath) > 0 {
		list := make(cdn.FrontdoorRuleConditionsUrlPathConditionArray, 0, len(conditions.UrlPath))
		for _, condition := range conditions.UrlPath {
			list = append(list, &cdn.FrontdoorRuleConditionsUrlPathConditionArgs{
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.UrlPathConditions = list
	}

	if len(conditions.UrlFileExtension) > 0 {
		list := make(cdn.FrontdoorRuleConditionsUrlFileExtensionConditionArray, 0, len(conditions.UrlFileExtension))
		for _, condition := range conditions.UrlFileExtension {
			list = append(list, &cdn.FrontdoorRuleConditionsUrlFileExtensionConditionArgs{
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.UrlFileExtensionConditions = list
	}

	if len(conditions.UrlFilename) > 0 {
		list := make(cdn.FrontdoorRuleConditionsUrlFilenameConditionArray, 0, len(conditions.UrlFilename))
		for _, condition := range conditions.UrlFilename {
			list = append(list, &cdn.FrontdoorRuleConditionsUrlFilenameConditionArgs{
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.UrlFilenameConditions = list
	}

	if len(conditions.HttpVersion) > 0 {
		list := make(cdn.FrontdoorRuleConditionsHttpVersionConditionArray, 0, len(conditions.HttpVersion))
		for _, condition := range conditions.HttpVersion {
			list = append(list, &cdn.FrontdoorRuleConditionsHttpVersionConditionArgs{
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
			})
		}
		args.HttpVersionConditions = list
	}

	if len(conditions.Cookies) > 0 {
		list := make(cdn.FrontdoorRuleConditionsCookiesConditionArray, 0, len(conditions.Cookies))
		for _, condition := range conditions.Cookies {
			list = append(list, &cdn.FrontdoorRuleConditionsCookiesConditionArgs{
				CookieName:      pulumi.String(condition.CookieName),
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.CookiesConditions = list
	}

	if len(conditions.IsDevice) > 0 {
		list := make(cdn.FrontdoorRuleConditionsIsDeviceConditionArray, 0, len(conditions.IsDevice))
		for _, condition := range conditions.IsDevice {
			// Single value (the bridge flattens the provider's
			// one-element list); the spec requires it.
			list = append(list, &cdn.FrontdoorRuleConditionsIsDeviceConditionArgs{
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.String(condition.MatchValue),
			})
		}
		args.IsDeviceConditions = list
	}

	if len(conditions.SocketAddress) > 0 {
		list := make(cdn.FrontdoorRuleConditionsSocketAddressConditionArray, 0, len(conditions.SocketAddress))
		for _, condition := range conditions.SocketAddress {
			list = append(list, &cdn.FrontdoorRuleConditionsSocketAddressConditionArgs{
				Operator:        pulumi.String(addressOperator(condition.Operator)),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
			})
		}
		args.SocketAddressConditions = list
	}

	if len(conditions.ClientPort) > 0 {
		list := make(cdn.FrontdoorRuleConditionsClientPortConditionArray, 0, len(conditions.ClientPort))
		for _, condition := range conditions.ClientPort {
			list = append(list, &cdn.FrontdoorRuleConditionsClientPortConditionArgs{
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
			})
		}
		args.ClientPortConditions = list
	}

	if len(conditions.ServerPort) > 0 {
		list := make(cdn.FrontdoorRuleConditionsServerPortConditionArray, 0, len(conditions.ServerPort))
		for _, condition := range conditions.ServerPort {
			list = append(list, &cdn.FrontdoorRuleConditionsServerPortConditionArgs{
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
			})
		}
		args.ServerPortConditions = list
	}

	if len(conditions.HostName) > 0 {
		list := make(cdn.FrontdoorRuleConditionsHostNameConditionArray, 0, len(conditions.HostName))
		for _, condition := range conditions.HostName {
			list = append(list, &cdn.FrontdoorRuleConditionsHostNameConditionArgs{
				Operator:        pulumi.String(operatorStrings[condition.Operator]),
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
				Transforms:      transformArray(condition.Transforms),
			})
		}
		args.HostNameConditions = list
	}

	if len(conditions.SslProtocol) > 0 {
		list := make(cdn.FrontdoorRuleConditionsSslProtocolConditionArray, 0, len(conditions.SslProtocol))
		for _, condition := range conditions.SslProtocol {
			list = append(list, &cdn.FrontdoorRuleConditionsSslProtocolConditionArgs{
				NegateCondition: pulumi.Bool(condition.NegateCondition),
				MatchValues:     pulumi.ToStringArray(condition.MatchValues),
			})
		}
		args.SslProtocolConditions = list
	}

	return args
}

// addressOperator maps an address-condition operator, materializing the
// documented IPMatch default when unspecified (IaC inputs never carry
// proto defaults).
func addressOperator(operator azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator) string {
	if operator == azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_azure_front_door_rule_operator_unspecified {
		return "IPMatch"
	}
	return operatorStrings[operator]
}

// transformArray maps transform enums to their ARM values.
func transformArray(transforms []azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleTransform) pulumi.StringArrayInput {
	if len(transforms) == 0 {
		return nil
	}
	values := make(pulumi.StringArray, 0, len(transforms))
	for _, transform := range transforms {
		values = append(values, pulumi.String(transformStrings[transform]))
	}
	return values
}
