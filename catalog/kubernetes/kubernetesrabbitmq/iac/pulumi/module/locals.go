package module

import (
	"fmt"
	"strconv"

	kubernetesrabbitmqv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesrabbitmq/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	KubernetesRabbitMq *kubernetesrabbitmqv1alpha1.KubernetesRabbitMq
	Spec               *kubernetesrabbitmqv1alpha1.KubernetesRabbitMqSpec

	// Resource-identity labels stamped on the module-created objects
	// (namespace, RabbitmqCluster). The operator derives ITS objects'
	// identity from the cluster name; these labels tie the family back
	// to the Planton resource.
	Labels map[string]string

	// Namespace the cluster lives in (resolved literal from the spec's
	// value-or-ref).
	Namespace string

	// ClusterName is metadata.name — the naming root the operator
	// derives every object from: the client Service `<name>`, the
	// headless Service `<name>-nodes`, the credentials Secret
	// `<name>-default-user`, the StatefulSet `<name>-server`.
	ClusterName string

	// Replicas is the resolved node count (spec default 1).
	Replicas int

	// TlsEnabled / PlainListenersClosed resolve the TLS posture: TLS is
	// on when a certificate Secret is referenced; the plain listeners
	// close only when disable_non_tls_listeners additionally asks.
	TlsEnabled           bool
	PlainListenersClosed bool

	// ServiceName / HeadlessServiceName follow the operator naming
	// contract (`<name>` / `<name>-nodes`).
	ServiceName         string
	HeadlessServiceName string

	// AmqpEndpoint / ManagementEndpoint are the in-cluster client
	// endpoints on the effective ports (TLS ports when the plain
	// listeners are closed).
	AmqpEndpoint       string
	ManagementEndpoint string

	// DefaultUserSecretName is the operator-generated credentials
	// Secret `<name>-default-user` — empty when the Vault secret
	// backend owns the credentials instead.
	DefaultUserSecretName string

	PortForwardCommand string
}

func initializeLocals(_ *pulumi.Context, iacInput *kubernetesrabbitmqv1alpha1.KubernetesRabbitMqIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesRabbitMq.String(),
	}
	if target.Metadata.Id != "" {
		labels[kuberneteslabelkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		labels[kuberneteslabelkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		labels[kuberneteslabelkeys.Environment] = target.Metadata.Env
	}

	namespace := spec.Namespace.GetValue()
	clusterName := target.Metadata.Name

	replicas := 1
	if spec.Replicas != nil && spec.GetReplicas() > 0 {
		replicas = int(spec.GetReplicas())
	}

	tlsEnabled := spec.GetTls().GetSecretName().GetValue() != ""
	plainListenersClosed := tlsEnabled && spec.GetTls().GetDisableNonTlsListeners()

	amqpPort := vars.AmqpPort
	managementPort := vars.ManagementPort
	managementScheme := "http"
	if plainListenersClosed {
		amqpPort = vars.AmqpTlsPort
		managementPort = vars.ManagementTlsPort
		managementScheme = "https"
	}

	// The Vault backend replaces the operator-generated Kubernetes
	// Secret entirely — credentials live at the Vault path instead.
	defaultUserSecretName := clusterName + vars.DefaultUserSecretSuffix
	if spec.GetSecretBackend().GetVault() != nil {
		defaultUserSecretName = ""
	}

	serviceName := clusterName
	headlessServiceName := clusterName + vars.HeadlessServiceSuffix

	return &Locals{
		KubernetesRabbitMq:   target,
		Spec:                 spec,
		Labels:               labels,
		Namespace:            namespace,
		ClusterName:          clusterName,
		Replicas:             replicas,
		TlsEnabled:           tlsEnabled,
		PlainListenersClosed: plainListenersClosed,
		ServiceName:          serviceName,
		HeadlessServiceName:  headlessServiceName,
		AmqpEndpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%d",
			serviceName, namespace, amqpPort),
		ManagementEndpoint: fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d",
			managementScheme, serviceName, namespace, managementPort),
		DefaultUserSecretName: defaultUserSecretName,
		PortForwardCommand: fmt.Sprintf("kubectl port-forward svc/%s -n %s %d:%d",
			serviceName, namespace, managementPort, managementPort),
	}
}
