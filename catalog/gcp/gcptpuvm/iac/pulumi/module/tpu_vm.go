package module

import (
	"strings"

	"github.com/pkg/errors"
	gcptpuvmv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcptpuvm/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/tpu"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Compute references (a GcpSubnetwork, a GcpComputeDisk) arrive as
// self-links; Cloud TPU takes the relative path.
const computeSelfLinkPrefix = "https://www.googleapis.com/compute/v1/"

// tpuVm creates the Cloud TPU VM. pulumi-gcp serves Google's beta-only
// google_tpu_v2_vm from its single provider, so no second provider is
// involved. Only the description, labels, metadata, tags, and data disks
// update in place; everything else replaces the TPU. With neither
// accelerator_type nor accelerator_config set, the provider asks Google for
// a v2-8 -- the Terraform module's behavior.
func tpuVm(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpTpuVm.Spec
	metadata := locals.GcpTpuVm.Metadata

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
		"gcptpu-tpu.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable tpu.googleapis.com api")
	}

	// The TPU's id defaults to metadata.name -- identical to the Terraform
	// module.
	nodeId := spec.NodeId
	if nodeId == "" {
		nodeId = metadata.Name
	}

	args := &tpu.V2VmArgs{
		Zone:           pulumi.String(spec.Zone),
		Name:           pulumi.String(nodeId),
		RuntimeVersion: pulumi.String(spec.RuntimeVersion),
		Labels:         pulumi.ToStringMap(locals.GcpLabels),
	}
	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.AcceleratorType != "" {
		args.AcceleratorType = pulumi.String(spec.AcceleratorType)
	}
	if c := spec.AcceleratorConfig; c != nil {
		args.AcceleratorConfig = &tpu.V2VmAcceleratorConfigArgs{
			Type:     pulumi.String(c.Type),
			Topology: pulumi.String(c.Topology),
		}
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	if spec.CidrBlock != "" {
		args.CidrBlock = pulumi.String(spec.CidrBlock)
	}
	if len(spec.Metadata) > 0 {
		args.Metadata = pulumi.ToStringMap(spec.Metadata)
	}
	if len(spec.Tags) > 0 {
		args.Tags = pulumi.ToStringArray(spec.Tags)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	// One interface renders the singular block; several render the repeated
	// one -- the spec allows only one of the two, as Google does.
	if nc := spec.NetworkConfig; nc != nil {
		args.NetworkConfig = buildNetworkConfig(nc)
	} else if len(spec.NetworkConfigs) > 0 {
		configs := tpu.V2VmNetworkConfigArray{}
		for _, nc := range spec.NetworkConfigs {
			configs = append(configs, buildNetworkConfig(nc))
		}
		args.NetworkConfigs = configs
	}

	if sa := spec.ServiceAccount; sa != nil {
		account := &tpu.V2VmServiceAccountArgs{}
		if sa.Email.GetValue() != "" {
			account.Email = pulumi.String(sa.Email.GetValue())
		}
		if len(sa.Scopes) > 0 {
			account.Scopes = pulumi.ToStringArray(sa.Scopes)
		}
		args.ServiceAccount = account
	}
	if s := spec.SchedulingConfig; s != nil {
		args.SchedulingConfig = &tpu.V2VmSchedulingConfigArgs{
			Preemptible: pulumi.Bool(s.Preemptible),
			Spot:        pulumi.Bool(s.Spot),
			Reserved:    pulumi.Bool(s.Reserved),
		}
	}
	if len(spec.DataDisks) > 0 {
		disks := tpu.V2VmDataDiskArray{}
		for _, d := range spec.DataDisks {
			disk := &tpu.V2VmDataDiskArgs{
				SourceDisk: pulumi.String(strings.TrimPrefix(d.SourceDisk.GetValue(), computeSelfLinkPrefix)),
			}
			if d.Mode != "" {
				disk.Mode = pulumi.String(d.Mode)
			}
			disks = append(disks, disk)
		}
		args.DataDisks = disks
	}
	// The spec lifts Google's one-leaf shielded_instance_config wrapper; the
	// block is sent only when Secure Boot is on.
	if spec.EnableSecureBoot {
		args.ShieldedInstanceConfig = &tpu.V2VmShieldedInstanceConfigArgs{EnableSecureBoot: pulumi.Bool(true)}
	}

	createdVm, err := tpu.NewV2Vm(ctx, metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create tpu vm")
	}

	ctx.Export(OpName, createdVm.ID())
	ctx.Export(OpNodeId, createdVm.Name)
	ctx.Export(OpZone, createdVm.Zone)
	return nil
}

func buildNetworkConfig(nc *gcptpuvmv1alpha1.GcpTpuVmNetworkConfig) *tpu.V2VmNetworkConfigArgs {
	args := &tpu.V2VmNetworkConfigArgs{
		EnableExternalIps: pulumi.Bool(nc.EnableExternalIps),
		CanIpForward:      pulumi.Bool(nc.CanIpForward),
	}
	if nc.Network.GetValue() != "" {
		args.Network = pulumi.String(nc.Network.GetValue())
	}
	if nc.Subnetwork.GetValue() != "" {
		args.Subnetwork = pulumi.String(strings.TrimPrefix(nc.Subnetwork.GetValue(), computeSelfLinkPrefix))
	}
	if nc.QueueCount != 0 {
		args.QueueCount = pulumi.Int(int(nc.QueueCount))
	}
	return args
}
