package module

import (
	"fmt"
	"strconv"

	kubernetes "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	kuberneteskafkav1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kuberneteskafka/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	Namespace       string
	KubernetesKafka *kuberneteskafkav1.KubernetesKafka
	Labels          map[string]string

	IngressCertClusterIssuerName string
	// bootstrap
	IngressExternalBootstrapHostname string
	IngressInternalBootstrapHostname string
	IngressExternalBrokerHostnames   []string
	IngressInternalBrokerHostnames   []string
	IngressHostnames                 []string
	BootstrapKubeServiceFqdn         string
	BootstrapKubeServiceName         string

	// Prefixed resource names to avoid conflicts when multiple Kafka instances share a namespace
	KafkaClusterName              string
	KafkaIngressCertName          string
	KafkaIngressCertSecretName    string
	AdminUsername                 string
	AdminPasswordSecretName       string
	SchemaRegistryDeploymentName  string
	SchemaRegistryKubeServiceName string
	KowlConfigMapName             string
	KowlDeploymentName            string
	KowlKubeServiceName           string

	// schema registry
	IngressSchemaRegistryCertSecretName   string
	IngressExternalSchemaRegistryHostname string
	IngressInternalSchemaRegistryHostname string
	IngressSchemaRegistryHostnames        []string
	SchemaRegistryKubeServiceFqdn         string

	// kowl dashboard
	IngressKowlCertSecretName                              string
	IngressExternalKowlHostname                            string
	KowlKubeServiceFqdn                                    string
	KafkaIngressPrivateListenerLoadBalancerAnnotationKey   string
	KafkaIngressPrivateListenerLoadBalancerAnnotationValue string
	KafkaIngressPublicListenerLoadBalancerAnnotationKey    string
	KafkaIngressPublicListenerLoadBalancerAnnotationValue  string
}

