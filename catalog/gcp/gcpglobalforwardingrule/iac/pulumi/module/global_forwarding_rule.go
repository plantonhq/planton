package module

import (
	"strconv"

	"github.com/pkg/errors"
	gcpglobalforwardingrulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpglobalforwardingrule/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// globalForwardingRule provisions the global Compute Engine forwarding rule
// — the VIP node where traffic enters a global load balancer (or, with the
// PSC form, where a VPC's private path to Google APIs / a producer service
// begins). It binds an IP address and port to a target proxy; everything
// behind it is wiring.
//
// target and labels update in place (GCP repoints the target via a
// dedicated setTarget call — the zero-downtime frontend swap); every other
// field is immutable and forces destroy-and-recreate. The VIP itself
// survives recreation only when ip_address references a reserved static
// address, which is why production frontends reserve one.
func globalForwardingRule(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpGlobalForwardingRule.Spec

	// Enable the Compute Engine API first so a fresh project works on the
	// first deploy. disable_on_destroy stays false: tearing down one rule
	// must never disable the API for everything else in the project.
	serviceArgs := &projects.ServiceArgs{
		Service:                  pulumi.String("compute.googleapis.com"),
		DisableDependentServices: pulumi.BoolPtr(true),
	}
	if spec.ProjectId.GetValue() != "" {
		serviceArgs.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	createdProjectService, err := projects.NewService(ctx,
		"globalforwardingrule-compute.googleapis.com", serviceArgs, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to enable compute.googleapis.com api")
	}

	args := &compute.GlobalForwardingRuleArgs{
		Name: pulumi.String(locals.ForwardingRuleName),
		// The target ref arrives resolved to a literal: a proxy self-link,
		// a PSC bundle name (all-apis / vpc-sc), or a service attachment URI.
		Target: pulumi.String(spec.Target.GetValue()),
	}

	// An empty project falls back to the provider's default project — the
	// ambient-project contract every GCP kind honors.
	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	// Omitted → GCP assigns an ephemeral IP. The ref default resolves a
	// GcpGlobalAddress to its literal IP (the API reads back the IP number,
	// so passing the number keeps state drift-free).
	if spec.IpAddress.GetValue() != "" {
		args.IpAddress = pulumi.String(spec.IpAddress.GetValue())
	}
	// The middleware default (TCP) matches GCP's own default, so an unset
	// value can simply be omitted — the API computes TCP either way.
	if spec.IpProtocol != nil && spec.GetIpProtocol() != "" {
		args.IpProtocol = pulumi.String(spec.GetIpProtocol())
	}
	if spec.IpVersion != "" {
		args.IpVersion = pulumi.String(spec.IpVersion)
	}
	// The PSC form (spec NONE) must SEND the empty scheme explicitly. An
	// empty spec value becomes EXTERNAL, never an omission: the spec's
	// default is EXTERNAL (the classic global external ALB), the provider's
	// own default is EXTERNAL_MANAGED, and the scheme is immutable -- letting
	// the provider decide would replace every existing classic frontend the
	// next time it was applied. The manifest defaults applier normally fills
	// EXTERNAL first; this guard covers every path that bypasses it. The
	// Terraform module makes the same choice.
	if locals.IsPrivateServiceConnect {
		args.LoadBalancingScheme = pulumi.String("")
	} else if locals.LoadBalancingScheme != "" {
		args.LoadBalancingScheme = pulumi.String(locals.LoadBalancingScheme)
	} else {
		args.LoadBalancingScheme = pulumi.String("EXTERNAL")
	}
	if spec.PortRange != "" {
		args.PortRange = pulumi.String(spec.PortRange)
	}
	if spec.Network.GetValue() != "" {
		args.Network = pulumi.String(spec.Network.GetValue())
	}
	if spec.Subnetwork.GetValue() != "" {
		args.Subnetwork = pulumi.String(spec.Subnetwork.GetValue())
	}
	// Global rules are PREMIUM-only (spec CEL enforces it); sending the
	// value only when set keeps the API's computed default in charge.
	if spec.NetworkTier != "" {
		args.NetworkTier = pulumi.String(spec.NetworkTier)
	}
	if len(spec.MetadataFilters) > 0 {
		args.MetadataFilters = buildMetadataFilters(spec.MetadataFilters)
	}
	if spec.ServiceDirectoryRegistration != nil {
		args.ServiceDirectoryRegistrations = buildServiceDirectoryRegistration(spec.ServiceDirectoryRegistration)
	}
	// Only meaningful for PSC; the API default (auto-create the DNS zone)
	// applies unless explicitly disabled.
	if spec.NoAutomateDnsZone {
		args.NoAutomateDnsZone = pulumi.Bool(true)
	}
	if len(spec.Labels) > 0 {
		labels := pulumi.StringMap{}
		for key, value := range spec.Labels {
			labels[key] = pulumi.String(value)
		}
		args.Labels = labels
	}
	if spec.ExternalManagedBackendBucketMigrationState != "" {
		args.ExternalManagedBackendBucketMigrationState = pulumi.String(spec.ExternalManagedBackendBucketMigrationState)
	}
	if spec.ExternalManagedBackendBucketMigrationTestingPercentage != 0 {
		args.ExternalManagedBackendBucketMigrationTestingPercentage = pulumi.Float64(spec.ExternalManagedBackendBucketMigrationTestingPercentage)
	}

	// What destroy does to the frontend: DELETE (default), PREVENT (refuse),
	// or ABANDON (drop from state, keep serving traffic).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdRule, err := compute.NewGlobalForwardingRule(ctx, "global-forwarding-rule", args,
		pulumi.Provider(gcpProvider), pulumi.DependsOn([]pulumi.Resource{createdProjectService}))
	if err != nil {
		return errors.Wrap(err, "failed to create global forwarding rule")
	}

	ctx.Export(OpIpAddress, createdRule.IpAddress)
	ctx.Export(OpSelfLink, createdRule.SelfLink)
	ctx.Export(OpForwardingRuleName, createdRule.Name)
	ctx.Export(OpForwardingRuleId, createdRule.ForwardingRuleId.ApplyT(func(id int) string {
		return strconv.Itoa(id)
	}).(pulumi.StringOutput))
	ctx.Export(OpPscConnectionId, createdRule.PscConnectionId)
	ctx.Export(OpPscConnectionStatus, createdRule.PscConnectionStatus)

	return nil
}

func buildMetadataFilters(filters []*gcpglobalforwardingrulev1alpha1.GcpGlobalForwardingRuleMetadataFilter) compute.GlobalForwardingRuleMetadataFilterArray {
	result := compute.GlobalForwardingRuleMetadataFilterArray{}
	for _, filter := range filters {
		labels := compute.GlobalForwardingRuleMetadataFilterFilterLabelArray{}
		for _, label := range filter.FilterLabels {
			labels = append(labels, &compute.GlobalForwardingRuleMetadataFilterFilterLabelArgs{
				Name:  pulumi.String(label.Name),
				Value: pulumi.String(label.Value),
			})
		}
		result = append(result, &compute.GlobalForwardingRuleMetadataFilterArgs{
			FilterMatchCriteria: pulumi.String(filter.FilterMatchCriteria),
			FilterLabels:        labels,
		})
	}
	return result
}

func buildServiceDirectoryRegistration(registration *gcpglobalforwardingrulev1alpha1.GcpGlobalForwardingRuleServiceDirectoryRegistration) *compute.GlobalForwardingRuleServiceDirectoryRegistrationsArgs {
	args := &compute.GlobalForwardingRuleServiceDirectoryRegistrationsArgs{}
	if registration.Namespace != "" {
		args.Namespace = pulumi.String(registration.Namespace)
	}
	if registration.ServiceDirectoryRegion != "" {
		args.ServiceDirectoryRegion = pulumi.String(registration.ServiceDirectoryRegion)
	}
	return args
}
