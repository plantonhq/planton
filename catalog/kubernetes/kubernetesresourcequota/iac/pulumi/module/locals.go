package module

import (
	"strconv"

	kubernetesresourcequotav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesresourcequota/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across the module.
type Locals struct {
	Context     *pulumi.Context
	Spec        *kubernetesresourcequotav1alpha1.KubernetesResourceQuotaSpec
	Namespace   string
	Name        string
	Labels      map[string]string
	Annotations map[string]string

	// The companion LimitRange's name — empty when the spec sets no
	// limit_defaults (no LimitRange is created). It shares the quota's name:
	// one governance pair, one identity.
	LimitRangeName string
}

// initializeLocals extracts and transforms spec fields into module-local values.
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetesresourcequotav1alpha1.KubernetesResourceQuotaIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to what
	// the Terraform module stamps for the same manifest. User labels merge in
	// afterwards and cannot override the identity keys.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesResourceQuota.String(),
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
	for k, v := range spec.GetLabels() {
		if _, isIdentityKey := labels[k]; !isIdentityKey {
			labels[k] = v
		}
	}

	annotations := make(map[string]string)
	for k, v := range spec.GetAnnotations() {
		annotations[k] = v
	}

	// namespace is a StringValueOrRef foreign key. References are resolved to
	// literal strings before the module runs, so GetValue() returns the final
	// namespace name. When omitted entirely, fall back to the cluster's
	// "default" namespace — the same behavior as kubectl without a namespace flag.
	namespace := spec.GetNamespace().GetValue()
	if namespace == "" {
		namespace = "default"
	}

	limitRangeName := ""
	if len(spec.GetLimitDefaults()) > 0 {
		limitRangeName = spec.GetName()
	}

	return &Locals{
		Context:        ctx,
		Spec:           spec,
		Namespace:      namespace,
		Name:           spec.GetName(),
		Labels:         labels,
		Annotations:    annotations,
		LimitRangeName: limitRangeName,
	}
}

// scopeApiString maps the proto scope enum to the Kubernetes API string.
func scopeApiString(s kubernetesresourcequotav1alpha1.KubernetesResourceQuotaSpec_KubernetesResourceQuotaScope) string {
	switch s {
	case kubernetesresourcequotav1alpha1.KubernetesResourceQuotaSpec_terminating:
		return "Terminating"
	case kubernetesresourcequotav1alpha1.KubernetesResourceQuotaSpec_not_terminating:
		return "NotTerminating"
	case kubernetesresourcequotav1alpha1.KubernetesResourceQuotaSpec_best_effort:
		return "BestEffort"
	case kubernetesresourcequotav1alpha1.KubernetesResourceQuotaSpec_not_best_effort:
		return "NotBestEffort"
	case kubernetesresourcequotav1alpha1.KubernetesResourceQuotaSpec_priority_class:
		return "PriorityClass"
	case kubernetesresourcequotav1alpha1.KubernetesResourceQuotaSpec_cross_namespace_pod_affinity:
		return "CrossNamespacePodAffinity"
	case kubernetesresourcequotav1alpha1.KubernetesResourceQuotaSpec_volume_attributes_class:
		return "VolumeAttributesClass"
	default:
		return ""
	}
}

// limitTypeApiString maps the proto limit type enum to the Kubernetes API string.
func limitTypeApiString(t kubernetesresourcequotav1alpha1.KubernetesResourceQuotaLimitDefaults_KubernetesResourceQuotaLimitType) string {
	switch t {
	case kubernetesresourcequotav1alpha1.KubernetesResourceQuotaLimitDefaults_pod:
		return "Pod"
	case kubernetesresourcequotav1alpha1.KubernetesResourceQuotaLimitDefaults_persistent_volume_claim:
		return "PersistentVolumeClaim"
	default:
		return "Container"
	}
}
