package module

import (
	"github.com/pkg/errors"
	azurecognitiveaccountv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecognitiveaccount/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cognitive"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecognitiveaccountv1alpha1.AzureCognitiveAccountIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCognitiveAccount.Spec

	// Create the Azure AI services account -- the container Azure AI
	// capabilities are provisioned and billed through (Azure OpenAI,
	// the multi-service AIServices account, the single-service
	// accounts). The spec's CEL contracts already enforce every
	// kind-gated rule the provider checks at apply time. Deletion is a
	// SOFT delete: the ghost keeps holding the account name until
	// purged (the provider purges by default).
	accountArgs := &cognitive.AccountArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// Both vocabularies are already wire values in the spec.
		Kind:    pulumi.String(spec.Kind),
		SkuName: pulumi.String(spec.SkuName),
		// Only legal on kind "AIServices" (spec CEL).
		ProjectManagementEnabled:        pulumi.Bool(spec.ProjectManagementEnabled),
		DynamicThrottlingEnabled:        pulumi.Bool(spec.DynamicThrottlingEnabled),
		OutboundNetworkAccessRestricted: pulumi.Bool(spec.OutboundNetworkAccessRestricted),
		Tags:                            pulumi.ToStringMap(locals.AzureTags),
	}

	// Set-once: the provider replaces the account only when CHANGING an
	// existing subdomain, not when adding one to an account without it.
	if spec.CustomSubdomainName != "" {
		accountArgs.CustomSubdomainName = pulumi.String(spec.CustomSubdomainName)
	}

	if spec.CustomerManagedKey != nil {
		customerManagedKeyArgs := &cognitive.AccountCustomerManagedKeyTypeArgs{
			// A Key Vault key data-plane URL; versionless follows rotation.
			KeyVaultKeyId: pulumi.String(spec.CustomerManagedKey.KeyVaultKeyId.GetValue()),
		}
		if spec.CustomerManagedKey.IdentityClientId != "" {
			customerManagedKeyArgs.IdentityClientId = pulumi.String(spec.CustomerManagedKey.IdentityClientId)
		}
		accountArgs.CustomerManagedKey = customerManagedKeyArgs
	}

	// The outbound FQDN allowlist (with outbound_network_access_restricted).
	if len(spec.Fqdns) > 0 {
		accountArgs.Fqdns = pulumi.ToStringArray(spec.Fqdns)
	}

	if spec.Identity != nil {
		identityArgs := &cognitive.AccountIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		accountArgs.Identity = identityArgs
	}

	// Optional-with-default-true on the provider: omit when the spec
	// leaves it unset so the provider default applies.
	if spec.LocalAuthEnabled != nil {
		accountArgs.LocalAuthEnabled = pulumi.Bool(*spec.LocalAuthEnabled)
	}
	if spec.PublicNetworkAccessEnabled != nil {
		accountArgs.PublicNetworkAccessEnabled = pulumi.Bool(*spec.PublicNetworkAccessEnabled)
	}

	// MetricsAdvisor-kind only (spec CEL); all four are ForceNew.
	if spec.MetricsAdvisorAadClientId != "" {
		accountArgs.MetricsAdvisorAadClientId = pulumi.String(spec.MetricsAdvisorAadClientId)
	}
	if spec.MetricsAdvisorAadTenantId != "" {
		accountArgs.MetricsAdvisorAadTenantId = pulumi.String(spec.MetricsAdvisorAadTenantId)
	}
	if spec.MetricsAdvisorSuperUserName != "" {
		accountArgs.MetricsAdvisorSuperUserName = pulumi.String(spec.MetricsAdvisorSuperUserName)
	}
	if spec.MetricsAdvisorWebsiteName != "" {
		accountArgs.MetricsAdvisorWebsiteName = pulumi.String(spec.MetricsAdvisorWebsiteName)
	}

	if spec.NetworkAcls != nil {
		networkAclsArgs := &cognitive.AccountNetworkAclsArgs{
			// Already a wire value ("Allow"/"Deny") in the spec.
			DefaultAction: pulumi.String(spec.NetworkAcls.DefaultAction),
		}
		if len(spec.NetworkAcls.IpRules) > 0 {
			networkAclsArgs.IpRules = pulumi.ToStringArray(spec.NetworkAcls.IpRules)
		}
		if len(spec.NetworkAcls.VirtualNetworkRules) > 0 {
			virtualNetworkRules := cognitive.AccountNetworkAclsVirtualNetworkRuleArray{}
			for _, rule := range spec.NetworkAcls.VirtualNetworkRules {
				virtualNetworkRules = append(virtualNetworkRules, &cognitive.AccountNetworkAclsVirtualNetworkRuleArgs{
					SubnetId:                         pulumi.String(rule.SubnetId.GetValue()),
					IgnoreMissingVnetServiceEndpoint: pulumi.Bool(rule.IgnoreMissingVnetServiceEndpoint),
				})
			}
			networkAclsArgs.VirtualNetworkRules = virtualNetworkRules
		}
		// Enum name -> wire value; unspecified omits the property so ARM
		// applies its default. Only legal on the AI kinds (spec CEL).
		if bypass, ok := networkAclsBypassWire[spec.NetworkAcls.Bypass]; ok {
			networkAclsArgs.Bypass = pulumi.String(bypass)
		}
		accountArgs.NetworkAcls = networkAclsArgs
	}

	// AIServices-kind only (spec CEL): inject agent workloads into the
	// given delegated subnet. NOTE: after the account deletes, ARM
	// removes the subnet's service association link asynchronously --
	// the provider waits for that before finishing the destroy.
	if spec.NetworkInjection != nil {
		accountArgs.NetworkInjection = &cognitive.AccountNetworkInjectionArgs{
			Scenario: pulumi.String(spec.NetworkInjection.Scenario),
			SubnetId: pulumi.String(spec.NetworkInjection.SubnetId.GetValue()),
		}
	}

	// QnAMaker-kind only (spec CEL).
	if spec.QnaRuntimeEndpoint != "" {
		accountArgs.QnaRuntimeEndpoint = pulumi.String(spec.QnaRuntimeEndpoint)
	}

	// TextAnalytics-kind only (spec CEL). The id stays a plain string
	// until AzureSearchService registers (recorded in-place upgrade);
	// the key is sensitive -- resolved from a secret reference, masked
	// in state/preview by the provider schema.
	if spec.CustomQuestionAnsweringSearchServiceId != "" {
		accountArgs.CustomQuestionAnsweringSearchServiceId = pulumi.String(spec.CustomQuestionAnsweringSearchServiceId)
	}
	if spec.CustomQuestionAnsweringSearchServiceKey.GetValue() != "" {
		accountArgs.CustomQuestionAnsweringSearchServiceKey = pulumi.String(spec.CustomQuestionAnsweringSearchServiceKey.GetValue())
	}

	if len(spec.Storage) > 0 {
		storages := cognitive.AccountStorageArray{}
		for _, storage := range spec.Storage {
			storageArgs := &cognitive.AccountStorageArgs{
				StorageAccountId: pulumi.String(storage.StorageAccountId.GetValue()),
			}
			if storage.IdentityClientId != "" {
				storageArgs.IdentityClientId = pulumi.String(storage.IdentityClientId)
			}
			storages = append(storages, storageArgs)
		}
		accountArgs.Storages = storages
	}

	createdAccount, err := cognitive.NewAccount(ctx,
		spec.Name,
		accountArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create cognitive account %s", spec.Name)
	}

	// The composed responsible-AI blocklists: standalone ARM children
	// of the account, one per spec entry, keyed by name (named
	// containers for custom blocked content; their ITEMS are
	// data-plane). The rai_blocklist_ids output publishes each
	// blocklist's ARM id.
	raiBlocklistIds := pulumi.Map{}
	createdBlocklists := []pulumi.Resource{}
	for _, blocklist := range spec.RaiBlocklists {
		blocklistArgs := &cognitive.AccountRaiBlocklistArgs{
			Name:               pulumi.String(blocklist.Name),
			CognitiveAccountId: createdAccount.ID(),
		}
		if blocklist.Description != "" {
			blocklistArgs.Description = pulumi.String(blocklist.Description)
		}
		if len(blocklist.Tags) > 0 {
			blocklistArgs.Tags = pulumi.ToStringMap(blocklist.Tags)
		}

		createdBlocklist, err := cognitive.NewAccountRaiBlocklist(ctx,
			spec.Name+"-"+blocklist.Name,
			blocklistArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdAccount))
		if err != nil {
			return errors.Wrapf(err, "failed to create rai blocklist %s", blocklist.Name)
		}
		raiBlocklistIds[blocklist.Name] = createdBlocklist.ID()
		createdBlocklists = append(createdBlocklists, createdBlocklist)
	}

	// The composed responsible-AI (content-filter) policies: standalone
	// ARM children of the account, one per spec entry, keyed by name.
	// Model deployments select a policy by NAME via their
	// rai_policy_name. A policy's content filter may reference a
	// blocklist defined in the same spec by its name -- hence the
	// explicit dependency on every blocklist child.
	raiPolicyIds := pulumi.Map{}
	for _, policy := range spec.RaiPolicies {
		contentFilters := cognitive.AccountRaiPolicyContentFilterArray{}
		for _, filter := range policy.ContentFilters {
			// PARITY-EXCEPTION: the classic SDK bridges the pre-v5
			// provider, which marks severity_threshold REQUIRED on
			// every content filter; azurerm v5 made it optional and
			// REJECTS it on the binary filters (Jailbreak, Indirect
			// Attack, Protected Material Text/Code). A severity-less
			// filter is therefore inexpressible on this engine --
			// fail loudly rather than invent a severity ARM never
			// asked for. Unblock: a v5-bridged pulumi-azure major.
			severity, ok := raiContentLevelWire[filter.SeverityThreshold]
			if !ok {
				return errors.Errorf(
					"rai policy %s content filter %s carries no severity_threshold: "+
						"the classic Azure SDK requires one on every content filter, so "+
						"severity-less filters (the binary filters, or a graded filter "+
						"relying on ARM's default) deploy through Terraform only "+
						"(PARITY-EXCEPTION -- see the spec field comment)",
					policy.Name, filter.Name)
			}
			filterArgs := &cognitive.AccountRaiPolicyContentFilterArgs{
				Name:          pulumi.String(filter.Name),
				FilterEnabled: pulumi.Bool(filter.FilterEnabled),
				BlockEnabled:  pulumi.Bool(filter.BlockEnabled),
				// Already a wire value in the spec.
				Source:            pulumi.String(filter.Source),
				SeverityThreshold: pulumi.String(severity),
			}
			contentFilters = append(contentFilters, filterArgs)
		}

		policyArgs := &cognitive.AccountRaiPolicyArgs{
			Name:               pulumi.String(policy.Name),
			CognitiveAccountId: createdAccount.ID(),
			BasePolicyName:     pulumi.String(policy.BasePolicyName),
			ContentFilters:     contentFilters,
		}
		// Enum name -> wire value; unspecified omits the property so
		// ARM applies its default.
		if mode, ok := raiPolicyModeWire[policy.Mode]; ok {
			policyArgs.Mode = pulumi.String(mode)
		}
		if len(policy.Tags) > 0 {
			policyArgs.Tags = pulumi.ToStringMap(policy.Tags)
		}

		createdPolicy, err := cognitive.NewAccountRaiPolicy(ctx,
			spec.Name+"-"+policy.Name,
			policyArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdAccount),
			pulumi.DependsOn(createdBlocklists))
		if err != nil {
			return errors.Wrapf(err, "failed to create rai policy %s", policy.Name)
		}
		raiPolicyIds[policy.Name] = createdPolicy.ID()
	}

	ctx.Export(OpCognitiveAccountId, createdAccount.ID())
	ctx.Export(OpCognitiveAccountName, createdAccount.Name)
	ctx.Export(OpEndpoint, createdAccount.Endpoint)
	// Sensitive on the provider schema; additionally marked secret so
	// the exported outputs mask them (empty when local auth is
	// disabled).
	ctx.Export(OpPrimaryAccessKey, pulumi.ToSecret(createdAccount.PrimaryAccessKey))
	ctx.Export(OpSecondaryAccessKey, pulumi.ToSecret(createdAccount.SecondaryAccessKey))
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdAccount.Identity.PrincipalId())
	ctx.Export(OpRaiBlocklistIds, raiBlocklistIds)
	ctx.Export(OpRaiPolicyIds, raiPolicyIds)

	return nil
}
