package module

import (
	"strconv"
	"strings"

	kubernetesnetworkpolicyv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesnetworkpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds computed values derived from the IaC input for use across the module.
type Locals struct {
	Context     *pulumi.Context
	Spec        *kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicySpec
	Namespace   string
	Name        string
	Labels      map[string]string
	Annotations map[string]string

	// The directions this policy governs, as the Kubernetes API strings
	// ("Ingress"/"Egress"). Explicit when the spec sets policy_types, otherwise
	// inferred with the API server's own rule: ingress always, egress only when
	// egress rules exist — computed here so the exported output reflects the
	// deployed truth even when the spec omitted the field.
	PolicyTypes []string
}

// initializeLocals extracts and transforms spec fields into module-local values.
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicyIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	// Resource-identity labels: the kuberneteslabelkeys set, identical to what
	// the Terraform module stamps for the same manifest. User labels merge in
	// afterwards and cannot override the identity keys.
	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesNetworkPolicy.String(),
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

	return &Locals{
		Context:     ctx,
		Spec:        spec,
		Namespace:   namespace,
		Name:        spec.GetName(),
		Labels:      labels,
		Annotations: annotations,
		PolicyTypes: resolvePolicyTypes(spec),
	}
}

// resolvePolicyTypes returns the API string forms of the governed directions,
// applying the API server's inference when the spec omits policy_types.
func resolvePolicyTypes(spec *kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicySpec) []string {
	if len(spec.GetPolicyTypes()) > 0 {
		types := make([]string, 0, len(spec.GetPolicyTypes()))
		for _, t := range spec.GetPolicyTypes() {
			if t == kubernetesnetworkpolicyv1alpha1.KubernetesNetworkPolicySpec_egress {
				types = append(types, "Egress")
			} else {
				types = append(types, "Ingress")
			}
		}
		return types
	}
	// The API server's inference: ingress always; egress only with egress rules.
	types := []string{"Ingress"}
	if len(spec.GetEgressRules()) > 0 {
		types = append(types, "Egress")
	}
	return types
}

// policyTypesString renders the governed directions for the outputs contract
// ("Ingress", "Egress", or "Ingress,Egress").
func policyTypesString(types []string) string {
	return strings.Join(types, ",")
}
