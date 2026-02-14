package module

import (
	"github.com/pkg/errors"
	azurelinuxwebappv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurelinuxwebapp/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/appservice"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurelinuxwebappv1.AzureLinuxWebAppStackInput) error {
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

	spec := locals.AzureLinuxWebApp.Spec

	// Build site_config
	siteConfigArgs := buildSiteConfig(spec)

	// Build the Linux Web App arguments
	webAppArgs := &appservice.LinuxWebAppArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		ServicePlanId:     pulumi.String(spec.ServicePlanId.GetValue()),
		SiteConfig:        siteConfigArgs,
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Merge app settings with Application Insights connection string
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

	// Connection strings
	if len(spec.ConnectionStrings) > 0 {
		webAppArgs.ConnectionStrings = buildConnectionStrings(spec.ConnectionStrings)
	}

	// HTTPS only
	if spec.HttpsOnly != nil {
		webAppArgs.HttpsOnly = pulumi.BoolPtr(spec.GetHttpsOnly())
	}

	// Public network access
	if spec.PublicNetworkAccessEnabled != nil {
		webAppArgs.PublicNetworkAccessEnabled = pulumi.BoolPtr(spec.GetPublicNetworkAccessEnabled())
	}

	// Enabled
	if spec.Enabled != nil {
		webAppArgs.Enabled = pulumi.BoolPtr(spec.GetEnabled())
	}

	// Client affinity
	if spec.ClientAffinityEnabled != nil {
		webAppArgs.ClientAffinityEnabled = pulumi.BoolPtr(spec.GetClientAffinityEnabled())
	}

	// VNet integration
	if spec.VirtualNetworkSubnetId != nil {
		webAppArgs.VirtualNetworkSubnetId = pulumi.StringPtr(spec.VirtualNetworkSubnetId.GetValue())
	}

	// Identity
	if spec.Identity != nil {
		webAppArgs.Identity = buildIdentity(spec.Identity)
	}

	// Key Vault reference identity
	if spec.KeyVaultReferenceIdentityId != nil {
		webAppArgs.KeyVaultReferenceIdentityId = pulumi.StringPtr(spec.KeyVaultReferenceIdentityId.GetValue())
	}

	// Client certificate settings
	if spec.ClientCertificateEnabled != nil {
		webAppArgs.ClientCertificateEnabled = pulumi.BoolPtr(spec.GetClientCertificateEnabled())
	}
	if spec.ClientCertificateMode != nil {
		webAppArgs.ClientCertificateMode = pulumi.StringPtr(spec.GetClientCertificateMode())
	}
	if spec.ClientCertificateExclusionPaths != "" {
		webAppArgs.ClientCertificateExclusionPaths = pulumi.StringPtr(spec.ClientCertificateExclusionPaths)
	}

	// Storage mounts
	if len(spec.StorageMounts) > 0 {
		webAppArgs.StorageAccounts = buildStorageAccounts(spec.StorageMounts)
	}

	// Logs
	if spec.Logs != nil {
		webAppArgs.Logs = buildLogs(spec.Logs)
	}

	// Create the Linux Web App
	webApp, err := appservice.NewLinuxWebApp(ctx,
		spec.Name,
		webAppArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Linux Web App %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpWebAppId, webApp.ID())
	ctx.Export(OpDefaultHostname, webApp.DefaultHostname)
	ctx.Export(OpOutboundIpAddresses, webApp.OutboundIpAddresses)
	ctx.Export(OpCustomDomainVerificationId, webApp.CustomDomainVerificationId)
	ctx.Export(OpKind, webApp.Kind)

	// Export identity outputs conditionally
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

	return nil
}

// ---------------------------------------------------------------------------
// Site Config
// ---------------------------------------------------------------------------

func buildSiteConfig(spec *azurelinuxwebappv1.AzureLinuxWebAppSpec) *appservice.LinuxWebAppSiteConfigArgs {
	sc := spec.GetSiteConfig()
	siteConfig := &appservice.LinuxWebAppSiteConfigArgs{}

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

	// Health check eviction time
	if sc.HealthCheckEvictionTimeInMin != nil {
		siteConfig.HealthCheckEvictionTimeInMin = pulumi.IntPtr(int(sc.GetHealthCheckEvictionTimeInMin()))
	}

	// TLS versions
	if sc.MinimumTlsVersion != nil {
		siteConfig.MinimumTlsVersion = pulumi.StringPtr(sc.GetMinimumTlsVersion())
	}
	if sc.ScmMinimumTlsVersion != nil {
		siteConfig.ScmMinimumTlsVersion = pulumi.StringPtr(sc.GetScmMinimumTlsVersion())
	}

	// Worker count
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

	// Container registry managed identity
	if sc.ContainerRegistryUseManagedIdentity != nil {
		siteConfig.ContainerRegistryUseManagedIdentity = pulumi.BoolPtr(sc.GetContainerRegistryUseManagedIdentity())
	}
	if sc.ContainerRegistryManagedIdentityClientId != "" {
		siteConfig.ContainerRegistryManagedIdentityClientId = pulumi.StringPtr(sc.ContainerRegistryManagedIdentityClientId)
	}

	return siteConfig
}

// ---------------------------------------------------------------------------
// Application Stack
// ---------------------------------------------------------------------------

