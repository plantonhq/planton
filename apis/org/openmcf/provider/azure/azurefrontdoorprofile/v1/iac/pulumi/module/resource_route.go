package module

import (
	"fmt"

	azurefrontdoorprofilev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurefrontdoorprofile/v1"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createRoutes(
	ctx *pulumi.Context,
	locals *Locals,
	azureProvider *azure.Provider,
	endpoints map[string]*createdEndpoint,
	originGroups map[string]*createdOriginGroup,
	origins map[string]*createdOrigin,
) error {
	spec := locals.AzureFrontDoorProfile.Spec

	for _, route := range spec.GetRoutes() {
		if err := createRoute(ctx, route, endpoints, originGroups, origins, azureProvider); err != nil {
			return err
		}
	}

	return nil
}

func createRoute(
	ctx *pulumi.Context,
	route *azurefrontdoorprofilev1.AzureFrontDoorRoute,
	endpoints map[string]*createdEndpoint,
	originGroups map[string]*createdOriginGroup,
	origins map[string]*createdOrigin,
	azureProvider *azure.Provider,
) error {
	// Look up the endpoint by name.
	ep, ok := endpoints[route.GetEndpointName()]
	if !ok {
		return fmt.Errorf("route %s references unknown endpoint %s", route.Name, route.GetEndpointName())
	}

	// Look up the origin group by name.
	og, ok := originGroups[route.GetOriginGroupName()]
	if !ok {
		return fmt.Errorf("route %s references unknown origin group %s", route.Name, route.GetOriginGroupName())
	}

	// Collect all origin IDs from the referenced origin group.
	originIds := pulumi.StringArray{}
	for _, o := range og.Spec.GetOrigins() {
		key := fmt.Sprintf("%s/%s", route.GetOriginGroupName(), o.Name)
		if created, exists := origins[key]; exists {
			originIds = append(originIds, created.Resource.ID().ToStringOutput())
		}
	}

	args := &cdn.FrontdoorRouteArgs{
		Name:                      pulumi.String(route.Name),
		CdnFrontdoorEndpointId:    ep.Resource.ID(),
		CdnFrontdoorOriginGroupId: og.Resource.ID(),
		CdnFrontdoorOriginIds:     originIds,
		PatternsToMatches:         pulumi.ToStringArray(route.GetPatternsToMatch()),
		SupportedProtocols:        pulumi.ToStringArray(route.GetSupportedProtocols()),
		ForwardingProtocol:        pulumi.StringPtr(route.GetForwardingProtocol()),
		HttpsRedirectEnabled:      pulumi.BoolPtr(route.GetHttpsRedirectEnabled()),
		LinkToDefaultDomain:       pulumi.BoolPtr(route.GetLinkToDefaultDomain()),
		Enabled:                   pulumi.BoolPtr(route.GetEnabled()),
	}

	// Optional cache configuration.
	if route.GetCache() != nil {
		args.Cache = buildRouteCache(route.GetCache())
	}

	// Build dependencies: endpoint, origin group, and all origins in the group.
	deps := []pulumi.Resource{ep.Resource, og.Resource}
	for _, o := range og.Spec.GetOrigins() {
		key := fmt.Sprintf("%s/%s", route.GetOriginGroupName(), o.Name)
		if created, exists := origins[key]; exists {
			deps = append(deps, created.Resource)
		}
	}

	_, err := cdn.NewFrontdoorRoute(ctx,
		fmt.Sprintf("route-%s", route.Name),
		args,
		pulumi.Provider(azureProvider),
		pulumi.DependsOn(deps))
	if err != nil {
		return errors.Wrapf(err, "failed to create Front Door route %s", route.Name)
	}

	return nil
}

func buildRouteCache(cache *azurefrontdoorprofilev1.AzureFrontDoorRouteCache) cdn.FrontdoorRouteCachePtrInput {
	args := cdn.FrontdoorRouteCacheArgs{
		QueryStringCachingBehavior: pulumi.StringPtr(cache.GetQueryStringCachingBehavior()),
		CompressionEnabled:         pulumi.Bool(cache.GetCompressionEnabled()),
	}

	if len(cache.GetQueryStrings()) > 0 {
		args.QueryStrings = pulumi.ToStringArray(cache.GetQueryStrings())
	}

	if len(cache.GetContentTypesToCompress()) > 0 {
		args.ContentTypesToCompresses = pulumi.ToStringArray(cache.GetContentTypesToCompress())
	}

	return &args
}
