package module

import (
	kubernetesrbacv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesrbac/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Kubernetes RBAC object kinds and the API group shared by every roleRef and
// user/group subject.
const (
	kindRole               = "Role"
	kindClusterRole        = "ClusterRole"
	kindRoleBinding        = "RoleBinding"
	kindClusterRoleBinding = "ClusterRoleBinding"
	rbacApiGroup           = "rbac.authorization.k8s.io"
)

// Locals holds all derived configuration for the module. Everything that both the
// role and the binding need (scope, names, kinds, labels) is resolved once here so
// the resource files contain no decision logic of their own.
type Locals struct {
	// Context for Pulumi operations
	Ctx *pulumi.Context

	// IaC input containing the target resource
	IacInput *kubernetesrbacv1alpha1.KubernetesRbacIacInput

	// Target RBAC grant resource
	Target *kubernetesrbacv1alpha1.KubernetesRbac

	// Spec from the target
	Spec *kubernetesrbacv1alpha1.KubernetesRbacSpec

	// Whether the grant is namespace-scoped (Role/RoleBinding) as opposed to
	// cluster-scoped (ClusterRole/ClusterRoleBinding)
	IsNamespaceScoped bool

	// The namespace the grant applies to; empty for cluster-scoped grants
	Namespace string

	// The name of the role in the grant: the created role's name, or the
	// existing role being bound to
	RoleName string

	// "Role" or "ClusterRole". This is both the kind of the role in the grant
	// and the kind used in the binding's roleRef, so the two can never diverge.
	RoleKind string

	// The name of the binding; empty when there are no subjects (no binding)
	BindingName string

	// "RoleBinding" or "ClusterRoleBinding"; empty when no binding is created
	BindingKind string

	// Combined labels (standard Planton labels + spec labels)
	Labels map[string]string

	// Annotations from the spec
	Annotations map[string]string
}

// initializeLocals resolves the three orthogonal choices of the spec (scope, role
// source, subjects) into concrete Kubernetes object names and kinds.
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetesrbacv1alpha1.KubernetesRbacIacInput) (*Locals, error) {
	locals := &Locals{
		Ctx:      ctx,
		IacInput: iacInput,
		Target:   iacInput.Target,
		Spec:     iacInput.Target.Spec,
	}

	spec := locals.Spec

	// Scope: namespace_scope produces namespaced Role/RoleBinding objects;
	// cluster_scope produces ClusterRole/ClusterRoleBinding objects. Spec
	// validation guarantees exactly one is set.
	locals.IsNamespaceScoped = spec.GetNamespaceScope() != nil

	// An omitted namespace lands the grant in the cluster's "default" namespace,
	// matching kubectl behavior. Cluster-scoped grants have no namespace at all.
	if locals.IsNamespaceScoped {
		locals.Namespace = spec.GetNamespaceScope().GetNamespace().GetValue()
		if locals.Namespace == "" {
			locals.Namespace = "default"
		}
	}

	// Role source: either we create the role (and know its kind from the scope),
	// or we bind to an existing one (whose kind the spec tells us).
	if createRole := spec.GetCreateRole(); createRole != nil {
		// The created role defaults its name to the component's own metadata.name
		// so simple grants need no explicit role name.
		locals.RoleName = createRole.GetName()
		if locals.RoleName == "" {
			locals.RoleName = locals.Target.Metadata.Name
		}
		if locals.IsNamespaceScoped {
			locals.RoleKind = kindRole
		} else {
			locals.RoleKind = kindClusterRole
		}
	} else {
		existingRole := spec.GetExistingRole()
		locals.RoleName = existingRole.GetName()
		// In namespace scope a RoleBinding may point at either a namespaced Role
		// or a ClusterRole (how built-in roles like "view" are granted
		// per-namespace); in cluster scope the reference is always a ClusterRole.
		if locals.IsNamespaceScoped && !existingRole.GetIsClusterRole() {
			locals.RoleKind = kindRole
		} else {
			locals.RoleKind = kindClusterRole
		}
	}

	// A binding exists only when there are subjects to bind. Subjects absent
	// means the grant only publishes a role definition for later bindings.
	if len(spec.GetSubjects()) > 0 {
		locals.BindingName = locals.Target.Metadata.Name
		if locals.IsNamespaceScoped {
			locals.BindingKind = kindRoleBinding
		} else {
			locals.BindingKind = kindClusterRoleBinding
		}
	}

	locals.Labels = buildLabels(locals)
	locals.Annotations = buildAnnotations(locals)

	return locals, nil
}

// buildLabels combines spec labels with standard Planton labels. Applied to every
// created RBAC object.
func buildLabels(locals *Locals) map[string]string {
	labels := make(map[string]string)

	labels["managed-by"] = "planton"
	labels["resource"] = locals.Target.Metadata.Name
	labels["resource-kind"] = "KubernetesRbac"

	for k, v := range locals.Spec.Labels {
		labels[k] = v
	}

	return labels
}

// buildAnnotations copies spec annotations. Applied to every created RBAC object.
func buildAnnotations(locals *Locals) map[string]string {
	annotations := make(map[string]string)

	for k, v := range locals.Spec.Annotations {
		annotations[k] = v
	}

	return annotations
}
