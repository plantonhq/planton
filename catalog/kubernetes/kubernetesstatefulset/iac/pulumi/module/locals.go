package module

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kubernetesstatefulsetv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstatefulset/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/workloadpod"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	KubernetesStatefulSet *kubernetesstatefulsetv1alpha1.KubernetesStatefulSet
	Namespace             string

	// SelectorLabels are the immutable pod-selection identity of this workload.
	// They key off metadata.name so multiple stateful sets coexist in one namespace
	// without cross-selecting each other's pods. Selectors are immutable on
	// apps/v1 StatefulSets, so nothing mutable may ever join this set.
	SelectorLabels map[string]string

	// Labels are the full governance label set stamped on every created object
	// (selector labels + Planton resource tracking).
	Labels map[string]string

	KubeServiceName        string
	KubeServiceFqdn        string
	KubePortForwardCommand string

	// Computed satellite resource names, prefixed with metadata.name so multiple
	// instances sharing a namespace never collide.
	EnvSecretName       string
	ImagePullSecretName string
	ImagePullSecretData map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *kubernetesstatefulsetv1alpha1.KubernetesStatefulSetIacInput) (*Locals, error) {
	locals := &Locals{}

	locals.KubernetesStatefulSet = iacInput.Target
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
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesStatefulSet.String(),
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
	ctx.Export(OpStatefulSetName, pulumi.String(target.Metadata.Name))

	// Selector labels exported as a deterministic "k=v,k=v" string so downstream
	// resources (NetworkPolicies, sibling workloads' anti-affinity, kubectl -l) can
	// consume them without re-deriving the labeling convention.
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

	// The headless governing Service always exists — StatefulSets require one for
	// per-pod DNS regardless of whether the app exposes ports — so every
	// service-derived output is unconditionally populated.
	locals.KubeServiceName = target.Metadata.Name
	ctx.Export(OpService, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)
	ctx.Export(OpKubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:8080",
		locals.Namespace, locals.KubeServiceName)
	ctx.Export(OpPortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	// Per-replica DNS template: {ordinal} is a literal placeholder the consumer
	// substitutes with the replica index (e.g. "0") to address a specific member —
	// this is how clustered clients build their member lists.
	ctx.Export(OpPodDnsTemplate, pulumi.String(
		fmt.Sprintf("%s-{ordinal}.%s.%s.svc.cluster.local",
			target.Metadata.Name, locals.KubeServiceName, locals.Namespace)))

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
