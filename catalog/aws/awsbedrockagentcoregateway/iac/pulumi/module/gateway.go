package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	awsbedrockagentcoregatewayv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbedrockagentcoregateway/v1alpha1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/bedrock"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/types/known/structpb"
)

// gateway creates the AgentCore gateway and its folded target satellites
// and exports outputs.
//
// Lifecycle facts the renders below depend on:
//   - targets attach to the gateway by ID; AWS deletes a gateway's
//     targets before the gateway itself at destroy (provider-managed);
//   - protocol_type has exactly one legal value (MCP) -- the provider
//     computes it; the module never sends it.
func gateway(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	args := &bedrock.AgentcoreGatewayArgs{
		// Create-time naming basis; doubles as the Name tag. metadata.name
		// on both engines. Changing it replaces the gateway.
		Name: pulumi.String(locals.GatewayName),
		// Required by AWS: the role the gateway assumes to reach its
		// targets and how inbound callers authenticate.
		RoleArn:        pulumi.String(spec.RoleArn.GetValue()),
		AuthorizerType: pulumi.String(spec.AuthorizerType),
		Tags:           pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.KmsKeyArn.GetValue() != "" {
		args.KmsKeyArn = pulumi.String(spec.KmsKeyArn.GetValue())
	}
	// DEBUG is the only exception level AWS defines -- the spec models it
	// as a bool and the module owns the constant.
	if spec.ExposeDebugExceptions {
		args.ExceptionLevel = pulumi.String("DEBUG")
	}

	// OIDC token validation -- required by AWS exactly when
	// authorizer_type is CUSTOM_JWT (spec-validated).
	if spec.CustomJwtAuthorizer != nil {
		args.AuthorizerConfiguration = &bedrock.AgentcoreGatewayAuthorizerConfigurationArgs{
			CustomJwtAuthorizer: jwtAuthorizerArgs(spec.CustomJwtAuthorizer),
		}
	}

	// Lambda interceptors in the request/response path (max 2).
	var interceptors bedrock.AgentcoreGatewayInterceptorConfigurationArray
	for _, i := range spec.Interceptors {
		interceptor := &bedrock.AgentcoreGatewayInterceptorConfigurationArgs{
			InterceptionPoints: pulumi.ToStringArray(i.InterceptionPoints),
			Interceptor: &bedrock.AgentcoreGatewayInterceptorConfigurationInterceptorArgs{
				Lambda: &bedrock.AgentcoreGatewayInterceptorConfigurationInterceptorLambdaArgs{
					Arn: pulumi.String(i.LambdaArn.GetValue()),
				},
			},
		}
		if i.PassRequestHeaders != nil {
			interceptor.InputConfiguration = &bedrock.AgentcoreGatewayInterceptorConfigurationInputConfigurationArgs{
				PassRequestHeaders: pulumi.Bool(*i.PassRequestHeaders),
			}
		}
		interceptors = append(interceptors, interceptor)
	}
	if len(interceptors) > 0 {
		args.InterceptorConfigurations = interceptors
	}

	// Cedar policy-engine evaluation of every tool call.
	if spec.PolicyEngine != nil {
		args.PolicyEngineConfiguration = &bedrock.AgentcoreGatewayPolicyEngineConfigurationArgs{
			Arn:  pulumi.String(spec.PolicyEngine.PolicyEngineArn.GetValue()),
			Mode: pulumi.String(spec.PolicyEngine.Mode),
		}
	}

	// MCP protocol tuning. SEMANTIC is the only search type AWS defines
	// -- the spec models it as a bool and the module owns the constant.
	if spec.Mcp != nil {
		mcp := &bedrock.AgentcoreGatewayProtocolConfigurationMcpArgs{}
		if spec.Mcp.Instructions != "" {
			mcp.Instructions = pulumi.String(spec.Mcp.Instructions)
		}
		if spec.Mcp.EnableSemanticSearch {
			mcp.SearchType = pulumi.String("SEMANTIC")
		}
		if len(spec.Mcp.SupportedVersions) > 0 {
			mcp.SupportedVersions = pulumi.ToStringArray(spec.Mcp.SupportedVersions)
		}
		if spec.Mcp.SessionTimeoutSeconds != 0 {
			mcp.SessionConfiguration = &bedrock.AgentcoreGatewayProtocolConfigurationMcpSessionConfigurationArgs{
				SessionTimeoutInSeconds: pulumi.Int(int(spec.Mcp.SessionTimeoutSeconds)),
			}
		}
		if spec.Mcp.EnableResponseStreaming {
			mcp.StreamingConfiguration = &bedrock.AgentcoreGatewayProtocolConfigurationMcpStreamingConfigurationArgs{
				EnableResponseStreaming: pulumi.Bool(true),
			}
		}
		args.ProtocolConfiguration = &bedrock.AgentcoreGatewayProtocolConfigurationArgs{
			Mcp: mcp,
		}
	}

	createdGateway, err := bedrock.NewAgentcoreGateway(ctx, locals.GatewayName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "create gateway")
	}

	ctx.Export(OpGatewayId, createdGateway.GatewayId)
	ctx.Export(OpGatewayArn, createdGateway.GatewayArn)
	ctx.Export(OpGatewayUrl, createdGateway.GatewayUrl)
	ctx.Export(OpWorkloadIdentityArn, createdGateway.WorkloadIdentityDetails.Index(pulumi.Int(0)).WorkloadIdentityArn())

	// Targets keyed by their stable entry names. Iteration is name-sorted
	// for deterministic previews.
	targetIds := pulumi.StringMap{}
	for _, t := range sortedTargets(spec.Targets) {
		targetArgs, err := targetArgs(t)
		if err != nil {
			return errors.Wrapf(err, "render target %q", t.Name)
		}
		targetArgs.GatewayIdentifier = createdGateway.GatewayId
		createdTarget, err := bedrock.NewAgentcoreGatewayTarget(ctx, "target-"+t.Name, targetArgs,
			pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{createdGateway}))
		if err != nil {
			return errors.Wrapf(err, "create target %q", t.Name)
		}
		targetIds[t.Name] = createdTarget.TargetId
	}
	ctx.Export(OpTargetIds, targetIds)

	return nil
}

