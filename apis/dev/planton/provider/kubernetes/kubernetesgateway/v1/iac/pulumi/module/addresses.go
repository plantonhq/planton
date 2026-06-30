package module

import (
	kubernetesgatewayv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesgateway/v1"
	gatewayv1 "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func buildAddresses(addresses []*kubernetesgatewayv1.KubernetesGatewayAddress) gatewayv1.GatewaySpecAddressesArray {
	arr := gatewayv1.GatewaySpecAddressesArray{}
	for _, a := range addresses {
		args := gatewayv1.GatewaySpecAddressesArgs{}
		if t := a.GetType(); t != "" {
			args.Type = pulumi.String(t)
		}
		if v := a.GetValue(); v != "" {
			args.Value = pulumi.String(v)
		}
		arr = append(arr, args)
	}
	return arr
}
