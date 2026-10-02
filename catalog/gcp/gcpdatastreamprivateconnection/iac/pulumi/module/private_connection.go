package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/datastream"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// privateConnection creates the Datastream private connection with exactly
// one connectivity block (the spec's rule).
func privateConnection(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpDatastreamPrivateConnection.Spec
	resourceName := locals.GcpDatastreamPrivateConnection.Metadata.Name

	// Enable the Datastream API first so a fresh project works on the first
	// deploy. DisableOnDestroy stays false: tearing down one private
	// connection must never disable the API for every profile and stream in
	// the project.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("datastream.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if spec.ProjectId.GetValue() != "" {
		apiArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdApi, err := projects.NewService(ctx,
		"gcpdspc-datastream.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable datastream.googleapis.com api")
	}

	args := &datastream.PrivateConnectionArgs{
		Location:            pulumi.String(spec.Location),
		PrivateConnectionId: pulumi.String(locals.PrivateConnectionId),
		DisplayName:         pulumi.String(locals.DisplayName),
		Labels:              pulumi.ToStringMap(locals.GcpLabels),
	}

	// An empty project falls back to the provider's default project -- the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.CreateWithoutValidation {
		args.CreateWithoutValidation = pulumi.BoolPtr(true)
	}
	if peering := spec.VpcPeeringConfig; peering != nil {
		args.VpcPeeringConfig = &datastream.PrivateConnectionVpcPeeringConfigArgs{
			Vpc:    pulumi.String(peering.Vpc.GetValue()),
			Subnet: pulumi.String(peering.Subnet),
		}
	}
	if psc := spec.PscInterfaceConfig; psc != nil {
		args.PscInterfaceConfig = &datastream.PrivateConnectionPscInterfaceConfigArgs{
			NetworkAttachment: pulumi.String(psc.NetworkAttachment),
		}
	}

	// Engine-side destroy stance: FORCE (the provider's default, which also
	// removes Datastream's routes), DELETE, PREVENT, or ABANDON. Sent only
	// when set so the provider default stays in charge.
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.String(spec.DeletionPolicy)
	}

	created, err := datastream.NewPrivateConnection(ctx, resourceName, args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create datastream private connection")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpPrivateConnectionId, created.PrivateConnectionId)
	return nil
}
