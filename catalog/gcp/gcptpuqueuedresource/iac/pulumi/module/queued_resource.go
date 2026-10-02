package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/tpu"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// A GcpSubnetwork reference arrives as a compute self-link; Cloud TPU takes
// the relative path.
const computeSelfLinkPrefix = "https://www.googleapis.com/compute/v1/"

// queuedResource creates the Cloud TPU queued resource. pulumi-gcp serves
// Google's beta-only google_tpu_v2_queued_resource from its single
// provider. Everything is immutable; the request waits for capacity, then
// provisions every node it describes.
func queuedResource(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpTpuQueuedResource.Spec
	metadata := locals.GcpTpuQueuedResource.Metadata

	// Google wants each node's parent as projects/{project}/locations/{zone}.
	// The project is the spec's, or the provider's own when the spec names
	// none -- the Terraform module's google_client_config twin.
	project := spec.ProjectId.GetValue()
	if project == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to resolve the provider's default project for the node parent")
		}
		if clientConfig.Project == "" {
			return errors.New("the request names no project and the provider has no default project -- set spec.project_id or configure a project")
		}
		project = clientConfig.Project
	}
	nodeParent := "projects/" + project + "/locations/" + spec.Zone

	// Enable the Cloud TPU API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("tpu.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcptpuq-tpu.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable tpu.googleapis.com api")
	}

	// The request's id defaults to metadata.name -- identical to the
	// Terraform module.
	queuedResourceId := spec.QueuedResourceId
	if queuedResourceId == "" {
		queuedResourceId = metadata.Name
	}

	nodeSpecs := tpu.V2QueuedResourceTpuNodeSpecArray{}
	for _, ns := range spec.NodeSpecs {
		node := &tpu.V2QueuedResourceTpuNodeSpecNodeArgs{
			RuntimeVersion: pulumi.String(ns.Node.RuntimeVersion),
		}
		if ns.Node.AcceleratorType != "" {
			node.AcceleratorType = pulumi.String(ns.Node.AcceleratorType)
		}
		if ns.Node.Description != "" {
			node.Description = pulumi.String(ns.Node.Description)
		}
		if nc := ns.Node.NetworkConfig; nc != nil {
			network := &tpu.V2QueuedResourceTpuNodeSpecNodeNetworkConfigArgs{
				EnableExternalIps: pulumi.Bool(nc.EnableExternalIps),
				CanIpForward:      pulumi.Bool(nc.CanIpForward),
			}
			if nc.Network.GetValue() != "" {
				network.Network = pulumi.String(nc.Network.GetValue())
			}
			if nc.Subnetwork.GetValue() != "" {
				network.Subnetwork = pulumi.String(strings.TrimPrefix(nc.Subnetwork.GetValue(), computeSelfLinkPrefix))
			}
			if nc.QueueCount != 0 {
				network.QueueCount = pulumi.Int(int(nc.QueueCount))
			}
			node.NetworkConfig = network
		}
		nodeSpec := &tpu.V2QueuedResourceTpuNodeSpecArgs{
			Parent: pulumi.String(nodeParent),
			Node:   node,
		}
		if ns.NodeId != "" {
			nodeSpec.NodeId = pulumi.String(ns.NodeId)
		}
		nodeSpecs = append(nodeSpecs, nodeSpec)
	}

	// The spec lifts Google's one-field tpu wrapper to node_specs.
	args := &tpu.V2QueuedResourceArgs{
		Zone: pulumi.String(spec.Zone),
		Name: pulumi.String(queuedResourceId),
		Tpu:  &tpu.V2QueuedResourceTpuArgs{NodeSpecs: nodeSpecs},
	}
	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdRequest, err := tpu.NewV2QueuedResource(ctx, metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create tpu queued resource")
	}

	ctx.Export(OpName, createdRequest.ID())
	ctx.Export(OpQueuedResourceId, createdRequest.Name)
	ctx.Export(OpZone, createdRequest.Zone)
	return nil
}
