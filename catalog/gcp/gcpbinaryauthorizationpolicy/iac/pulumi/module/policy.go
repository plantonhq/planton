package module

import (
	"strings"

	"github.com/pkg/errors"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/binaryauthorization"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/organizations"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// policy applies the project's one Binary Authorization policy. Every
// create and update is a full PUT (this block replaces whatever policy the
// project had); destroy under DELETE writes Google's default policy back --
// allow every image.
func policy(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpBinaryAuthorizationPolicy.Spec

	// An empty project means the provider's default project -- the
	// Terraform module's google_client_config twin.
	project := strings.TrimPrefix(spec.GetProjectId().GetValue(), "projects/")
	if project == "" {
		clientConfig, err := organizations.GetClientConfig(ctx, pulumi.Provider(gcpProvider))
		if err != nil {
			return errors.Wrap(err, "failed to resolve the provider's default project for the policy")
		}
		if clientConfig.Project == "" {
			return errors.New("the policy names no project and the provider has no default project -- set spec.project_id or configure a project")
		}
		project = clientConfig.Project
	}

	// disable_on_destroy is false: destroying the policy writes Google's
	// default policy back, and GKE clusters keep evaluating it through the
	// API.
	api, err := projects.NewService(ctx, "gcpbapol-binaryauthorization.googleapis.com", &projects.ServiceArgs{
		Project:                  pulumi.String(project),
		Service:                  pulumi.String("binaryauthorization.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
		DisableOnDestroy:         pulumi.BoolPtr(false),
	}, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable binaryauthorization.googleapis.com api")
	}

	def := spec.DefaultAdmissionRule
	args := &binaryauthorization.PolicyArgs{
		Project: pulumi.StringPtr(project),
		DefaultAdmissionRule: &binaryauthorization.PolicyDefaultAdmissionRuleArgs{
			EvaluationMode:          pulumi.String(def.EvaluationMode),
			EnforcementMode:         pulumi.String(def.EnforcementMode),
			RequireAttestationsBies: attestors(def.RequireAttestationsBy),
		},
	}
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}
	if spec.GlobalPolicyEvaluationMode != "" {
		args.GlobalPolicyEvaluationMode = pulumi.StringPtr(spec.GlobalPolicyEvaluationMode)
	}
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}
	// The spec lifts Google's one-field pattern blocks to strings.
	if len(spec.AdmissionWhitelistPatterns) > 0 {
		patterns := binaryauthorization.PolicyAdmissionWhitelistPatternArray{}
		for _, pattern := range spec.AdmissionWhitelistPatterns {
			patterns = append(patterns, &binaryauthorization.PolicyAdmissionWhitelistPatternArgs{NamePattern: pulumi.String(pattern)})
		}
		args.AdmissionWhitelistPatterns = patterns
	}
	if len(spec.ClusterAdmissionRules) > 0 {
		rules := binaryauthorization.PolicyClusterAdmissionRuleArray{}
		for _, rule := range spec.ClusterAdmissionRules {
			rules = append(rules, &binaryauthorization.PolicyClusterAdmissionRuleArgs{
				Cluster:                 pulumi.String(rule.Cluster),
				EvaluationMode:          pulumi.String(rule.EvaluationMode),
				EnforcementMode:         pulumi.String(rule.EnforcementMode),
				RequireAttestationsBies: attestors(rule.RequireAttestationsBy),
			})
		}
		args.ClusterAdmissionRules = rules
	}

	created, err := binaryauthorization.NewPolicy(ctx, locals.GcpBinaryAuthorizationPolicy.Metadata.Name, args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{api}))
	if err != nil {
		return errors.Wrap(err, "failed to apply binary authorization policy")
	}

	ctx.Export(OpName, created.ID().ApplyT(func(id pulumi.ID) string { return string(id) + "/policy" }).(pulumi.StringOutput))
	ctx.Export(OpProjectId, created.Project)
	return nil
}

// attestors renders a rule's attestor list, or nil when it is empty so the
// field is not sent (Google requires it empty unless the rule requires
// attestation).
func attestors(refs []*foreignkeyv1.StringValueOrRef) pulumi.StringArrayInput {
	if len(refs) == 0 {
		return nil
	}
	values := pulumi.StringArray{}
	for _, ref := range refs {
		values = append(values, pulumi.String(ref.GetValue()))
	}
	return values
}
