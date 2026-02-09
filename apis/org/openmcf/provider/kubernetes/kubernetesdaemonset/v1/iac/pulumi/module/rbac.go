package module

import (
	"github.com/pkg/errors"
	kubernetesdaemonsetv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesdaemonset/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	kubernetesrbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// rbac creates RBAC resources (ClusterRole, ClusterRoleBinding, Role, RoleBinding)
// based on the spec.rbac configuration.
func rbac(ctx *pulumi.Context, locals *Locals, serviceAccountName string, kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
	spec := locals.KubernetesDaemonSet.Spec

	// Skip if no RBAC configuration or no service account
	if spec.Rbac == nil || !spec.CreateServiceAccount {
		return nil
	}

	resourceName := locals.KubernetesDaemonSet.Metadata.Name

	// Create ClusterRole and ClusterRoleBinding for cluster-wide rules
	if len(spec.Rbac.ClusterRules) > 0 {
		// Build policy rules
		clusterRules := buildPolicyRules(spec.Rbac.ClusterRules)

		// Create ClusterRole
		crOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		clusterRole, err := kubernetesrbacv1.NewClusterRole(ctx,
			resourceName,
			&kubernetesrbacv1.ClusterRoleArgs{
				Metadata: &kubernetesmetav1.ObjectMetaArgs{
					Name:   pulumi.String(resourceName),
					Labels: pulumi.ToStringMap(locals.Labels),
				},
				Rules: clusterRules,
			},
			crOpts...,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create cluster role")
		}

		// Create ClusterRoleBinding
		crbOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		_, err = kubernetesrbacv1.NewClusterRoleBinding(ctx,
			resourceName,
			&kubernetesrbacv1.ClusterRoleBindingArgs{
				Metadata: &kubernetesmetav1.ObjectMetaArgs{
					Name:   pulumi.String(resourceName),
					Labels: pulumi.ToStringMap(locals.Labels),
				},
				Subjects: kubernetesrbacv1.SubjectArray{
					&kubernetesrbacv1.SubjectArgs{
						Kind:      pulumi.String("ServiceAccount"),
						Name:      pulumi.String(serviceAccountName),
						Namespace: pulumi.String(locals.Namespace),
					},
				},
				RoleRef: &kubernetesrbacv1.RoleRefArgs{
					Kind:     pulumi.String("ClusterRole"),
					Name:     clusterRole.Metadata.Name().Elem(),
					ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
				},
			},
			crbOpts...,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create cluster role binding")
		}
	}

	// Create Role and RoleBinding for namespace-scoped rules
	if len(spec.Rbac.NamespaceRules) > 0 {
		// Build policy rules
		namespaceRules := buildPolicyRules(spec.Rbac.NamespaceRules)

		// Create Role
		roleOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		role, err := kubernetesrbacv1.NewRole(ctx,
			resourceName,
			&kubernetesrbacv1.RoleArgs{
				Metadata: &kubernetesmetav1.ObjectMetaArgs{
					Name:      pulumi.String(resourceName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Rules: namespaceRules,
			},
			roleOpts...,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create role")
		}

		// Create RoleBinding
		rbOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		_, err = kubernetesrbacv1.NewRoleBinding(ctx,
			resourceName,
			&kubernetesrbacv1.RoleBindingArgs{
				Metadata: &kubernetesmetav1.ObjectMetaArgs{
					Name:      pulumi.String(resourceName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				Subjects: kubernetesrbacv1.SubjectArray{
					&kubernetesrbacv1.SubjectArgs{
						Kind:      pulumi.String("ServiceAccount"),
						Name:      pulumi.String(serviceAccountName),
						Namespace: pulumi.String(locals.Namespace),
					},
				},
				RoleRef: &kubernetesrbacv1.RoleRefArgs{
					Kind:     pulumi.String("Role"),
					Name:     role.Metadata.Name().Elem(),
					ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
				},
			},
			rbOpts...,
		)
		if err != nil {
			return errors.Wrap(err, "failed to create role binding")
		}
	}

	return nil
}

// buildPolicyRules converts proto RBAC rules to Pulumi PolicyRuleArray
func buildPolicyRules(rules []*kubernetesdaemonsetv1.KubernetesDaemonSetRbacRule) kubernetesrbacv1.PolicyRuleArray {
	policyRules := make(kubernetesrbacv1.PolicyRuleArray, 0, len(rules))

	for _, rule := range rules {
		policyRule := &kubernetesrbacv1.PolicyRuleArgs{
			ApiGroups: pulumi.ToStringArray(rule.ApiGroups),
			Resources: pulumi.ToStringArray(rule.Resources),
			Verbs:     pulumi.ToStringArray(rule.Verbs),
		}
		if len(rule.ResourceNames) > 0 {
			policyRule.ResourceNames = pulumi.ToStringArray(rule.ResourceNames)
		}
		policyRules = append(policyRules, policyRule)
	}

	return policyRules
}