func initializeLocals(ctx *pulumi.Context, stackInput *kuberneteskafkav1.KubernetesKafkaStackInput) *Locals {
	locals := &Locals{}

	locals.KubernetesKafka = stackInput.Target

	target := stackInput.Target

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_KubernetesKafka.String(),
	}

	if target.Metadata.Id != "" {
		locals.Labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.Labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.Labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	// get namespace from spec, it is required field
	locals.Namespace = target.Spec.Namespace.GetValue()

	// export namespace as an output
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))

	// Prefixed resource names to avoid conflicts when multiple Kafka instances share a namespace
	locals.KafkaClusterName = target.Metadata.Name
	locals.KafkaIngressCertName = fmt.Sprintf("%s-kafka-ingress", target.Metadata.Name)
	locals.KafkaIngressCertSecretName = fmt.Sprintf("cert-%s-kafka-ingress", target.Metadata.Name)
	locals.AdminUsername = fmt.Sprintf("%s-admin", target.Metadata.Name)
	locals.AdminPasswordSecretName = fmt.Sprintf("%s-admin", target.Metadata.Name)
	locals.SchemaRegistryDeploymentName = fmt.Sprintf("%s-schema-registry", target.Metadata.Name)
	locals.SchemaRegistryKubeServiceName = fmt.Sprintf("%s-sr", target.Metadata.Name)
	locals.KowlConfigMapName = fmt.Sprintf("%s-kowl", target.Metadata.Name)
	locals.KowlDeploymentName = fmt.Sprintf("%s-kowl", target.Metadata.Name)
	locals.KowlKubeServiceName = fmt.Sprintf("%s-kowl", target.Metadata.Name)

	ctx.Export(OpUsername, pulumi.String(locals.AdminUsername))
	ctx.Export(OpPasswordSecretName, pulumi.String(locals.AdminPasswordSecretName))
	ctx.Export(OpPasswordSecretKey, pulumi.String(vars.SaslPasswordKeyInSecret))

	locals.BootstrapKubeServiceName = fmt.Sprintf("%s-kafka-bootstrap", target.Metadata.Name)

	locals.BootstrapKubeServiceFqdn = fmt.Sprintf("%s.%s.svc", locals.BootstrapKubeServiceName, locals.Namespace)

	if target.Spec.Ingress == nil ||
		!target.Spec.Ingress.Enabled ||
		target.Spec.Ingress.Hostname == "" {
		return locals
	}

	// schema registry related locals data
	if locals.KubernetesKafka.Spec.SchemaRegistryContainer != nil &&
		locals.KubernetesKafka.Spec.SchemaRegistryContainer.IsEnabled {

		locals.IngressSchemaRegistryCertSecretName = fmt.Sprintf("cert-%s-schema-registry", target.Metadata.Name)

		locals.IngressExternalSchemaRegistryHostname = fmt.Sprintf("schema-registry-%s", target.Spec.Ingress.Hostname)

		locals.IngressInternalSchemaRegistryHostname = fmt.Sprintf("internal-schema-registry-%s", target.Spec.Ingress.Hostname)

		ctx.Export(OpSchemaRegistryExternalUrl, pulumi.Sprintf("https://%s", locals.IngressExternalSchemaRegistryHostname))
		ctx.Export(OpSchemaRegistryInternalUrl, pulumi.Sprintf("https://%s", locals.IngressInternalSchemaRegistryHostname))

		locals.IngressSchemaRegistryHostnames = []string{
			locals.IngressExternalSchemaRegistryHostname,
			locals.IngressInternalSchemaRegistryHostname,
		}
		locals.SchemaRegistryKubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.SchemaRegistryKubeServiceName, locals.Namespace)
	}

	// kowl related locals data
	if locals.KubernetesKafka.Spec.IsDeployKafkaUi {

		locals.IngressKowlCertSecretName = fmt.Sprintf("cert-%s-kowl", target.Metadata.Name)

		locals.IngressExternalKowlHostname = fmt.Sprintf("ui-%s", target.Spec.Ingress.Hostname)

		ctx.Export(OpKafkaUiExternalUrl, pulumi.Sprintf("https://%s", locals.IngressExternalKowlHostname))

		locals.KowlKubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KowlKubeServiceName, locals.Namespace)
	}

	locals.IngressExternalBootstrapHostname = target.Spec.Ingress.Hostname

	locals.IngressInternalBootstrapHostname = fmt.Sprintf("internal-%s", target.Spec.Ingress.Hostname)

	ctx.Export(OpBootstrapServerExternalHostname, pulumi.String(locals.IngressExternalBootstrapHostname))
	ctx.Export(OpBootstrapServerInternalHostname, pulumi.String(locals.IngressInternalBootstrapHostname))

	// Creating internal broker hostnames
	ingressInternalBrokerHostnames := make([]string, int(target.Spec.BrokerContainer.Replicas))
	for i := 0; i < int(target.Spec.BrokerContainer.Replicas); i++ {
		ingressInternalBrokerHostnames[i] = fmt.Sprintf("internal-broker-%d-%s", i, target.Spec.Ingress.Hostname)
	}
	locals.IngressInternalBrokerHostnames = ingressInternalBrokerHostnames

	// Creating external broker hostnames
	ingressExternalBrokerHostnames := make([]string, int(target.Spec.BrokerContainer.Replicas))
	for i := 0; i < int(target.Spec.BrokerContainer.Replicas); i++ {
		ingressExternalBrokerHostnames[i] = fmt.Sprintf("broker-%d-%s", i, target.Spec.Ingress.Hostname)
	}
	locals.IngressExternalBrokerHostnames = ingressExternalBrokerHostnames

	var ingressHostnames = []string{
		locals.IngressInternalBootstrapHostname,
		locals.IngressExternalBootstrapHostname,
	}

	ingressHostnames = append(ingressHostnames, locals.IngressInternalBrokerHostnames...)
	ingressHostnames = append(ingressHostnames, locals.IngressExternalBrokerHostnames...)
	locals.IngressHostnames = ingressHostnames

	//export ingress hostnames
	//ctx.Export(IngressExternalHostname, pulumi.String(locals.IngressExternalHostname))
	//ctx.Export(IngressInternalHostname, pulumi.String(locals.IngressInternalHostname))

	//note: a ClusterIssuer resource should have already exist on the kubernetes-cluster.
	//this is typically taken care of by the kubernetes cluster administrator.
	//if the kubernetes-cluster is created using Planton, then the cluster-issuer name will be
	//same as the ingress-domain-name as long as the same ingress-domain-name is added to the list of
	//ingress-domain-names for the GkeCluster/EksCluster/AksCluster spec.
	// Extract the domain from hostname for certificate issuer name
	dnsDomain := extractDomainFromHostname(target.Spec.Ingress.Hostname)
	locals.IngressCertClusterIssuerName = dnsDomain

	switch stackInput.ProviderConfig.Provider {
	case kubernetes.KubernetesProvider_gcp_gke:
		locals.KafkaIngressPrivateListenerLoadBalancerAnnotationKey = "cloud.google.com/load-balancer-type"
		locals.KafkaIngressPrivateListenerLoadBalancerAnnotationValue = "Internal"
		locals.KafkaIngressPublicListenerLoadBalancerAnnotationKey = "cloud.google.com/load-balancer-type"
		locals.KafkaIngressPublicListenerLoadBalancerAnnotationValue = "External"
	}

	return locals
}

// extractDomainFromHostname extracts the domain from a hostname
// Example: "kafka.example.com" -> "example.com"
func extractDomainFromHostname(hostname string) string {
	// Split by dots and take everything after the first part
	// This is a simple implementation - assumes standard domain structure
	parts := []rune(hostname)
	firstDotIndex := -1
	for i, char := range parts {
		if char == '.' {
			firstDotIndex = i
			break
		}
	}
	if firstDotIndex > 0 && firstDotIndex < len(hostname)-1 {
		return hostname[firstDotIndex+1:]
	}
	// If no dot found or dot is at the end, return the hostname as-is
	return hostname
}
