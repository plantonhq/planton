package module

import (
	"strconv"

	"github.com/pkg/errors"
	gcpglobalforwardingrulev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpglobalforwardingrule/v1alpha1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// regionalForwardingRule builds the regional twin (spec.region set): the
// front door of the regional external and internal Application Load
// Balancers (target = a regional proxy), of the internal and external
// passthrough Network Load Balancers (backend_service, no proxy), and of a
// Private Service Connect consumer endpoint (target = a service attachment,
// empty scheme).
func regionalForwardingRule(ctx *pulumi.Context, locals *Locals, opts []pulumi.ResourceOption) error {
	spec := locals.GcpGlobalForwardingRule.Spec

	args := &compute.ForwardingRuleArgs{
		Name:   pulumi.String(locals.ForwardingRuleName),
		Region: pulumi.String(spec.Region),
	}

	// Exactly one of the two sinks is set (spec CEL): a regional proxy
	// self-link or a service attachment URI, or the regional backend service
	// of a passthrough NLB.
	if spec.Target.GetValue() != "" {
		args.Target = pulumi.String(spec.Target.GetValue())
	}
	if spec.BackendService.GetValue() != "" {
		args.BackendService = pulumi.String(spec.BackendService.GetValue())
	}

	if spec.ProjectId.GetValue() != "" {
		args.Project = pulumi.String(spec.ProjectId.GetValue())
	}
	if spec.Description != "" {
		args.Description = pulumi.String(spec.Description)
	}
	// A regional GcpAddress reference (or a literal); omitted → ephemeral IP.
	if spec.IpAddress.GetValue() != "" {
		args.IpAddress = pulumi.String(spec.IpAddress.GetValue())
	}
	// L3_DEFAULT (every protocol) is admitted here alone; the spec pairs it
	// with all_ports.
	if spec.IpProtocol != nil && spec.GetIpProtocol() != "" {
		args.IpProtocol = pulumi.String(spec.GetIpProtocol())
	}
	if spec.IpVersion != "" {
		args.IpVersion = pulumi.String(spec.IpVersion)
	}
	// The regional provider default is EXTERNAL too, so the explicit send
	// changes nothing here; it keeps one contract on both scopes.
	args.LoadBalancingScheme = loadBalancingScheme(locals)

	// One of three port forms (spec CEL); the unchosen ones stay unset so
	// the API's own exclusivity never sees two.
	if spec.PortRange != "" {
		args.PortRange = pulumi.String(spec.PortRange)
	}
	if len(spec.Ports) > 0 {
		args.Ports = pulumi.ToStringArray(spec.Ports)
	}
	if spec.AllPorts {
		args.AllPorts = pulumi.Bool(true)
	}

	// Required for the internal schemes and PSC, and for the regional
	// external ALB whose proxy-only subnet lives in the network.
	if spec.Network.GetValue() != "" {
		args.Network = pulumi.String(spec.Network.GetValue())
	}
	if spec.Subnetwork.GetValue() != "" {
		args.Subnetwork = pulumi.String(spec.Subnetwork.GetValue())
	}
	// PREMIUM or STANDARD; unset keeps the API's computed default (PREMIUM).
	if spec.NetworkTier != "" {
		args.NetworkTier = pulumi.String(spec.NetworkTier)
	}
	// Service Directory registration for a PSC consumer endpoint: namespace
	// and service (the region field is the global rule's, spec CEL).
	if spec.ServiceDirectoryRegistration != nil {
		args.ServiceDirectoryRegistrations = buildRegionalServiceDirectoryRegistration(spec.ServiceDirectoryRegistration)
	}
	if spec.NoAutomateDnsZone {
		args.NoAutomateDnsZone = pulumi.Bool(true)
	}
	if len(spec.Labels) > 0 {
		args.Labels = buildLabels(spec.Labels)
	}

	// The passthrough and PSC-consumer levers; each omitted when unset so
	// the API default stands and a re-plan stays clean. recreate_closed_psc
	// carries a provider default (false) and is sent explicitly.
	if spec.AllowGlobalAccess {
		args.AllowGlobalAccess = pulumi.Bool(true)
	}
	if spec.AllowPscGlobalAccess {
		args.AllowPscGlobalAccess = pulumi.Bool(true)
	}
	if spec.ServiceLabel != "" {
		args.ServiceLabel = pulumi.String(spec.ServiceLabel)
	}
	if spec.IsMirroringCollector {
		args.IsMirroringCollector = pulumi.Bool(true)
	}
	if spec.IpCollection != "" {
		args.IpCollection = pulumi.String(spec.IpCollection)
	}
	args.RecreateClosedPsc = pulumi.Bool(spec.RecreateClosedPsc)
	if len(spec.SourceIpRanges) > 0 {
		args.SourceIpRanges = pulumi.ToStringArray(spec.SourceIpRanges)
	}

	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	createdRule, err := compute.NewForwardingRule(ctx, "forwarding-rule", args, opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create regional forwarding rule")
	}

	ctx.Export(OpIpAddress, createdRule.IpAddress)
	ctx.Export(OpSelfLink, createdRule.SelfLink)
	ctx.Export(OpForwardingRuleName, createdRule.Name)
	ctx.Export(OpForwardingRuleId, createdRule.ForwardingRuleId.ApplyT(func(id int) string {
		return strconv.Itoa(id)
	}).(pulumi.StringOutput))
	ctx.Export(OpPscConnectionId, createdRule.PscConnectionId)
	ctx.Export(OpPscConnectionStatus, createdRule.PscConnectionStatus)
	ctx.Export(OpRegion, pulumi.String(spec.Region))
	ctx.Export(OpServiceName, createdRule.ServiceName)

	return nil
}

// buildRegionalServiceDirectoryRegistration renders the regional rule's
// registration: namespace and service.
func buildRegionalServiceDirectoryRegistration(registration *gcpglobalforwardingrulev1alpha1.GcpGlobalForwardingRuleServiceDirectoryRegistration) *compute.ForwardingRuleServiceDirectoryRegistrationsArgs {
	args := &compute.ForwardingRuleServiceDirectoryRegistrationsArgs{}
	if registration.Namespace != "" {
		args.Namespace = pulumi.String(registration.Namespace)
	}
	if registration.Service != "" {
		args.Service = pulumi.String(registration.Service)
	}
	return args
}
