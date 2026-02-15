package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// routes creates API Gateway routes, wiring each route to its deduplicated
// integration and optional authorizer.
func routes(
	ctx *pulumi.Context,
	locals *Locals,
	createdApi *apigatewayv2.Api,
	integrationMap map[string]*apigatewayv2.Integration,
	authorizerMap map[string]*apigatewayv2.Authorizer,
	provider *aws.Provider,
) error {
	for i, route := range locals.Spec.Routes {
		// Build a safe resource name from the route key.
		// "GET /users" -> "get-users", "$default" -> "default"
		safeName := sanitizeRouteKey(route.RouteKey)
		resourceName := fmt.Sprintf("%s-route-%s", locals.ApiName, safeName)

		// Look up the deduplicated integration for this route.
		key := integrationKey(route.Integration)
		integration, ok := integrationMap[key]
		if !ok {
			return fmt.Errorf("integration not found for route %q (index %d)", route.RouteKey, i)
		}

		args := &apigatewayv2.RouteArgs{
			ApiId:    createdApi.ID(),
			RouteKey: pulumi.String(route.RouteKey),
			// Target format: "integrations/{integrationId}"
			Target: integration.ID().ApplyT(func(id string) string {
				return fmt.Sprintf("integrations/%s", id)
			}).(pulumi.StringOutput),
		}

		// Authorization
		if route.AuthorizationType != "" && route.AuthorizationType != "NONE" {
			args.AuthorizationType = pulumi.StringPtr(route.AuthorizationType)

			// Wire authorizer reference
			if route.AuthorizerName != "" {
				authorizer, authOk := authorizerMap[route.AuthorizerName]
				if authOk {
					args.AuthorizerId = authorizer.ID().ApplyT(func(id string) string {
						return id
					}).(pulumi.StringOutput)
				}
			}

			// Authorization scopes (JWT)
			if len(route.AuthorizationScopes) > 0 {
				args.AuthorizationScopes = pulumi.ToStringArray(route.AuthorizationScopes)
			}
		}

		if _, err := apigatewayv2.NewRoute(ctx, resourceName, args, pulumi.Provider(provider)); err != nil {
			return errors.Wrapf(err, "failed to create route %q (index %d)", route.RouteKey, i)
		}
	}

	return nil
}

// sanitizeRouteKey converts a route key into a safe resource name component.
// "GET /users/{id}" -> "get-users-id"
// "$default" -> "default"
func sanitizeRouteKey(routeKey string) string {
	s := strings.ToLower(routeKey)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	s = strings.ReplaceAll(s, "$", "")
	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}
