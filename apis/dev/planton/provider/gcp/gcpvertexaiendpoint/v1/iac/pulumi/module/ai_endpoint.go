package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func aiEndpoint(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiEndpoint.Spec

	args := &vertex.AiEndpointArgs{
		DisplayName: pulumi.String(spec.DisplayName),
		Location:    pulumi.String(spec.Location),
		Project:     pulumi.StringPtr(spec.ProjectId.GetValue()),
		Labels:      pulumi.ToStringMap(locals.GcpLabels),
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// Endpoint name (numeric GCP resource identifier).
	if spec.EndpointName != "" {
		args.Name = pulumi.StringPtr(spec.EndpointName)
	}

	// VPC-peered private networking.
	if spec.Network != nil && spec.Network.GetValue() != "" {
		args.Network = pulumi.StringPtr(spec.Network.GetValue())
	}

	// CMEK encryption.
	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &vertex.AiEndpointEncryptionSpecArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}

	// Dedicated endpoint DNS.
	if spec.DedicatedEndpointEnabled {
		args.DedicatedEndpointEnabled = pulumi.BoolPtr(true)
	}

	// Private Service Connect configuration.
	if spec.PrivateServiceConnectConfig != nil {
		pscArgs := &vertex.AiEndpointPrivateServiceConnectConfigArgs{
			EnablePrivateServiceConnect: pulumi.Bool(true),
		}
		if len(spec.PrivateServiceConnectConfig.ProjectAllowlist) > 0 {
			pscArgs.ProjectAllowlists = pulumi.ToStringArray(spec.PrivateServiceConnectConfig.ProjectAllowlist)
		}
		args.PrivateServiceConnectConfig = pscArgs
	}

	createdEndpoint, err := vertex.NewAiEndpoint(ctx, "vertex-ai-endpoint", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create vertex ai endpoint")
	}

	ctx.Export(OpEndpointId, createdEndpoint.ID())
	ctx.Export(OpDisplayName, createdEndpoint.DisplayName)
	ctx.Export(OpDedicatedEndpointDns, createdEndpoint.DedicatedEndpointDns)
	ctx.Export(OpCreateTime, createdEndpoint.CreateTime)

	return nil
}
