package module

import (
	"fmt"

	"github.com/pkg/errors"
	ociapigatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociapigateway/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/apigateway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func deploymentResource(
	ctx *pulumi.Context,
	locals *Locals,
	provider *oci.Provider,
	gateway *apigateway.Gateway,
) error {
	spec := locals.OciApiGateway.Spec
	deploy := spec.Deployment

	deployDisplayName := deploy.DisplayName
	if deployDisplayName == "" {
		deployDisplayName = fmt.Sprintf("%s-deployment", locals.DisplayName)
	}

	specArg := buildSpecification(deploy)

	args := &apigateway.DeploymentArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		GatewayId:     gateway.ID(),
		PathPrefix:    pulumi.String(deploy.PathPrefix),
		DisplayName:   pulumi.String(deployDisplayName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
		Specification: specArg,
	}

	createdDeployment, err := apigateway.NewDeployment(
		ctx,
		deployDisplayName,
		args,
		pulumiOciOpt(provider),
		pulumi.DependsOn([]pulumi.Resource{gateway}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create api deployment")
	}

	ctx.Export(OpDeploymentEndpoint, createdDeployment.Endpoint)

	return nil
}

func buildSpecification(deploy *ociapigatewayv1.OciApiGatewaySpec_Deployment) *apigateway.DeploymentSpecificationArgs {
	specArgs := &apigateway.DeploymentSpecificationArgs{
		Routes: buildRoutes(deploy.Routes),
	}

	if deploy.LoggingPolicies != nil {
		specArgs.LoggingPolicies = buildLoggingPolicies(deploy.LoggingPolicies)
	}

	if deploy.RequestPolicies != nil {
		specArgs.RequestPolicies = buildRequestPolicies(deploy.RequestPolicies)
	}

	return specArgs
}

func buildLoggingPolicies(lp *ociapigatewayv1.OciApiGatewaySpec_LoggingPolicies) *apigateway.DeploymentSpecificationLoggingPoliciesArgs {
	args := &apigateway.DeploymentSpecificationLoggingPoliciesArgs{}

	if lp.AccessLog != nil {
		args.AccessLog = &apigateway.DeploymentSpecificationLoggingPoliciesAccessLogArgs{
			IsEnabled: pulumi.Bool(lp.AccessLog.IsEnabled),
		}
	}

	if lp.ExecutionLog != nil {
		execArgs := &apigateway.DeploymentSpecificationLoggingPoliciesExecutionLogArgs{
			IsEnabled: pulumi.Bool(lp.ExecutionLog.IsEnabled),
		}
		if lvl, ok := logLevelMap[lp.ExecutionLog.LogLevel]; ok {
			execArgs.LogLevel = pulumi.String(lvl)
		}
		args.ExecutionLog = execArgs
	}

	return args
}

func buildRequestPolicies(rp *ociapigatewayv1.OciApiGatewaySpec_RequestPolicies) *apigateway.DeploymentSpecificationRequestPoliciesArgs {
	args := &apigateway.DeploymentSpecificationRequestPoliciesArgs{}

	if rp.Authentication != nil {
		args.Authentication = buildAuthentication(rp.Authentication)
	}

	if rp.Cors != nil {
		args.Cors = buildCors(rp.Cors)
	}

	if rp.RateLimiting != nil {
		args.RateLimiting = buildRateLimiting(rp.RateLimiting)
	}

	return args
}

func buildAuthentication(auth *ociapigatewayv1.OciApiGatewaySpec_Authentication) *apigateway.DeploymentSpecificationRequestPoliciesAuthenticationArgs {
	args := &apigateway.DeploymentSpecificationRequestPoliciesAuthenticationArgs{
		Type: pulumi.String("JWT_AUTHENTICATION"),
	}

	if len(auth.Issuers) > 0 {
		args.Issuers = pulumi.ToStringArray(auth.Issuers)
	}

	if len(auth.Audiences) > 0 {
		args.Audiences = pulumi.ToStringArray(auth.Audiences)
	}

	if auth.TokenHeader != "" {
		args.TokenHeader = pulumi.String(auth.TokenHeader)
	}

	if auth.TokenQueryParam != "" {
		args.TokenQueryParam = pulumi.String(auth.TokenQueryParam)
	}

	if auth.TokenAuthScheme != "" {
		args.TokenAuthScheme = pulumi.String(auth.TokenAuthScheme)
	}

	if auth.MaxClockSkewInSeconds != nil {
		args.MaxClockSkewInSeconds = pulumi.Float64(float64(*auth.MaxClockSkewInSeconds))
	}

	if auth.IsAnonymousAccessAllowed != nil {
		args.IsAnonymousAccessAllowed = pulumi.Bool(*auth.IsAnonymousAccessAllowed)
	}

	if auth.PublicKeys != nil {
		args.PublicKeys = buildPublicKeys(auth.PublicKeys)
	}

	if len(auth.VerifyClaims) > 0 {
		claims := make(apigateway.DeploymentSpecificationRequestPoliciesAuthenticationVerifyClaimArray, len(auth.VerifyClaims))
		for i, vc := range auth.VerifyClaims {
			claimArgs := &apigateway.DeploymentSpecificationRequestPoliciesAuthenticationVerifyClaimArgs{}
			if vc.Key != "" {
				claimArgs.Key = pulumi.String(vc.Key)
			}
			if len(vc.Values) > 0 {
				claimArgs.Values = pulumi.ToStringArray(vc.Values)
			}
			if vc.IsRequired != nil {
				claimArgs.IsRequired = pulumi.Bool(*vc.IsRequired)
			}
			claims[i] = claimArgs
		}
		args.VerifyClaims = claims
	}

	return args
}

func buildPublicKeys(pk *ociapigatewayv1.OciApiGatewaySpec_PublicKeys) *apigateway.DeploymentSpecificationRequestPoliciesAuthenticationPublicKeysArgs {
	args := &apigateway.DeploymentSpecificationRequestPoliciesAuthenticationPublicKeysArgs{
		Type: pulumi.String(publicKeyTypeMap[pk.Type]),
	}

	if pk.Uri != "" {
		args.Uri = pulumi.String(pk.Uri)
	}

	if pk.IsSslVerifyDisabled != nil {
		args.IsSslVerifyDisabled = pulumi.Bool(*pk.IsSslVerifyDisabled)
	}

	if pk.MaxCacheDurationInHours != nil {
		args.MaxCacheDurationInHours = pulumi.Int(int(*pk.MaxCacheDurationInHours))
	}

	if len(pk.Keys) > 0 {
		keys := make(apigateway.DeploymentSpecificationRequestPoliciesAuthenticationPublicKeysKeyArray, len(pk.Keys))
		for i, k := range pk.Keys {
			keyArgs := &apigateway.DeploymentSpecificationRequestPoliciesAuthenticationPublicKeysKeyArgs{
				Kid:    pulumi.String(k.Kid),
				Format: pulumi.String(keyFormatMap[k.Format]),
			}
			if k.Key != "" {
				keyArgs.Key = pulumi.String(k.Key)
			}
			if k.Kty != "" {
				keyArgs.Kty = pulumi.String(k.Kty)
			}
			if k.Alg != "" {
				keyArgs.Alg = pulumi.String(k.Alg)
			}
			if k.N != "" {
				keyArgs.N = pulumi.String(k.N)
			}
			if k.E != "" {
				keyArgs.E = pulumi.String(k.E)
			}
			if k.Use != "" {
				keyArgs.Use = pulumi.String(k.Use)
			}
			keys[i] = keyArgs
		}
		args.Keys = keys
	}

	return args
}

func buildCors(cors *ociapigatewayv1.OciApiGatewaySpec_CorsPolicy) *apigateway.DeploymentSpecificationRequestPoliciesCorsArgs {
	args := &apigateway.DeploymentSpecificationRequestPoliciesCorsArgs{
		AllowedOrigins: pulumi.ToStringArray(cors.AllowedOrigins),
	}

	if len(cors.AllowedMethods) > 0 {
		args.AllowedMethods = pulumi.ToStringArray(cors.AllowedMethods)
	}

	if len(cors.AllowedHeaders) > 0 {
		args.AllowedHeaders = pulumi.ToStringArray(cors.AllowedHeaders)
	}

	if len(cors.ExposedHeaders) > 0 {
		args.ExposedHeaders = pulumi.ToStringArray(cors.ExposedHeaders)
	}

	if cors.IsAllowCredentialsEnabled != nil {
		args.IsAllowCredentialsEnabled = pulumi.Bool(*cors.IsAllowCredentialsEnabled)
	}

	if cors.MaxAgeInSeconds != nil {
		args.MaxAgeInSeconds = pulumi.Int(int(*cors.MaxAgeInSeconds))
	}

	return args
}

func buildRateLimiting(rl *ociapigatewayv1.OciApiGatewaySpec_RateLimiting) *apigateway.DeploymentSpecificationRequestPoliciesRateLimitingArgs {
	return &apigateway.DeploymentSpecificationRequestPoliciesRateLimitingArgs{
		RateInRequestsPerSecond: pulumi.Int(int(rl.RateInRequestsPerSecond)),
		RateKey:                 pulumi.String(rateKeyMap[rl.RateKey]),
	}
}

func buildRoutes(routes []*ociapigatewayv1.OciApiGatewaySpec_Route) apigateway.DeploymentSpecificationRouteArray {
	result := make(apigateway.DeploymentSpecificationRouteArray, len(routes))
	for i, r := range routes {
		routeArgs := &apigateway.DeploymentSpecificationRouteArgs{
			Path:    pulumi.String(r.Path),
			Backend: buildBackend(r.Backend),
		}

		if len(r.Methods) > 0 {
			routeArgs.Methods = pulumi.ToStringArray(r.Methods)
		}

		if r.Authorization != nil {
			routeArgs.RequestPolicies = buildRouteRequestPolicies(r.Authorization)
		}

		if r.LoggingPolicies != nil {
			routeArgs.LoggingPolicies = buildRouteLoggingPolicies(r.LoggingPolicies)
		}

		result[i] = routeArgs
	}
	return result
}

func buildBackend(b *ociapigatewayv1.OciApiGatewaySpec_Backend) *apigateway.DeploymentSpecificationRouteBackendArgs {
	args := &apigateway.DeploymentSpecificationRouteBackendArgs{
		Type: pulumi.String(backendTypeMap[b.Type]),
	}

	if b.Url != "" {
		args.Url = pulumi.String(b.Url)
	}

	if b.FunctionId != "" {
		args.FunctionId = pulumi.String(b.FunctionId)
	}

	if b.Status != 0 {
		args.Status = pulumi.Int(int(b.Status))
	}

	if b.Body != "" {
		args.Body = pulumi.String(b.Body)
	}

	if b.ConnectTimeoutInSeconds != nil {
		args.ConnectTimeoutInSeconds = pulumi.Float64(float64(*b.ConnectTimeoutInSeconds))
	}

	if b.ReadTimeoutInSeconds != nil {
		args.ReadTimeoutInSeconds = pulumi.Float64(float64(*b.ReadTimeoutInSeconds))
	}

	if b.SendTimeoutInSeconds != nil {
		args.SendTimeoutInSeconds = pulumi.Float64(float64(*b.SendTimeoutInSeconds))
	}

	if b.IsSslVerifyDisabled != nil {
		args.IsSslVerifyDisabled = pulumi.Bool(*b.IsSslVerifyDisabled)
	}

	if len(b.Headers) > 0 {
		headers := make(apigateway.DeploymentSpecificationRouteBackendHeaderArray, len(b.Headers))
		for i, h := range b.Headers {
			headers[i] = &apigateway.DeploymentSpecificationRouteBackendHeaderArgs{
				Name:  pulumi.String(h.Name),
				Value: pulumi.String(h.Value),
			}
		}
		args.Headers = headers
	}

	return args
}

func buildRouteRequestPolicies(authz *ociapigatewayv1.OciApiGatewaySpec_RouteAuthorization) *apigateway.DeploymentSpecificationRouteRequestPoliciesArgs {
	authzArgs := &apigateway.DeploymentSpecificationRouteRequestPoliciesAuthorizationArgs{}

	if t, ok := authorizationTypeMap[authz.Type]; ok {
		authzArgs.Type = pulumi.String(t)
	}

	if len(authz.AllowedScope) > 0 {
		authzArgs.AllowedScopes = pulumi.ToStringArray(authz.AllowedScope)
	}

	return &apigateway.DeploymentSpecificationRouteRequestPoliciesArgs{
		Authorization: authzArgs,
	}
}

func buildRouteLoggingPolicies(lp *ociapigatewayv1.OciApiGatewaySpec_LoggingPolicies) *apigateway.DeploymentSpecificationRouteLoggingPoliciesArgs {
	args := &apigateway.DeploymentSpecificationRouteLoggingPoliciesArgs{}

	if lp.AccessLog != nil {
		args.AccessLog = &apigateway.DeploymentSpecificationRouteLoggingPoliciesAccessLogArgs{
			IsEnabled: pulumi.Bool(lp.AccessLog.IsEnabled),
		}
	}

	if lp.ExecutionLog != nil {
		execArgs := &apigateway.DeploymentSpecificationRouteLoggingPoliciesExecutionLogArgs{
			IsEnabled: pulumi.Bool(lp.ExecutionLog.IsEnabled),
		}
		if lvl, ok := logLevelMap[lp.ExecutionLog.LogLevel]; ok {
			execArgs.LogLevel = pulumi.String(lvl)
		}
		args.ExecutionLog = execArgs
	}

	return args
}
