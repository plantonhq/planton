package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudnatgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudnatgateway/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/ecs"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/vpc"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudnatgatewayv1.AliCloudNatGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudNatGateway.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	// Look up the EIP's public IP address from its allocation ID.
	eipId := spec.EipId.GetValue()
	eipLookup, err := ecs.GetEipAddresses(ctx, &ecs.GetEipAddressesArgs{
		Ids: []string{eipId},
	}, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to look up EIP %s", eipId)
	}
	if len(eipLookup.Addresses) == 0 {
		return fmt.Errorf("EIP %s not found", eipId)
	}
	snatIp := eipLookup.Addresses[0].IpAddress

	natGatewayArgs := &vpc.NatGatewayArgs{
		NatGatewayName:     pulumi.String(spec.NatGatewayName),
		VpcId:              pulumi.String(spec.VpcId.GetValue()),
		VswitchId:          pulumi.String(spec.VswitchId.GetValue()),
		NatType:            pulumi.String(natType(spec)),
		PaymentType:        pulumi.String(paymentType(spec)),
		InternetChargeType: pulumi.String(internetChargeType(spec)),
		Tags:               pulumi.ToStringMap(locals.Tags),
	}

	if spec.Description != "" {
		natGatewayArgs.Description = pulumi.String(spec.Description)
	}

	if spec.Specification != nil && *spec.Specification != "" {
		natGatewayArgs.Specification = pulumi.String(*spec.Specification)
	}

	if spec.DeletionProtection != nil {
		natGatewayArgs.DeletionProtection = pulumi.Bool(*spec.DeletionProtection)
	}

	natGateway, err := vpc.NewNatGateway(ctx, spec.NatGatewayName, natGatewayArgs,
		pulumi.Provider(alicloudProvider),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to create NAT gateway %s", spec.NatGatewayName)
	}

	// Associate the EIP with the NAT Gateway.
	_, err = ecs.NewEipAssociation(ctx, fmt.Sprintf("%s-eip", spec.NatGatewayName), &ecs.EipAssociationArgs{
		AllocationId: pulumi.String(eipId),
		InstanceId:   natGateway.ID(),
		InstanceType: pulumi.String("Nat"),
	},
		pulumi.Provider(alicloudProvider),
		pulumi.Parent(natGateway),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to associate EIP %s with NAT gateway %s", eipId, spec.NatGatewayName)
	}

	// Create SNAT entries.
	for i, entry := range spec.SnatEntries {
		if err := snatEntry(ctx, alicloudProvider, natGateway, spec.NatGatewayName, snatIp, i, entry); err != nil {
			return err
		}
	}

	ctx.Export(OpNatGatewayId, natGateway.ID())
	ctx.Export(OpNatGatewayName, natGateway.NatGatewayName)
	ctx.Export(OpSnatTableId, natGateway.SnatTableIds)
	ctx.Export(OpForwardTableId, natGateway.ForwardTableIds)

	return nil
}
