package module

import (
	"github.com/pkg/errors"
	azuresearchservicev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuresearchservice/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/search"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuresearchservicev1alpha1.AzureSearchServiceIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureSearchService.Spec

	// Create the Azure AI Search service. Capacity is sku x
	// partitions x replicas; the spec's CEL contracts already enforce
	// the provider's per-SKU caps and pairing rules (high-density on
	// standard3 only, failure mode only with local auth, semantic not
	// on free), so by the time this module runs the shape is legal.
	// The SKU changes in place ONLY along basic -> standard ->
	// standard2 -> standard3 (the provider's update contract) --
	// every other SKU change replaces the service.
	serviceArgs := &search.ServiceArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Sku:               pulumi.String(spec.Sku),
		// Provider default false -- an explicit false is the same wire.
		CustomerManagedKeyEnforcementEnabled: pulumi.Bool(spec.CustomerManagedKeyEnforcementEnabled),
		Tags:                                 pulumi.ToStringMap(locals.AzureTags),
	}

	// Optional-with-default-1 in the spec; presence-guard with the
	// proto default so manifest-driven IaC inputs (nil optional)
	// send the same value the Terraform module's optional(number, 1)
	// carries -- the provider range-validates these, so sending 0
	// would hard-fail the deploy.
	if spec.ReplicaCount != nil {
		serviceArgs.ReplicaCount = pulumi.Int(int(*spec.ReplicaCount))
	} else {
		serviceArgs.ReplicaCount = pulumi.Int(1)
	}
	if spec.PartitionCount != nil {
		serviceArgs.PartitionCount = pulumi.Int(int(*spec.PartitionCount))
	} else {
		serviceArgs.PartitionCount = pulumi.Int(1)
	}

	// Enum name -> wire value; unspecified omits the property so the
	// provider applies its default, "default". ForceNew.
	if hostingMode, ok := hostingModeWire[spec.HostingMode]; ok {
		serviceArgs.HostingMode = pulumi.String(hostingMode)
	}

	// Optional-with-default-true: presence-guard with the proto
	// default (matches the provider default). Setting false is the
	// RBAC-only posture -- admin/query keys stop working.
	if spec.LocalAuthenticationEnabled != nil {
		serviceArgs.LocalAuthenticationEnabled = pulumi.Bool(*spec.LocalAuthenticationEnabled)
	} else {
		serviceArgs.LocalAuthenticationEnabled = pulumi.Bool(true)
	}

	// Setting a failure mode is what enables RBAC alongside API keys;
	// omitted, the service stays in API-keys-only mode.
	if spec.AuthenticationFailureMode != "" {
		serviceArgs.AuthenticationFailureMode = pulumi.String(spec.AuthenticationFailureMode)
	}

	if spec.PublicNetworkAccessEnabled != nil {
		serviceArgs.PublicNetworkAccessEnabled = pulumi.Bool(*spec.PublicNetworkAccessEnabled)
	} else {
		serviceArgs.PublicNetworkAccessEnabled = pulumi.Bool(true)
	}

	// Omitted, the provider sends "disabled" -- semantic ranking off.
	if spec.SemanticSearchSku != "" {
		serviceArgs.SemanticSearchSku = pulumi.String(spec.SemanticSearchSku)
	}

	if len(spec.AllowedIps) > 0 {
		allowedIps := pulumi.StringArray{}
		for _, ip := range spec.AllowedIps {
			allowedIps = append(allowedIps, pulumi.String(ip))
		}
		serviceArgs.AllowedIps = allowedIps
	}

	// Unspecified omits the property so the provider applies its
	// default, "None".
	if spec.NetworkRuleBypassOption != "" {
		serviceArgs.NetworkRuleBypassOption = pulumi.String(spec.NetworkRuleBypassOption)
	}

	if spec.Identity != nil {
		identityArgs := &search.ServiceIdentityArgs{
			Type: pulumi.String(identityTypeWire[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		serviceArgs.Identity = identityArgs
	}

	createdService, err := search.NewService(ctx,
		spec.Name,
		serviceArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create search service %s", spec.Name)
	}

	// The composed shared private links: standalone ARM children
	// (.../sharedPrivateLinkResources/{name}), one per spec entry,
	// keyed by name (uniqueness is spec CEL). Each link sits
	// "Pending" until the target resource's owner approves it --
	// creating the link never requires the target side's consent.
	sharedPrivateLinkServiceIds := pulumi.Map{}
	for _, link := range spec.SharedPrivateLinkServices {
		linkArgs := &search.SharedPrivateLinkServiceArgs{
			Name:             pulumi.String(link.Name),
			SearchServiceId:  createdService.ID(),
			SubresourceName:  pulumi.String(link.SubresourceName),
			TargetResourceId: pulumi.String(link.TargetResourceId.GetValue()),
		}
		if link.RequestMessage != "" {
			linkArgs.RequestMessage = pulumi.String(link.RequestMessage)
		}
		createdLink, err := search.NewSharedPrivateLinkService(ctx,
			spec.Name+"-"+link.Name,
			linkArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdService))
		if err != nil {
			return errors.Wrapf(err, "failed to create shared private link %s", link.Name)
		}
		sharedPrivateLinkServiceIds[link.Name] = createdLink.ID()
	}

	ctx.Export(OpSearchServiceId, createdService.ID())
	ctx.Export(OpSearchServiceName, createdService.Name)
	ctx.Export(OpEndpoint, createdService.Endpoint)
	// The service-minted credentials (no vault indirection exists) --
	// exported as secrets so both engines mask them. Empty when local
	// authentication is disabled.
	ctx.Export(OpPrimaryKey, pulumi.ToSecret(createdService.PrimaryKey))
	ctx.Export(OpSecondaryKey, pulumi.ToSecret(createdService.SecondaryKey))
	// The service creates exactly ONE query key at provisioning, with
	// an empty name -- exporting the single key beats a name-keyed map
	// whose only key would be the empty string.
	ctx.Export(OpDefaultQueryKey, pulumi.ToSecret(createdService.QueryKeys.Index(pulumi.Int(0)).Key().Elem()))
	ctx.Export(OpCmkEncryptionComplianceStatus, createdService.CustomerManagedKeyEncryptionComplianceStatus)
	ctx.Export(OpSystemAssignedIdentityPrincipalId, createdService.Identity.PrincipalId())
	ctx.Export(OpSharedPrivateLinkServiceIds, sharedPrivateLinkServiceIds)

	return nil
}
