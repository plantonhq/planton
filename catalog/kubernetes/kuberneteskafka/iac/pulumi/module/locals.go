package module

import (
	"fmt"
	"strconv"

	kuberneteskafkav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafka/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteskafkav1alpha1.KubernetesKafkaSpec

	// Resource-identity labels stamped on every module-created object
	// (the namespace, the Kafka and KafkaNodePool CRs, the metrics
	// ConfigMap). Strimzi's own strimzi.io/cluster binding label is
	// added per node pool on top of these.
	Labels map[string]string

	// Namespace the cluster deploys into (resolved literal from the
	// spec's value-or-ref). KafkaTopic/KafkaUser declarations for this
	// cluster must live HERE — the entity operators watch only this
	// namespace.
	Namespace string

	// ClusterName is metadata.name — the Kafka CR name and the value of
	// the strimzi.io/cluster label on every dependent resource.
	ClusterName string

	// BootstrapServiceName is the operator's naming contract for the
	// internal bootstrap Service: <cluster>-kafka-bootstrap.
	BootstrapServiceName string

	// InternalBootstrapEndpoint is the in-cluster bootstrap address for
	// the FIRST internal-type listener (empty when the spec declares no
	// internal listener) — the value workloads put in bootstrap.servers.
	InternalBootstrapEndpoint string

	// ClusterCaCertSecretName is the operator's naming contract for the
	// cluster CA certificate Secret: <cluster>-cluster-ca-cert (key
	// ca.crt) — what TLS clients trust.
	ClusterCaCertSecretName string

	// MetricsConfigMapName names the module-owned JMX exporter rules
	// ConfigMap rendered when spec.metrics.enabled:
	// <cluster>-kafka-metrics.
	MetricsConfigMapName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskafkav1alpha1.KubernetesKafkaIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKafka.String(),
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

	clusterName := target.Metadata.Name
	namespace := spec.Namespace.GetValue()
	bootstrapServiceName := fmt.Sprintf("%s-kafka-bootstrap", clusterName)

	// The first internal listener (explicit "internal" or unset type,
	// which defaults to internal) supplies the in-cluster bootstrap
	// endpoint. Clusters exposing ONLY external listeners export an
	// empty endpoint — an honest signal that in-cluster clients have no
	// plain path.
	internalBootstrapEndpoint := ""
	for _, listener := range spec.GetListeners() {
		listenerType := listener.GetType()
		if listenerType == "" || listenerType == "internal" {
			internalBootstrapEndpoint = fmt.Sprintf("%s.%s.svc.cluster.local:%d",
				bootstrapServiceName, namespace, listener.GetPort())
			break
		}
	}

	return &Locals{
		Spec:                      spec,
		Labels:                    labels,
		Namespace:                 namespace,
		ClusterName:               clusterName,
		BootstrapServiceName:      bootstrapServiceName,
		InternalBootstrapEndpoint: internalBootstrapEndpoint,
		ClusterCaCertSecretName:   fmt.Sprintf("%s-cluster-ca-cert", clusterName),
		MetricsConfigMapName:      fmt.Sprintf("%s-kafka-metrics", clusterName),
	}
}
