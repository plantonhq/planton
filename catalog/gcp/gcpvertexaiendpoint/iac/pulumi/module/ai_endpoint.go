package module

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func aiEndpoint(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiEndpoint.Spec

	// Enable the Vertex AI API — the control plane that owns endpoints.
	// DisableOnDestroy stays false: tearing down one endpoint must never
	// disable the API for everything else in the project (other endpoints
	// keep serving).
	aiplatformApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	// Honor the spec contract: an empty project_id falls back to the
	// provider's default project.
	if spec.ProjectId.GetValue() != "" {
		aiplatformApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdAiplatformApi, err := projects.NewService(ctx,
		"gcpvep-aiplatform.googleapis.com", aiplatformApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	// The Vertex AI endpoint — the durable serving surface models deploy
	// onto. `Name` is ALWAYS sent: the API requires a numeric ID (no
	// leading zeros, max 10 digits) and never generates one, and the
	// engine's own auto-naming would produce a non-numeric name the API
	// rejects. locals.EndpointName is the spec's explicit value or the
	// identity-derived ID shared byte-for-byte with the Terraform module.
	args := &vertex.AiEndpointArgs{
		Name:        pulumi.StringPtr(locals.EndpointName),
		DisplayName: pulumi.String(spec.DisplayName),
		Location:    pulumi.String(spec.Location),

		// The provider resolves the Vertex AI API host from `region`
		// (https://{region}-aiplatform.googleapis.com), never from
		// `location` — without it, deploys fail with "Cannot determine
		// region" unless the provider config happens to carry one. The two
		// fields are the same axis for this regional API, so the module
		// pins region to location and the spec keeps a single honest field.
		Region: pulumi.StringPtr(spec.Location),
		Labels: pulumi.ToStringMap(locals.GcpLabels),
	}
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.StringPtr(spec.ProjectId.GetValue())
	}

	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// VPC-peered private networking (requires Private Services Access on
	// the network; mutually exclusive with PSC — enforced pre-deploy by
	// the spec's CEL rule).
	if spec.Network != nil && spec.Network.GetValue() != "" {
		args.Network = pulumi.StringPtr(spec.Network.GetValue())
	}

	// CMEK: the endpoint and all sub-resources encrypted under this key.
	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &vertex.AiEndpointEncryptionSpecArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}

	// Dedicated endpoint DNS: isolated hostname with better performance
	// and reliability than the shared regional DNS.
	if spec.DedicatedEndpointEnabled {
		args.DedicatedEndpointEnabled = pulumi.BoolPtr(true)
	}

	// Private Service Connect: the endpoint exposed via a service
	// attachment. Secure PSC (IAM authorization on top of network
	// reachability) is not offered: the GA provider does not expose it,
	// and GA is the catalog's parity baseline.
	// The block's switch (on by default once declared, filled by the platform
	// before this runs) decides whether the PSC block renders at all: the
	// API rejects the block with its flag off, so "off" is the absent block.
	if spec.PrivateServiceConnectConfig != nil && spec.PrivateServiceConnectConfig.GetEnabled() {
		pscArgs := &vertex.AiEndpointPrivateServiceConnectConfigArgs{
			EnablePrivateServiceConnect: pulumi.Bool(true),
		}
		if len(spec.PrivateServiceConnectConfig.ProjectAllowlist) > 0 {
			pscArgs.ProjectAllowlists = pulumi.ToStringArray(spec.PrivateServiceConnectConfig.ProjectAllowlist)
		}
		// PSC endpoints Vertex AI provisions automatically in consumer
		// projects. The API wants the relative network form; references
		// arrive as self-links and are normalized here (mirrors the
		// Terraform module).
		if len(spec.PrivateServiceConnectConfig.PscAutomationConfigs) > 0 {
			automationConfigs := vertex.AiEndpointPrivateServiceConnectConfigPscAutomationConfigArray{}
			for _, config := range spec.PrivateServiceConnectConfig.PscAutomationConfigs {
				automationConfigs = append(automationConfigs,
					&vertex.AiEndpointPrivateServiceConnectConfigPscAutomationConfigArgs{
						Network: pulumi.String(strings.TrimPrefix(
							config.Network.GetValue(), "https://www.googleapis.com/compute/v1/")),
						ProjectId: pulumi.String(config.ProjectId.GetValue()),
					})
			}
			pscArgs.PscAutomationConfigs = automationConfigs
		}
		args.PrivateServiceConnectConfig = pscArgs
	}

	// Traffic routing across deployed models: the provider takes the split
	// as a JSON string; json.Marshal renders map keys in sorted order, so
	// the same spec always produces the same string (matching the Terraform
	// module's jsonencode). Empty means "no traffic accepted" and is
	// deliberately omitted — GCP rejects IDs that are not currently
	// deployed.
	if len(spec.TrafficSplit) > 0 {
		trafficSplitJson, err := json.Marshal(spec.TrafficSplit)
		if err != nil {
			return errors.Wrap(err, "failed to marshal traffic_split to json")
		}
		args.TrafficSplit = pulumi.StringPtr(string(trafficSplitJson))
	}

	// Request/response logging: samples online predictions into BigQuery —
	// the raw material for drift monitoring and audit. sampling_rate 0
	// means "not set" (the API then applies its own default).
	if spec.RequestResponseLoggingConfig != nil {
		loggingArgs := &vertex.AiEndpointPredictRequestResponseLoggingConfigArgs{}
		if spec.RequestResponseLoggingConfig.Enabled {
			loggingArgs.Enabled = pulumi.BoolPtr(true)
		}
		if spec.RequestResponseLoggingConfig.SamplingRate != 0 {
			loggingArgs.SamplingRate = pulumi.Float64Ptr(spec.RequestResponseLoggingConfig.SamplingRate)
		}
		if spec.RequestResponseLoggingConfig.BigqueryDestinationUri != "" {
			loggingArgs.BigqueryDestination = &vertex.AiEndpointPredictRequestResponseLoggingConfigBigqueryDestinationArgs{
				OutputUri: pulumi.StringPtr(spec.RequestResponseLoggingConfig.BigqueryDestinationUri),
			}
		}
		args.PredictRequestResponseLoggingConfig = loggingArgs
	}

	createdEndpoint, err := vertex.NewAiEndpoint(ctx, "vertex-ai-endpoint", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAiplatformApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create vertex ai endpoint")
	}

	ctx.Export(OpEndpointId, createdEndpoint.ID())
	ctx.Export(OpDisplayName, createdEndpoint.DisplayName)
	ctx.Export(OpDedicatedEndpointDns, createdEndpoint.DedicatedEndpointDns)
	ctx.Export(OpCreateTime, createdEndpoint.CreateTime)
	ctx.Export(OpEndpointName, createdEndpoint.Name)

	return nil
}
