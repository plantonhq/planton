package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/apigateway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func gatewayResource(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciApiGateway.Spec

	args := &apigateway.GatewayArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		EndpointType:  pulumi.String(endpointTypeMap[spec.EndpointType]),
		SubnetId:      pulumi.String(spec.SubnetId.GetValue()),
		DisplayName:   pulumi.String(locals.DisplayName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.CertificateId != "" {
		args.CertificateId = pulumi.String(spec.CertificateId)
	}

	if len(spec.NetworkSecurityGroupIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(spec.NetworkSecurityGroupIds))
		for i, n := range spec.NetworkSecurityGroupIds {
			nsgIds[i] = pulumi.String(n.GetValue())
		}
		args.NetworkSecurityGroupIds = nsgIds
	}

	createdGateway, err := apigateway.NewGateway(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create api gateway")
	}

	ctx.Export(OpGatewayId, createdGateway.ID())
	ctx.Export(OpHostname, createdGateway.Hostname)

	if err := deploymentResource(ctx, locals, provider, createdGateway); err != nil {
		return errors.Wrap(err, "failed to create api deployment")
	}

	return nil
}
