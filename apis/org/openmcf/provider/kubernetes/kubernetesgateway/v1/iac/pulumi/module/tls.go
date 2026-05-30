package module

import (
	kubernetesapis "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	kubernetesgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesgateway/v1"
	gatewayv1 "github.com/plantonhq/openmcf/pkg/kubernetes/kubernetestypes/gatewayapis/kubernetes/gateway/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// buildGatewayTls maps the gateway-wide TLS configuration: backend
// (Gateway-as-client) certificate material and frontend (inbound) client
// certificate validation. The CRD generates separate Go types for the default
// validation and the per-port validation, so each is mapped by a dedicated
// builder below.
func buildGatewayTls(tls *kubernetesgatewayv1.KubernetesGatewayTlsConfig) gatewayv1.GatewaySpecTlsArgs {
	args := gatewayv1.GatewaySpecTlsArgs{}
	if backend := tls.GetBackend(); backend != nil {
		args.Backend = buildBackendTls(backend)
	}
	if frontend := tls.GetFrontend(); frontend != nil {
		args.Frontend = buildFrontendTls(frontend)
	}
	return args
}

func buildBackendTls(backend *kubernetesgatewayv1.KubernetesGatewayBackendTls) gatewayv1.GatewaySpecTlsBackendArgs {
	args := gatewayv1.GatewaySpecTlsBackendArgs{}
	if ref := backend.GetClientCertificateRef(); ref != nil {
		clientCertRef := gatewayv1.GatewaySpecTlsBackendClientCertificateRefArgs{
			Name: pulumi.String(ref.GetName()),
		}
		if group := ref.GetGroup(); group != "" {
			clientCertRef.Group = pulumi.String(group)
		}
		if kind := ref.GetKind(); kind != "" {
			clientCertRef.Kind = pulumi.String(kind)
		}
		if namespace := ref.GetNamespace(); namespace != "" {
			clientCertRef.Namespace = pulumi.String(namespace)
		}
		args.ClientCertificateRef = clientCertRef
	}
	return args
}

func buildFrontendTls(frontend *kubernetesgatewayv1.KubernetesGatewayFrontendTlsConfig) gatewayv1.GatewaySpecTlsFrontendArgs {
	args := gatewayv1.GatewaySpecTlsFrontendArgs{}
	if def := frontend.GetDefault(); def != nil {
		args.Default = buildFrontendDefault(def)
	}
	if perPort := frontend.GetPerPort(); len(perPort) > 0 {
		perPortArr := gatewayv1.GatewaySpecTlsFrontendPerPortArray{}
		for _, p := range perPort {
			perPortArgs := gatewayv1.GatewaySpecTlsFrontendPerPortArgs{
				Port: pulumi.Int(int(p.GetPort())),
			}
			if tls := p.GetTls(); tls != nil {
				perPortArgs.Tls = buildPerPortTls(tls)
			}
			perPortArr = append(perPortArr, perPortArgs)
		}
		args.PerPort = perPortArr
	}
	return args
}

func buildFrontendDefault(def *kubernetesgatewayv1.KubernetesGatewayFrontendTlsValidationConfig) gatewayv1.GatewaySpecTlsFrontendDefaultArgs {
	args := gatewayv1.GatewaySpecTlsFrontendDefaultArgs{}
	validation := def.GetValidation()
	if validation == nil {
		return args
	}
	validationArgs := gatewayv1.GatewaySpecTlsFrontendDefaultValidationArgs{}
	if mode := validation.GetMode(); mode != "" {
		validationArgs.Mode = pulumi.String(mode)
	}
	if refs := validation.GetCaCertificateRefs(); len(refs) > 0 {
		caRefs := gatewayv1.GatewaySpecTlsFrontendDefaultValidationCaCertificateRefsArray{}
		for _, ref := range refs {
			caRefs = append(caRefs, buildDefaultCaCertificateRef(ref))
		}
		validationArgs.CaCertificateRefs = caRefs
	}
	args.Validation = validationArgs
	return args
}

func buildPerPortTls(tls *kubernetesgatewayv1.KubernetesGatewayFrontendTlsValidationConfig) gatewayv1.GatewaySpecTlsFrontendPerPortTlsArgs {
	args := gatewayv1.GatewaySpecTlsFrontendPerPortTlsArgs{}
	validation := tls.GetValidation()
	if validation == nil {
		return args
	}
	validationArgs := gatewayv1.GatewaySpecTlsFrontendPerPortTlsValidationArgs{}
	if mode := validation.GetMode(); mode != "" {
		validationArgs.Mode = pulumi.String(mode)
	}
	if refs := validation.GetCaCertificateRefs(); len(refs) > 0 {
		caRefs := gatewayv1.GatewaySpecTlsFrontendPerPortTlsValidationCaCertificateRefsArray{}
		for _, ref := range refs {
			caRefs = append(caRefs, buildPerPortCaCertificateRef(ref))
		}
		validationArgs.CaCertificateRefs = caRefs
	}
	args.Validation = validationArgs
	return args
}

func buildDefaultCaCertificateRef(ref *kubernetesapis.KubernetesGatewayApiObjectReference) gatewayv1.GatewaySpecTlsFrontendDefaultValidationCaCertificateRefsArgs {
	args := gatewayv1.GatewaySpecTlsFrontendDefaultValidationCaCertificateRefsArgs{
		Group: pulumi.String(ref.GetGroup()),
		Kind:  pulumi.String(ref.GetKind()),
		Name:  pulumi.String(ref.GetName()),
	}
	if namespace := ref.GetNamespace(); namespace != "" {
		args.Namespace = pulumi.String(namespace)
	}
	return args
}

func buildPerPortCaCertificateRef(ref *kubernetesapis.KubernetesGatewayApiObjectReference) gatewayv1.GatewaySpecTlsFrontendPerPortTlsValidationCaCertificateRefsArgs {
	args := gatewayv1.GatewaySpecTlsFrontendPerPortTlsValidationCaCertificateRefsArgs{
		Group: pulumi.String(ref.GetGroup()),
		Kind:  pulumi.String(ref.GetKind()),
		Name:  pulumi.String(ref.GetName()),
	}
	if namespace := ref.GetNamespace(); namespace != "" {
		args.Namespace = pulumi.String(namespace)
	}
	return args
}
