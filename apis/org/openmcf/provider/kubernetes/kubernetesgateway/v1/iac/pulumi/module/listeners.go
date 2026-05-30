package module

import (
	kubernetesapis "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	kubernetesgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesgateway/v1"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildListeners maps the OpenMCF listeners onto the typed crd2pulumi listener
// args array. Optional fields are only set when present so upstream/controller
// defaults flow through unchanged.
func buildListeners(listeners []*kubernetesgatewayv1.KubernetesGatewayListener) gatewayv1.GatewaySpecListenersArray {
	arr := gatewayv1.GatewaySpecListenersArray{}
	for _, l := range listeners {
		args := gatewayv1.GatewaySpecListenersArgs{
			Name:     pulumi.String(l.GetName()),
			Port:     pulumi.Int(int(l.GetPort())),
			Protocol: pulumi.String(l.GetProtocol()),
		}
		if hostname := l.GetHostname(); hostname != "" {
			args.Hostname = pulumi.String(hostname)
		}
		if tls := l.GetTls(); tls != nil {
			args.Tls = buildListenerTls(tls)
		}
		if allowedRoutes := l.GetAllowedRoutes(); allowedRoutes != nil {
			args.AllowedRoutes = buildAllowedRoutes(allowedRoutes)
		}
		arr = append(arr, args)
	}
	return arr
}

func buildListenerTls(tls *kubernetesgatewayv1.KubernetesGatewayListenerTlsConfig) gatewayv1.GatewaySpecListenersTlsArgs {
	args := gatewayv1.GatewaySpecListenersTlsArgs{}
	if mode := tls.GetMode(); mode != "" {
		args.Mode = pulumi.String(mode)
	}
	if refs := tls.GetCertificateRefs(); len(refs) > 0 {
		certRefs := gatewayv1.GatewaySpecListenersTlsCertificateRefsArray{}
		for _, ref := range refs {
			certRefs = append(certRefs, buildListenerCertificateRef(ref))
		}
		args.CertificateRefs = certRefs
	}
	if options := tls.GetOptions(); len(options) > 0 {
		args.Options = pulumi.ToStringMap(options)
	}
	return args
}

func buildListenerCertificateRef(ref *kubernetesapis.KubernetesGatewayApiSecretObjectReference) gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs {
	args := gatewayv1.GatewaySpecListenersTlsCertificateRefsArgs{
		Name: pulumi.String(ref.GetName()),
	}
	if group := ref.GetGroup(); group != "" {
		args.Group = pulumi.String(group)
	}
	if kind := ref.GetKind(); kind != "" {
		args.Kind = pulumi.String(kind)
	}
	if namespace := ref.GetNamespace(); namespace != "" {
		args.Namespace = pulumi.String(namespace)
	}
	return args
}

func buildAllowedRoutes(allowedRoutes *kubernetesgatewayv1.KubernetesGatewayAllowedRoutes) gatewayv1.GatewaySpecListenersAllowedRoutesArgs {
	args := gatewayv1.GatewaySpecListenersAllowedRoutesArgs{}
	if namespaces := allowedRoutes.GetNamespaces(); namespaces != nil {
		nsArgs := gatewayv1.GatewaySpecListenersAllowedRoutesNamespacesArgs{}
		if from := namespaces.GetFrom(); from != "" {
			nsArgs.From = pulumi.String(from)
		}
		if selector := namespaces.GetSelector(); selector != nil {
			nsArgs.Selector = buildAllowedRoutesSelector(selector)
		}
		args.Namespaces = nsArgs
	}
	if kinds := allowedRoutes.GetKinds(); len(kinds) > 0 {
		kindArr := gatewayv1.GatewaySpecListenersAllowedRoutesKindsArray{}
		for _, k := range kinds {
			kindArgs := gatewayv1.GatewaySpecListenersAllowedRoutesKindsArgs{
				Kind: pulumi.String(k.GetKind()),
			}
			if group := k.GetGroup(); group != "" {
				kindArgs.Group = pulumi.String(group)
			}
			kindArr = append(kindArr, kindArgs)
		}
		args.Kinds = kindArr
	}
	return args
}
