package module

import (
	"github.com/pkg/errors"
	azurefunctionappflexconsumptionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefunctionappflexconsumption/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFunctionAppFlexConsumption.Spec

	// The Flex Consumption Function App: Azure's newest serverless
	// Functions hosting model, on an FC1-SKU plan (the provider verifies
	// the plan's SKU at create time and rejects anything else). Name,
	// region, resource group, and service plan are ForceNew.
	functionAppArgs := &appservice.AppFlexConsumptionArgs{
		Name:              pulumi.String(spec.FunctionAppName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		ServicePlanId:     pulumi.String(spec.ServicePlanId.GetValue()),

		// The deployment storage binding. blobContainer is the type's
		// single legal value, so the spec omits it and the module sends
		// the constant.
		StorageContainerType:     pulumi.String("blobContainer"),
		StorageContainerEndpoint: pulumi.String(spec.StorageContainerEndpoint),

		// The per-mode requiredness (key with connection-string auth,
		// identity id with user-assigned auth) is spec-enforced,
		// mirroring the provider's own create-time checks.
		StorageAuthenticationType: pulumi.String(storageAuthenticationTypeStrings[spec.StorageAuthenticationType]),

		// The flat runtime declaration (Flex Consumption has no
		// application_stack block and no container form).
		RuntimeName:    pulumi.String(runtimeNameStrings[spec.RuntimeName]),
		RuntimeVersion: pulumi.String(spec.RuntimeVersion),

		SiteConfig: buildSiteConfig(spec),
		Tags:       pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.StorageAccessKey != nil {
		functionAppArgs.StorageAccessKey = pulumi.StringPtr(spec.StorageAccessKey.GetValue())
	}
	if spec.StorageUserAssignedIdentityId != nil {
		functionAppArgs.StorageUserAssignedIdentityId = pulumi.StringPtr(spec.StorageUserAssignedIdentityId.GetValue())
	}

	// Presence-guarded proto defaults: IaC inputs never materialize
	// them, so an unset field must deploy the spec's documented default,
	// not the Go zero value.

	// Scale: per-instance memory, the fan-out ceiling, and optional
	// per-instance HTTP concurrency (absent lets Azure pick the
	// runtime's default for the memory size).
	instanceMemory := 2048
	if spec.InstanceMemoryInMb != nil {
		instanceMemory = int(spec.GetInstanceMemoryInMb())
	}
	functionAppArgs.InstanceMemoryInMb = pulumi.IntPtr(instanceMemory)

	maximumInstanceCount := 100
	if spec.MaximumInstanceCount != nil {
		maximumInstanceCount = int(spec.GetMaximumInstanceCount())
	}
	functionAppArgs.MaximumInstanceCount = pulumi.IntPtr(maximumInstanceCount)

	if spec.HttpConcurrency != nil {
		functionAppArgs.HttpConcurrency = pulumi.IntPtr(int(spec.GetHttpConcurrency()))
	}

	// Always-ready pools: the counts' sum must stay within
	// maximum_instance_count (Azure enforces at apply time). Azure
	// lower-cases pool names on save.
	if len(spec.AlwaysReady) > 0 {
		pools := make(appservice.AppFlexConsumptionAlwaysReadyArray, 0, len(spec.AlwaysReady))
		for _, pool := range spec.AlwaysReady {
			poolArgs := appservice.AppFlexConsumptionAlwaysReadyArgs{
				Name: pulumi.String(pool.Name),
			}
			if pool.InstanceCount != nil {
				poolArgs.InstanceCount = pulumi.IntPtr(int(pool.GetInstanceCount()))
			}
			pools = append(pools, poolArgs)
		}
		functionAppArgs.AlwaysReadies = pools
	}

	enabled := true
	if spec.Enabled != nil {
		enabled = spec.GetEnabled()
	}
	functionAppArgs.Enabled = pulumi.BoolPtr(enabled)

	httpsOnly := true
	if spec.HttpsOnly != nil {
		httpsOnly = spec.GetHttpsOnly()
	}
	functionAppArgs.HttpsOnly = pulumi.BoolPtr(httpsOnly)

	publicNetworkAccess := true
	if spec.PublicNetworkAccessEnabled != nil {
		publicNetworkAccess = spec.GetPublicNetworkAccessEnabled()
	}
	functionAppArgs.PublicNetworkAccessEnabled = pulumi.BoolPtr(publicNetworkAccess)

	if spec.VirtualNetworkSubnetId != nil {
		functionAppArgs.VirtualNetworkSubnetId = pulumi.StringPtr(spec.VirtualNetworkSubnetId.GetValue())
	}

	// Client certificate (mutual TLS) posture.
	clientCertEnabled := false
	if spec.ClientCertificateEnabled != nil {
		clientCertEnabled = spec.GetClientCertificateEnabled()
	}
	functionAppArgs.ClientCertificateEnabled = pulumi.BoolPtr(clientCertEnabled)

	clientCertMode := "Optional"
	if spec.ClientCertificateMode != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionClientCertificateMode_azure_function_app_flex_consumption_client_certificate_mode_unspecified {
		clientCertMode = clientCertificateModeStrings[spec.ClientCertificateMode]
	}
	functionAppArgs.ClientCertificateMode = pulumi.StringPtr(clientCertMode)

	if spec.ClientCertificateExclusionPaths != "" {
		functionAppArgs.ClientCertificateExclusionPaths = pulumi.StringPtr(spec.ClientCertificateExclusionPaths)
	}

	// Basic-auth publishing toggle: disabling it closes the classic
	// credential-based deployment path and forces identity-based
	// deployment.
	webdeployBasicAuth := true
	if spec.WebdeployPublishBasicAuthenticationEnabled != nil {
		webdeployBasicAuth = spec.GetWebdeployPublishBasicAuthenticationEnabled()
	}
	functionAppArgs.WebdeployPublishBasicAuthenticationEnabled = pulumi.BoolPtr(webdeployBasicAuth)

	if spec.ZipDeployFile != "" {
		functionAppArgs.ZipDeployFile = pulumi.StringPtr(spec.ZipDeployFile)
	}

	// App settings; auth provider secrets referenced by setting name land
	// here too.
	if len(spec.AppSettings) > 0 {
		functionAppArgs.AppSettings = pulumi.ToStringMap(spec.AppSettings)
	}

	if len(spec.ConnectionStrings) > 0 {
		functionAppArgs.ConnectionStrings = buildConnectionStrings(spec.ConnectionStrings)
	}

	// Settings pinned to the production slot during slot swaps.
	if spec.StickySettings != nil {
		stickyArgs := &appservice.AppFlexConsumptionStickySettingsArgs{}
		if len(spec.StickySettings.AppSettingNames) > 0 {
			stickyArgs.AppSettingNames = pulumi.ToStringArray(spec.StickySettings.AppSettingNames)
		}
		if len(spec.StickySettings.ConnectionStringNames) > 0 {
			stickyArgs.ConnectionStringNames = pulumi.ToStringArray(spec.StickySettings.ConnectionStringNames)
		}
		functionAppArgs.StickySettings = stickyArgs
	}

	if spec.Identity != nil {
		functionAppArgs.Identity = buildIdentity(spec.Identity)
	}

	if spec.AuthSettingsV2 != nil {
		functionAppArgs.AuthSettingsV2 = buildAuthSettingsV2(spec.AuthSettingsV2)
	}

	// Create the Flex Consumption Function App
	functionApp, err := appservice.NewAppFlexConsumption(ctx,
		spec.FunctionAppName,
		functionAppArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Flex Consumption Function App %s", spec.FunctionAppName)
	}

	// Export outputs. The outbound IP sets are exported as real
	// lists so they flatten onto the repeated proto outputs identically
	// on both engines.
	ctx.Export(OpFunctionAppId, functionApp.ID())
	ctx.Export(OpDefaultHostname, functionApp.DefaultHostname)
	ctx.Export(OpOutboundIpAddresses, functionApp.OutboundIpAddressLists)
	ctx.Export(OpPossibleOutboundIpAddresses, functionApp.PossibleOutboundIpAddressLists)
	ctx.Export(OpCustomDomainVerificationId, pulumi.Unsecret(functionApp.CustomDomainVerificationId))
	ctx.Export(OpKind, functionApp.Kind)

	// The identity outputs populate only when a system-assigned identity
	// exists.
	ctx.Export(OpIdentityPrincipalId, functionApp.Identity.ApplyT(func(identity *appservice.AppFlexConsumptionIdentity) string {
		if identity != nil && identity.PrincipalId != nil {
			return *identity.PrincipalId
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpIdentityTenantId, functionApp.Identity.ApplyT(func(identity *appservice.AppFlexConsumptionIdentity) string {
		if identity != nil && identity.TenantId != nil {
			return *identity.TenantId
		}
		return ""
	}).(pulumi.StringOutput))

	// The site-level publishing credential -- grants deploy access while
	// basic-auth publishing is enabled. The provider marks the whole
	// credential block secret; the name is a username, so it is unwrapped,
	// and the password is exported as a secret.
	ctx.Export(OpSiteCredentialName, pulumi.Unsecret(functionApp.SiteCredentials.ApplyT(func(creds []appservice.AppFlexConsumptionSiteCredential) string {
		if len(creds) > 0 && creds[0].Name != nil {
			return *creds[0].Name
		}
		return ""
	}).(pulumi.StringOutput)))

	ctx.Export(OpSiteCredentialPassword, pulumi.ToSecret(functionApp.SiteCredentials.ApplyT(func(creds []appservice.AppFlexConsumptionSiteCredential) string {
		if len(creds) > 0 && creds[0].Password != nil {
			return *creds[0].Password
		}
		return ""
	}).(pulumi.StringOutput)))

	return nil
}

// ---------------------------------------------------------------------------
// Site Config
// ---------------------------------------------------------------------------

func buildSiteConfig(spec *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionSpec) *appservice.AppFlexConsumptionSiteConfigArgs {
	sc := spec.GetSiteConfig()
	siteConfig := &appservice.AppFlexConsumptionSiteConfigArgs{}

	if sc.ApiManagementApiId != "" {
		siteConfig.ApiManagementApiId = pulumi.StringPtr(sc.ApiManagementApiId)
	}
	if sc.ApiDefinitionUrl != "" {
		siteConfig.ApiDefinitionUrl = pulumi.StringPtr(sc.ApiDefinitionUrl)
	}
	if sc.AppCommandLine != "" {
		siteConfig.AppCommandLine = pulumi.StringPtr(sc.AppCommandLine)
	}

	// Both App Insights values travel as app settings on the wire; the
	// connection string lives on the parent spec (a typed reference to
	// AzureApplicationInsights), the classic key here.
	if sc.ApplicationInsightsKey != "" {
		siteConfig.ApplicationInsightsKey = pulumi.StringPtr(sc.ApplicationInsightsKey)
	}
	if spec.ApplicationInsightsConnectionString != nil {
		siteConfig.ApplicationInsightsConnectionString = pulumi.StringPtr(spec.ApplicationInsightsConnectionString.GetValue())
	}

	// Azure applies this block on update operations only and never
	// returns it on read.
	if sc.AppServiceLogs != nil {
		logsArgs := &appservice.AppFlexConsumptionSiteConfigAppServiceLogsArgs{}
		diskQuota := 35
		if sc.AppServiceLogs.DiskQuotaMb != nil {
			diskQuota = int(sc.AppServiceLogs.GetDiskQuotaMb())
		}
		logsArgs.DiskQuotaMb = pulumi.IntPtr(diskQuota)
		if sc.AppServiceLogs.RetentionPeriodDays != nil {
			logsArgs.RetentionPeriodDays = pulumi.IntPtr(int(sc.AppServiceLogs.GetRetentionPeriodDays()))
		}
		siteConfig.AppServiceLogs = logsArgs
	}

	useManagedIdentityAcr := false
	if sc.ContainerRegistryUseManagedIdentity != nil {
		useManagedIdentityAcr = sc.GetContainerRegistryUseManagedIdentity()
	}
	siteConfig.ContainerRegistryUseManagedIdentity = pulumi.BoolPtr(useManagedIdentityAcr)

	if len(sc.DefaultDocuments) > 0 {
		siteConfig.DefaultDocuments = pulumi.ToStringArray(sc.DefaultDocuments)
	}

	// Accepted by Azure but never read back on this hosting model
	// (always_ready is the flex-native warm-instance mechanism).
	if sc.ElasticInstanceMinimum != nil {
		siteConfig.ElasticInstanceMinimum = pulumi.IntPtr(int(sc.GetElasticInstanceMinimum()))
	}

	http2 := false
	if sc.Http2Enabled != nil {
		http2 = sc.GetHttp2Enabled()
	}
	siteConfig.Http2Enabled = pulumi.BoolPtr(http2)

	websockets := false
	if sc.WebsocketsEnabled != nil {
		websockets = sc.GetWebsocketsEnabled()
	}
	siteConfig.WebsocketsEnabled = pulumi.BoolPtr(websockets)

	vnetRouteAll := false
	if sc.VnetRouteAllEnabled != nil {
		vnetRouteAll = sc.GetVnetRouteAllEnabled()
	}
	siteConfig.VnetRouteAllEnabled = pulumi.BoolPtr(vnetRouteAll)

	// Enum-mapped dials: unset deploys the documented defaults.
	loadBalancing := "LeastRequests"
	if sc.LoadBalancingMode != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionLoadBalancingMode_azure_function_app_flex_consumption_load_balancing_mode_unspecified {
		loadBalancing = loadBalancingModeStrings[sc.LoadBalancingMode]
	}
	siteConfig.LoadBalancingMode = pulumi.StringPtr(loadBalancing)

	pipelineMode := "Integrated"
	if sc.ManagedPipelineMode != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionManagedPipelineMode_azure_function_app_flex_consumption_managed_pipeline_mode_unspecified {
		pipelineMode = managedPipelineModeStrings[sc.ManagedPipelineMode]
	}
	siteConfig.ManagedPipelineMode = pulumi.StringPtr(pipelineMode)

	remoteDebugging := false
	if sc.RemoteDebuggingEnabled != nil {
		remoteDebugging = sc.GetRemoteDebuggingEnabled()
	}
	siteConfig.RemoteDebuggingEnabled = pulumi.BoolPtr(remoteDebugging)

	if sc.RemoteDebuggingVersion != "" {
		siteConfig.RemoteDebuggingVersion = pulumi.StringPtr(sc.RemoteDebuggingVersion)
	}

	if sc.RuntimeScaleMonitoringEnabled != nil {
		siteConfig.RuntimeScaleMonitoringEnabled = pulumi.BoolPtr(sc.GetRuntimeScaleMonitoringEnabled())
	}

	if sc.HealthCheckPath != "" {
		siteConfig.HealthCheckPath = pulumi.StringPtr(sc.HealthCheckPath)
	}
	if sc.HealthCheckEvictionTimeInMin != nil {
		siteConfig.HealthCheckEvictionTimeInMin = pulumi.IntPtr(int(sc.GetHealthCheckEvictionTimeInMin()))
	}

	if sc.WorkerCount != nil {
		siteConfig.WorkerCount = pulumi.IntPtr(int(sc.GetWorkerCount()))
	}

	// TLS floors: unset deploys 1.2 on both the main and SCM sites.
	minTls := "1.2"
	if sc.MinimumTlsVersion != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionTlsVersion_azure_function_app_flex_consumption_tls_version_unspecified {
		minTls = tlsVersionStrings[sc.MinimumTlsVersion]
	}
	siteConfig.MinimumTlsVersion = pulumi.StringPtr(minTls)

	scmMinTls := "1.2"
	if sc.ScmMinimumTlsVersion != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionTlsVersion_azure_function_app_flex_consumption_tls_version_unspecified {
		scmMinTls = tlsVersionStrings[sc.ScmMinimumTlsVersion]
	}
	siteConfig.ScmMinimumTlsVersion = pulumi.StringPtr(scmMinTls)

	if sc.Cors != nil {
		siteConfig.Cors = buildCors(sc.Cors)
	}

	if len(sc.IpRestrictions) > 0 {
		siteConfig.IpRestrictions = buildIpRestrictions(sc.IpRestrictions)
	}
	ipDefaultAction := "Allow"
	if sc.IpRestrictionDefaultAction != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestrictionAction_azure_function_app_flex_consumption_ip_restriction_action_unspecified {
		ipDefaultAction = ipRestrictionActionStrings[sc.IpRestrictionDefaultAction]
	}
	siteConfig.IpRestrictionDefaultAction = pulumi.StringPtr(ipDefaultAction)

	scmUseMain := false
	if sc.ScmUseMainIpRestriction != nil {
		scmUseMain = sc.GetScmUseMainIpRestriction()
	}
	siteConfig.ScmUseMainIpRestriction = pulumi.BoolPtr(scmUseMain)

	if len(sc.ScmIpRestrictions) > 0 {
		siteConfig.ScmIpRestrictions = buildScmIpRestrictions(sc.ScmIpRestrictions)
	}
	scmDefaultAction := "Allow"
	if sc.ScmIpRestrictionDefaultAction != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestrictionAction_azure_function_app_flex_consumption_ip_restriction_action_unspecified {
		scmDefaultAction = ipRestrictionActionStrings[sc.ScmIpRestrictionDefaultAction]
	}
	siteConfig.ScmIpRestrictionDefaultAction = pulumi.StringPtr(scmDefaultAction)

	return siteConfig
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func buildIdentity(spec *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIdentity) *appservice.AppFlexConsumptionIdentityArgs {
	identity := &appservice.AppFlexConsumptionIdentityArgs{
		Type: pulumi.String(identityTypeStrings[spec.Type]),
	}

	if len(spec.IdentityIds) > 0 {
		ids := make(pulumi.StringArray, 0, len(spec.IdentityIds))
		for _, ref := range spec.IdentityIds {
			ids = append(ids, pulumi.String(ref.GetValue()))
		}
		identity.IdentityIds = ids
	}

	return identity
}

// ---------------------------------------------------------------------------
// Connection Strings
// ---------------------------------------------------------------------------

func buildConnectionStrings(specs []*azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionString) appservice.AppFlexConsumptionConnectionStringArray {
	connStrings := make(appservice.AppFlexConsumptionConnectionStringArray, 0, len(specs))
	for _, cs := range specs {
		connStrings = append(connStrings, appservice.AppFlexConsumptionConnectionStringArgs{
			Name:  pulumi.String(cs.Name),
			Type:  pulumi.String(connectionStringTypeStrings[cs.Type]),
			Value: pulumi.String(cs.Value.GetValue()),
		})
	}
	return connStrings
}

// ---------------------------------------------------------------------------
// CORS
// ---------------------------------------------------------------------------

func buildCors(spec *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionCorsSettings) *appservice.AppFlexConsumptionSiteConfigCorsArgs {
	cors := &appservice.AppFlexConsumptionSiteConfigCorsArgs{
		AllowedOrigins: pulumi.ToStringArray(spec.AllowedOrigins),
	}
	supportCredentials := false
	if spec.SupportCredentials != nil {
		supportCredentials = spec.GetSupportCredentials()
	}
	cors.SupportCredentials = pulumi.BoolPtr(supportCredentials)
	return cors
}

// ---------------------------------------------------------------------------
// IP Restrictions
// ---------------------------------------------------------------------------

func buildIpRestrictions(specs []*azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestriction) appservice.AppFlexConsumptionSiteConfigIpRestrictionArray {
	restrictions := make(appservice.AppFlexConsumptionSiteConfigIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := appservice.AppFlexConsumptionSiteConfigIpRestrictionArgs{}

		if r.Name != "" {
			restriction.Name = pulumi.StringPtr(r.Name)
		}
		if r.Priority != nil {
			restriction.Priority = pulumi.IntPtr(int(r.GetPriority()))
		}
		action := "Allow"
		if r.Action != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestrictionAction_azure_function_app_flex_consumption_ip_restriction_action_unspecified {
			action = ipRestrictionActionStrings[r.Action]
		}
		restriction.Action = pulumi.StringPtr(action)
		if r.IpAddress != "" {
			restriction.IpAddress = pulumi.StringPtr(r.IpAddress)
		}
		if r.ServiceTag != "" {
			restriction.ServiceTag = pulumi.StringPtr(r.ServiceTag)
		}
		if r.VirtualNetworkSubnetId != nil {
			restriction.VirtualNetworkSubnetId = pulumi.StringPtr(r.VirtualNetworkSubnetId.GetValue())
		}
		if r.Description != "" {
			restriction.Description = pulumi.StringPtr(r.Description)
		}
		if r.Headers != nil {
			restriction.Headers = buildIpRestrictionHeaders(r.Headers)
		}

		restrictions = append(restrictions, restriction)
	}
	return restrictions
}

func buildScmIpRestrictions(specs []*azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestriction) appservice.AppFlexConsumptionSiteConfigScmIpRestrictionArray {
	restrictions := make(appservice.AppFlexConsumptionSiteConfigScmIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := appservice.AppFlexConsumptionSiteConfigScmIpRestrictionArgs{}

		if r.Name != "" {
			restriction.Name = pulumi.StringPtr(r.Name)
		}
		if r.Priority != nil {
			restriction.Priority = pulumi.IntPtr(int(r.GetPriority()))
		}
		action := "Allow"
		if r.Action != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestrictionAction_azure_function_app_flex_consumption_ip_restriction_action_unspecified {
			action = ipRestrictionActionStrings[r.Action]
		}
		restriction.Action = pulumi.StringPtr(action)
		if r.IpAddress != "" {
			restriction.IpAddress = pulumi.StringPtr(r.IpAddress)
		}
		if r.ServiceTag != "" {
			restriction.ServiceTag = pulumi.StringPtr(r.ServiceTag)
		}
		if r.VirtualNetworkSubnetId != nil {
			restriction.VirtualNetworkSubnetId = pulumi.StringPtr(r.VirtualNetworkSubnetId.GetValue())
		}
		if r.Description != "" {
			restriction.Description = pulumi.StringPtr(r.Description)
		}
		if r.Headers != nil {
			restriction.Headers = buildScmIpRestrictionHeaders(r.Headers)
		}

		restrictions = append(restrictions, restriction)
	}
	return restrictions
}

// refValues unwraps a repeated StringValueOrRef into the resolved literal
// strings (the platform resolves valueFrom references before the module
// runs, so GetValue() always returns the resolved literal).
func refValues(refs []*foreignkeyv1.StringValueOrRef) []string {
	values := make([]string, 0, len(refs))
	for _, ref := range refs {
		values = append(values, ref.GetValue())
	}
	return values
}

func buildIpRestrictionHeaders(h *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestrictionHeaders) *appservice.AppFlexConsumptionSiteConfigIpRestrictionHeadersArgs {
	headers := &appservice.AppFlexConsumptionSiteConfigIpRestrictionHeadersArgs{}

	if len(h.XForwardedFor) > 0 {
		headers.XForwardedFors = pulumi.ToStringArray(h.XForwardedFor)
	}
	if len(h.XForwardedHost) > 0 {
		headers.XForwardedHosts = pulumi.ToStringArray(h.XForwardedHost)
	}
	// FDID values reference AzureFrontDoorProfile.resource_guid by
	// default -- the origin-lockdown seam pinning the app to specific
	// Front Door instances.
	if len(h.XAzureFdid) > 0 {
		headers.XAzureFdids = pulumi.ToStringArray(refValues(h.XAzureFdid))
	}
	if len(h.XFdHealthProbe) > 0 {
		headers.XFdHealthProbe = pulumi.StringPtr(h.XFdHealthProbe[0])
	}

	return headers
}

func buildScmIpRestrictionHeaders(h *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestrictionHeaders) *appservice.AppFlexConsumptionSiteConfigScmIpRestrictionHeadersArgs {
	headers := &appservice.AppFlexConsumptionSiteConfigScmIpRestrictionHeadersArgs{}

	if len(h.XForwardedFor) > 0 {
		headers.XForwardedFors = pulumi.ToStringArray(h.XForwardedFor)
	}
	if len(h.XForwardedHost) > 0 {
		headers.XForwardedHosts = pulumi.ToStringArray(h.XForwardedHost)
	}
	if len(h.XAzureFdid) > 0 {
		headers.XAzureFdids = pulumi.ToStringArray(refValues(h.XAzureFdid))
	}
	if len(h.XFdHealthProbe) > 0 {
		headers.XFdHealthProbe = pulumi.StringPtr(h.XFdHealthProbe[0])
	}

	return headers
}

// ---------------------------------------------------------------------------
// Easy Auth v2
// ---------------------------------------------------------------------------

// buildAuthSettingsV2 maps App Service built-in authentication. Provider
// client secrets are referenced by APP SETTING NAME -- the secret values
// live in app_settings (or Key Vault references), never in this block.
func buildAuthSettingsV2(spec *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionAuthSettingsV2) *appservice.AppFlexConsumptionAuthSettingsV2Args {
	args := &appservice.AppFlexConsumptionAuthSettingsV2Args{
		Login: buildAuthV2Login(spec.Login),
	}

	authEnabled := false
	if spec.AuthEnabled != nil {
		authEnabled = spec.GetAuthEnabled()
	}
	args.AuthEnabled = pulumi.BoolPtr(authEnabled)

	runtimeVersion := "~1"
	if spec.RuntimeVersion != nil && spec.GetRuntimeVersion() != "" {
		runtimeVersion = spec.GetRuntimeVersion()
	}
	args.RuntimeVersion = pulumi.StringPtr(runtimeVersion)

	if spec.ConfigFilePath != "" {
		args.ConfigFilePath = pulumi.StringPtr(spec.ConfigFilePath)
	}

	requireAuth := false
	if spec.RequireAuthentication != nil {
		requireAuth = spec.GetRequireAuthentication()
	}
	args.RequireAuthentication = pulumi.BoolPtr(requireAuth)

	unauthenticatedAction := "RedirectToLoginPage"
	if spec.UnauthenticatedAction != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionUnauthenticatedAction_azure_function_app_flex_consumption_unauthenticated_action_unspecified {
		unauthenticatedAction = unauthenticatedActionStrings[spec.UnauthenticatedAction]
	}
	args.UnauthenticatedAction = pulumi.StringPtr(unauthenticatedAction)

	if spec.DefaultProvider != "" {
		args.DefaultProvider = pulumi.StringPtr(spec.DefaultProvider)
	}
	if len(spec.ExcludedPaths) > 0 {
		args.ExcludedPaths = pulumi.ToStringArray(spec.ExcludedPaths)
	}

	requireHttps := true
	if spec.RequireHttps != nil {
		requireHttps = spec.GetRequireHttps()
	}
	args.RequireHttps = pulumi.BoolPtr(requireHttps)

	httpRoutePrefix := "/.auth"
	if spec.HttpRouteApiPrefix != nil && spec.GetHttpRouteApiPrefix() != "" {
		httpRoutePrefix = spec.GetHttpRouteApiPrefix()
	}
	args.HttpRouteApiPrefix = pulumi.StringPtr(httpRoutePrefix)

	forwardProxy := "NoProxy"
	if spec.ForwardProxyConvention != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionForwardProxyConvention_azure_function_app_flex_consumption_forward_proxy_convention_unspecified {
		forwardProxy = forwardProxyConventionStrings[spec.ForwardProxyConvention]
	}
	args.ForwardProxyConvention = pulumi.StringPtr(forwardProxy)

	if spec.ForwardProxyCustomHostHeaderName != "" {
		args.ForwardProxyCustomHostHeaderName = pulumi.StringPtr(spec.ForwardProxyCustomHostHeaderName)
	}
	if spec.ForwardProxyCustomSchemeHeaderName != "" {
		args.ForwardProxyCustomSchemeHeaderName = pulumi.StringPtr(spec.ForwardProxyCustomSchemeHeaderName)
	}

	if spec.AppleV2 != nil {
		args.AppleV2 = &appservice.AppFlexConsumptionAuthSettingsV2AppleV2Args{
			ClientId:                pulumi.String(spec.AppleV2.ClientId),
			ClientSecretSettingName: pulumi.String(spec.AppleV2.ClientSecretSettingName),
		}
	}

	if spec.ActiveDirectoryV2 != nil {
		aad := spec.ActiveDirectoryV2
		aadArgs := &appservice.AppFlexConsumptionAuthSettingsV2ActiveDirectoryV2Args{
			ClientId:           pulumi.String(aad.ClientId),
			TenantAuthEndpoint: pulumi.String(aad.TenantAuthEndpoint),
		}
		if aad.ClientSecretSettingName != "" {
			aadArgs.ClientSecretSettingName = pulumi.StringPtr(aad.ClientSecretSettingName)
		}
		if aad.ClientSecretCertificateThumbprint != "" {
			aadArgs.ClientSecretCertificateThumbprint = pulumi.StringPtr(aad.ClientSecretCertificateThumbprint)
		}
		if len(aad.LoginParameters) > 0 {
			aadArgs.LoginParameters = pulumi.ToStringMap(aad.LoginParameters)
		}
		wwwAuthDisabled := false
		if aad.WwwAuthenticationDisabled != nil {
			wwwAuthDisabled = aad.GetWwwAuthenticationDisabled()
		}
		aadArgs.WwwAuthenticationDisabled = pulumi.BoolPtr(wwwAuthDisabled)
		if len(aad.JwtAllowedGroups) > 0 {
			aadArgs.JwtAllowedGroups = pulumi.ToStringArray(aad.JwtAllowedGroups)
		}
		if len(aad.JwtAllowedClientApplications) > 0 {
			aadArgs.JwtAllowedClientApplications = pulumi.ToStringArray(aad.JwtAllowedClientApplications)
		}
		if len(aad.AllowedGroups) > 0 {
			aadArgs.AllowedGroups = pulumi.ToStringArray(aad.AllowedGroups)
		}
		if len(aad.AllowedIdentities) > 0 {
			aadArgs.AllowedIdentities = pulumi.ToStringArray(aad.AllowedIdentities)
		}
		if len(aad.AllowedApplications) > 0 {
			aadArgs.AllowedApplications = pulumi.ToStringArray(aad.AllowedApplications)
		}
		if len(aad.AllowedAudiences) > 0 {
			aadArgs.AllowedAudiences = pulumi.ToStringArray(aad.AllowedAudiences)
		}
		args.ActiveDirectoryV2 = aadArgs
	}

	if spec.AzureStaticWebAppV2 != nil {
		args.AzureStaticWebAppV2 = &appservice.AppFlexConsumptionAuthSettingsV2AzureStaticWebAppV2Args{
			ClientId: pulumi.String(spec.AzureStaticWebAppV2.ClientId),
		}
	}

	if len(spec.CustomOidcV2) > 0 {
		oidcs := make(appservice.AppFlexConsumptionAuthSettingsV2CustomOidcV2Array, 0, len(spec.CustomOidcV2))
		for _, oidc := range spec.CustomOidcV2 {
			oidcArgs := appservice.AppFlexConsumptionAuthSettingsV2CustomOidcV2Args{
				Name:                        pulumi.String(oidc.Name),
				ClientId:                    pulumi.String(oidc.ClientId),
				OpenidConfigurationEndpoint: pulumi.String(oidc.OpenidConfigurationEndpoint),
			}
			if oidc.NameClaimType != "" {
				oidcArgs.NameClaimType = pulumi.StringPtr(oidc.NameClaimType)
			}
			if len(oidc.Scopes) > 0 {
				oidcArgs.Scopes = pulumi.ToStringArray(oidc.Scopes)
			}
			oidcs = append(oidcs, oidcArgs)
		}
		args.CustomOidcV2s = oidcs
	}

	if spec.FacebookV2 != nil {
		fb := spec.FacebookV2
		fbArgs := &appservice.AppFlexConsumptionAuthSettingsV2FacebookV2Args{
			AppId:                pulumi.String(fb.AppId),
			AppSecretSettingName: pulumi.String(fb.AppSecretSettingName),
		}
		if fb.GraphApiVersion != "" {
			fbArgs.GraphApiVersion = pulumi.StringPtr(fb.GraphApiVersion)
		}
		if len(fb.LoginScopes) > 0 {
			fbArgs.LoginScopes = pulumi.ToStringArray(fb.LoginScopes)
		}
		args.FacebookV2 = fbArgs
	}

	if spec.GithubV2 != nil {
		gh := spec.GithubV2
		ghArgs := &appservice.AppFlexConsumptionAuthSettingsV2GithubV2Args{
			ClientId:                pulumi.String(gh.ClientId),
			ClientSecretSettingName: pulumi.String(gh.ClientSecretSettingName),
		}
		if len(gh.LoginScopes) > 0 {
			ghArgs.LoginScopes = pulumi.ToStringArray(gh.LoginScopes)
		}
		args.GithubV2 = ghArgs
	}

	if spec.GoogleV2 != nil {
		g := spec.GoogleV2
		gArgs := &appservice.AppFlexConsumptionAuthSettingsV2GoogleV2Args{
			ClientId:                pulumi.String(g.ClientId),
			ClientSecretSettingName: pulumi.String(g.ClientSecretSettingName),
		}
		if len(g.AllowedAudiences) > 0 {
			gArgs.AllowedAudiences = pulumi.ToStringArray(g.AllowedAudiences)
		}
		if len(g.LoginScopes) > 0 {
			gArgs.LoginScopes = pulumi.ToStringArray(g.LoginScopes)
		}
		args.GoogleV2 = gArgs
	}

	if spec.MicrosoftV2 != nil {
		ms := spec.MicrosoftV2
		msArgs := &appservice.AppFlexConsumptionAuthSettingsV2MicrosoftV2Args{
			ClientId:                pulumi.String(ms.ClientId),
			ClientSecretSettingName: pulumi.String(ms.ClientSecretSettingName),
		}
		if len(ms.AllowedAudiences) > 0 {
			msArgs.AllowedAudiences = pulumi.ToStringArray(ms.AllowedAudiences)
		}
		if len(ms.LoginScopes) > 0 {
			msArgs.LoginScopes = pulumi.ToStringArray(ms.LoginScopes)
		}
		args.MicrosoftV2 = msArgs
	}

	if spec.TwitterV2 != nil {
		args.TwitterV2 = &appservice.AppFlexConsumptionAuthSettingsV2TwitterV2Args{
			ConsumerKey:               pulumi.String(spec.TwitterV2.ConsumerKey),
			ConsumerSecretSettingName: pulumi.String(spec.TwitterV2.ConsumerSecretSettingName),
		}
	}

	return args
}

func buildAuthV2Login(spec *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionAuthV2Login) *appservice.AppFlexConsumptionAuthSettingsV2LoginArgs {
	login := &appservice.AppFlexConsumptionAuthSettingsV2LoginArgs{}

	if spec.LogoutEndpoint != "" {
		login.LogoutEndpoint = pulumi.StringPtr(spec.LogoutEndpoint)
	}

	tokenStore := false
	if spec.TokenStoreEnabled != nil {
		tokenStore = spec.GetTokenStoreEnabled()
	}
	login.TokenStoreEnabled = pulumi.BoolPtr(tokenStore)

	tokenRefreshExtension := float64(72)
	if spec.TokenRefreshExtensionTime != nil {
		tokenRefreshExtension = spec.GetTokenRefreshExtensionTime()
	}
	login.TokenRefreshExtensionTime = pulumi.Float64Ptr(tokenRefreshExtension)

	if spec.TokenStorePath != "" {
		login.TokenStorePath = pulumi.StringPtr(spec.TokenStorePath)
	}
	if spec.TokenStoreSasSettingName != "" {
		login.TokenStoreSasSettingName = pulumi.StringPtr(spec.TokenStoreSasSettingName)
	}

	preserveFragments := false
	if spec.PreserveUrlFragmentsForLogins != nil {
		preserveFragments = spec.GetPreserveUrlFragmentsForLogins()
	}
	login.PreserveUrlFragmentsForLogins = pulumi.BoolPtr(preserveFragments)

	if len(spec.AllowedExternalRedirectUrls) > 0 {
		login.AllowedExternalRedirectUrls = pulumi.ToStringArray(spec.AllowedExternalRedirectUrls)
	}

	cookieConvention := "FixedTime"
	if spec.CookieExpirationConvention != azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionCookieExpirationConvention_azure_function_app_flex_consumption_cookie_expiration_convention_unspecified {
		cookieConvention = cookieExpirationConventionStrings[spec.CookieExpirationConvention]
	}
	login.CookieExpirationConvention = pulumi.StringPtr(cookieConvention)

	cookieExpiration := "08:00:00"
	if spec.CookieExpirationTime != nil && spec.GetCookieExpirationTime() != "" {
		cookieExpiration = spec.GetCookieExpirationTime()
	}
	login.CookieExpirationTime = pulumi.StringPtr(cookieExpiration)

	validateNonce := true
	if spec.ValidateNonce != nil {
		validateNonce = spec.GetValidateNonce()
	}
	login.ValidateNonce = pulumi.BoolPtr(validateNonce)

	nonceExpiration := "00:05:00"
	if spec.NonceExpirationTime != nil && spec.GetNonceExpirationTime() != "" {
		nonceExpiration = spec.GetNonceExpirationTime()
	}
	login.NonceExpirationTime = pulumi.StringPtr(nonceExpiration)

	return login
}
