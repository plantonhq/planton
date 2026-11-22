package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/certificateauthority"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// certificateManagerAgentGrant creates the Certificate Manager service agent
// (service-<project_number>@gcp-sa-certificatemanager.iam.gserviceaccount.com)
// and grants it roles/privateca.certificateRequester on the pool the spec
// names. Google checks that grant when the issuance config is created and
// refuses the config without it ("Config validation failed"), and it creates
// the agent only on request (services.generateServiceIdentity), never when
// the API is turned on -- a grant naming the agent before then is refused as
// an invalid member. Creating an agent that already exists returns it, and
// deleting the identity is a no-op, so the agent outlives this config. The
// returned grant is what the config depends on.
//
// The grant lands in the pool's own project and location, parsed from its
// full name, so a pool in another project works -- identical parse to the
// Terraform module. It is a member grant: two configs over one pool write
// the same binding, and destroying either removes it until the other is
// applied again.
func certificateManagerAgentGrant(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, projectService pulumi.Resource) (pulumi.Resource, error) {
	spec := locals.GcpCertManagerIssuanceConfig.Spec
	name := locals.GcpCertManagerIssuanceConfig.Metadata.Name

	identityArgs := &projects.ServiceIdentityArgs{
		Service: pulumi.String("certificatemanager.googleapis.com"),
	}
	if locals.ProjectId != "" {
		identityArgs.Project = pulumi.StringPtr(locals.ProjectId)
	}
	createdAgent, err := projects.NewServiceIdentity(ctx, name+"-certificatemanager-agent", identityArgs,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{projectService}))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the certificate manager service agent")
	}

	// projects/{project}/locations/{location}/caPools/{pool}, which the
	// spec's ca_pool.format rule guarantees for a literal and the
	// GcpPrivateCaPool name output always has.
	poolParts := strings.Split(spec.CaPool.GetValue(), "/")
	if len(poolParts) != 6 {
		return nil, errors.Errorf("ca_pool %q is not projects/{project}/locations/{location}/caPools/{pool}", spec.CaPool.GetValue())
	}

	createdGrant, err := certificateauthority.NewCaPoolIamMember(ctx, name+"-agent-requests-certificates", &certificateauthority.CaPoolIamMemberArgs{
		Project:  pulumi.StringPtr(poolParts[1]),
		Location: pulumi.StringPtr(poolParts[3]),
		CaPool:   pulumi.String(poolParts[5]),
		Role:     pulumi.String("roles/privateca.certificateRequester"),
		Member:   createdAgent.Member,
	}, pulumi.Provider(gcpProvider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to grant the certificate manager service agent on the ca pool")
	}
	return createdGrant, nil
}
