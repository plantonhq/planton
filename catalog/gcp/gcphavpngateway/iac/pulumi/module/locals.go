package module

import (
	"strconv"
	"strings"

	gcphavpngatewayv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcphavpngateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// target plus the derivations both engines share -- the gateway and router
// names and the merged label set.
type Locals struct {
	GcpHaVpnGateway *gcphavpngatewayv1alpha1.GcpHaVpnGateway

	// GatewayName is the spec's gateway_name, or metadata.name when the
	// spec leaves it empty -- the same naming basis every kind uses.
	GatewayName string

	// RouterName is the spec's router.name, or the gateway's name when the
	// spec leaves it empty: a router and a gateway are different resource
	// types, so the shared name never collides.
	RouterName string

	// GcpLabels is the gateway's label set: user labels first, then the
	// platform attribution labels, which win on key conflicts -- identical
	// merge order to the Terraform module.
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *gcphavpngatewayv1alpha1.GcpHaVpnGatewayStackInput) *Locals {
	target := stackInput.Target

	gatewayName := target.Spec.GatewayName
	if gatewayName == "" {
		gatewayName = target.Metadata.Name
	}
	routerName := target.Spec.Router.GetName()
	if routerName == "" {
		routerName = gatewayName
	}

	labels := map[string]string{}
	for key, value := range target.Spec.Labels {
		labels[key] = value
	}
	labels[gcplabelkeys.Resource] = strconv.FormatBool(true)
	labels[gcplabelkeys.ResourceName] = gatewayName
	labels[gcplabelkeys.ResourceKind] = strings.ToLower(cloudresourcekind.CloudResourceKind_GcpHaVpnGateway.String())
	if target.Metadata.Org != "" {
		labels[gcplabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[gcplabelkeys.Environment] = target.Metadata.Env
	}
	if target.Metadata.Id != "" {
		labels[gcplabelkeys.ResourceId] = target.Metadata.Id
	}

	return &Locals{
		GcpHaVpnGateway: target,
		GatewayName:     gatewayName,
		RouterName:      routerName,
		GcpLabels:       labels,
	}
}
