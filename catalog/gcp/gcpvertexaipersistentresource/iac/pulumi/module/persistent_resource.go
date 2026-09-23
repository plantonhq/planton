package module

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	gcpvertexaipersistentresourcev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpvertexaipersistentresource/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/vertex"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var numericProject = regexp.MustCompile(`^[0-9]+$`)

// persistentResource creates the persistent resource: pools of machines
// Vertex AI keeps provisioned for training jobs. Everything but the display
// name, labels, and each pool's replica_count is immutable.
//
// Send posture (parity with the Terraform module): the Optional+Computed
// levers -- a pool's id and its disk spec -- are sent only when set, and
// counts travel as decimal strings, the provider's wire type.
func persistentResource(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVertexAiPersistentResource.Spec

	// Enable the Vertex AI API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one persistent
	// resource must never disable the API for everything else in the
	// project.
	aiplatformApiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("aiplatform.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		aiplatformApiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdAiplatformApi, err := projects.NewService(ctx,
		"gcpvpr-aiplatform.googleapis.com", aiplatformApiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable aiplatform.googleapis.com api")
	}

	args := &vertex.AiPersistentResourceArgs{
		Location:      pulumi.String(spec.Location),
		Name:          pulumi.String(locals.PersistentResourceId),
		Labels:        pulumi.ToStringMap(locals.GcpLabels),
		ResourcePools: buildResourcePools(spec.ResourcePools),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.String(spec.DisplayName)
	}
	if spec.Network.GetValue() != "" {
		network, err := resolveNetwork(ctx, gcpProvider, spec.Network.GetValue())
		if err != nil {
			return err
		}
		args.Network = pulumi.String(network)
	}
	if len(spec.ReservedIpRanges) > 0 {
		args.ReservedIpRanges = pulumi.ToStringArray(spec.ReservedIpRanges)
	}
	if psc := spec.PscInterfaceConfig; psc != nil {
		pscArgs := &vertex.AiPersistentResourcePscInterfaceConfigArgs{}
		if psc.NetworkAttachment != "" {
			pscArgs.NetworkAttachment = pulumi.String(psc.NetworkAttachment)
		}
		if len(psc.DnsPeeringConfigs) > 0 {
			peerings := vertex.AiPersistentResourcePscInterfaceConfigDnsPeeringConfigArray{}
			for _, peering := range psc.DnsPeeringConfigs {
				peerings = append(peerings, &vertex.AiPersistentResourcePscInterfaceConfigDnsPeeringConfigArgs{
					Domain:        pulumi.String(peering.Domain),
					TargetProject: pulumi.String(peering.TargetProject.GetValue()),
					TargetNetwork: pulumi.String(peering.TargetNetwork.GetValue()),
				})
			}
			pscArgs.DnsPeeringConfigs = peerings
		}
		args.PscInterfaceConfig = pscArgs
	}
	// The spec lifts the two one-leaf wrappers into one bool; false and an
	// absent block are the same fact to Google, so the block is sent only
	// when true.
	if spec.EnableCustomServiceAccount {
		args.ResourceRuntimeSpec = &vertex.AiPersistentResourceResourceRuntimeSpecArgs{
			ServiceAccountSpec: &vertex.AiPersistentResourceResourceRuntimeSpecServiceAccountSpecArgs{
				EnableCustomServiceAccount: pulumi.Bool(true),
			},
		}
	}
	if spec.KmsKeyName.GetValue() != "" {
		args.EncryptionSpec = &vertex.AiPersistentResourceEncryptionSpecArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}

	// Engine-side destroy stance: DELETE releases the machines, PREVENT
	// fails destroys, ABANDON removes from management and keeps them
	// running. Sent only when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	createdResource, err := vertex.NewAiPersistentResource(ctx,
		locals.GcpVertexAiPersistentResource.Metadata.Name, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdAiplatformApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create vertex ai persistent resource")
	}

	ctx.Export(OpName, createdResource.ID().ToStringOutput())
	ctx.Export(OpPersistentResourceId, createdResource.Name)
	ctx.Export(OpLocation, createdResource.Location.Elem())
	ctx.Export(OpState, createdResource.State)
	return nil
}

