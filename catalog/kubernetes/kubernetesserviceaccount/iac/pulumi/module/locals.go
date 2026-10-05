// Derived values for the KubernetesServiceAccount module: label/annotation merging,
// namespace defaulting, image-pull-secret resolution, and the workload-identity
// annotation translation shared verbatim (in behavior) with the Terraform module.
package module

import (
	"fmt"

	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesserviceaccountv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesserviceaccount/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ServiceAccount annotation keys each cloud's workload-identity webhook/agent watches.
// These exact strings are the contract between the ServiceAccount and the cloud-side
// trust configuration; they must stay in sync with the Terraform module.
const (
	gkeServiceAccountAnnotation = "iam.gke.io/gcp-service-account"
	eksRoleArnAnnotation        = "eks.amazonaws.com/role-arn"
	aksClientIdAnnotation       = "azure.workload.identity/client-id"
	aksTenantIdAnnotation       = "azure.workload.identity/tenant-id"
)

// Locals holds all derived configuration and state for the module
type Locals struct {
	// Context for Pulumi operations
	Ctx *pulumi.Context

	// IaC input containing the target resource
	IacInput *kubernetesserviceaccountv1alpha1.KubernetesServiceAccountIacInput

	// Target service account resource
	Target *kubernetesserviceaccountv1alpha1.KubernetesServiceAccount

	// Spec from the target
	Spec *kubernetesserviceaccountv1alpha1.KubernetesServiceAccountSpec

	// ServiceAccount name
	ServiceAccountName string

	// ServiceAccount namespace (defaults to "default" when unset)
	Namespace string

	// Combined labels (spec labels + standard labels)
	Labels map[string]string

	// Combined annotations (spec annotations + workload-identity annotations;
	// workload-identity annotations win on collision)
	Annotations map[string]string

	// Resolved image-pull secret names (from StringValueOrRef entries)
	ImagePullSecretNames []string

	// Tri-state token automount: nil means "unset" and the field is omitted from
	// the ServiceAccount entirely, deferring to the cluster default
	AutomountServiceAccountToken *bool

	// Fully-qualified RBAC subject: "system:serviceaccount:<namespace>:<name>"
	RbacSubject string

	// The configured cloud identity handle (email/ARN/client-id), or "" when
	// workload identity is not configured
	WorkloadIdentityHandle string
}

// initializeLocals creates and populates the Locals struct
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetesserviceaccountv1alpha1.KubernetesServiceAccountIacInput) *Locals {
	locals := &Locals{
		Ctx:      ctx,
		IacInput: iacInput,
		Target:   iacInput.Target,
		Spec:     iacInput.Target.Spec,
	}

	locals.ServiceAccountName = locals.Spec.Name

	// Namespace is a StringValueOrRef; references are resolved to literals before the
	// module runs, so GetValue() is the resolved name. Empty means the user omitted
	// it — fall back to "default", matching kubectl behavior without a namespace flag.
	locals.Namespace = locals.Spec.GetNamespace().GetValue()
	if locals.Namespace == "" {
		locals.Namespace = "default"
	}

	locals.Labels = buildLabels(locals)

	workloadIdentityAnnotations, workloadIdentityHandle := translateWorkloadIdentity(locals.Spec.WorkloadIdentity)
	locals.WorkloadIdentityHandle = workloadIdentityHandle
	locals.Annotations = buildAnnotations(locals, workloadIdentityAnnotations)

	// Each image-pull-secret entry is a StringValueOrRef resolved to the secret's
	// name in the same namespace. An entry that resolved to nothing is dropped
	// rather than rendered as `{name: ""}` — the API server rejects the empty
	// name, and a reference that did not resolve is refused upstream, not here.
	for _, secretRef := range locals.Spec.ImagePullSecrets {
		if name := secretRef.GetValue(); name != "" {
			locals.ImagePullSecretNames = append(locals.ImagePullSecretNames, name)
		}
	}

	// Preserve the tri-state as-is: nil (unset) must NOT collapse to false, because
	// the resource creation step omits the field entirely when nil.
	locals.AutomountServiceAccountToken = locals.Spec.AutomountServiceAccountToken

	// The exact string cloud trust configuration (IAM trust policies, federated
	// credentials) matches on — exported so downstream never re-assembles it.
	locals.RbacSubject = fmt.Sprintf("system:serviceaccount:%s:%s", locals.Namespace, locals.ServiceAccountName)

	return locals
}

// buildLabels combines spec labels with standard labels
func buildLabels(locals *Locals) map[string]string {
	labels := make(map[string]string)

	// Add standard labels
	labels["managed-by"] = "planton"
	labels["resource"] = locals.Target.Metadata.Name
	labels["resource-kind"] = "KubernetesServiceAccount"

	// Add spec labels
	for k, v := range locals.Spec.Labels {
		labels[k] = v
	}

	return labels
}

// buildAnnotations merges user annotations with workload-identity annotations.
// Workload-identity annotations are applied last so they win on collision: the
// typed workload_identity field is the authoritative expression of the cloud
// binding, and a stray user annotation must not silently override it.
func buildAnnotations(locals *Locals, workloadIdentityAnnotations map[string]string) map[string]string {
	annotations := make(map[string]string)

	for k, v := range locals.Spec.Annotations {
		annotations[k] = v
	}

	for k, v := range workloadIdentityAnnotations {
		annotations[k] = v
	}

	return annotations
}

// translateWorkloadIdentity converts the workload-identity oneof into the exact
// ServiceAccount annotations each cloud's webhook expects, and returns the bound
// identity handle for the outputs.
//
// The translation is intentionally the ONLY place cloud specifics appear: the
// ServiceAccount itself is cloud-agnostic, and the annotation is the entire
// Kubernetes-side half of the federation. The cloud-side half (IAM binding,
// trust policy, federated credential) is owned by the referenced cloud identity
// resource, not this module.
//
//   - GKE Workload Identity: `iam.gke.io/gcp-service-account` carries the GCP
//     service account email the pod's tokens impersonate.
//   - EKS IRSA: `eks.amazonaws.com/role-arn` carries the IAM role ARN the EKS
//     pod-identity webhook injects credentials for.
//   - Azure AD Workload Identity: `azure.workload.identity/client-id` carries the
//     managed-identity/Entra-app client ID; `azure.workload.identity/tenant-id` is
//     added only when explicitly set (cross-tenant scenarios) — when omitted, the
//     Azure webhook falls back to its default tenant.
//
// Returns an empty map and "" when workload identity is not configured.
func translateWorkloadIdentity(workloadIdentity *kubernetes.KubernetesWorkloadIdentity) (map[string]string, string) {
	annotations := make(map[string]string)

	switch {
	case workloadIdentity.GetGke() != nil:
		email := workloadIdentity.GetGke().GetServiceAccountEmail().GetValue()
		annotations[gkeServiceAccountAnnotation] = email
		return annotations, email

	case workloadIdentity.GetEks() != nil:
		roleArn := workloadIdentity.GetEks().GetRoleArn().GetValue()
		annotations[eksRoleArnAnnotation] = roleArn
		return annotations, roleArn

	case workloadIdentity.GetAks() != nil:
		aks := workloadIdentity.GetAks()
		clientId := aks.GetClientId().GetValue()
		annotations[aksClientIdAnnotation] = clientId
		// tenant_id is an optional-string pointer: only annotate when the user set it,
		// so the Azure webhook's default-tenant behavior is preserved otherwise.
		if aks.TenantId != nil {
			annotations[aksTenantIdAnnotation] = aks.GetTenantId()
		}
		return annotations, clientId

	default:
		return annotations, ""
	}
}
