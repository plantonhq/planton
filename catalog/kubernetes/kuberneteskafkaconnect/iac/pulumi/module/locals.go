package module

import (
	"fmt"
	"strconv"

	kuberneteskafkaconnectv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkaconnect/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteskafkaconnectv1alpha1.KubernetesKafkaConnectSpec

	// Resource-identity labels stamped on every module-created object
	// (the namespace, the KafkaConnect CR, the metrics ConfigMap).
	Labels map[string]string

	// Namespace the Connect cluster deploys into (resolved literal from
	// the spec's value-or-ref). KafkaConnector declarations binding to
	// this cluster must live HERE — the operator matches them by
	// namespace + the strimzi.io/cluster label.
	Namespace string

	// ConnectName is metadata.name — the KafkaConnect CR name and the
	// value KubernetesKafkaConnector resources bind to (rendered as
	// their strimzi.io/cluster label).
	ConnectName string

	// GroupId is the Connect group identity the workers share. Defaults
	// from metadata.name — MUST be unique among Connect clusters
	// (including MirrorMaker 2 instances) sharing a Kafka cluster; two
	// clusters sharing a group.id corrupt each other's state.
	GroupId string

	// The three Connect-internal storage topics. Same uniqueness
	// contract as GroupId — the defaults derive from metadata.name so
	// two same-named clusters never collide by accident.
	ConfigStorageTopic string
	StatusStorageTopic string
	OffsetStorageTopic string

	// RestApiServiceName is the operator's naming contract for the
	// Connect REST API Service: <cluster>-connect-api.
	RestApiServiceName string

	// RestApiEndpoint is the in-cluster Connect REST API address (port
	// 8083) — read-only inspection only; connector management stays
	// declarative through KubernetesKafkaConnector.
	RestApiEndpoint string

	// MetricsConfigMapName names the module-owned JMX exporter rules
	// ConfigMap rendered when spec.metrics.enabled:
	// <cluster>-connect-metrics.
	MetricsConfigMapName string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskafkaconnectv1alpha1.KubernetesKafkaConnectIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKafkaConnect.String(),
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

	connectName := target.Metadata.Name
	namespace := spec.Namespace.GetValue()

	// The group identity quartet defaults from metadata.name (the spec's
	// documented contract) so distinct Connect clusters get distinct
	// identities without the author having to invent four names.
	groupId := spec.GetGroupId()
	if groupId == "" {
		groupId = connectName
	}
	configStorageTopic := spec.GetConfigStorageTopic()
	if configStorageTopic == "" {
		configStorageTopic = fmt.Sprintf("%s-connect-configs", connectName)
	}
	statusStorageTopic := spec.GetStatusStorageTopic()
	if statusStorageTopic == "" {
		statusStorageTopic = fmt.Sprintf("%s-connect-status", connectName)
	}
	offsetStorageTopic := spec.GetOffsetStorageTopic()
	if offsetStorageTopic == "" {
		offsetStorageTopic = fmt.Sprintf("%s-connect-offsets", connectName)
	}

	restApiServiceName := fmt.Sprintf("%s-connect-api", connectName)

	return &Locals{
		Spec:               spec,
		Labels:             labels,
		Namespace:          namespace,
		ConnectName:        connectName,
		GroupId:            groupId,
		ConfigStorageTopic: configStorageTopic,
		StatusStorageTopic: statusStorageTopic,
		OffsetStorageTopic: offsetStorageTopic,
		RestApiServiceName: restApiServiceName,
		RestApiEndpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:8083",
			restApiServiceName, namespace),
		MetricsConfigMapName: fmt.Sprintf("%s-connect-metrics", connectName),
	}
}