func jwtAuthorizerArgs(jwt *awsbedrockagentcoregatewayv1alpha1.AwsBedrockAgentCoreGatewayJwtAuthorizer) *bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerArgs {
	authorizer := &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerArgs{
		DiscoveryUrl: pulumi.String(jwt.DiscoveryUrl),
	}
	if len(jwt.AllowedAudience) > 0 {
		authorizer.AllowedAudiences = pulumi.ToStringArray(jwt.AllowedAudience)
	}
	if len(jwt.AllowedClients) > 0 {
		authorizer.AllowedClients = pulumi.ToStringArray(jwt.AllowedClients)
	}
	if len(jwt.AllowedScopes) > 0 {
		authorizer.AllowedScopes = pulumi.ToStringArray(jwt.AllowedScopes)
	}
	if jwt.AllowedWorkloads != nil {
		workloads := &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerAllowedWorkloadConfigurationArgs{}
		if len(jwt.AllowedWorkloads.WorkloadIdentities) > 0 {
			workloads.WorkloadIdentities = pulumi.ToStringArray(jwt.AllowedWorkloads.WorkloadIdentities)
		}
		var environments bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerAllowedWorkloadConfigurationHostingEnvironmentArray
		for _, arn := range jwt.AllowedWorkloads.HostingEnvironmentArns {
			environments = append(environments, &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerAllowedWorkloadConfigurationHostingEnvironmentArgs{
				Arn: pulumi.String(arn),
			})
		}
		if len(environments) > 0 {
			workloads.HostingEnvironments = environments
		}
		authorizer.AllowedWorkloadConfiguration = workloads
	}
	var claims bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerCustomClaimArray
	for _, c := range jwt.CustomClaims {
		matchValue := &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerCustomClaimAuthorizingClaimMatchValueClaimMatchValueArgs{}
		if c.MatchValue != "" {
			matchValue.MatchValueString = pulumi.String(c.MatchValue)
		}
		if len(c.MatchValues) > 0 {
			matchValue.MatchValueStringLists = pulumi.ToStringArray(c.MatchValues)
		}
		claims = append(claims, &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerCustomClaimArgs{
			InboundTokenClaimName:      pulumi.String(c.ClaimName),
			InboundTokenClaimValueType: pulumi.String(c.ValueType),
			AuthorizingClaimMatchValue: &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerCustomClaimAuthorizingClaimMatchValueArgs{
				ClaimMatchOperator: pulumi.String(c.MatchOperator),
				ClaimMatchValue:    matchValue,
			},
		})
	}
	if len(claims) > 0 {
		authorizer.CustomClaims = claims
	}
	if jwt.PrivateEndpoint != nil {
		endpoint := &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerPrivateEndpointArgs{}
		if jwt.PrivateEndpoint.ManagedVpc != nil {
			managed := jwt.PrivateEndpoint.ManagedVpc
			managedArgs := &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerPrivateEndpointManagedVpcResourceArgs{
				VpcIdentifier:         pulumi.String(managed.VpcId.GetValue()),
				SubnetIds:             svrsToStringArray(managed.SubnetIds),
				EndpointIpAddressType: pulumi.String(managed.EndpointIpAddressType),
			}
			if len(managed.SecurityGroupIds) > 0 {
				managedArgs.SecurityGroupIds = svrsToStringArray(managed.SecurityGroupIds)
			}
			if managed.RoutingDomain != "" {
				managedArgs.RoutingDomain = pulumi.String(managed.RoutingDomain)
			}
			if len(managed.Tags) > 0 {
				managedArgs.Tags = pulumi.ToStringMap(managed.Tags)
			}
			endpoint.ManagedVpcResource = managedArgs
		}
		if jwt.PrivateEndpoint.SelfManagedLattice != nil {
			endpoint.SelfManagedLatticeResource = &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerPrivateEndpointSelfManagedLatticeResourceArgs{
				ResourceConfigurationIdentifier: pulumi.String(jwt.PrivateEndpoint.SelfManagedLattice.ResourceConfigurationId),
			}
		}
		authorizer.PrivateEndpoint = endpoint
	}
	var overrides bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerPrivateEndpointOverrideArray
	for _, o := range jwt.PrivateEndpointOverrides {
		override := &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerPrivateEndpointOverrideArgs{
			Domain: pulumi.String(o.Domain),
		}
		endpoint := &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerPrivateEndpointOverridePrivateEndpointArgs{}
		if o.PrivateEndpoint.ManagedVpc != nil {
			managed := o.PrivateEndpoint.ManagedVpc
			managedArgs := &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerPrivateEndpointOverridePrivateEndpointManagedVpcResourceArgs{
				VpcIdentifier:         pulumi.String(managed.VpcId.GetValue()),
				SubnetIds:             svrsToStringArray(managed.SubnetIds),
				EndpointIpAddressType: pulumi.String(managed.EndpointIpAddressType),
			}
			if len(managed.SecurityGroupIds) > 0 {
				managedArgs.SecurityGroupIds = svrsToStringArray(managed.SecurityGroupIds)
			}
			if managed.RoutingDomain != "" {
				managedArgs.RoutingDomain = pulumi.String(managed.RoutingDomain)
			}
			if len(managed.Tags) > 0 {
				managedArgs.Tags = pulumi.ToStringMap(managed.Tags)
			}
			endpoint.ManagedVpcResource = managedArgs
		}
		if o.PrivateEndpoint.SelfManagedLattice != nil {
			endpoint.SelfManagedLatticeResource = &bedrock.AgentcoreGatewayAuthorizerConfigurationCustomJwtAuthorizerPrivateEndpointOverridePrivateEndpointSelfManagedLatticeResourceArgs{
				ResourceConfigurationIdentifier: pulumi.String(o.PrivateEndpoint.SelfManagedLattice.ResourceConfigurationId),
			}
		}
		override.PrivateEndpoint = endpoint
		overrides = append(overrides, override)
	}
	if len(overrides) > 0 {
		authorizer.PrivateEndpointOverrides = overrides
	}
	return authorizer
}

