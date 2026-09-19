package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// peeringRoutesConfig is the ROUTES-CONFIG form: the route exchange of a
// peering that already exists on `network` under peering_name -- typically
// the `servicenetworking-googleapis-com` peering Google creates for Cloud
// SQL private IP and other private-services-access products.
//
// The provider REQUIRES both custom-route flags here, so they are always
// sent. The two public-IP subnet-route flags are always sent as well, with
// the spec's defaults (true / false) when unset -- the manifest loader
// applies the proto defaults before either engine runs, so the manifest
// states the route exchange in full and a re-plan is clean. Destroying this
// resource is a no-op in GCP -- the peering keeps whatever flags it last
// had -- so there is no deletion_policy argument to send.
func peeringRoutesConfig(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpVpcPeering.Spec

	args := &compute.NetworkPeeringRoutesConfigArgs{
		Peering:            pulumi.String(locals.PeeringName),
		Network:            pulumi.String(networkNameFromSelfLink(spec.Network.GetValue())),
		ExportCustomRoutes: pulumi.Bool(spec.ExportCustomRoutes),
		ImportCustomRoutes: pulumi.Bool(spec.ImportCustomRoutes),
	}
	args.ExportSubnetRoutesWithPublicIp = pulumi.BoolPtr(exportSubnetRoutesWithPublicIp(spec.ExportSubnetRoutesWithPublicIp))
	args.ImportSubnetRoutesWithPublicIp = pulumi.BoolPtr(spec.GetImportSubnetRoutesWithPublicIp())
	// The routes-config resource is project-scoped (unlike the peering,
	// which is addressed by the network's self link), so the project is
	// taken from the network's self link when it names one; otherwise the
	// provider's default project applies.
	if project := projectFromSelfLink(spec.Network.GetValue()); project != "" {
		args.Project = pulumi.String(project)
	}

	created, err := compute.NewNetworkPeeringRoutesConfig(ctx, locals.GcpVpcPeering.Metadata.Name, args,
		pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to configure network peering routes")
	}

	ctx.Export(OpPeeringName, created.Peering)
	ctx.Export(OpNetwork, pulumi.String(spec.Network.GetValue()))
	// The routes-config resource does not own the peering entry and reports
	// no state; both engines export empty strings so the outputs message
	// has the same shape in either form.
	ctx.Export(OpState, pulumi.String(""))
	ctx.Export(OpStateDetails, pulumi.String(""))

	return nil
}

// networkNameFromSelfLink returns the last path segment of a network self
// link (`.../global/networks/{name}` -> `{name}`); a bare name is returned
// as is. The routes-config resource accepts either form, but the Terraform
// module sends the bare name, so both engines send the same string.
func networkNameFromSelfLink(network string) string {
	if idx := strings.LastIndex(network, "/"); idx >= 0 {
		return network[idx+1:]
	}
	return network
}

// projectFromSelfLink extracts `{project}` from
// `[https://.../]projects/{project}/global/networks/{name}`; empty when
// the value carries no project segment.
func projectFromSelfLink(network string) string {
	parts := strings.Split(network, "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "projects" {
			return parts[i+1]
		}
	}
	return ""
}
