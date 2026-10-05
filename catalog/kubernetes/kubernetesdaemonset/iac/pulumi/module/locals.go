package module

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kubernetesdaemonsetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesdaemonset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadpod"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesDaemonSet *kubernetesdaemonsetv1alpha1.KubernetesDaemonSet
	Namespace           string

	// SelectorLabels are the immutable pod-selection identity of this workload.
	// They key off metadata.name so multiple daemon sets coexist in one namespace
	// without cross-selecting each other's pods. Selectors are immutable on
	// apps/v1 DaemonSets, so nothing mutable may ever join this set.
	SelectorLabels map[string]string

	// Labels are the full governance label set stamped on every created object
	// (selector labels + Planton resource tracking).
	Labels map[string]string

	// Computed satellite resource names, prefixed with metadata.name so multiple
	// instances sharing a namespace never collide.
	EnvSecretName       string
	ImagePullSecretName string
	ImagePullSecretData map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *kubernetesdaemonsetv1alpha1.KubernetesDaemonSetIacInput) (*Locals, error) {
	locals := &Locals{}

	locals.KubernetesDaemonSet = iacInput.Target
	target := iacInput.Target

	if target.Spec.Container == nil || target.Spec.Container.App == nil || target.Spec.Container.App.Image == nil {
		return nil, errors.New("spec.container.app.image is required")
	}

	locals.SelectorLabels = map[string]string{
		"app":                            target.Metadata.Name,
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
	}

	locals.Labels = map[string]string{
		"app":                            target.Metadata.Name,
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesDaemonSet.String(),
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

	locals.Namespace = target.Spec.Namespace.GetValue()
	ctx.Export(OpNamespace, pulumi.String(locals.Namespace))
	ctx.Export(OpDaemonSetName, pulumi.String(target.Metadata.Name))

	// Selector labels exported as a deterministic "k=v,k=v" string so downstream
	// resources (NetworkPolicies, sibling workloads' anti-affinity, kubectl -l) can
	// consume them without re-deriving the labeling convention. DaemonSets have no
	// Service, so this and the object identity are the whole composition surface.
	ctx.Export(OpSelectorLabels, pulumi.String(formatSelector(locals.SelectorLabels)))

	locals.EnvSecretName = fmt.Sprintf("%s-env-secrets", target.Metadata.Name)
	locals.ImagePullSecretName = fmt.Sprintf("%s-image-pull", target.Metadata.Name)

	// The image-pull Secret's data comes from the workload's OWN spec — the registry
	// logins declared on pod.image_registries — and from nowhere else. Nil when the
	// pod declares none: a public image, or a same-cloud registry the cluster's own
	// identity reaches, or a Secret declared beside the workload and named in
	// pod.image_pull_secrets, all need no Secret from this module.
	imagePullSecretData, err := workloadpod.BuildImagePullSecretData(target.Spec.Pod)
	if err != nil {
		return nil, err
	}
	locals.ImagePullSecretData = imagePullSecretData

	return locals, nil
}

// formatSelector renders selector labels as a deterministic, sorted "k=v,k=v"
// string — the exact syntax kubectl's -l flag and NetworkPolicy tooling accept.
func formatSelector(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, labels[k]))
	}
	return strings.Join(pairs, ",")
}
