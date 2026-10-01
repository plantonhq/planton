package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/gkehub"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// fleet declares the project's one fleet ("default" in "global") with its
// display name and the defaults every cluster in it inherits. Fleet labels
// and the compliance posture default are not sent: the pinned SDK lacks
// both, so neither engine sends them.
func fleet(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpGkeFleet.Spec

	// An empty project means the provider's default project -- the
	// Terraform module's google_client_config twin. The fleet exports its
	// project as the value every fleet child references, so it is always
	// concrete.
	project := strings.TrimPrefix(spec.GetProjectId().GetValue(), "projects/")
	if project == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to resolve the provider's default project for the fleet")
		}
		if clientConfig.Project == "" {
			return errors.New("the fleet names no project and the provider has no default project -- set spec.project_id or configure a project")
		}
		project = clientConfig.Project
	}

	// The Fleet API (GKE Hub). DisableOnDestroy stays false: tearing down
	// the fleet's settings must never disable the API for the clusters and
	// features that still use it.
	createdGkehubApi, err := projects.NewService(ctx, "gcpfleet-gkehub.googleapis.com", &projects.ServiceArgs{
		Project:                  pulumi.String(project),
		Service:                  pulumi.String("gkehub.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable gkehub.googleapis.com")
	}

	args := &gkehub.FleetArgs{
		Project: pulumi.StringPtr(project),
	}
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	if defaults := spec.DefaultClusterConfig; defaults != nil &&
		(defaults.BinaryAuthorizationConfig != nil || defaults.SecurityPostureConfig != nil) {
		defaultArgs := &gkehub.FleetDefaultClusterConfigArgs{}
		if binauthz := defaults.BinaryAuthorizationConfig; binauthz != nil {
			binauthzArgs := &gkehub.FleetDefaultClusterConfigBinaryAuthorizationConfigArgs{}
			if binauthz.EvaluationMode != "" {
				binauthzArgs.EvaluationMode = pulumi.StringPtr(binauthz.EvaluationMode)
			}
			policyBindings := gkehub.FleetDefaultClusterConfigBinaryAuthorizationConfigPolicyBindingArray{}
			for _, name := range binauthz.PolicyBindings {
				policyBindings = append(policyBindings, &gkehub.FleetDefaultClusterConfigBinaryAuthorizationConfigPolicyBindingArgs{
					Name: pulumi.StringPtr(name),
				})
			}
			if len(policyBindings) > 0 {
				binauthzArgs.PolicyBindings = policyBindings
			}
			defaultArgs.BinaryAuthorizationConfig = binauthzArgs
		}
		if posture := defaults.SecurityPostureConfig; posture != nil {
			postureArgs := &gkehub.FleetDefaultClusterConfigSecurityPostureConfigArgs{}
			if posture.Mode != "" {
				postureArgs.Mode = pulumi.StringPtr(posture.Mode)
			}
			if posture.VulnerabilityMode != "" {
				postureArgs.VulnerabilityMode = pulumi.StringPtr(posture.VulnerabilityMode)
			}
			defaultArgs.SecurityPostureConfig = postureArgs
		}
		args.DefaultClusterConfig = defaultArgs
	}

	created, err := gkehub.NewFleet(ctx, locals.GcpGkeFleet.Metadata.Name, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdGkehubApi}))
	if err != nil {
		return errors.Wrap(err, "failed to create fleet")
	}

	ctx.Export(OpProjectId, created.Project)
	ctx.Export(OpName, created.ID())
	ctx.Export(OpUid, created.Uid)
	return nil
}
