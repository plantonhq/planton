package module

import (
	"fmt"
	"strconv"

	kuberneteskafkamirrormaker2v1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkamirrormaker2/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across
// the module. Every resolution here has an exact twin in the Terraform
// module's locals.tf — keep them in lockstep.
type Locals struct {
	Spec *kuberneteskafkamirrormaker2v1alpha1.KubernetesKafkaMirrorMaker2Spec

	// Resource-identity labels stamped on every module-created object
	// (the namespace, the KafkaMirrorMaker2 CR, the metrics ConfigMap).
	Labels map[string]string

	// Namespace the engine deploys into (resolved literal from the
	// spec's value-or-ref). Must be watched by a Strimzi operator
	// installation.
	Namespace string

	// MirrorMakerName is metadata.name — the KafkaMirrorMaker2 CR name
	// and the stem of every operator naming contract below.
	MirrorMakerName string

	// TargetAlias identifies the target cluster in the replication flow
	// (spec.target.alias, "target" when unset). Mirrored topic names are
	// prefixed with the SOURCE alias under the default replication
	// policy — the target alias only names the cluster in connector
	// configs and metrics.
	TargetAlias string

	// GroupId is the engine's Connect-protocol group ID on the target
	// (spec.target.group_id, metadata.name when unset). MUST be unique
	// among Connect-protocol workloads sharing the target cluster —
	// the same contract as the storage topics below.
	GroupId string

	// The three engine-state topics on the target cluster
	// (spec.target.*_storage_topic, <name>-mirrormaker2-* when unset).
	ConfigStorageTopic string
	StatusStorageTopic string
	OffsetStorageTopic string

	// MetricsConfigMapName names the module-owned JMX exporter rules
	// ConfigMap rendered when spec.metrics.enabled: <name>-mm2-metrics.
	MetricsConfigMapName string

	// RestApiEndpoint is the operator's naming contract for the engine's
	// Connect REST API Service:
	// http://<name>-mirrormaker2-api.<namespace>.svc.cluster.local:8083.
	RestApiEndpoint string
}

// initializeLocals extracts and transforms spec fields into module-local
// values.
func initializeLocals(_ *pulumi.Context, iacInput *kuberneteskafkamirrormaker2v1alpha1.KubernetesKafkaMirrorMaker2IacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKafkaMirrorMaker2.String(),
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

	mirrorMakerName := target.Metadata.Name
	namespace := spec.Namespace.GetValue()

	// The group-identity fallbacks (twins of TF's coalesce chain): the
	// spec defaults the alias to "target" and derives the group ID and
	// the three storage topics from metadata.name — one engine identity
	// per resource by construction.
	targetAlias := spec.GetTarget().GetAlias()
	if targetAlias == "" {
		targetAlias = "target"
	}
	groupId := spec.GetTarget().GetGroupId()
	if groupId == "" {
		groupId = mirrorMakerName
	}
	configStorageTopic := spec.GetTarget().GetConfigStorageTopic()
	if configStorageTopic == "" {
		configStorageTopic = fmt.Sprintf("%s-mirrormaker2-configs", mirrorMakerName)
	}
	statusStorageTopic := spec.GetTarget().GetStatusStorageTopic()
	if statusStorageTopic == "" {
		statusStorageTopic = fmt.Sprintf("%s-mirrormaker2-status", mirrorMakerName)
	}
	offsetStorageTopic := spec.GetTarget().GetOffsetStorageTopic()
	if offsetStorageTopic == "" {
		offsetStorageTopic = fmt.Sprintf("%s-mirrormaker2-offsets", mirrorMakerName)
	}

	return &Locals{
		Spec:                 spec,
		Labels:               labels,
		Namespace:            namespace,
		MirrorMakerName:      mirrorMakerName,
		TargetAlias:          targetAlias,
		GroupId:              groupId,
		ConfigStorageTopic:   configStorageTopic,
		StatusStorageTopic:   statusStorageTopic,
		OffsetStorageTopic:   offsetStorageTopic,
		MetricsConfigMapName: fmt.Sprintf("%s-mm2-metrics", mirrorMakerName),
		RestApiEndpoint: fmt.Sprintf("http://%s-mirrormaker2-api.%s.svc.cluster.local:8083",
			mirrorMakerName, namespace),
	}
}
