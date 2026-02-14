package module

import (
	"github.com/pkg/errors"
	azurefunctionappv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurefunctionapp/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurefunctionappv1.AzureFunctionAppStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx, "azure", &azure.ProviderArgs{
		ClientId:       pulumi.String(azureProviderConfig.ClientId),
		ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
		SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
		TenantId:       pulumi.String(azureProviderConfig.TenantId),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFunctionApp.Spec

	// Build site_config
	siteConfigArgs := buildSiteConfig(spec)

	// Build the Linux Function App arguments
	functionAppArgs := &appservice.LinuxFunctionAppArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		ServicePlanId:     pulumi.String(spec.ServicePlanId.GetValue()),
		StorageAccountName: pulumi.String(spec.StorageAccountName.GetValue()),
		SiteConfig:        siteConfigArgs,
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Storage authentication: access key or managed identity
	if spec.StorageAccountAccessKey != nil {
		functionAppArgs.StorageAccountAccessKey = pulumi.StringPtr(spec.StorageAccountAccessKey.GetValue())
	}
	if spec.StorageUsesManagedIdentity != nil {
		functionAppArgs.StorageUsesManagedIdentity = pulumi.BoolPtr(spec.GetStorageUsesManagedIdentity())
	}

	// Functions extension version
	if spec.FunctionsExtensionVersion != nil {
		functionAppArgs.FunctionsExtensionVersion = pulumi.StringPtr(spec.GetFunctionsExtensionVersion())
	}

	// App settings
	if len(spec.AppSettings) > 0 {
		functionAppArgs.AppSettings = pulumi.ToStringMap(spec.AppSettings)
	}

	// Connection strings
	if len(spec.ConnectionStrings) > 0 {
		functionAppArgs.ConnectionStrings = buildConnectionStrings(spec.ConnectionStrings)
	}

	// HTTPS only
	if spec.HttpsOnly != nil {
		functionAppArgs.HttpsOnly = pulumi.BoolPtr(spec.GetHttpsOnly())
	}

	// Public network access
	if spec.PublicNetworkAccessEnabled != nil {
		functionAppArgs.PublicNetworkAccessEnabled = pulumi.BoolPtr(spec.GetPublicNetworkAccessEnabled())
	}

	// Built-in logging
	if spec.BuiltinLoggingEnabled != nil {
		functionAppArgs.BuiltinLoggingEnabled = pulumi.BoolPtr(spec.GetBuiltinLoggingEnabled())
	}

	// VNet integration
	if spec.VirtualNetworkSubnetId != nil {
		functionAppArgs.VirtualNetworkSubnetId = pulumi.StringPtr(spec.VirtualNetworkSubnetId.GetValue())
	}

	// Identity
	if spec.Identity != nil {
		functionAppArgs.Identity = buildIdentity(spec.Identity)
	}

	// Key Vault reference identity
	if spec.KeyVaultReferenceIdentityId != nil {
		functionAppArgs.KeyVaultReferenceIdentityId = pulumi.StringPtr(spec.KeyVaultReferenceIdentityId.GetValue())
	}

	// Client certificate settings
	if spec.ClientCertificateEnabled != nil {
		functionAppArgs.ClientCertificateEnabled = pulumi.BoolPtr(spec.GetClientCertificateEnabled())
	}
	if spec.ClientCertificateMode != nil {
		functionAppArgs.ClientCertificateMode = pulumi.StringPtr(spec.GetClientCertificateMode())
	}
	if spec.ClientCertificateExclusionPaths != "" {
		functionAppArgs.ClientCertificateExclusionPaths = pulumi.StringPtr(spec.ClientCertificateExclusionPaths)
	}

	// Content share force disabled
	if spec.ContentShareForceDisabled != nil {
		functionAppArgs.ContentShareForceDisabled = pulumi.BoolPtr(spec.GetContentShareForceDisabled())
	}

	// Storage mounts
	if len(spec.StorageMounts) > 0 {
		functionAppArgs.StorageAccounts = buildStorageAccounts(spec.StorageMounts)
	}

	// Create the Linux Function App
	functionApp, err := appservice.NewLinuxFunctionApp(ctx,
		spec.Name,
		functionAppArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Linux Function App %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpFunctionAppId, functionApp.ID())
	ctx.Export(OpDefaultHostname, functionApp.DefaultHostname)
	ctx.Export(OpOutboundIpAddresses, functionApp.OutboundIpAddresses)
	ctx.Export(OpCustomDomainVerificationId, functionApp.CustomDomainVerificationId)
	ctx.Export(OpKind, functionApp.Kind)

	// Export identity outputs conditionally
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

	return nil
}

// ---------------------------------------------------------------------------
// Site Config
// ---------------------------------------------------------------------------

func buildSiteConfig(spec *azurefunctionappv1.AzureFunctionAppSpec) *appservice.LinuxFunctionAppSiteConfigArgs {
	sc := spec.GetSiteConfig()
	siteConfig := &appservice.LinuxFunctionAppSiteConfigArgs{}

	// Application stack
	if sc.GetApplicationStack() != nil {
		siteConfig.ApplicationStack = buildApplicationStack(sc.ApplicationStack)
	}

	// Always on
	if sc.AlwaysOn != nil {
		siteConfig.AlwaysOn = pulumi.BoolPtr(sc.GetAlwaysOn())
	}

	// App command line
	if sc.AppCommandLine != "" {
		siteConfig.AppCommandLine = pulumi.StringPtr(sc.AppCommandLine)
	}

	// Health check path
	if sc.HealthCheckPath != "" {
		siteConfig.HealthCheckPath = pulumi.StringPtr(sc.HealthCheckPath)
	}

	// TLS versions
	if sc.MinimumTlsVersion != nil {
		siteConfig.MinimumTlsVersion = pulumi.StringPtr(sc.GetMinimumTlsVersion())
	}
	if sc.ScmMinimumTlsVersion != nil {
		siteConfig.ScmMinimumTlsVersion = pulumi.StringPtr(sc.GetScmMinimumTlsVersion())
	}

	// Scaling settings
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

	// Protocol and worker settings
	if sc.Http2Enabled != nil {
		siteConfig.Http2Enabled = pulumi.BoolPtr(sc.GetHttp2Enabled())
	}
	if sc.WebsocketsEnabled != nil {
		siteConfig.WebsocketsEnabled = pulumi.BoolPtr(sc.GetWebsocketsEnabled())
	}
	if sc.Use_32BitWorker != nil {
		siteConfig.Use32BitWorker = pulumi.BoolPtr(sc.GetUse_32BitWorker())
	}
	if sc.VnetRouteAllEnabled != nil {
		siteConfig.VnetRouteAllEnabled = pulumi.BoolPtr(sc.GetVnetRouteAllEnabled())
	}

	// FTPS and load balancing
	if sc.FtpsState != nil {
		siteConfig.FtpsState = pulumi.StringPtr(sc.GetFtpsState())
	}
	if sc.LoadBalancingMode != nil {
		siteConfig.LoadBalancingMode = pulumi.StringPtr(sc.GetLoadBalancingMode())
	}

	// Runtime scale monitoring
	if sc.RuntimeScaleMonitoringEnabled != nil {
		siteConfig.RuntimeScaleMonitoringEnabled = pulumi.BoolPtr(sc.GetRuntimeScaleMonitoringEnabled())
	}

	// CORS
	if sc.Cors != nil {
		siteConfig.Cors = buildCors(sc.Cors)
	}

	// IP restrictions
	if len(sc.IpRestrictions) > 0 {
		siteConfig.IpRestrictions = buildIpRestrictions(sc.IpRestrictions)
	}
	if sc.IpRestrictionDefaultAction != nil {
		siteConfig.IpRestrictionDefaultAction = pulumi.StringPtr(sc.GetIpRestrictionDefaultAction())
	}

	// SCM IP restrictions
	if sc.ScmUseMainIpRestriction != nil {
		siteConfig.ScmUseMainIpRestriction = pulumi.BoolPtr(sc.GetScmUseMainIpRestriction())
	}
	if len(sc.ScmIpRestrictions) > 0 {
		siteConfig.ScmIpRestrictions = buildScmIpRestrictions(sc.ScmIpRestrictions)
	}
	if sc.ScmIpRestrictionDefaultAction != nil {
		siteConfig.ScmIpRestrictionDefaultAction = pulumi.StringPtr(sc.GetScmIpRestrictionDefaultAction())
	}

	// App service logs
	if sc.AppServiceLogs != nil {
		siteConfig.AppServiceLogs = buildAppServiceLogs(sc.AppServiceLogs)
	}

	// Default documents
	if len(sc.DefaultDocuments) > 0 {
		siteConfig.DefaultDocuments = pulumi.ToStringArray(sc.DefaultDocuments)
	}

	// Container registry managed identity
	if sc.ContainerRegistryUseManagedIdentity != nil {
		siteConfig.ContainerRegistryUseManagedIdentity = pulumi.BoolPtr(sc.GetContainerRegistryUseManagedIdentity())
	}
	if sc.ContainerRegistryManagedIdentityClientId != "" {
		siteConfig.ContainerRegistryManagedIdentityClientId = pulumi.StringPtr(sc.ContainerRegistryManagedIdentityClientId)
	}

	// Application Insights key (classic, via site_config)
	if sc.ApplicationInsightsKey != "" {
		siteConfig.ApplicationInsightsKey = pulumi.StringPtr(sc.ApplicationInsightsKey)
	}

	// Application Insights connection string (modern, from parent spec)
	if spec.ApplicationInsightsConnectionString != nil {
		siteConfig.ApplicationInsightsConnectionString = pulumi.StringPtr(spec.ApplicationInsightsConnectionString.GetValue())
	}

	return siteConfig
}

// ---------------------------------------------------------------------------
// Application Stack
// ---------------------------------------------------------------------------

func buildApplicationStack(stack *azurefunctionappv1.AzureFunctionAppApplicationStack) *appservice.LinuxFunctionAppSiteConfigApplicationStackArgs {
	appStack := &appservice.LinuxFunctionAppSiteConfigApplicationStackArgs{}

	if stack.DotnetVersion != "" {
		appStack.DotnetVersion = pulumi.StringPtr(stack.DotnetVersion)
	}
	if stack.UseDotnetIsolatedRuntime != nil {
		appStack.UseDotnetIsolatedRuntime = pulumi.BoolPtr(stack.GetUseDotnetIsolatedRuntime())
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

	// Docker container configuration
	if stack.Docker != nil {
		docker := stack.Docker
		appStack.Dockers = appservice.LinuxFunctionAppSiteConfigApplicationStackDockerArray{
			appservice.LinuxFunctionAppSiteConfigApplicationStackDockerArgs{
				RegistryUrl: pulumi.String(docker.RegistryUrl),
				ImageName:   pulumi.String(docker.ImageName),
				ImageTag:    pulumi.String(docker.ImageTag),
			},
		}
		// Set optional Docker registry credentials
		if docker.RegistryUsername != "" {
			appStack.Dockers = appservice.LinuxFunctionAppSiteConfigApplicationStackDockerArray{
				buildDockerArgs(docker),
			}
		}
	}

	return appStack
}

func buildDockerArgs(docker *azurefunctionappv1.AzureFunctionAppDockerConfig) appservice.LinuxFunctionAppSiteConfigApplicationStackDockerArgs {
	args := appservice.LinuxFunctionAppSiteConfigApplicationStackDockerArgs{
		RegistryUrl: pulumi.String(docker.RegistryUrl),
		ImageName:   pulumi.String(docker.ImageName),
		ImageTag:    pulumi.String(docker.ImageTag),
	}
	if docker.RegistryUsername != "" {
		args.RegistryUsername = pulumi.StringPtr(docker.RegistryUsername)
	}
	if docker.RegistryPassword != nil {
		args.RegistryPassword = pulumi.StringPtr(docker.RegistryPassword.GetValue())
	}
	return args
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func buildIdentity(spec *azurefunctionappv1.AzureFunctionAppIdentity) *appservice.LinuxFunctionAppIdentityArgs {
	identity := &appservice.LinuxFunctionAppIdentityArgs{
		Type: pulumi.String(spec.Type),
	}

	// Resolve identity IDs from StringValueOrRef
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

func buildConnectionStrings(specs []*azurefunctionappv1.AzureFunctionAppConnectionString) appservice.LinuxFunctionAppConnectionStringArray {
	connStrings := make(appservice.LinuxFunctionAppConnectionStringArray, 0, len(specs))
	for _, cs := range specs {
		connStrings = append(connStrings, appservice.LinuxFunctionAppConnectionStringArgs{
			Name:  pulumi.String(cs.Name),
			Type:  pulumi.String(cs.Type),
			Value: pulumi.String(cs.Value.GetValue()),
		})
	}
	return connStrings
}

// ---------------------------------------------------------------------------
// Storage Accounts (Mounts)
// ---------------------------------------------------------------------------

func buildStorageAccounts(specs []*azurefunctionappv1.AzureFunctionAppStorageMount) appservice.LinuxFunctionAppStorageAccountArray {
	accounts := make(appservice.LinuxFunctionAppStorageAccountArray, 0, len(specs))
	for _, sm := range specs {
		account := appservice.LinuxFunctionAppStorageAccountArgs{
			Name:        pulumi.String(sm.Name),
			Type:        pulumi.String(sm.Type),
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

func buildCors(spec *azurefunctionappv1.AzureFunctionAppCorsSettings) *appservice.LinuxFunctionAppSiteConfigCorsArgs {
	cors := &appservice.LinuxFunctionAppSiteConfigCorsArgs{
		AllowedOrigins: pulumi.ToStringArray(spec.AllowedOrigins),
	}
	if spec.SupportCredentials != nil {
		cors.SupportCredentials = pulumi.BoolPtr(spec.GetSupportCredentials())
	}
	return cors
}

// ---------------------------------------------------------------------------
// IP Restrictions
// ---------------------------------------------------------------------------

func buildIpRestrictions(specs []*azurefunctionappv1.AzureFunctionAppIpRestriction) appservice.LinuxFunctionAppSiteConfigIpRestrictionArray {
	restrictions := make(appservice.LinuxFunctionAppSiteConfigIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := buildSingleIpRestriction(r)
		restrictions = append(restrictions, restriction)
	}
	return restrictions
}

func buildScmIpRestrictions(specs []*azurefunctionappv1.AzureFunctionAppIpRestriction) appservice.LinuxFunctionAppSiteConfigScmIpRestrictionArray {
	restrictions := make(appservice.LinuxFunctionAppSiteConfigScmIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := appservice.LinuxFunctionAppSiteConfigScmIpRestrictionArgs{}

		if r.Name != "" {
			restriction.Name = pulumi.StringPtr(r.Name)
		}
		if r.Priority != nil {
			restriction.Priority = pulumi.IntPtr(int(r.GetPriority()))
		}
		if r.Action != nil {
			restriction.Action = pulumi.StringPtr(r.GetAction())
		}
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

func buildSingleIpRestriction(r *azurefunctionappv1.AzureFunctionAppIpRestriction) appservice.LinuxFunctionAppSiteConfigIpRestrictionArgs {
	restriction := appservice.LinuxFunctionAppSiteConfigIpRestrictionArgs{}

	if r.Name != "" {
		restriction.Name = pulumi.StringPtr(r.Name)
	}
	if r.Priority != nil {
		restriction.Priority = pulumi.IntPtr(int(r.GetPriority()))
	}
	if r.Action != nil {
		restriction.Action = pulumi.StringPtr(r.GetAction())
	}
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

	return restriction
}

func buildIpRestrictionHeaders(h *azurefunctionappv1.AzureFunctionAppIpRestrictionHeaders) *appservice.LinuxFunctionAppSiteConfigIpRestrictionHeadersArgs {
	headers := &appservice.LinuxFunctionAppSiteConfigIpRestrictionHeadersArgs{}

	if len(h.XForwardedFor) > 0 {
		headers.XForwardedFors = pulumi.ToStringArray(h.XForwardedFor)
	}
	if len(h.XForwardedHost) > 0 {
		headers.XForwardedHosts = pulumi.ToStringArray(h.XForwardedHost)
	}
	if len(h.XAzureFdid) > 0 {
		headers.XAzureFdids = pulumi.ToStringArray(h.XAzureFdid)
	}
	if len(h.XFdHealthProbe) > 0 {
		headers.XFdHealthProbe = pulumi.StringPtr(h.XFdHealthProbe[0])
	}

	return headers
}

func buildScmIpRestrictionHeaders(h *azurefunctionappv1.AzureFunctionAppIpRestrictionHeaders) *appservice.LinuxFunctionAppSiteConfigScmIpRestrictionHeadersArgs {
	headers := &appservice.LinuxFunctionAppSiteConfigScmIpRestrictionHeadersArgs{}

	if len(h.XForwardedFor) > 0 {
		headers.XForwardedFors = pulumi.ToStringArray(h.XForwardedFor)
	}
	if len(h.XForwardedHost) > 0 {
		headers.XForwardedHosts = pulumi.ToStringArray(h.XForwardedHost)
	}
	if len(h.XAzureFdid) > 0 {
		headers.XAzureFdids = pulumi.ToStringArray(h.XAzureFdid)
	}
	if len(h.XFdHealthProbe) > 0 {
		headers.XFdHealthProbe = pulumi.StringPtr(h.XFdHealthProbe[0])
	}

	return headers
}

// ---------------------------------------------------------------------------
// App Service Logs
// ---------------------------------------------------------------------------

func buildAppServiceLogs(spec *azurefunctionappv1.AzureFunctionAppAppServiceLogs) *appservice.LinuxFunctionAppSiteConfigAppServiceLogsArgs {
	logs := &appservice.LinuxFunctionAppSiteConfigAppServiceLogsArgs{}

	if spec.DiskQuotaMb != nil {
		logs.DiskQuotaMb = pulumi.IntPtr(int(spec.GetDiskQuotaMb()))
	}
	if spec.RetentionPeriodDays != nil {
		logs.RetentionPeriodDays = pulumi.IntPtr(int(spec.GetRetentionPeriodDays()))
	}

	return logs
}
