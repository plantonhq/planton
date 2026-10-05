package module

import (
	"github.com/pkg/errors"
	azurefunctionappv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefunctionapp/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefunctionappv1alpha1.AzureFunctionAppIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFunctionApp.Spec

	// The Linux Function App: event-driven compute on an App Service
	// Plan. Name, region, and resource group are ForceNew; moving between
	// Dynamic (Consumption) and other tiers also forces recreation.
	functionAppArgs := &appservice.LinuxFunctionAppArgs{
		Name:              pulumi.String(spec.FunctionAppName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		ServicePlanId:     pulumi.String(spec.ServicePlanId.GetValue()),
		SiteConfig:        buildSiteConfig(spec),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// The runtime-state storage binding: exactly one of account name or
	// Key Vault secret ID applies (spec-enforced); the access key and
	// managed identity are mutually exclusive authentication methods.
	if spec.StorageAccountName != nil {
		functionAppArgs.StorageAccountName = pulumi.StringPtr(spec.StorageAccountName.GetValue())
	}
	if spec.StorageAccountAccessKey != nil {
		functionAppArgs.StorageAccountAccessKey = pulumi.StringPtr(spec.StorageAccountAccessKey.GetValue())
	}
	if spec.StorageUsesManagedIdentity != nil && spec.GetStorageUsesManagedIdentity() {
		functionAppArgs.StorageUsesManagedIdentity = pulumi.BoolPtr(true)
	}
	if spec.StorageKeyVaultSecretId != "" {
		functionAppArgs.StorageKeyVaultSecretId = pulumi.StringPtr(spec.StorageKeyVaultSecretId)
	}

	functionsExtensionVersion := "~4"
	if spec.FunctionsExtensionVersion != nil && spec.GetFunctionsExtensionVersion() != "" {
		functionsExtensionVersion = spec.GetFunctionsExtensionVersion()
	}
	functionAppArgs.FunctionsExtensionVersion = pulumi.StringPtr(functionsExtensionVersion)

	// The Consumption-plan cost circuit breaker (GB-seconds per day; 0 =
	// unlimited).
	if spec.DailyMemoryTimeQuota != nil {
		functionAppArgs.DailyMemoryTimeQuota = pulumi.IntPtr(int(spec.GetDailyMemoryTimeQuota()))
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
		stickyArgs := &appservice.LinuxFunctionAppStickySettingsArgs{}
		if len(spec.StickySettings.AppSettingNames) > 0 {
			stickyArgs.AppSettingNames = pulumi.ToStringArray(spec.StickySettings.AppSettingNames)
		}
		if len(spec.StickySettings.ConnectionStringNames) > 0 {
			stickyArgs.ConnectionStringNames = pulumi.ToStringArray(spec.StickySettings.ConnectionStringNames)
		}
		functionAppArgs.StickySettings = stickyArgs
	}

	// Presence-guarded proto defaults: IaC inputs never materialize
	// them, so an unset field must deploy the spec's documented default,
	// not the Go zero value.

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

	enabled := true
	if spec.Enabled != nil {
		enabled = spec.GetEnabled()
	}
	functionAppArgs.Enabled = pulumi.BoolPtr(enabled)

	builtinLogging := true
	if spec.BuiltinLoggingEnabled != nil {
		builtinLogging = spec.GetBuiltinLoggingEnabled()
	}
	functionAppArgs.BuiltinLoggingEnabled = pulumi.BoolPtr(builtinLogging)

	contentShareForceDisabled := false
	if spec.ContentShareForceDisabled != nil {
		contentShareForceDisabled = spec.GetContentShareForceDisabled()
	}
	functionAppArgs.ContentShareForceDisabled = pulumi.BoolPtr(contentShareForceDisabled)

	// Client certificate (mutual TLS) posture.
	clientCertEnabled := false
	if spec.ClientCertificateEnabled != nil {
		clientCertEnabled = spec.GetClientCertificateEnabled()
	}
	functionAppArgs.ClientCertificateEnabled = pulumi.BoolPtr(clientCertEnabled)

	clientCertMode := "Optional"
	if spec.ClientCertificateMode != azurefunctionappv1alpha1.AzureFunctionAppClientCertificateMode_azure_function_app_client_certificate_mode_unspecified {
		clientCertMode = clientCertificateModeStrings[spec.ClientCertificateMode]
	}
	functionAppArgs.ClientCertificateMode = pulumi.StringPtr(clientCertMode)

	if spec.ClientCertificateExclusionPaths != "" {
		functionAppArgs.ClientCertificateExclusionPaths = pulumi.StringPtr(spec.ClientCertificateExclusionPaths)
	}

	// VNet integration. Image pulls and backup/restore traffic only ride
	// the VNet when their dedicated toggles say so (spec-gated to
	// require the subnet).
	if spec.VirtualNetworkSubnetId != nil {
		functionAppArgs.VirtualNetworkSubnetId = pulumi.StringPtr(spec.VirtualNetworkSubnetId.GetValue())
	}
	vnetImagePull := false
	if spec.VnetImagePullEnabled != nil {
		vnetImagePull = spec.GetVnetImagePullEnabled()
	}
	functionAppArgs.VnetImagePullEnabled = pulumi.BoolPtr(vnetImagePull)

	vnetBackupRestore := false
	if spec.VirtualNetworkBackupRestoreEnabled != nil {
		vnetBackupRestore = spec.GetVirtualNetworkBackupRestoreEnabled()
	}
	functionAppArgs.VirtualNetworkBackupRestoreEnabled = pulumi.BoolPtr(vnetBackupRestore)

	if spec.Identity != nil {
		functionAppArgs.Identity = buildIdentity(spec.Identity)
	}

	if spec.KeyVaultReferenceIdentityId != nil {
		functionAppArgs.KeyVaultReferenceIdentityId = pulumi.StringPtr(spec.KeyVaultReferenceIdentityId.GetValue())
	}

	// Basic-auth publishing toggles: disabling both closes the classic
	// credential-based deployment paths (Web Deploy + FTP) and forces
	// identity-based deployment.
	webdeployBasicAuth := true
	if spec.WebdeployPublishBasicAuthenticationEnabled != nil {
		webdeployBasicAuth = spec.GetWebdeployPublishBasicAuthenticationEnabled()
	}
	functionAppArgs.WebdeployPublishBasicAuthenticationEnabled = pulumi.BoolPtr(webdeployBasicAuth)

	ftpBasicAuth := true
	if spec.FtpPublishBasicAuthenticationEnabled != nil {
		ftpBasicAuth = spec.GetFtpPublishBasicAuthenticationEnabled()
	}
	functionAppArgs.FtpPublishBasicAuthenticationEnabled = pulumi.BoolPtr(ftpBasicAuth)

	if spec.ZipDeployFile != "" {
		functionAppArgs.ZipDeployFile = pulumi.StringPtr(spec.ZipDeployFile)
	}

	if len(spec.StorageMounts) > 0 {
		functionAppArgs.StorageAccounts = buildStorageAccounts(spec.StorageMounts)
	}

	if spec.Backup != nil {
		functionAppArgs.Backup = buildBackup(spec.Backup)
	}

	if spec.AuthSettingsV2 != nil {
		functionAppArgs.AuthSettingsV2 = buildAuthSettingsV2(spec.AuthSettingsV2)
	}

	// Create the Linux Function App
	functionApp, err := appservice.NewLinuxFunctionApp(ctx,
		spec.FunctionAppName,
		functionAppArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Linux Function App %s", spec.FunctionAppName)
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
	ctx.Export(OpHostingEnvironmentId, functionApp.HostingEnvironmentId)

	// The identity outputs populate only when a system-assigned identity
	// exists.
	ctx.Export(OpIdentityPrincipalId, functionApp.Identity.ApplyT(func(identity *appservice.LinuxFunctionAppIdentity) string {
		if identity != nil && identity.PrincipalId != nil {
			return *identity.PrincipalId
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpIdentityTenantId, functionApp.Identity.ApplyT(func(identity *appservice.LinuxFunctionAppIdentity) string {
		if identity != nil && identity.TenantId != nil {
			return *identity.TenantId
		}
		return ""
	}).(pulumi.StringOutput))

	// The site-level publishing credential -- grants deploy access while
	// basic-auth publishing is enabled. The provider marks the whole
	// credential block secret; the name is a username, so it is unwrapped,
	// and the password is exported as a secret.
	ctx.Export(OpSiteCredentialName, pulumi.Unsecret(functionApp.SiteCredentials.ApplyT(func(creds []appservice.LinuxFunctionAppSiteCredential) string {
		if len(creds) > 0 && creds[0].Name != nil {
			return *creds[0].Name
		}
		return ""
	}).(pulumi.StringOutput)))

	ctx.Export(OpSiteCredentialPassword, pulumi.ToSecret(functionApp.SiteCredentials.ApplyT(func(creds []appservice.LinuxFunctionAppSiteCredential) string {
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

func buildSiteConfig(spec *azurefunctionappv1alpha1.AzureFunctionAppSpec) *appservice.LinuxFunctionAppSiteConfigArgs {
	sc := spec.GetSiteConfig()
	siteConfig := &appservice.LinuxFunctionAppSiteConfigArgs{}

	if sc.GetApplicationStack() != nil {
		siteConfig.ApplicationStack = buildApplicationStack(sc.ApplicationStack)
	}

	// Azure rejects always_on on Free/Shared plans and manages it itself
	// on serverless tiers (the plan's SKU is not visible here).
	if sc.AlwaysOn != nil {
		siteConfig.AlwaysOn = pulumi.BoolPtr(sc.GetAlwaysOn())
	}

	if sc.AppCommandLine != "" {
		siteConfig.AppCommandLine = pulumi.StringPtr(sc.AppCommandLine)
	}

	if sc.ApiManagementApiId != "" {
		siteConfig.ApiManagementApiId = pulumi.StringPtr(sc.ApiManagementApiId)
	}
	if sc.ApiDefinitionUrl != "" {
		siteConfig.ApiDefinitionUrl = pulumi.StringPtr(sc.ApiDefinitionUrl)
	}

	if len(sc.DefaultDocuments) > 0 {
		siteConfig.DefaultDocuments = pulumi.ToStringArray(sc.DefaultDocuments)
	}

	if sc.HealthCheckPath != "" {
		siteConfig.HealthCheckPath = pulumi.StringPtr(sc.HealthCheckPath)
	}
	if sc.HealthCheckEvictionTimeInMin != nil {
		siteConfig.HealthCheckEvictionTimeInMin = pulumi.IntPtr(int(sc.GetHealthCheckEvictionTimeInMin()))
	}

	// TLS floors: unset deploys 1.2 on both the main and SCM sites.
	minTls := "1.2"
	if sc.MinimumTlsVersion != azurefunctionappv1alpha1.AzureFunctionAppTlsVersion_azure_function_app_tls_version_unspecified {
		minTls = tlsVersionStrings[sc.MinimumTlsVersion]
	}
	siteConfig.MinimumTlsVersion = pulumi.StringPtr(minTls)

	scmMinTls := "1.2"
	if sc.ScmMinimumTlsVersion != azurefunctionappv1alpha1.AzureFunctionAppTlsVersion_azure_function_app_tls_version_unspecified {
		scmMinTls = tlsVersionStrings[sc.ScmMinimumTlsVersion]
	}
	siteConfig.ScmMinimumTlsVersion = pulumi.StringPtr(scmMinTls)

	// Absent accepts Azure's platform default cipher set.
	if sc.MinimumTlsCipherSuite != "" {
		siteConfig.MinimumTlsCipherSuite = pulumi.StringPtr(sc.MinimumTlsCipherSuite)
	}

	// Serverless scaling dials (Consumption / Elastic Premium).
	if sc.AppScaleLimit != nil {
		siteConfig.AppScaleLimit = pulumi.IntPtr(int(sc.GetAppScaleLimit()))
	}
	if sc.ElasticInstanceMinimum != nil {
		siteConfig.ElasticInstanceMinimum = pulumi.IntPtr(int(sc.GetElasticInstanceMinimum()))
	}
	if sc.PreWarmedInstanceCount != nil {
		siteConfig.PreWarmedInstanceCount = pulumi.IntPtr(int(sc.GetPreWarmedInstanceCount()))
	}

	if sc.WorkerCount != nil {
		siteConfig.WorkerCount = pulumi.IntPtr(int(sc.GetWorkerCount()))
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

	use32Bit := false
	if sc.Use_32BitWorker != nil {
		use32Bit = sc.GetUse_32BitWorker()
	}
	siteConfig.Use32BitWorker = pulumi.BoolPtr(use32Bit)

	vnetRouteAll := false
	if sc.VnetRouteAllEnabled != nil {
		vnetRouteAll = sc.GetVnetRouteAllEnabled()
	}
	siteConfig.VnetRouteAllEnabled = pulumi.BoolPtr(vnetRouteAll)

	// Enum-mapped dials: unset deploys the documented defaults.
	ftpsState := "Disabled"
	if sc.FtpsState != azurefunctionappv1alpha1.AzureFunctionAppFtpsState_azure_function_app_ftps_state_unspecified {
		ftpsState = ftpsStateStrings[sc.FtpsState]
	}
	siteConfig.FtpsState = pulumi.StringPtr(ftpsState)

	loadBalancing := "LeastRequests"
	if sc.LoadBalancingMode != azurefunctionappv1alpha1.AzureFunctionAppLoadBalancingMode_azure_function_app_load_balancing_mode_unspecified {
		loadBalancing = loadBalancingModeStrings[sc.LoadBalancingMode]
	}
	siteConfig.LoadBalancingMode = pulumi.StringPtr(loadBalancing)

	pipelineMode := "Integrated"
	if sc.ManagedPipelineMode != azurefunctionappv1alpha1.AzureFunctionAppManagedPipelineMode_azure_function_app_managed_pipeline_mode_unspecified {
		pipelineMode = managedPipelineModeStrings[sc.ManagedPipelineMode]
	}
	siteConfig.ManagedPipelineMode = pulumi.StringPtr(pipelineMode)

	// Remote debugging: Azure supports the current Visual Studio
	// generation only, so the debugger version is left to the platform.
	remoteDebugging := false
	if sc.RemoteDebuggingEnabled != nil {
		remoteDebugging = sc.GetRemoteDebuggingEnabled()
	}
	siteConfig.RemoteDebuggingEnabled = pulumi.BoolPtr(remoteDebugging)

	if sc.RuntimeScaleMonitoringEnabled != nil {
		siteConfig.RuntimeScaleMonitoringEnabled = pulumi.BoolPtr(sc.GetRuntimeScaleMonitoringEnabled())
	}

	if sc.Cors != nil {
		siteConfig.Cors = buildCors(sc.Cors)
	}

	if len(sc.IpRestrictions) > 0 {
		siteConfig.IpRestrictions = buildIpRestrictions(sc.IpRestrictions)
	}
	ipDefaultAction := "Allow"
	if sc.IpRestrictionDefaultAction != azurefunctionappv1alpha1.AzureFunctionAppIpRestrictionAction_azure_function_app_ip_restriction_action_unspecified {
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
	if sc.ScmIpRestrictionDefaultAction != azurefunctionappv1alpha1.AzureFunctionAppIpRestrictionAction_azure_function_app_ip_restriction_action_unspecified {
		scmDefaultAction = ipRestrictionActionStrings[sc.ScmIpRestrictionDefaultAction]
	}
	siteConfig.ScmIpRestrictionDefaultAction = pulumi.StringPtr(scmDefaultAction)

	if sc.AppServiceLogs != nil {
		logsArgs := &appservice.LinuxFunctionAppSiteConfigAppServiceLogsArgs{}
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

	if sc.ContainerRegistryManagedIdentityClientId != "" {
		siteConfig.ContainerRegistryManagedIdentityClientId = pulumi.StringPtr(sc.ContainerRegistryManagedIdentityClientId)
	}

	if sc.ApplicationInsightsKey != "" {
		siteConfig.ApplicationInsightsKey = pulumi.StringPtr(sc.ApplicationInsightsKey)
	}
	if spec.ApplicationInsightsConnectionString != nil {
		siteConfig.ApplicationInsightsConnectionString = pulumi.StringPtr(spec.ApplicationInsightsConnectionString.GetValue())
	}

	return siteConfig
}

// ---------------------------------------------------------------------------
// Application Stack
// ---------------------------------------------------------------------------

func buildApplicationStack(stack *azurefunctionappv1alpha1.AzureFunctionAppApplicationStack) *appservice.LinuxFunctionAppSiteConfigApplicationStackArgs {
	appStack := &appservice.LinuxFunctionAppSiteConfigApplicationStackArgs{}

	if stack.DotnetVersion != "" {
		appStack.DotnetVersion = pulumi.StringPtr(stack.DotnetVersion)
		useDotnetIsolated := false
		if stack.UseDotnetIsolatedRuntime != nil {
			useDotnetIsolated = stack.GetUseDotnetIsolatedRuntime()
		}
		appStack.UseDotnetIsolatedRuntime = pulumi.BoolPtr(useDotnetIsolated)
	}
	if stack.NodeVersion != "" {
		appStack.NodeVersion = pulumi.StringPtr(stack.NodeVersion)
	}
	if stack.PythonVersion != "" {
		appStack.PythonVersion = pulumi.StringPtr(stack.PythonVersion)
	}
	if stack.JavaVersion != "" {
		appStack.JavaVersion = pulumi.StringPtr(stack.JavaVersion)
	}
	if stack.PowershellCoreVersion != "" {
		appStack.PowershellCoreVersion = pulumi.StringPtr(stack.PowershellCoreVersion)
	}
	if stack.UseCustomRuntime != nil {
		appStack.UseCustomRuntime = pulumi.BoolPtr(stack.GetUseCustomRuntime())
	}

	if stack.Docker != nil {
		docker := stack.Docker
		dockerArgs := appservice.LinuxFunctionAppSiteConfigApplicationStackDockerArgs{
			RegistryUrl: pulumi.String(docker.RegistryUrl),
			ImageName:   pulumi.String(docker.ImageName),
			ImageTag:    pulumi.String(docker.ImageTag),
		}
		if docker.RegistryUsername != "" {
			dockerArgs.RegistryUsername = pulumi.StringPtr(docker.RegistryUsername)
		}
		if docker.RegistryPassword != nil {
			dockerArgs.RegistryPassword = pulumi.StringPtr(docker.RegistryPassword.GetValue())
		}
		appStack.Dockers = appservice.LinuxFunctionAppSiteConfigApplicationStackDockerArray{dockerArgs}
	}

	return appStack
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func buildIdentity(spec *azurefunctionappv1alpha1.AzureFunctionAppIdentity) *appservice.LinuxFunctionAppIdentityArgs {
	identity := &appservice.LinuxFunctionAppIdentityArgs{
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

func buildConnectionStrings(specs []*azurefunctionappv1alpha1.AzureFunctionAppConnectionString) appservice.LinuxFunctionAppConnectionStringArray {
	connStrings := make(appservice.LinuxFunctionAppConnectionStringArray, 0, len(specs))
	for _, cs := range specs {
		connStrings = append(connStrings, appservice.LinuxFunctionAppConnectionStringArgs{
			Name:  pulumi.String(cs.Name),
			Type:  pulumi.String(connectionStringTypeStrings[cs.Type]),
			Value: pulumi.String(cs.Value.GetValue()),
		})
	}
	return connStrings
}

// ---------------------------------------------------------------------------
// Storage Accounts (Mounts)
// ---------------------------------------------------------------------------

func buildStorageAccounts(specs []*azurefunctionappv1alpha1.AzureFunctionAppStorageMount) appservice.LinuxFunctionAppStorageAccountArray {
	accounts := make(appservice.LinuxFunctionAppStorageAccountArray, 0, len(specs))
	for _, sm := range specs {
		account := appservice.LinuxFunctionAppStorageAccountArgs{
			Name:        pulumi.String(sm.Name),
			Type:        pulumi.String(storageMountTypeStrings[sm.Type]),
			AccountName: pulumi.String(sm.AccountName),
			ShareName:   pulumi.String(sm.ShareName),
			AccessKey:   pulumi.String(sm.AccessKey.GetValue()),
		}
		if sm.MountPath != "" {
			account.MountPath = pulumi.StringPtr(sm.MountPath)
		}
		accounts = append(accounts, account)
	}
	return accounts
}

// ---------------------------------------------------------------------------
// CORS
// ---------------------------------------------------------------------------

func buildCors(spec *azurefunctionappv1alpha1.AzureFunctionAppCorsSettings) *appservice.LinuxFunctionAppSiteConfigCorsArgs {
	cors := &appservice.LinuxFunctionAppSiteConfigCorsArgs{
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

func buildIpRestrictions(specs []*azurefunctionappv1alpha1.AzureFunctionAppIpRestriction) appservice.LinuxFunctionAppSiteConfigIpRestrictionArray {
	restrictions := make(appservice.LinuxFunctionAppSiteConfigIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := appservice.LinuxFunctionAppSiteConfigIpRestrictionArgs{}

		if r.Name != "" {
			restriction.Name = pulumi.StringPtr(r.Name)
		}
		if r.Priority != nil {
			restriction.Priority = pulumi.IntPtr(int(r.GetPriority()))
		}
		action := "Allow"
		if r.Action != azurefunctionappv1alpha1.AzureFunctionAppIpRestrictionAction_azure_function_app_ip_restriction_action_unspecified {
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

func buildScmIpRestrictions(specs []*azurefunctionappv1alpha1.AzureFunctionAppIpRestriction) appservice.LinuxFunctionAppSiteConfigScmIpRestrictionArray {
	restrictions := make(appservice.LinuxFunctionAppSiteConfigScmIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := appservice.LinuxFunctionAppSiteConfigScmIpRestrictionArgs{}

		if r.Name != "" {
			restriction.Name = pulumi.StringPtr(r.Name)
		}
		if r.Priority != nil {
			restriction.Priority = pulumi.IntPtr(int(r.GetPriority()))
		}
		action := "Allow"
		if r.Action != azurefunctionappv1alpha1.AzureFunctionAppIpRestrictionAction_azure_function_app_ip_restriction_action_unspecified {
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

func buildIpRestrictionHeaders(h *azurefunctionappv1alpha1.AzureFunctionAppIpRestrictionHeaders) *appservice.LinuxFunctionAppSiteConfigIpRestrictionHeadersArgs {
	headers := &appservice.LinuxFunctionAppSiteConfigIpRestrictionHeadersArgs{}

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

func buildScmIpRestrictionHeaders(h *azurefunctionappv1alpha1.AzureFunctionAppIpRestrictionHeaders) *appservice.LinuxFunctionAppSiteConfigScmIpRestrictionHeadersArgs {
	headers := &appservice.LinuxFunctionAppSiteConfigScmIpRestrictionHeadersArgs{}

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
// Backup
// ---------------------------------------------------------------------------

// buildBackup maps the scheduled-backup block. Azure rejects backup on
// Consumption and Basic plans at apply time.
func buildBackup(spec *azurefunctionappv1alpha1.AzureFunctionAppBackup) *appservice.LinuxFunctionAppBackupArgs {
	schedule := &appservice.LinuxFunctionAppBackupScheduleArgs{
		FrequencyInterval: pulumi.Int(int(spec.Schedule.FrequencyInterval)),
		FrequencyUnit:     pulumi.String(backupFrequencyUnitStrings[spec.Schedule.FrequencyUnit]),
	}
	keepOne := false
	if spec.Schedule.KeepAtLeastOneBackup != nil {
		keepOne = spec.Schedule.GetKeepAtLeastOneBackup()
	}
	schedule.KeepAtLeastOneBackup = pulumi.BoolPtr(keepOne)

	retentionDays := 30
	if spec.Schedule.RetentionPeriodDays != nil {
		retentionDays = int(spec.Schedule.GetRetentionPeriodDays())
	}
	schedule.RetentionPeriodDays = pulumi.IntPtr(retentionDays)

	if spec.Schedule.StartTime != "" {
		schedule.StartTime = pulumi.StringPtr(spec.Schedule.StartTime)
	}

	backup := &appservice.LinuxFunctionAppBackupArgs{
		Name:              pulumi.String(spec.Name),
		StorageAccountUrl: pulumi.String(spec.StorageAccountUrl.GetValue()),
		Schedule:          schedule,
	}
	enabled := true
	if spec.Enabled != nil {
		enabled = spec.GetEnabled()
	}
	backup.Enabled = pulumi.BoolPtr(enabled)

	return backup
}

// ---------------------------------------------------------------------------
// Easy Auth v2
// ---------------------------------------------------------------------------

// buildAuthSettingsV2 maps App Service built-in authentication. Provider
// client secrets are referenced by APP SETTING NAME -- the secret values
// live in app_settings (or Key Vault references), never in this block.
func buildAuthSettingsV2(spec *azurefunctionappv1alpha1.AzureFunctionAppAuthSettingsV2) *appservice.LinuxFunctionAppAuthSettingsV2Args {
	args := &appservice.LinuxFunctionAppAuthSettingsV2Args{
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
	if spec.UnauthenticatedAction != azurefunctionappv1alpha1.AzureFunctionAppUnauthenticatedAction_azure_function_app_unauthenticated_action_unspecified {
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
	if spec.ForwardProxyConvention != azurefunctionappv1alpha1.AzureFunctionAppForwardProxyConvention_azure_function_app_forward_proxy_convention_unspecified {
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
		args.AppleV2 = &appservice.LinuxFunctionAppAuthSettingsV2AppleV2Args{
			ClientId:                pulumi.String(spec.AppleV2.ClientId),
			ClientSecretSettingName: pulumi.String(spec.AppleV2.ClientSecretSettingName),
		}
	}

	if spec.ActiveDirectoryV2 != nil {
		aad := spec.ActiveDirectoryV2
		aadArgs := &appservice.LinuxFunctionAppAuthSettingsV2ActiveDirectoryV2Args{
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
		args.AzureStaticWebAppV2 = &appservice.LinuxFunctionAppAuthSettingsV2AzureStaticWebAppV2Args{
			ClientId: pulumi.String(spec.AzureStaticWebAppV2.ClientId),
		}
	}

	if len(spec.CustomOidcV2) > 0 {
		oidcs := make(appservice.LinuxFunctionAppAuthSettingsV2CustomOidcV2Array, 0, len(spec.CustomOidcV2))
		for _, oidc := range spec.CustomOidcV2 {
			oidcArgs := appservice.LinuxFunctionAppAuthSettingsV2CustomOidcV2Args{
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
		fbArgs := &appservice.LinuxFunctionAppAuthSettingsV2FacebookV2Args{
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
		ghArgs := &appservice.LinuxFunctionAppAuthSettingsV2GithubV2Args{
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
		gArgs := &appservice.LinuxFunctionAppAuthSettingsV2GoogleV2Args{
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
		msArgs := &appservice.LinuxFunctionAppAuthSettingsV2MicrosoftV2Args{
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
		args.TwitterV2 = &appservice.LinuxFunctionAppAuthSettingsV2TwitterV2Args{
			ConsumerKey:               pulumi.String(spec.TwitterV2.ConsumerKey),
			ConsumerSecretSettingName: pulumi.String(spec.TwitterV2.ConsumerSecretSettingName),
		}
	}

	return args
}

func buildAuthV2Login(spec *azurefunctionappv1alpha1.AzureFunctionAppAuthV2Login) *appservice.LinuxFunctionAppAuthSettingsV2LoginArgs {
	login := &appservice.LinuxFunctionAppAuthSettingsV2LoginArgs{}

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
	if spec.CookieExpirationConvention != azurefunctionappv1alpha1.AzureFunctionAppCookieExpirationConvention_azure_function_app_cookie_expiration_convention_unspecified {
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