func buildApplicationStack(stack *azurelinuxwebappv1.AzureLinuxWebAppApplicationStack) *appservice.LinuxWebAppSiteConfigApplicationStackArgs {
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
	if stack.RubyVersion != "" {
		appStack.RubyVersion = pulumi.StringPtr(stack.RubyVersion)
	}
	if stack.GoVersion != "" {
		appStack.GoVersion = pulumi.StringPtr(stack.GoVersion)
	}
	if stack.JavaVersion != "" {
		appStack.JavaVersion = pulumi.StringPtr(stack.JavaVersion)
	}
	if stack.JavaServer != "" {
		appStack.JavaServer = pulumi.StringPtr(stack.JavaServer)
	}
	if stack.JavaServerVersion != "" {
		appStack.JavaServerVersion = pulumi.StringPtr(stack.JavaServerVersion)
	}

	// Docker container configuration
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

func buildIdentity(spec *azurelinuxwebappv1.AzureLinuxWebAppIdentity) *appservice.LinuxWebAppIdentityArgs {
	identity := &appservice.LinuxWebAppIdentityArgs{
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

func buildConnectionStrings(specs []*azurelinuxwebappv1.AzureLinuxWebAppConnectionString) appservice.LinuxWebAppConnectionStringArray {
	connStrings := make(appservice.LinuxWebAppConnectionStringArray, 0, len(specs))
	for _, cs := range specs {
		connStrings = append(connStrings, appservice.LinuxWebAppConnectionStringArgs{
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

func buildStorageAccounts(specs []*azurelinuxwebappv1.AzureLinuxWebAppStorageMount) appservice.LinuxWebAppStorageAccountArray {
	accounts := make(appservice.LinuxWebAppStorageAccountArray, 0, len(specs))
	for _, sm := range specs {
		account := appservice.LinuxWebAppStorageAccountArgs{
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

func buildCors(spec *azurelinuxwebappv1.AzureLinuxWebAppCorsSettings) *appservice.LinuxWebAppSiteConfigCorsArgs {
	cors := &appservice.LinuxWebAppSiteConfigCorsArgs{
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

func buildIpRestrictions(specs []*azurelinuxwebappv1.AzureLinuxWebAppIpRestriction) appservice.LinuxWebAppSiteConfigIpRestrictionArray {
	restrictions := make(appservice.LinuxWebAppSiteConfigIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := buildSingleIpRestriction(r)
		restrictions = append(restrictions, restriction)
	}
	return restrictions
}

func buildScmIpRestrictions(specs []*azurelinuxwebappv1.AzureLinuxWebAppIpRestriction) appservice.LinuxWebAppSiteConfigScmIpRestrictionArray {
	restrictions := make(appservice.LinuxWebAppSiteConfigScmIpRestrictionArray, 0, len(specs))
	for _, r := range specs {
		restriction := appservice.LinuxWebAppSiteConfigScmIpRestrictionArgs{}

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

func buildSingleIpRestriction(r *azurelinuxwebappv1.AzureLinuxWebAppIpRestriction) appservice.LinuxWebAppSiteConfigIpRestrictionArgs {
	restriction := appservice.LinuxWebAppSiteConfigIpRestrictionArgs{}

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

func buildIpRestrictionHeaders(h *azurelinuxwebappv1.AzureLinuxWebAppIpRestrictionHeaders) *appservice.LinuxWebAppSiteConfigIpRestrictionHeadersArgs {
	headers := &appservice.LinuxWebAppSiteConfigIpRestrictionHeadersArgs{}

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

func buildScmIpRestrictionHeaders(h *azurelinuxwebappv1.AzureLinuxWebAppIpRestrictionHeaders) *appservice.LinuxWebAppSiteConfigScmIpRestrictionHeadersArgs {
	headers := &appservice.LinuxWebAppSiteConfigScmIpRestrictionHeadersArgs{}

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
// Logs
// ---------------------------------------------------------------------------

func buildLogs(spec *azurelinuxwebappv1.AzureLinuxWebAppLogs) *appservice.LinuxWebAppLogsArgs {
	logs := &appservice.LinuxWebAppLogsArgs{}

	if spec.ApplicationLogs != nil {
		appLogs := appservice.LinuxWebAppLogsApplicationLogsArgs{}
		if spec.ApplicationLogs.FileSystemLevel != nil {
			appLogs.FileSystemLevel = pulumi.String(spec.ApplicationLogs.GetFileSystemLevel())
		}
		logs.ApplicationLogs = &appLogs
	}

	if spec.HttpLogs != nil {
		fsArgs := appservice.LinuxWebAppLogsHttpLogsFileSystemArgs{}
		if spec.HttpLogs.RetentionInMb != nil {
			fsArgs.RetentionInMb = pulumi.Int(int(spec.HttpLogs.GetRetentionInMb()))
		}
		if spec.HttpLogs.RetentionInDays != nil {
			fsArgs.RetentionInDays = pulumi.Int(int(spec.HttpLogs.GetRetentionInDays()))
		}
		httpLogs := appservice.LinuxWebAppLogsHttpLogsArgs{
			FileSystem: &fsArgs,
		}
		logs.HttpLogs = &httpLogs
	}

	if spec.FailedRequestTracing != nil {
		logs.FailedRequestTracing = pulumi.BoolPtr(spec.GetFailedRequestTracing())
	}
	if spec.DetailedErrorMessages != nil {
		logs.DetailedErrorMessages = pulumi.BoolPtr(spec.GetDetailedErrorMessages())
	}

	return logs
}
