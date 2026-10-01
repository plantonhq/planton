package module

import (
	"github.com/pkg/errors"
	azurefrontdoororiginv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoororigin/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cdn"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurefrontdoororiginv1alpha1.AzureFrontDoorOriginStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFrontDoorOrigin.Spec

	// Certificate name checking defaults ON (the spec's documented
	// default; stack inputs never carry proto defaults) -- the provider
	// requires the value explicitly, and Azure requires it to be true
	// when Private Link is configured.
	certificateNameCheckEnabled := true
	if spec.CertificateNameCheckEnabled != nil {
		certificateNameCheckEnabled = spec.GetCertificateNameCheckEnabled()
	}

	// The origin addresses its parent by the origin group's full ARM id
	// -- the provider derives the resource group, profile, and group
	// names from it. ARM does not support tags on origins. host_name is
	// a StringValueOrRef; the platform resolves valueFrom references
	// before the module runs (e.g. a web app's default_hostname or a
	// storage account's primary_web_host), so GetValue() is the resolved
	// literal.
	originArgs := &cdn.FrontdoorOriginArgs{
		Name:                        pulumi.String(spec.OriginName),
		CdnFrontdoorOriginGroupId:   pulumi.String(locals.OriginGroupId),
		HostName:                    pulumi.String(spec.HostName.GetValue()),
		CertificateNameCheckEnabled: pulumi.Bool(certificateNameCheckEnabled),
	}

	// Optional dials are sent only when set: Azure's own defaults (ports
	// 80/443, priority 1, weight 500, enabled) apply when omitted, and
	// the platform materializes the documented defaults centrally.
	if spec.OriginHostHeader != nil && spec.OriginHostHeader.GetValue() != "" {
		originArgs.OriginHostHeader = pulumi.String(spec.OriginHostHeader.GetValue())
	}
	if spec.HttpPort != nil {
		originArgs.HttpPort = pulumi.Int(int(spec.GetHttpPort()))
	}
	if spec.HttpsPort != nil {
		originArgs.HttpsPort = pulumi.Int(int(spec.GetHttpsPort()))
	}
	if spec.Priority != nil {
		originArgs.Priority = pulumi.Int(int(spec.GetPriority()))
	}
	if spec.Weight != nil {
		originArgs.Weight = pulumi.Int(int(spec.GetWeight()))
	}
	if spec.Enabled != nil {
		originArgs.Enabled = pulumi.Bool(spec.GetEnabled())
	}

	// Private Link keeps origin traffic off the public internet.
	// PREMIUM-profile only -- Azure rejects it at apply on STANDARD (the
	// SKU lives on a different resource, so the spec cannot check it
	// statically). target_type is omitted for Private Link Service
	// targets, whose ARM id is itself the attachment point (the spec's
	// CEL enforces that pairing).
	if spec.PrivateLink != nil {
		privateLinkArgs := &cdn.FrontdoorOriginPrivateLinkArgs{
			Location:            pulumi.String(spec.PrivateLink.Location),
			PrivateLinkTargetId: pulumi.String(spec.PrivateLink.PrivateLinkTargetId.GetValue()),
		}
		if spec.PrivateLink.TargetType != azurefrontdoororiginv1alpha1.AzureFrontDoorOriginPrivateLinkTargetType_azure_front_door_origin_private_link_target_type_unspecified {
			privateLinkArgs.TargetType = pulumi.String(privateLinkTargetTypeStrings[spec.PrivateLink.TargetType])
		}
		if spec.PrivateLink.RequestMessage != nil {
			privateLinkArgs.RequestMessage = pulumi.String(spec.PrivateLink.GetRequestMessage())
		}
		originArgs.PrivateLink = privateLinkArgs
	}

	createdOrigin, err := cdn.NewFrontdoorOrigin(ctx,
		spec.OriginName,
		originArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create front door origin %s", spec.OriginName)
	}

	// Export stack outputs. origin_id is what AzureFrontDoorRoute's
	// origin_ids list references to sequence deployment.
	ctx.Export(OpOriginId, createdOrigin.ID())
	ctx.Export(OpOriginName, createdOrigin.Name)

	return nil
}
