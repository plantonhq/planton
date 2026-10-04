package module

import (
	"fmt"

	"github.com/pkg/errors"
	azuremssqlserverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremssqlserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mssql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremssqlserverv1alpha1.AzureMssqlServerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMssqlServer.Spec

	// The deploying credential's context -- the tenant fallback for the
	// Microsoft Entra administrator grant when the spec does not pin one.
	clientConfig, err := core.GetClientConfig(ctx, pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrap(err, "failed to read the azure client configuration")
	}

	serverArgs := &mssql.ServerArgs{
		Name:              pulumi.String(spec.ServerName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Version is presence-guarded to the spec default ("12.0") -- IaC
	// inputs built from a manifest do NOT materialize proto defaults, and
	// azurerm requires the version.
	if spec.Version != nil {
		serverArgs.Version = pulumi.String(spec.GetVersion())
	} else {
		serverArgs.Version = pulumi.String("12.0")
	}

	// SQL-auth credentials -- omitted on an Entra-only server. The login
	// is fixed once set; ARM rejects a password change while
	// azuread_authentication_only is true.
	if spec.AdministratorLogin != "" {
		serverArgs.AdministratorLogin = pulumi.String(spec.AdministratorLogin)
	}
	if spec.AdministratorPassword.GetValue() != "" {
		serverArgs.AdministratorLoginPassword = pulumi.String(spec.AdministratorPassword.GetValue())
	}

	// The Microsoft Entra administrator. With azuread_authentication_only
	// SQL logins are disabled server-wide. The tenant falls back to the
	// deploying credential's.
	if spec.AzureadAdministrator != nil {
		tenantId := clientConfig.TenantId
		if spec.AzureadAdministrator.TenantId != nil && spec.AzureadAdministrator.GetTenantId() != "" {
			tenantId = spec.AzureadAdministrator.GetTenantId()
		}
		serverArgs.AzureadAdministrator = &mssql.ServerAzureadAdministratorArgs{
			LoginUsername:             pulumi.String(spec.AzureadAdministrator.LoginUsername),
			ObjectId:                  pulumi.String(spec.AzureadAdministrator.ObjectId.GetValue()),
			TenantId:                  pulumi.String(tenantId),
			AzureadAuthenticationOnly: pulumi.Bool(spec.AzureadAdministrator.AzureadAuthenticationOnly),
		}
	}

	// The server's managed identity -- unwraps the TDE customer-managed
	// key. The primary user-assigned identity is the one ARM uses for Key
	// Vault access when several are attached.
	if spec.Identity != nil {
		identityArgs := &mssql.ServerIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := make([]string, 0, len(spec.Identity.IdentityIds))
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, identityId.GetValue())
			}
			identityArgs.IdentityIds = pulumi.ToStringArray(identityIds)
		}
		serverArgs.Identity = identityArgs
	}
	if spec.PrimaryUserAssignedIdentityId.GetValue() != "" {
		serverArgs.PrimaryUserAssignedIdentityId = pulumi.String(spec.PrimaryUserAssignedIdentityId.GetValue())
	}

	// Server-level TDE customer-managed key (VERSIONED Key Vault key id --
	// ARM pins the exact version at the server level).
	if spec.TransparentDataEncryptionKeyVaultKeyId.GetValue() != "" {
		serverArgs.TransparentDataEncryptionKeyVaultKeyId = pulumi.String(spec.TransparentDataEncryptionKeyVaultKeyId.GetValue())
	}

	// Unspecified connection_policy is not sent at all, letting Azure
	// apply its Default policy -- mirroring the Terraform module's null.
	if spec.ConnectionPolicy != azuremssqlserverv1alpha1.AzureMssqlServerConnectionPolicy_azure_mssql_server_connection_policy_unspecified {
		serverArgs.ConnectionPolicy = pulumi.String(connectionPolicyStrings[spec.ConnectionPolicy])
	}

	// The TLS floor and the public-endpoint dial, presence-guarded to
	// their spec defaults.
	if spec.MinimumTlsVersion != nil {
		serverArgs.MinimumTlsVersion = pulumi.String(spec.GetMinimumTlsVersion())
	} else {
		serverArgs.MinimumTlsVersion = pulumi.String("1.2")
	}
	if spec.PublicNetworkAccessEnabled != nil {
		serverArgs.PublicNetworkAccessEnabled = pulumi.Bool(spec.GetPublicNetworkAccessEnabled())
	} else {
		serverArgs.PublicNetworkAccessEnabled = pulumi.Bool(true)
	}

	// Outbound restriction: the allowlist itself is the outbound
	// firewall-rule resources below.
	serverArgs.OutboundNetworkRestrictionEnabled = pulumi.Bool(spec.OutboundNetworkRestrictionEnabled)

	// Microsoft Defender's agentless SQL scanning (no storage account
	// needed, unlike the classic vulnerability assessment).
	serverArgs.ExpressVulnerabilityAssessmentEnabled = pulumi.Bool(spec.ExpressVulnerabilityAssessmentEnabled)

	server, err := mssql.NewServer(ctx,
		spec.ServerName,
		serverArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create mssql server %s", spec.ServerName)
	}

	// Public-endpoint firewall allowlist. Only meaningful while public
	// network access is enabled. 0.0.0.0-0.0.0.0 admits Azure-internal
	// services only.
	for _, rule := range spec.FirewallRules {
		if _, err := mssql.NewFirewallRule(ctx,
			fmt.Sprintf("%s-%s", spec.ServerName, rule.Name),
			&mssql.FirewallRuleArgs{
				Name:           pulumi.String(rule.Name),
				ServerId:       server.ID(),
				StartIpAddress: pulumi.String(rule.StartIpAddress),
				EndIpAddress:   pulumi.String(rule.EndIpAddress),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrapf(err, "failed to create firewall rule %s", rule.Name)
		}
	}

	// Subnet allowlist through Microsoft.Sql service endpoints -- the
	// classic (non-Private-Link) way to keep traffic on the Azure
	// backbone.
	for _, rule := range spec.VirtualNetworkRules {
		if _, err := mssql.NewVirtualNetworkRule(ctx,
			fmt.Sprintf("%s-%s", spec.ServerName, rule.Name),
			&mssql.VirtualNetworkRuleArgs{
				Name:                             pulumi.String(rule.Name),
				ServerId:                         server.ID(),
				SubnetId:                         pulumi.String(rule.SubnetId.GetValue()),
				IgnoreMissingVnetServiceEndpoint: pulumi.Bool(rule.IgnoreMissingVnetServiceEndpoint),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrapf(err, "failed to create virtual network rule %s", rule.Name)
		}
	}

	// The FQDNs the server may reach OUT to while outbound restriction is
	// enabled (elastic queries, linked external tables).
	for _, fqdn := range spec.OutboundFirewallRules {
		if _, err := mssql.NewOutboundFirewallRule(ctx,
			fmt.Sprintf("%s-outbound-%s", spec.ServerName, fqdn),
			&mssql.OutboundFirewallRuleArgs{
				Name:     pulumi.String(fqdn),
				ServerId: server.ID(),
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrapf(err, "failed to create outbound firewall rule %s", fqdn)
		}
	}

	// Server-level SQL auditing: audit events for every database on the
	// server go to blob storage (storage_endpoint + key) and/or Azure
	// Monitor (log_monitoring_enabled, consumed through diagnostic
	// settings). Every optional dial is presence-guarded to its
	// documented default.
	if spec.ExtendedAuditing != nil {
		auditing := spec.ExtendedAuditing
		auditingArgs := &mssql.ServerExtendedAuditingPolicyArgs{
			ServerId: server.ID(),
		}
		if auditing.StorageEndpoint != "" {
			auditingArgs.StorageEndpoint = pulumi.String(auditing.StorageEndpoint)
		}
		if auditing.StorageAccountAccessKey.GetValue() != "" {
			auditingArgs.StorageAccountAccessKey = pulumi.String(auditing.StorageAccountAccessKey.GetValue())
		}
		auditingArgs.StorageAccountAccessKeyIsSecondary = pulumi.Bool(auditing.StorageAccountAccessKeyIsSecondary)
		if auditing.RetentionInDays != nil {
			auditingArgs.RetentionInDays = pulumi.Int(int(auditing.GetRetentionInDays()))
		}
		if auditing.LogMonitoringEnabled != nil {
			auditingArgs.LogMonitoringEnabled = pulumi.Bool(auditing.GetLogMonitoringEnabled())
		} else {
			auditingArgs.LogMonitoringEnabled = pulumi.Bool(true)
		}
		if auditing.StorageAccountSubscriptionId != nil && auditing.GetStorageAccountSubscriptionId() != "" {
			auditingArgs.StorageAccountSubscriptionId = pulumi.String(auditing.GetStorageAccountSubscriptionId())
		}
		if auditing.PredicateExpression != "" {
			auditingArgs.PredicateExpression = pulumi.String(auditing.PredicateExpression)
		}
		if len(auditing.AuditActionsAndGroups) > 0 {
			auditingArgs.AuditActionsAndGroups = pulumi.ToStringArray(auditing.AuditActionsAndGroups)
		}
		if _, err := mssql.NewServerExtendedAuditingPolicy(ctx,
			fmt.Sprintf("%s-auditing", spec.ServerName),
			auditingArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrap(err, "failed to create extended auditing policy")
		}
	}

	// Microsoft Defender for SQL threat detection at the server scope.
	// The resource addresses the server by name + resource group rather
	// than by id -- its azurerm contract, not a choice here.
	if spec.SecurityAlertPolicy != nil {
		policy := spec.SecurityAlertPolicy
		disabledAlerts := make([]string, 0, len(policy.DisabledAlerts))
		for _, alert := range policy.DisabledAlerts {
			disabledAlerts = append(disabledAlerts, alertTypeStrings[alert])
		}
		alertArgs := &mssql.ServerSecurityAlertPolicyArgs{
			ResourceGroupName:  pulumi.String(locals.ResourceGroupName),
			ServerName:         server.Name,
			State:              pulumi.String(alertPolicyStateStrings[policy.State]),
			EmailAccountAdmins: pulumi.Bool(policy.EmailAccountAdmins),
		}
		if len(disabledAlerts) > 0 {
			alertArgs.DisabledAlerts = pulumi.ToStringArray(disabledAlerts)
		}
		if len(policy.EmailAddresses) > 0 {
			alertArgs.EmailAddresses = pulumi.ToStringArray(policy.EmailAddresses)
		}
		if policy.RetentionDays != nil {
			alertArgs.RetentionDays = pulumi.Int(int(policy.GetRetentionDays()))
		}
		if policy.StorageEndpoint != "" {
			alertArgs.StorageEndpoint = pulumi.String(policy.StorageEndpoint)
		}
		if policy.StorageAccountAccessKey.GetValue() != "" {
			alertArgs.StorageAccountAccessKey = pulumi.String(policy.StorageAccountAccessKey.GetValue())
		}
		if _, err := mssql.NewServerSecurityAlertPolicy(ctx,
			fmt.Sprintf("%s-alert-policy", spec.ServerName),
			alertArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(server)); err != nil {
			return errors.Wrap(err, "failed to create security alert policy")
		}
	}

	// Export outputs from the created resources.
	ctx.Export(OpServerId, server.ID())
	ctx.Export(OpServerName, server.Name)
	ctx.Export(OpFqdn, server.FullyQualifiedDomainName)
	ctx.Export(OpAdministratorLogin, server.AdministratorLogin)

	// The system-assigned identity's principal ID -- empty unless the
	// identity type includes SYSTEM_ASSIGNED, matching the Terraform
	// module's conditional output.
	hasSystemIdentity := spec.Identity != nil &&
		(spec.Identity.Type == azuremssqlserverv1alpha1.AzureMssqlServerIdentityType_SYSTEM_ASSIGNED ||
			spec.Identity.Type == azuremssqlserverv1alpha1.AzureMssqlServerIdentityType_SYSTEM_AND_USER_ASSIGNED)
	if hasSystemIdentity {
		ctx.Export(OpIdentityPrincipalId, server.Identity.PrincipalId().Elem())
	} else {
		ctx.Export(OpIdentityPrincipalId, pulumi.String(""))
	}

	return nil
}