// targetArgs renders one target entry (everything except the gateway
// identifier, which the caller wires from the created gateway).
func targetArgs(t *awsbedrockagentcoregatewayv1alpha1.AwsBedrockAgentCoreGatewayTarget) (*bedrock.AgentcoreGatewayTargetArgs, error) {
	args := &bedrock.AgentcoreGatewayTargetArgs{
		Name: pulumi.String(t.Name),
	}
	if t.Description != "" {
		args.Description = pulumi.String(t.Description)
	}

	// Exactly one backend arm (spec-validated).
	configuration := &bedrock.AgentcoreGatewayTargetTargetConfigurationArgs{}
	backend := t.Backend
	if backend.AgentcoreRuntime != nil {
		runtime := &bedrock.AgentcoreGatewayTargetTargetConfigurationHttpAgentcoreRuntimeArgs{
			Arn: pulumi.String(backend.AgentcoreRuntime.AgentRuntimeArn.GetValue()),
		}
		if backend.AgentcoreRuntime.Qualifier != "" {
			runtime.Qualifier = pulumi.String(backend.AgentcoreRuntime.Qualifier)
		}
		configuration.Http = &bedrock.AgentcoreGatewayTargetTargetConfigurationHttpArgs{
			AgentcoreRuntime: runtime,
		}
	} else {
		mcp := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpArgs{}
		if backend.ApiGateway != nil {
			apiGateway := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpApiGatewayArgs{
				RestApiId: pulumi.String(backend.ApiGateway.RestApiId),
				Stage:     pulumi.String(backend.ApiGateway.Stage),
			}
			if len(backend.ApiGateway.ToolFilters) > 0 || len(backend.ApiGateway.ToolOverrides) > 0 {
				tools := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpApiGatewayApiGatewayToolConfigurationArgs{}
				var filters bedrock.AgentcoreGatewayTargetTargetConfigurationMcpApiGatewayApiGatewayToolConfigurationToolFilterArray
				for _, f := range backend.ApiGateway.ToolFilters {
					filters = append(filters, &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpApiGatewayApiGatewayToolConfigurationToolFilterArgs{
						FilterPath: pulumi.String(f.FilterPath),
						Methods:    pulumi.ToStringArray(f.Methods),
					})
				}
				if len(filters) > 0 {
					tools.ToolFilters = filters
				}
				var toolOverrides bedrock.AgentcoreGatewayTargetTargetConfigurationMcpApiGatewayApiGatewayToolConfigurationToolOverrideArray
				for _, o := range backend.ApiGateway.ToolOverrides {
					entry := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpApiGatewayApiGatewayToolConfigurationToolOverrideArgs{
						Path:   pulumi.String(o.Path),
						Method: pulumi.String(o.Method),
						Name:   pulumi.String(o.Name),
					}
					if o.Description != "" {
						entry.Description = pulumi.String(o.Description)
					}
					toolOverrides = append(toolOverrides, entry)
				}
				if len(toolOverrides) > 0 {
					tools.ToolOverrides = toolOverrides
				}
				apiGateway.ApiGatewayToolConfiguration = tools
			}
			mcp.ApiGateway = apiGateway
		}
		if backend.Lambda != nil {
			toolSchema := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpLambdaToolSchemaArgs{}
			var payloads bedrock.AgentcoreGatewayTargetTargetConfigurationMcpLambdaToolSchemaInlinePayloadArray
			for _, tool := range backend.Lambda.Tools {
				inputSchema, err := inputSchemaArgs(tool.InputSchema)
				if err != nil {
					return nil, errors.Wrapf(err, "tool %q input schema", tool.Name)
				}
				payload := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpLambdaToolSchemaInlinePayloadArgs{
					Name:        pulumi.String(tool.Name),
					Description: pulumi.String(tool.Description),
					InputSchema: inputSchema,
				}
				if tool.OutputSchema != nil {
					outputSchema, err := outputSchemaArgs(tool.OutputSchema)
					if err != nil {
						return nil, errors.Wrapf(err, "tool %q output schema", tool.Name)
					}
					payload.OutputSchema = outputSchema
				}
				payloads = append(payloads, payload)
			}
			if len(payloads) > 0 {
				toolSchema.InlinePayloads = payloads
			}
			if backend.Lambda.ToolsS3 != nil {
				s3 := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpLambdaToolSchemaS3Args{
					Uri: pulumi.String(backend.Lambda.ToolsS3.Uri),
				}
				if backend.Lambda.ToolsS3.BucketOwnerAccountId != "" {
					s3.BucketOwnerAccountId = pulumi.String(backend.Lambda.ToolsS3.BucketOwnerAccountId)
				}
				toolSchema.S3 = s3
			}
			mcp.Lambda = &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpLambdaArgs{
				LambdaArn:  pulumi.String(backend.Lambda.LambdaArn.GetValue()),
				ToolSchema: toolSchema,
			}
		}
		if backend.McpServer != nil {
			server := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpMcpServerArgs{
				Endpoint: pulumi.String(backend.McpServer.Endpoint),
			}
			if backend.McpServer.ListingMode != "" {
				server.ListingMode = pulumi.String(backend.McpServer.ListingMode)
			}
			mcp.McpServer = server
		}
		if backend.OpenApiSchema != nil {
			schema := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpOpenApiSchemaArgs{}
			if backend.OpenApiSchema.InlinePayload != "" {
				schema.InlinePayload = &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpOpenApiSchemaInlinePayloadArgs{
					Payload: pulumi.String(backend.OpenApiSchema.InlinePayload),
				}
			}
			if backend.OpenApiSchema.S3 != nil {
				s3 := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpOpenApiSchemaS3Args{
					Uri: pulumi.String(backend.OpenApiSchema.S3.Uri),
				}
				if backend.OpenApiSchema.S3.BucketOwnerAccountId != "" {
					s3.BucketOwnerAccountId = pulumi.String(backend.OpenApiSchema.S3.BucketOwnerAccountId)
				}
				schema.S3 = s3
			}
			mcp.OpenApiSchema = schema
		}
		if backend.SmithyModel != nil {
			schema := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpSmithyModelArgs{}
			if backend.SmithyModel.InlinePayload != "" {
				schema.InlinePayload = &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpSmithyModelInlinePayloadArgs{
					Payload: pulumi.String(backend.SmithyModel.InlinePayload),
				}
			}
			if backend.SmithyModel.S3 != nil {
				s3 := &bedrock.AgentcoreGatewayTargetTargetConfigurationMcpSmithyModelS3Args{
					Uri: pulumi.String(backend.SmithyModel.S3.Uri),
				}
				if backend.SmithyModel.S3.BucketOwnerAccountId != "" {
					s3.BucketOwnerAccountId = pulumi.String(backend.SmithyModel.S3.BucketOwnerAccountId)
				}
				schema.S3 = s3
			}
			mcp.SmithyModel = schema
		}
		configuration.Mcp = mcp
	}
	args.TargetConfiguration = configuration

	// How the GATEWAY authenticates to this backend: the credentials
	// oneof carries at most one arm by construction. jwt_passthrough is an
	// empty member at the provider -- choosing the arm IS the configuration.
	if t.Credentials != nil {
		credentials := &bedrock.AgentcoreGatewayTargetCredentialProviderConfigurationArgs{}
		if t.Credentials.GetApiKey() != nil {
			apiKey := &bedrock.AgentcoreGatewayTargetCredentialProviderConfigurationApiKeyArgs{
				ProviderArn: pulumi.String(t.Credentials.GetApiKey().ProviderArn.GetValue()),
			}
			if t.Credentials.GetApiKey().CredentialLocation != "" {
				apiKey.CredentialLocation = pulumi.String(t.Credentials.GetApiKey().CredentialLocation)
			}
			if t.Credentials.GetApiKey().CredentialParameterName != "" {
				apiKey.CredentialParameterName = pulumi.String(t.Credentials.GetApiKey().CredentialParameterName)
			}
			if t.Credentials.GetApiKey().CredentialPrefix != "" {
				apiKey.CredentialPrefix = pulumi.String(t.Credentials.GetApiKey().CredentialPrefix)
			}
			credentials.ApiKey = apiKey
		}
		if t.Credentials.GetCallerIamCredentials() != nil {
			caller := &bedrock.AgentcoreGatewayTargetCredentialProviderConfigurationCallerIamCredentialsArgs{
				Service: pulumi.String(t.Credentials.GetCallerIamCredentials().Service),
			}
			if t.Credentials.GetCallerIamCredentials().Region != "" {
				caller.Region = pulumi.String(t.Credentials.GetCallerIamCredentials().Region)
			}
			credentials.CallerIamCredentials = caller
		}
		if t.Credentials.GetGatewayIamRole() != nil {
			role := &bedrock.AgentcoreGatewayTargetCredentialProviderConfigurationGatewayIamRoleArgs{}
			if t.Credentials.GetGatewayIamRole().Service != "" {
				role.Service = pulumi.String(t.Credentials.GetGatewayIamRole().Service)
			}
			if t.Credentials.GetGatewayIamRole().Region != "" {
				role.Region = pulumi.String(t.Credentials.GetGatewayIamRole().Region)
			}
			credentials.GatewayIamRole = role
		}
		if t.Credentials.GetJwtPassthrough() != nil {
			credentials.JwtPassthrough = &bedrock.AgentcoreGatewayTargetCredentialProviderConfigurationJwtPassthroughArgs{}
		}
		if t.Credentials.GetOauth() != nil {
			oauth := &bedrock.AgentcoreGatewayTargetCredentialProviderConfigurationOauthArgs{
				ProviderArn: pulumi.String(t.Credentials.GetOauth().ProviderArn.GetValue()),
				Scopes:      pulumi.ToStringArray(t.Credentials.GetOauth().Scopes),
			}
			if t.Credentials.GetOauth().GrantType != "" {
				oauth.GrantType = pulumi.String(t.Credentials.GetOauth().GrantType)
			}
			if t.Credentials.GetOauth().DefaultReturnUrl != "" {
				oauth.DefaultReturnUrl = pulumi.String(t.Credentials.GetOauth().DefaultReturnUrl)
			}
			if len(t.Credentials.GetOauth().CustomParameters) > 0 {
				oauth.CustomParameters = pulumi.ToStringMap(t.Credentials.GetOauth().CustomParameters)
			}
			credentials.Oauth = oauth
		}
		args.CredentialProviderConfiguration = credentials
	}

	// Caller metadata propagation (max 10 entries each); the block's own
	// switch (on by default once declared, filled by the platform before
	// this runs) decides whether it renders.
	if t.Metadata != nil && t.Metadata.GetEnabled() {
		metadata := &bedrock.AgentcoreGatewayTargetMetadataConfigurationArgs{}
		if len(t.Metadata.AllowedQueryParameters) > 0 {
			metadata.AllowedQueryParameters = pulumi.ToStringArray(t.Metadata.AllowedQueryParameters)
		}
		if len(t.Metadata.AllowedRequestHeaders) > 0 {
			metadata.AllowedRequestHeaders = pulumi.ToStringArray(t.Metadata.AllowedRequestHeaders)
		}
		if len(t.Metadata.AllowedResponseHeaders) > 0 {
			metadata.AllowedResponseHeaders = pulumi.ToStringArray(t.Metadata.AllowedResponseHeaders)
		}
		args.MetadataConfiguration = metadata
	}

	// Reach a PRIVATE backend through your VPC.
	if t.PrivateEndpoint != nil {
		endpoint := &bedrock.AgentcoreGatewayTargetPrivateEndpointArgs{}
		if t.PrivateEndpoint.ManagedVpc != nil {
			managed := t.PrivateEndpoint.ManagedVpc
			managedArgs := &bedrock.AgentcoreGatewayTargetPrivateEndpointManagedVpcResourceArgs{
				VpcIdentifier:         pulumi.String(managed.VpcId.GetValue()),
				SubnetIds:             svrsToStringArray(managed.SubnetIds),
				EndpointIpAddressType: pulumi.String(managed.EndpointIpAddressType),
			}
			if len(managed.SecurityGroupIds) > 0 {
				managedArgs.SecurityGroupIds = svrsToStringArray(managed.SecurityGroupIds)
			}
			if managed.RoutingDomain != "" {
				managedArgs.RoutingDomain = pulumi.String(managed.RoutingDomain)
			}
			if len(managed.Tags) > 0 {
				managedArgs.Tags = pulumi.ToStringMap(managed.Tags)
			}
			endpoint.ManagedVpcResource = managedArgs
		}
		if t.PrivateEndpoint.SelfManagedLattice != nil {
			endpoint.SelfManagedLatticeResource = &bedrock.AgentcoreGatewayTargetPrivateEndpointSelfManagedLatticeResourceArgs{
				ResourceConfigurationIdentifier: pulumi.String(t.PrivateEndpoint.SelfManagedLattice.ResourceConfigurationId),
			}
		}
		args.PrivateEndpoint = endpoint
	}

	return args, nil
}

// structToJson renders a raw JSON-schema leaf (a google.protobuf.Struct)
// as the provider's normalized-JSON string.
func structToJson(in *structpb.Struct) (string, error) {
	bytes, err := json.Marshal(in.AsMap())
	if err != nil {
		return "", errors.Wrap(err, "marshal json leaf")
	}
	return string(bytes), nil
}
