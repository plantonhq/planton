package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/gkehub"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// membership registers a cluster with the fleet. The ID, location,
// cluster, and issuer are create-time decisions; labels update in place.
func membership(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpGkeFleetMembership.Spec
	project := spec.GetProjectId().GetValue()

	// The Fleet API (GKE Hub). DisableOnDestroy stays false: unregistering
	// one cluster must never disable the API for the rest of the fleet.
	apiArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("gkehub.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}
	if project != "" {
		apiArgs.Project = pulumi.String(project)
	}
	createdGkehubApi, err := projects.NewService(ctx, "gcpflmb-gkehub.googleapis.com", apiArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable gkehub.googleapis.com")
	}

	// The membership ID defaults to metadata.name and the location to
	// "global" -- identical to the Terraform module.
	membershipId := spec.MembershipId
	if membershipId == "" {
		membershipId = locals.GcpGkeFleetMembership.Metadata.Name
	}
	location := spec.Location
	if location == "" {
		location = "global"
	}

	args := &gkehub.MembershipArgs{
		MembershipId: pulumi.String(membershipId),
		Location:     pulumi.StringPtr(location),
		Labels:       pulumi.ToStringMap(locals.GcpLabels),
	}
	if project != "" {
		args.Project = pulumi.StringPtr(project)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}
	if cluster := spec.GetGkeCluster().GetValue(); cluster != "" {
		args.Endpoint = &gkehub.MembershipEndpointArgs{
			GkeCluster: &gkehub.MembershipEndpointGkeClusterArgs{
				ResourceLink: pulumi.String(gkeClusterResourceLink(cluster)),
			},
		}
	}
	// Fleet Workload Identity: Google trusts this issuer's OIDC tokens
	// within the fleet's workload identity pool.
	if spec.Issuer != "" {
		args.Authority = &gkehub.MembershipAuthorityArgs{Issuer: pulumi.String(spec.Issuer)}
	}

	created, err := gkehub.NewMembership(ctx, locals.GcpGkeFleetMembership.Metadata.Name, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdGkehubApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create fleet membership")
	}

	ctx.Export(OpName, created.Name)
	ctx.Export(OpMembershipId, created.MembershipId)
	ctx.Export(OpLocation, created.Location)
	return nil
}

// gkeClusterResourceLink returns Google's resource link for a GKE cluster
// ("//container.googleapis.com/projects/{p}/locations/{l}/clusters/{n}")
// from the cluster's ID, so both engines send the string Google stores --
// the Terraform module's locals.tf twin.
func gkeClusterResourceLink(cluster string) string {
	if strings.HasPrefix(cluster, "//") {
		return cluster
	}
	return "//container.googleapis.com/" + strings.TrimPrefix(cluster, "/")
}