// buildResourcePools maps the machine pools. Counts are int64 in the spec
// and decimal strings on the wire; a pool's id and disk spec are sent only
// when set.
func buildResourcePools(pools []*gcpvertexaipersistentresourcev1alpha1.GcpVertexAiPersistentResourceResourcePool) vertex.AiPersistentResourceResourcePoolArray {
	result := vertex.AiPersistentResourceResourcePoolArray{}
	for _, pool := range pools {
		machine := &vertex.AiPersistentResourceResourcePoolMachineSpecArgs{}
		if pool.MachineSpec.MachineType != "" {
			machine.MachineType = pulumi.String(pool.MachineSpec.MachineType)
		}
		if pool.MachineSpec.AcceleratorType != "" {
			machine.AcceleratorType = pulumi.String(pool.MachineSpec.AcceleratorType)
		}
		if pool.MachineSpec.AcceleratorCount > 0 {
			machine.AcceleratorCount = pulumi.Int(int(pool.MachineSpec.AcceleratorCount))
		}

		poolArgs := &vertex.AiPersistentResourceResourcePoolArgs{MachineSpec: machine}
		if pool.Id != "" {
			poolArgs.Id = pulumi.String(pool.Id)
		}
		if pool.ReplicaCount != nil {
			poolArgs.ReplicaCount = pulumi.String(strconv.FormatInt(pool.GetReplicaCount(), 10))
		}
		if autoscaling := pool.AutoscalingSpec; autoscaling != nil {
			autoscalingArgs := &vertex.AiPersistentResourceResourcePoolAutoscalingSpecArgs{}
			if autoscaling.MinReplicaCount != nil {
				autoscalingArgs.MinReplicaCount = pulumi.String(strconv.FormatInt(autoscaling.GetMinReplicaCount(), 10))
			}
			if autoscaling.MaxReplicaCount != nil {
				autoscalingArgs.MaxReplicaCount = pulumi.String(strconv.FormatInt(autoscaling.GetMaxReplicaCount(), 10))
			}
			poolArgs.AutoscalingSpec = autoscalingArgs
		}
		if disk := pool.DiskSpec; disk != nil {
			diskArgs := &vertex.AiPersistentResourceResourcePoolDiskSpecArgs{}
			if disk.BootDiskSizeGb != nil {
				diskArgs.BootDiskSizeGb = pulumi.Int(int(disk.GetBootDiskSizeGb()))
			}
			if disk.BootDiskType != "" {
				diskArgs.BootDiskType = pulumi.String(disk.BootDiskType)
			}
			poolArgs.DiskSpec = diskArgs
		}
		result = append(result, poolArgs)
	}
	return result
}

// resolveNetwork turns a GcpVpcNetwork self-link or a literal path into
// the projects/{NUMBER}/global/networks/{name} form Google requires. When
// the project segment is not already a number, one project lookup resolves
// it -- the same guarded read the Terraform module performs.
func resolveNetwork(ctx *pulumi.Context, gcpProvider *gcp.Provider, value string) (string, error) {
	path := strings.TrimPrefix(value, "https://www.googleapis.com/compute/v1/")
	segments := strings.Split(path, "/")
	if len(segments) < 2 {
		return path, nil
	}
	project := segments[1]
	if numericProject.MatchString(project) {
		return path, nil
	}
	lookup, err := organizations.LookupProject(ctx,
		&organizations.LookupProjectArgs{ProjectId: pulumi.StringRef(project)},
		pulumi.Provider(gcpProvider))
	if err != nil {
		return "", errors.Wrapf(err, "failed to resolve the project number for network %s", value)
	}
	if lookup.Number == "" {
		return "", errors.Errorf("project %s has no number; cannot build the network path Google requires", project)
	}
	return "projects/" + lookup.Number + "/global/networks/" + segments[len(segments)-1], nil
}
