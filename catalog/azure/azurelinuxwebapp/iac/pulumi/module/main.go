package module

import (
	"strings"

	"github.com/pkg/errors"
	azurelinuxwebappv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurelinuxwebapp/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurelinuxwebappv1alpha1.AzureLinuxWebAppStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureLinuxWebApp.Spec

	// The Linux Web App: an always-on HTTP workload on an App Service
	// Plan. Name, region, and resource group are ForceNew; the plan
	// reference is not (apps can move between plans in the same region +
	// resource group).
	webAppArgs := &appservice.LinuxWebAppArgs{
		Name:              pulumi.String(spec.WebAppName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		ServicePlanId:     pulumi.String(spec.ServicePlanId.GetValue()),
		SiteConfig:        buildSiteConfig(spec),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Merge Application Insights connection string into app settings --
	// the connection-string form is how the App Service agent discovers
	// the telemetry destination.
	appSettings := make(map[string]string)
	for k, v := range spec.AppSettings {
		appSettings[k] = v
	}
	if spec.ApplicationInsightsConnectionString != nil {
		appSettings["APPLICATIONINSIGHTS_CONNECTION_STRING"] = spec.ApplicationInsightsConnectionString.GetValue()
	}
	if len(appSettings) > 0 {
		webAppArgs.AppSettings = pulumi.ToStringMap(appSettings)
	}

	if len(spec.ConnectionStrings) > 0 {
		webAppArgs.ConnectionStrings = buildConnectionStrings(spec.ConnectionStrings)
	}

	// Settings pinned to the production slot during slot swaps.
	if spec.StickySettings != nil {
		stickyArgs := &appservice.LinuxWebAppStickySettingsArgs{}
		if len(spec.StickySettings.AppSettingNames) > 0 {
			stickyArgs.AppSettingNames = pulumi.ToStringArray(spec.StickySettings.AppSettingNames)
		}
		if len(spec.StickySettings.ConnectionStringNames) > 0 {
			stickyArgs.ConnectionStringNames = pulumi.ToStringArray(spec.StickySettings.ConnectionStringNames)
		}
		webAppArgs.StickySettings = stickyArgs
	}

	// Presence-guarded proto defaults: stack inputs never materialize
	// them, so an unset field must deploy the spec's documented default,
	// not the Go zero value.

	httpsOnly := true
	if spec.HttpsOnly != nil {
		httpsOnly = spec.GetHttpsOnly()
	}
	webAppArgs.HttpsOnly = pulumi.BoolPtr(httpsOnly)

	publicNetworkAccess := true
	if spec.PublicNetworkAccessEnabled != nil {
		publicNetworkAccess = spec.GetPublicNetworkAccessEnabled()
	}
	webAppArgs.PublicNetworkAccessEnabled = pulumi.BoolPtr(publicNetworkAccess)

	enabled := true
	if spec.Enabled != nil {
		enabled = spec.GetEnabled()
	}
	webAppArgs.Enabled = pulumi.BoolPtr(enabled)

	clientAffinity := false
	if spec.ClientAffinityEnabled != nil {
		clientAffinity = spec.GetClientAffinityEnabled()
	}
	webAppArgs.ClientAffinityEnabled = pulumi.BoolPtr(clientAffinity)

	// VNet integration. Image pulls and backup/restore traffic only ride
	// the VNet when their dedicated toggles say so (spec-gated to
	// require the subnet).
	if spec.VirtualNetworkSubnetId != nil {
		webAppArgs.VirtualNetworkSubnetId = pulumi.StringPtr(spec.VirtualNetworkSubnetId.GetValue())
	}
	vnetImagePull := false
	if spec.VnetImagePullEnabled != nil {
		vnetImagePull = spec.GetVnetImagePullEnabled()
	}
	webAppArgs.VnetImagePullEnabled = pulumi.BoolPtr(vnetImagePull)

	vnetBackupRestore := false
	if spec.VirtualNetworkBackupRestoreEnabled != nil {
		vnetBackupRestore = spec.GetVirtualNetworkBackupRestoreEnabled()
	}
	webAppArgs.VirtualNetworkBackupRestoreEnabled = pulumi.BoolPtr(vnetBackupRestore)

	if spec.Identity != nil {
		webAppArgs.Identity = buildIdentity(spec.Identity)
	}

	if spec.KeyVaultReferenceIdentityId != nil {
		webAppArgs.KeyVaultReferenceIdentityId = pulumi.StringPtr(spec.KeyVaultReferenceIdentityId.GetValue())
	}

	// Client certificate (mutual TLS) posture.
	clientCertEnabled := false
	if spec.ClientCertificateEnabled != nil {
		clientCertEnabled = spec.GetClientCertificateEnabled()
	}
	webAppArgs.ClientCertificateEnabled = pulumi.BoolPtr(clientCertEnabled)

	clientCertMode := "Optional"
	if spec.ClientCertificateMode != azurelinuxwebappv1alpha1.AzureLinuxWebAppClientCertificateMode_azure_linux_web_app_client_certificate_mode_unspecified {
		clientCertMode = clientCertificateModeStrings[spec.ClientCertificateMode]
	}
	webAppArgs.ClientCertificateMode = pulumi.StringPtr(clientCertMode)

	if spec.ClientCertificateExclusionPaths != "" {
		webAppArgs.ClientCertificateExclusionPaths = pulumi.StringPtr(spec.ClientCertificateExclusionPaths)
	}

	// Basic-auth publishing toggles: disabling both closes the classic
	// credential-based deployment paths (Web Deploy + FTP) and forces
	// identity-based deployment.
	webdeployBasicAuth := true
	if spec.WebdeployPublishBasicAuthenticationEnabled != nil {
		webdeployBasicAuth = spec.GetWebdeployPublishBasicAuthenticationEnabled()
	}
	webAppArgs.WebdeployPublishBasicAuthenticationEnabled = pulumi.BoolPtr(webdeployBasicAuth)

	ftpBasicAuth := true
	if spec.FtpPublishBasicAuthenticationEnabled != nil {
		ftpBasicAuth = spec.GetFtpPublishBasicAuthenticationEnabled()
	}
	webAppArgs.FtpPublishBasicAuthenticationEnabled = pulumi.BoolPtr(ftpBasicAuth)

	if spec.ZipDeployFile != "" {
		webAppArgs.ZipDeployFile = pulumi.StringPtr(spec.ZipDeployFile)
	}

	if len(spec.StorageMounts) > 0 {
		webAppArgs.StorageAccounts = buildStorageAccounts(spec.StorageMounts)
	}

	if spec.Logs != nil {
		webAppArgs.Logs = buildLogs(spec.Logs)
	}

	if spec.Backup != nil {
		webAppArgs.Backup = buildBackup(spec.Backup)
	}

	if spec.AuthSettingsV2 != nil {
		webAppArgs.AuthSettingsV2 = buildAuthSettingsV2(spec.AuthSettingsV2)
	}

	// Create the Linux Web App
	webApp, err := appservice.NewLinuxWebApp(ctx,
		spec.WebAppName,
		webAppArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Linux Web App %s", spec.WebAppName)
	}

	// Export stack outputs. Azure reports the outbound IP sets as
	// comma-joined strings; both engines export them as real lists so
	// they flatten onto the repeated proto outputs identically.
	ctx.Export(OpWebAppId, webApp.ID())
	ctx.Export(OpDefaultHostname, webApp.DefaultHostname)
	ctx.Export(OpOutboundIpAddresses, webApp.OutboundIpAddresses.ApplyT(func(joined string) []string {
		if joined == "" {
			return []string{}
		}
		return strings.Split(joined, ",")
	}).(pulumi.StringArrayOutput))
	ctx.Export(OpPossibleOutboundIpAddresses, webApp.PossibleOutboundIpAddresses.ApplyT(func(joined string) []string {
		if joined == "" {
			return []string{}
		}
		return strings.Split(joined, ",")
	}).(pulumi.StringArrayOutput))
	ctx.Export(OpCustomDomainVerificationId, pulumi.Unsecret(webApp.CustomDomainVerificationId))
	ctx.Export(OpKind, webApp.Kind)
	ctx.Export(OpHostingEnvironmentId, webApp.HostingEnvironmentId)

	// The identity outputs populate only when a system-assigned identity
	// exists.
	ctx.Export(OpIdentityPrincipalId, webApp.Identity.ApplyT(func(identity *appservice.LinuxWebAppIdentity) string {
		if identity != nil && identity.PrincipalId != nil {
			return *identity.PrincipalId
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpIdentityTenantId, webApp.Identity.ApplyT(func(identity *appservice.LinuxWebAppIdentity) string {
		if identity != nil && identity.TenantId != nil {
			return *identity.TenantId
		}
		return ""
	}).(pulumi.StringOutput))

	// The site-level publishing credential -- grants deploy access while
	// basic-auth publishing is enabled. The provider marks the whole
	// credential block secret; the name is a username, so it is unwrapped,
	// and the password is exported as a secret.
	ctx.Export(OpSiteCredentialName, pulumi.Unsecret(webApp.SiteCredentials.ApplyT(func(creds []appservice.LinuxWebAppSiteCredential) string {
		if len(creds) > 0 && creds[0].Name != nil {
			return *creds[0].Name
		}
		return ""
	}).(pulumi.StringOutput)))

	ctx.Export(OpSiteCredentialPassword, pulumi.ToSecret(webApp.SiteCredentials.ApplyT(func(creds []appservice.LinuxWebAppSiteCredential) string {
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

func buildSiteConfig(spec *azurelinuxwebappv1alpha1.AzureLinuxWebAppSpec) *appservice.LinuxWebAppSiteConfigArgs {
	sc := spec.GetSiteConfig()
	siteConfig := &appservice.LinuxWebAppSiteConfigArgs{}

	if sc.GetApplicationStack() != nil {
		siteConfig.ApplicationStack = buildApplicationStack(sc.ApplicationStack)
	}

	// Azure rejects always_on on Free/Shared plans at apply time (the
	// plan's SKU is not visible here).
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
	if sc.MinimumTlsVersion != azurelinuxwebappv1alpha1.AzureLinuxWebAppTlsVersion_azure_linux_web_app_tls_version_unspecified {
		minTls = tlsVersionStrings[sc.MinimumTlsVersion]
	}
	siteConfig.MinimumTlsVersion = pulumi.StringPtr(minTls)

	scmMinTls := "1.2"
	if sc.ScmMinimumTlsVersion != azurelinuxwebappv1alpha1.AzureLinuxWebAppTlsVersion_azure_linux_web_app_tls_version_unspecified {
		scmMinTls = tlsVersionStrings[sc.ScmMinimumTlsVersion]
	}
	siteConfig.ScmMinimumTlsVersion = pulumi.StringPtr(scmMinTls)

	// Absent accepts Azure's platform default cipher set.
	if sc.MinimumTlsCipherSuite != "" {
		siteConfig.MinimumTlsCipherSuite = pulumi.StringPtr(sc.MinimumTlsCipherSuite)
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

	// The spec's default (false = 64-bit) deliberately overrides the
	// provider's own true default -- 64-bit workers for production.
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
	if sc.FtpsState != azurelinuxwebappv1alpha1.AzureLinuxWebAppFtpsState_azure_linux_web_app_ftps_state_unspecified {
		ftpsState = ftpsStateStrings[sc.FtpsState]
	}
	siteConfig.FtpsState = pulumi.StringPtr(ftpsState)

	loadBalancing := "LeastRequests"
	if sc.LoadBalancingMode != azurelinuxwebappv1alpha1.AzureLinuxWebAppLoadBalancingMode_azure_linux_web_app_load_balancing_mode_unspecified {
		loadBalancing = loadBalancingModeStrings[sc.LoadBalancingMode]
	}
	siteConfig.LoadBalancingMode = pulumi.StringPtr(loadBalancing)

	pipelineMode := "Integrated"
	if sc.ManagedPipelineMode != azurelinuxwebappv1alpha1.AzureLinuxWebAppManagedPipelineMode_azure_linux_web_app_managed_pipeline_mode_unspecified {
		pipelineMode = managedPipelineModeStrings[sc.ManagedPipelineMode]
	}
	siteConfig.ManagedPipelineMode = pulumi.StringPtr(pipelineMode)

	localMysql := false
	if sc.LocalMysqlEnabled != nil {
		localMysql = sc.GetLocalMysqlEnabled()
	}
	siteConfig.LocalMysqlEnabled = pulumi.BoolPtr(localMysql)

	// Remote debugging: Azure supports the current Visual Studio
	// generation only, so the debugger version is left to the platform.
	remoteDebugging := false
	if sc.RemoteDebuggingEnabled != nil {
		remoteDebugging = sc.GetRemoteDebuggingEnabled()
	}
	siteConfig.RemoteDebuggingEnabled = pulumi.BoolPtr(remoteDebugging)

	if sc.Cors != nil {
		siteConfig.Cors = buildCors(sc.Cors)
	}

	if len(sc.IpRestrictions) > 0 {
		siteConfig.IpRestrictions = buildIpRestrictions(sc.IpRestrictions)
	}
	ipDefaultAction := "Allow"
	if sc.IpRestrictionDefaultAction != azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestrictionAction_azure_linux_web_app_ip_restriction_action_unspecified {
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
	if sc.ScmIpRestrictionDefaultAction != azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestrictionAction_azure_linux_web_app_ip_restriction_action_unspecified {
		scmDefaultAction = ipRestrictionActionStrings[sc.ScmIpRestrictionDefaultAction]
	}
	siteConfig.ScmIpRestrictionDefaultAction = pulumi.StringPtr(scmDefaultAction)

	useManagedIdentityAcr := false
	if sc.ContainerRegistryUseManagedIdentity != nil {
		useManagedIdentityAcr = sc.GetContainerRegistryUseManagedIdentity()
	}
	siteConfig.ContainerRegistryUseManagedIdentity = pulumi.BoolPtr(useManagedIdentityAcr)

	if sc.ContainerRegistryManagedIdentityClientId != "" {
		siteConfig.ContainerRegistryManagedIdentityClientId = pulumi.StringPtr(sc.ContainerRegistryManagedIdentityClientId)
	}

	if sc.AutoHealSetting != nil {
		siteConfig.AutoHealSetting = buildAutoHealSetting(sc.AutoHealSetting)
	}

	return siteConfig
}

// ---------------------------------------------------------------------------
// Auto-heal
// ---------------------------------------------------------------------------

// buildAutoHealSetting maps the spec's trigger + uptime guard onto the
// provider's block. Linux supports exactly one heal action -- Recycle --
// so the module sends it implicitly.
func buildAutoHealSetting(spec *azurelinuxwebappv1alpha1.AzureLinuxWebAppAutoHealSetting) *appservice.LinuxWebAppSiteConfigAutoHealSettingArgs {
	trigger := &appservice.LinuxWebAppSiteConfigAutoHealSettingTriggerArgs{}

	if spec.Trigger.Requests != nil {
		trigger.Requests = &appservice.LinuxWebAppSiteConfigAutoHealSettingTriggerRequestsArgs{
			Count:    pulumi.Int(int(spec.Trigger.Requests.Count)),
			Interval: pulumi.String(spec.Trigger.Requests.Interval),
		}
	}

	if len(spec.Trigger.StatusCodes) > 0 {
		statusCodes := make(appservice.LinuxWebAppSiteConfigAutoHealSettingTriggerStatusCodeArray, 0, len(spec.Trigger.StatusCodes))
		for _, s := range spec.Trigger.StatusCodes {
			statusCode := appservice.LinuxWebAppSiteConfigAutoHealSettingTriggerStatusCodeArgs{
				StatusCodeRange: pulumi.String(s.StatusCodeRange),
				Count:           pulumi.Int(int(s.Count)),
				Interval:        pulumi.String(s.Interval),
			}
			if s.SubStatus != nil {
				statusCode.SubStatus = pulumi.IntPtr(int(s.GetSubStatus()))
			}
			if s.Win32StatusCode != nil {
				statusCode.Win32StatusCode = pulumi.IntPtr(int(s.GetWin32StatusCode()))
			}
			if s.Path != "" {
				statusCode.Path = pulumi.StringPtr(s.Path)
			}
			statusCodes = append(statusCodes, statusCode)
		}
		trigger.StatusCodes = statusCodes
	}

	if spec.Trigger.SlowRequest != nil {
		trigger.SlowRequest = &appservice.LinuxWebAppSiteConfigAutoHealSettingTriggerSlowRequestArgs{
			TimeTaken: pulumi.String(spec.Trigger.SlowRequest.TimeTaken),
			Interval:  pulumi.String(spec.Trigger.SlowRequest.Interval),
			Count:     pulumi.Int(int(spec.Trigger.SlowRequest.Count)),
		}
	}

	if len(spec.Trigger.SlowRequestWithPath) > 0 {
		slowWithPaths := make(appservice.LinuxWebAppSiteConfigAutoHealSettingTriggerSlowRequestWithPathArray, 0, len(spec.Trigger.SlowRequestWithPath))
		for _, s := range spec.Trigger.SlowRequestWithPath {
			slowWithPath := appservice.LinuxWebAppSiteConfigAutoHealSettingTriggerSlowRequestWithPathArgs{
				TimeTaken: pulumi.String(s.TimeTaken),
				Interval:  pulumi.String(s.Interval),
				Count:     pulumi.Int(int(s.Count)),
			}
			if s.Path != "" {
				slowWithPath.Path = pulumi.StringPtr(s.Path)
			}
			slowWithPaths = append(slowWithPaths, slowWithPath)
		}
		trigger.SlowRequestWithPaths = slowWithPaths
	}

	action := &appservice.LinuxWebAppSiteConfigAutoHealSettingActionArgs{
		ActionType: pulumi.String("Recycle"),
	}
	if spec.MinimumProcessExecutionTime != "" {
		action.MinimumProcessExecutionTime = pulumi.StringPtr(spec.MinimumProcessExecutionTime)
	}

	return &appservice.LinuxWebAppSiteConfigAutoHealSettingArgs{
		Trigger: trigger,
		Action:  action,
	}
}

// ---------------------------------------------------------------------------
// Application Stack
// ---------------------------------------------------------------------------

// buildApplicationStack maps the spec's nested docker message onto the
// provider's flattened docker_* attributes; image and tag combine into
// the single docker_image_name argument.
func buildApplicationStack(stack *azurelinuxwebappv1alpha1.AzureLinuxWebAppApplicationStack) *appservice.LinuxWebAppSiteConfigApplicationStackArgs {
	appStack := &appservice.LinuxWebAppSiteConfigApplicationStackArgs{}

	if stack.DotnetVersion != "" {
		appStack.DotnetVersion = pulumi.StringPtr(stack.DotnetVersion)
	}
	if stack.NodeVersion != "" {
		appStack.NodeVersion = pulumi.StringPtr(stack.NodeVersion)
	}
	if stack.PythonVersion != "" {
		appStack.PythonVersion = pulumi.StringPtr(stack.PythonVersion)
	}
	if stack.PhpVersion != "" {
		appStack.PhpVersion = pulumi.StringPtr(stack.PhpVersion)
	}
	if stack.GoVersion != "" {
		appStack.GoVersion = pulumi.StringPtr(stack.GoVersion)
	}
	if stack.JavaVersion != "" {
		appStack.JavaVersion = pulumi.StringPtr(stack.JavaVersion)
	}
	if stack.JavaServer != azurelinuxwebappv1alpha1.AzureLinuxWebAppJavaServer_azure_linux_web_app_java_server_unspecified {
		appStack.JavaServer = pulumi.StringPtr(javaServerStrings[stack.JavaServer])
	}
	if stack.JavaServerVersion != "" {
		appStack.JavaServerVersion = pulumi.StringPtr(stack.JavaServerVersion)
	}

	if stack.Docker != nil {
		docker := stack.Docker
		appStack.DockerImageName = pulumi.StringPtr(docker.ImageName + ":" + docker.ImageTag)
		appStack.DockerRegistryUrl = pulumi.StringPtr(docker.RegistryUrl)
		if docker.RegistryUsername != "" {
			appStack.DockerRegistryUsername = pulumi.StringPtr(docker.RegistryUsername)
		}
		if docker.RegistryPassword != nil {
			appStack.DockerRegistryPassword = pulumi.StringPtr(docker.RegistryPassword.GetValue())
		}
	}

	return appStack
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func buildIdentity(spec *azurelinuxwebappv1alpha1.AzureLinuxWebAppIdentity) *appservice.LinuxWebAppIdentityArgs {
	identity := &appservice.LinuxWebAppIdentityArgs{
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

func buildConnectionStrings(specs []*azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionString) appservice.LinuxWebAppConnectionStringArray {
	connStrings := make(appservice.LinuxWebAppConnectionStringArray, 0, len(specs))
	for _, cs := range specs {
		connStrings = append(connStrings, appservice.LinuxWebAppConnectionStringArgs{
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

func buildStorageAccounts(specs []*azurelinuxwebappv1alpha1.AzureLinuxWebAppStorageMount) appservice.LinuxWebAppStorageAccountArray {
	accounts := make(appservice.LinuxWebAppStorageAccountArray, 0, len(specs))
	for _, sm := range specs {
		account := appservice.LinuxWebAppStorageAccountArgs{
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

func buildCors(spec *azurelinuxwebappv1alpha1.AzureLinuxWebAppCorsSettings) *appservice.LinuxWebAppSiteConfigCorsArgs {
	cors := &appservice.LinuxWebAppSiteConfigCorsArgs{
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

func buildIpRestrictions(specs []*azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestriction) appservice.LinuxWebAppSiteConfigIpRestrictionArray {
	restrictions := make(appservice.LinuxWebAppSiteConfigIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := appservice.LinuxWebAppSiteConfigIpRestrictionArgs{}

		if r.Name != "" {
			restriction.Name = pulumi.StringPtr(r.Name)
		}
		if r.Priority != nil {
			restriction.Priority = pulumi.IntPtr(int(r.GetPriority()))
		}
		action := "Allow"
		if r.Action != azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestrictionAction_azure_linux_web_app_ip_restriction_action_unspecified {
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

func buildScmIpRestrictions(specs []*azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestriction) appservice.LinuxWebAppSiteConfigScmIpRestrictionArray {
	restrictions := make(appservice.LinuxWebAppSiteConfigScmIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := appservice.LinuxWebAppSiteConfigScmIpRestrictionArgs{}

		if r.Name != "" {
			restriction.Name = pulumi.StringPtr(r.Name)
		}
		if r.Priority != nil {
			restriction.Priority = pulumi.IntPtr(int(r.GetPriority()))
		}
		action := "Allow"
		if r.Action != azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestrictionAction_azure_linux_web_app_ip_restriction_action_unspecified {
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

func buildIpRestrictionHeaders(h *azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestrictionHeaders) *appservice.LinuxWebAppSiteConfigIpRestrictionHeadersArgs {
	headers := &appservice.LinuxWebAppSiteConfigIpRestrictionHeadersArgs{}

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

func buildScmIpRestrictionHeaders(h *azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestrictionHeaders) *appservice.LinuxWebAppSiteConfigScmIpRestrictionHeadersArgs {
	headers := &appservice.LinuxWebAppSiteConfigScmIpRestrictionHeadersArgs{}

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
// Logs
// ---------------------------------------------------------------------------

func buildLogs(spec *azurelinuxwebappv1alpha1.AzureLinuxWebAppLogs) *appservice.LinuxWebAppLogsArgs {
	logs := &appservice.LinuxWebAppLogsArgs{}

	if spec.ApplicationLogs != nil {
		appLogs := appservice.LinuxWebAppLogsApplicationLogsArgs{}
		fileSystemLevel := "Error"
		if spec.ApplicationLogs.FileSystemLevel != azurelinuxwebappv1alpha1.AzureLinuxWebAppLogLevel_azure_linux_web_app_log_level_unspecified {
			fileSystemLevel = logLevelStrings[spec.ApplicationLogs.FileSystemLevel]
		}
		appLogs.FileSystemLevel = pulumi.String(fileSystemLevel)

		if spec.ApplicationLogs.AzureBlobStorage != nil {
			blob := spec.ApplicationLogs.AzureBlobStorage
			blobArgs := &appservice.LinuxWebAppLogsApplicationLogsAzureBlobStorageArgs{
				Level:  pulumi.String(logLevelStrings[blob.Level]),
				SasUrl: pulumi.String(blob.SasUrl.GetValue()),
			}
			retentionDays := 0
			if blob.RetentionInDays != nil {
				retentionDays = int(blob.GetRetentionInDays())
			}
			blobArgs.RetentionInDays = pulumi.Int(retentionDays)
			appLogs.AzureBlobStorage = blobArgs
		}

		logs.ApplicationLogs = &appLogs
	}

	// Exactly one HTTP-log destination applies (spec-enforced XOR).
	if spec.HttpLogs != nil {
		httpLogs := appservice.LinuxWebAppLogsHttpLogsArgs{}

		if spec.HttpLogs.FileSystem != nil {
			fs := spec.HttpLogs.FileSystem
			fsArgs := &appservice.LinuxWebAppLogsHttpLogsFileSystemArgs{}
			retentionMb := 35
			if fs.RetentionInMb != nil {
				retentionMb = int(fs.GetRetentionInMb())
			}
			fsArgs.RetentionInMb = pulumi.Int(retentionMb)
			retentionDays := 0
			if fs.RetentionInDays != nil {
				retentionDays = int(fs.GetRetentionInDays())
			}
			fsArgs.RetentionInDays = pulumi.Int(retentionDays)
			httpLogs.FileSystem = fsArgs
		}

		if spec.HttpLogs.AzureBlobStorage != nil {
			blob := spec.HttpLogs.AzureBlobStorage
			blobArgs := &appservice.LinuxWebAppLogsHttpLogsAzureBlobStorageArgs{
				SasUrl: pulumi.String(blob.SasUrl.GetValue()),
			}
			if blob.RetentionInDays != nil {
				blobArgs.RetentionInDays = pulumi.IntPtr(int(blob.GetRetentionInDays()))
			}
			httpLogs.AzureBlobStorage = blobArgs
		}

		logs.HttpLogs = &httpLogs
	}

	failedRequestTracing := false
	if spec.FailedRequestTracing != nil {
		failedRequestTracing = spec.GetFailedRequestTracing()
	}
	logs.FailedRequestTracing = pulumi.BoolPtr(failedRequestTracing)

	detailedErrors := false
	if spec.DetailedErrorMessages != nil {
		detailedErrors = spec.GetDetailedErrorMessages()
	}
	logs.DetailedErrorMessages = pulumi.BoolPtr(detailedErrors)

	return logs
}

// ---------------------------------------------------------------------------
// Backup
// ---------------------------------------------------------------------------

// buildBackup maps the scheduled-backup block. Azure rejects backup on
// plans below Standard tier at apply time.
func buildBackup(spec *azurelinuxwebappv1alpha1.AzureLinuxWebAppBackup) *appservice.LinuxWebAppBackupArgs {
	schedule := &appservice.LinuxWebAppBackupScheduleArgs{
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

	backup := &appservice.LinuxWebAppBackupArgs{
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
func buildAuthSettingsV2(spec *azurelinuxwebappv1alpha1.AzureLinuxWebAppAuthSettingsV2) *appservice.LinuxWebAppAuthSettingsV2Args {
	args := &appservice.LinuxWebAppAuthSettingsV2Args{
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
	if spec.UnauthenticatedAction != azurelinuxwebappv1alpha1.AzureLinuxWebAppUnauthenticatedAction_azure_linux_web_app_unauthenticated_action_unspecified {
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
	if spec.ForwardProxyConvention != azurelinuxwebappv1alpha1.AzureLinuxWebAppForwardProxyConvention_azure_linux_web_app_forward_proxy_convention_unspecified {
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
		args.AppleV2 = &appservice.LinuxWebAppAuthSettingsV2AppleV2Args{
			ClientId:                pulumi.String(spec.AppleV2.ClientId),
			ClientSecretSettingName: pulumi.String(spec.AppleV2.ClientSecretSettingName),
		}
	}

	if spec.ActiveDirectoryV2 != nil {
		aad := spec.ActiveDirectoryV2
		aadArgs := &appservice.LinuxWebAppAuthSettingsV2ActiveDirectoryV2Args{
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
		args.AzureStaticWebAppV2 = &appservice.LinuxWebAppAuthSettingsV2AzureStaticWebAppV2Args{
			ClientId: pulumi.String(spec.AzureStaticWebAppV2.ClientId),
		}
	}

	if len(spec.CustomOidcV2) > 0 {
		oidcs := make(appservice.LinuxWebAppAuthSettingsV2CustomOidcV2Array, 0, len(spec.CustomOidcV2))
		for _, oidc := range spec.CustomOidcV2 {
			oidcArgs := appservice.LinuxWebAppAuthSettingsV2CustomOidcV2Args{
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
		fbArgs := &appservice.LinuxWebAppAuthSettingsV2FacebookV2Args{
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
		ghArgs := &appservice.LinuxWebAppAuthSettingsV2GithubV2Args{
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
		gArgs := &appservice.LinuxWebAppAuthSettingsV2GoogleV2Args{
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
		msArgs := &appservice.LinuxWebAppAuthSettingsV2MicrosoftV2Args{
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
		args.TwitterV2 = &appservice.LinuxWebAppAuthSettingsV2TwitterV2Args{
			ConsumerKey:               pulumi.String(spec.TwitterV2.ConsumerKey),
			ConsumerSecretSettingName: pulumi.String(spec.TwitterV2.ConsumerSecretSettingName),
		}
	}

	return args
}

func buildAuthV2Login(spec *azurelinuxwebappv1alpha1.AzureLinuxWebAppAuthV2Login) *appservice.LinuxWebAppAuthSettingsV2LoginArgs {
	login := &appservice.LinuxWebAppAuthSettingsV2LoginArgs{}

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
	if spec.CookieExpirationConvention != azurelinuxwebappv1alpha1.AzureLinuxWebAppCookieExpirationConvention_azure_linux_web_app_cookie_expiration_convention_unspecified {
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
