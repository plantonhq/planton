package module

import (
	"github.com/pkg/errors"
	azurefrontdoorroutev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorroute/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefrontdoorroutev1alpha1.AzureFrontDoorRouteIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorRoute.Spec

	supportedProtocols := make(pulumi.StringArray, 0, len(spec.SupportedProtocols))
	for _, protocol := range spec.SupportedProtocols {
		supportedProtocols = append(supportedProtocols, pulumi.String(protocolStrings[protocol]))
	}

	// The origin-leg protocol default (MatchRequest) is materialized here
	// because IaC inputs never carry proto defaults.
	forwardingProtocol := forwardingProtocolStrings[spec.ForwardingProtocol]
	if forwardingProtocol == "" {
		forwardingProtocol = "MatchRequest"
	}

	// The route addresses its ARM parent by the endpoint's full id and
	// its destination by the origin group's id -- the provider derives
	// resource group / profile / endpoint names from them. ARM does not
	// support tags on routes.
	routeArgs := &cdn.FrontdoorRouteArgs{
		Name:                      pulumi.String(spec.RouteName),
		CdnFrontdoorEndpointId:    pulumi.String(locals.EndpointId),
		CdnFrontdoorOriginGroupId: pulumi.String(locals.OriginGroupId),
		PatternsToMatches:         pulumi.ToStringArray(spec.PatternsToMatch),
		SupportedProtocols:        supportedProtocols,
		ForwardingProtocol:        pulumi.String(forwardingProtocol),
	}

	// Azure never receives the origin ids -- the provider uses them
	// purely to sequence route creation after the origins exist (ARM
	// rejects a route whose origin group has no origins yet).
	if len(locals.OriginIds) > 0 {
		routeArgs.CdnFrontdoorOriginIds = pulumi.ToStringArray(locals.OriginIds)
	}

	// Attached delivery policies and custom domains. Front Door treats
	// an EMPTY collection differently from an absent one (empty means
	// "disassociate", which only matters on update), so the lists are
	// sent only when populated -- absence and emptiness then agree.
	if len(locals.RuleSetIds) > 0 {
		routeArgs.CdnFrontdoorRuleSetIds = pulumi.ToStringArray(locals.RuleSetIds)
	}
	if len(locals.CustomDomainIds) > 0 {
		routeArgs.CdnFrontdoorCustomDomainIds = pulumi.ToStringArray(locals.CustomDomainIds)
	}

	// Booleans are sent only when explicitly set: Azure's defaults
	// (https redirect on, linked to the default domain, enabled) apply
	// when omitted, and the platform materializes the documented
	// defaults centrally.
	if spec.HttpsRedirectEnabled != nil {
		routeArgs.HttpsRedirectEnabled = pulumi.Bool(spec.GetHttpsRedirectEnabled())
	}
	if spec.LinkToDefaultDomain != nil {
		routeArgs.LinkToDefaultDomain = pulumi.Bool(spec.GetLinkToDefaultDomain())
	}
	if spec.Enabled != nil {
		routeArgs.Enabled = pulumi.Bool(spec.GetEnabled())
	}
	if spec.OriginPath != nil {
		routeArgs.CdnFrontdoorOriginPath = pulumi.String(spec.GetOriginPath())
	}

	// The cache block is sent only when configured: Front Door treats
	// ABSENT cache settings as caching disabled (the provider sends an
	// explicit null), so omitting the block is a real behavior switch.
	if spec.Cache != nil {
		cachingBehavior := queryStringCachingBehaviorStrings[spec.Cache.QueryStringCachingBehavior]
		if cachingBehavior == "" {
			cachingBehavior = "IgnoreQueryString"
		}
		cacheArgs := &cdn.FrontdoorRouteCacheArgs{
			QueryStringCachingBehavior: pulumi.String(cachingBehavior),
		}
		if len(spec.Cache.QueryStrings) > 0 {
			cacheArgs.QueryStrings = pulumi.ToStringArray(spec.Cache.QueryStrings)
		}
		if spec.Cache.CompressionEnabled != nil {
			cacheArgs.CompressionEnabled = pulumi.Bool(spec.Cache.GetCompressionEnabled())
		}
		if len(spec.Cache.ContentTypesToCompress) > 0 {
			cacheArgs.ContentTypesToCompresses = pulumi.ToStringArray(spec.Cache.ContentTypesToCompress)
		}
		routeArgs.Cache = cacheArgs
	}

	createdRoute, err := cdn.NewFrontdoorRoute(ctx,
		spec.RouteName,
		routeArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door route %s", spec.RouteName)
	}

	// Export outputs. No hostname output on purpose: the
	// client-facing hostname lives on the endpoint's outputs.
	ctx.Export(OpRouteId, createdRoute.ID())
	ctx.Export(OpRouteName, createdRoute.Name)

	return nil
}
