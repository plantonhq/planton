package module

import (
	"github.com/pkg/errors"
	azurevpnserverconfigurationv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevpnserverconfiguration/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevpnserverconfigurationv1alpha1.AzureVpnServerConfigurationIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVpnServerConfiguration.Spec

	// Create the VPN server configuration -- the reusable point-to-site
	// authentication policy (Entra ID / certificate / RADIUS, trusted
	// certificates, tunnel protocols) point-to-site VPN gateways attach
	// to. The object is free, deploys in seconds, and gateways pick up
	// in-place changes without redeploying. The spec's CEL contracts
	// already guarantee each enabled authentication type brings its
	// block -- the provider enforces the same three rules at apply time.
	configurationArgs := &network.VpnServerConfigurationArgs{
		Name:              pulumi.String(spec.Name),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// The wire values are the spec's own vocabulary ("AAD",
		// "Certificate", "Radius") -- no mapping needed.
		VpnAuthenticationTypes: pulumi.ToStringArray(spec.VpnAuthenticationTypes),
		Tags:                   pulumi.ToStringMap(locals.AzureTags),
	}

	if spec.AadAuthentication != nil {
		configurationArgs.AzureActiveDirectoryAuthentications = network.VpnServerConfigurationAzureActiveDirectoryAuthenticationArray{
			&network.VpnServerConfigurationAzureActiveDirectoryAuthenticationArgs{
				Audience: pulumi.String(spec.AadAuthentication.Audience),
				Issuer:   pulumi.String(spec.AadAuthentication.Issuer),
				Tenant:   pulumi.String(spec.AadAuthentication.Tenant),
			},
		}
	}

	if len(spec.ClientRootCertificates) > 0 {
		clientRootCertificates := network.VpnServerConfigurationClientRootCertificateArray{}
		for _, certificate := range spec.ClientRootCertificates {
			clientRootCertificates = append(clientRootCertificates, &network.VpnServerConfigurationClientRootCertificateArgs{
				Name:           pulumi.String(certificate.Name),
				PublicCertData: pulumi.String(certificate.PublicCertData),
			})
		}
		configurationArgs.ClientRootCertificates = clientRootCertificates
	}

	if len(spec.ClientRevokedCertificates) > 0 {
		clientRevokedCertificates := network.VpnServerConfigurationClientRevokedCertificateArray{}
		for _, certificate := range spec.ClientRevokedCertificates {
			clientRevokedCertificates = append(clientRevokedCertificates, &network.VpnServerConfigurationClientRevokedCertificateArgs{
				Name:       pulumi.String(certificate.Name),
				Thumbprint: pulumi.String(certificate.Thumbprint),
			})
		}
		configurationArgs.ClientRevokedCertificates = clientRevokedCertificates
	}

	// The spec requires every field of a configured proposal (no
	// partial pinning); the vocabularies are already wire values.
	if spec.IpsecPolicy != nil {
		configurationArgs.IpsecPolicy = &network.VpnServerConfigurationIpsecPolicyArgs{
			DhGroup:             pulumi.String(spec.IpsecPolicy.DhGroup),
			IkeEncryption:       pulumi.String(spec.IpsecPolicy.IkeEncryption),
			IkeIntegrity:        pulumi.String(spec.IpsecPolicy.IkeIntegrity),
			IpsecEncryption:     pulumi.String(spec.IpsecPolicy.IpsecEncryption),
			IpsecIntegrity:      pulumi.String(spec.IpsecPolicy.IpsecIntegrity),
			PfsGroup:            pulumi.String(spec.IpsecPolicy.PfsGroup),
			SaLifetimeSeconds:   pulumi.Int(int(spec.IpsecPolicy.SaLifetimeSeconds)),
			SaDataSizeKilobytes: pulumi.Int(int(spec.IpsecPolicy.SaDataSizeKilobytes)),
		}
	}

	if spec.Radius != nil {
		radiusArgs := &network.VpnServerConfigurationRadiusArgs{}

		if len(spec.Radius.Servers) > 0 {
			servers := network.VpnServerConfigurationRadiusServerArray{}
			for _, server := range spec.Radius.Servers {
				servers = append(servers, &network.VpnServerConfigurationRadiusServerArgs{
					Address: pulumi.String(server.Address),
					// Sensitive on the provider schema (masked in
					// state/preview); ARM never returns it on reads --
					// the import round-trip declares the matching
					// tolerance.
					Secret: pulumi.String(server.Secret.GetValue()),
					Score:  pulumi.Int(int(server.Score)),
				})
			}
			radiusArgs.Servers = servers
		}

		if len(spec.Radius.ClientRootCertificates) > 0 {
			clientRootCertificates := network.VpnServerConfigurationRadiusClientRootCertificateArray{}
			for _, certificate := range spec.Radius.ClientRootCertificates {
				clientRootCertificates = append(clientRootCertificates, &network.VpnServerConfigurationRadiusClientRootCertificateArgs{
					Name:       pulumi.String(certificate.Name),
					Thumbprint: pulumi.String(certificate.Thumbprint),
				})
			}
			radiusArgs.ClientRootCertificates = clientRootCertificates
		}

		if len(spec.Radius.ServerRootCertificates) > 0 {
			serverRootCertificates := network.VpnServerConfigurationRadiusServerRootCertificateArray{}
			for _, certificate := range spec.Radius.ServerRootCertificates {
				serverRootCertificates = append(serverRootCertificates, &network.VpnServerConfigurationRadiusServerRootCertificateArgs{
					Name:           pulumi.String(certificate.Name),
					PublicCertData: pulumi.String(certificate.PublicCertData),
				})
			}
			radiusArgs.ServerRootCertificates = serverRootCertificates
		}

		configurationArgs.Radius = radiusArgs
	}

	// Optional+Computed on the provider: omit when the spec leaves it
	// empty so ARM's default selection applies and reads don't drift.
	if len(spec.VpnProtocols) > 0 {
		configurationArgs.VpnProtocols = pulumi.ToStringArray(spec.VpnProtocols)
	}

	createdConfiguration, err := network.NewVpnServerConfiguration(ctx,
		spec.Name,
		configurationArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create vpn server configuration %s", spec.Name)
	}

	// The composed policy groups: standalone ARM children of the
	// configuration, one per spec entry, keyed by name (named
	// member-matching rules a point-to-site gateway maps to address
	// pools). The policy_group_ids output publishes each group's ARM
	// id.
	policyGroupIds := pulumi.Map{}
	for _, policyGroup := range spec.PolicyGroups {
		policies := network.VpnServerConfigurationPolicyGroupPolicyArray{}
		for _, policy := range policyGroup.Policies {
			policies = append(policies, &network.VpnServerConfigurationPolicyGroupPolicyArgs{
				Name:  pulumi.String(policy.Name),
				Type:  pulumi.String(policy.Type),
				Value: pulumi.String(policy.Value),
			})
		}

		createdPolicyGroup, err := network.NewVpnServerConfigurationPolicyGroup(ctx,
			spec.Name+"-"+policyGroup.Name,
			&network.VpnServerConfigurationPolicyGroupArgs{
				Name:                     pulumi.String(policyGroup.Name),
				VpnServerConfigurationId: createdConfiguration.ID(),
				// Both ForceNew on the group: is_default marks the
				// catch-all group.
				IsDefault: pulumi.Bool(policyGroup.IsDefault),
				Priority:  pulumi.Int(int(policyGroup.Priority)),
				Policies:  policies,
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdConfiguration))
		if err != nil {
			return errors.Wrapf(err, "failed to create policy group %s", policyGroup.Name)
		}
		policyGroupIds[policyGroup.Name] = createdPolicyGroup.ID()
	}

	ctx.Export(OpVpnServerConfigurationId, createdConfiguration.ID())
	ctx.Export(OpVpnServerConfigurationName, createdConfiguration.Name)
	ctx.Export(OpPolicyGroupIds, policyGroupIds)

	return nil
}
