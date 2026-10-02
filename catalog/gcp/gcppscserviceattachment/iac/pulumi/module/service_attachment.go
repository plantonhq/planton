package module

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// serviceAttachment builds the producer half of Private Service Connect:
// one service attachment in front of the producer's internal load balancer
// (target_service, a regional forwarding rule), translating consumer
// traffic into the PSC NAT subnets. The connection policy changes in
// place; the name, region, and domain names recreate the attachment.
func serviceAttachment(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpPscServiceAttachment.Spec

	natSubnets := make(pulumi.StringArray, 0, len(spec.NatSubnets))
	for _, s := range spec.NatSubnets {
		natSubnets = append(natSubnets, pulumi.String(s.GetValue()))
	}

	args := &compute.ServiceAttachmentArgs{
		Name:                 pulumi.String(locals.AttachmentName),
		Region:               pulumi.String(spec.Region),
		TargetService:        pulumi.String(spec.TargetService.GetValue()),
		NatSubnets:           natSubnets,
		ConnectionPreference: pulumi.String(spec.ConnectionPreference),
		// Required by the API: the manifest states it either way (proto
		// default false is a real answer).
		EnableProxyProtocol: pulumi.Bool(spec.EnableProxyProtocol),
	}

	// Empty project falls back to the provider's default project — the same
	// ambient contract the Terraform module honors.
	if locals.ProjectId != "" {
		args.Project = pulumi.StringPtr(locals.ProjectId)
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// The accept list: each consumer names exactly one of a project, a
	// network, or an endpoint (spec CEL); unset arms are left nil.
	if len(spec.ConsumerAcceptLists) > 0 {
		var accept compute.ServiceAttachmentConsumerAcceptListArray
		for _, c := range spec.ConsumerAcceptLists {
			entry := &compute.ServiceAttachmentConsumerAcceptListArgs{
				ConnectionLimit: pulumi.Int(int(c.ConnectionLimit)),
			}
			if v := c.ProjectId.GetValue(); v != "" {
				entry.ProjectIdOrNum = pulumi.StringPtr(v)
			}
			if v := c.Network.GetValue(); v != "" {
				entry.NetworkUrl = pulumi.StringPtr(v)
			}
			if c.EndpointUrl != "" {
				entry.EndpointUrl = pulumi.StringPtr(c.EndpointUrl)
			}
			accept = append(accept, entry)
		}
		args.ConsumerAcceptLists = accept
	}

	if len(spec.ConsumerRejectLists) > 0 {
		reject := make(pulumi.StringArray, 0, len(spec.ConsumerRejectLists))
		for _, p := range spec.ConsumerRejectLists {
			reject = append(reject, pulumi.String(p.GetValue()))
		}
		args.ConsumerRejectLists = reject
	}

	// Optional+Computed on the provider: sent only when the spec sets it, so
	// Google's default (false) is never fought.
	if spec.ReconcileConnections != nil {
		args.ReconcileConnections = pulumi.BoolPtr(spec.GetReconcileConnections())
	}

	if len(spec.DomainNames) > 0 {
		args.DomainNames = pulumi.ToStringArray(spec.DomainNames)
	}

	// propagated_connection_limit is tri-state in the spec: unset lets
	// Google apply its default of 250; an explicit 0 must reach the API as
	// 0, which the provider only sends when its send-if-zero twin is true --
	// derived here from the spec value, never a field of its own (PARITY
	// with the Terraform module).
	if spec.PropagatedConnectionLimit != nil {
		args.PropagatedConnectionLimit = pulumi.IntPtr(int(spec.GetPropagatedConnectionLimit()))
		if spec.GetPropagatedConnectionLimit() == 0 {
			args.SendPropagatedConnectionLimitIfZero = pulumi.BoolPtr(true)
		}
	}

	// Google's API currently ignores the flag; the module still sends the
	// stated intent so it takes effect the day the API honors it.
	if spec.ShowNatIps {
		args.ShowNatIps = pulumi.BoolPtr(true)
	}

	// What destroy does to a published service: DELETE (default), PREVENT
	// (refuse), or ABANDON (drop from state, keep serving).
	if spec.DeletionPolicy != "" {
		args.DeletionPolicy = pulumi.StringPtr(spec.DeletionPolicy)
	}

	created, err := compute.NewServiceAttachment(ctx, "service-attachment", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create service attachment")
	}

	ctx.Export(OpSelfLink, created.SelfLink)
	ctx.Export(OpAttachmentName, created.Name)
	ctx.Export(OpRegion, pulumi.String(spec.Region))
	ctx.Export(OpFingerprint, created.Fingerprint)
	ctx.Export(OpConnectedEndpointsCount, created.ConnectedEndpoints.ApplyT(func(eps []compute.ServiceAttachmentConnectedEndpoint) string {
		return strconv.Itoa(len(eps))
	}).(pulumi.StringOutput))

	return nil
}
