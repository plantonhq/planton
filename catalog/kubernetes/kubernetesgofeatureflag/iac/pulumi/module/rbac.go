package module

import (
	"sort"

	"github.com/pkg/errors"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	kubernetesrbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// flagReaderRbac grants the relay's ServiceAccount `get` on exactly the
// ConfigMaps its `config_map` retrievers read - one Role and RoleBinding,
// both `<metadata.name>-flag-reader`, in each namespace those ConfigMaps
// live in. The relay's retriever issues a plain GET per poll (no list, no
// watch), and resourceNames scopes the grant to the named ConfigMaps. The
// chart ships no RBAC of its own.
//
// Returns the created resources so the release can depend on them: the
// relay reads its flags on startup. Terraform twin: kubernetes_role_v1 and
// kubernetes_role_binding_v1 with for_each over the same namespaces.
func flagReaderRbac(ctx *pulumi.Context,
	locals *Locals,
	kubernetesProvider pulumi.ProviderResource,
	dependsOn []pulumi.ResourceOption,
) ([]pulumi.Resource, error) {
	namespaces := make([]string, 0, len(locals.Relay.ConfigMapGrants))
	for ns := range locals.Relay.ConfigMapGrants {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, dependsOn...)
	var created []pulumi.Resource
	for _, ns := range namespaces {
		role, err := kubernetesrbacv1.NewRole(ctx, locals.FlagReaderName+"-"+ns, &kubernetesrbacv1.RoleArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.FlagReaderName),
				Namespace: pulumi.String(ns),
				Labels:    pulumi.ToStringMap(locals.Labels),
			}),
			Rules: kubernetesrbacv1.PolicyRuleArray{
				kubernetesrbacv1.PolicyRuleArgs{
					ApiGroups:     pulumi.StringArray{pulumi.String("")},
					Resources:     pulumi.StringArray{pulumi.String("configmaps")},
					ResourceNames: pulumi.ToStringArray(locals.Relay.ConfigMapGrants[ns]),
					Verbs:         pulumi.StringArray{pulumi.String("get")},
				},
			},
		}, opts...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create flag-reader role in %s", ns)
		}

		binding, err := kubernetesrbacv1.NewRoleBinding(ctx, locals.FlagReaderName+"-"+ns, &kubernetesrbacv1.RoleBindingArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(locals.FlagReaderName),
				Namespace: pulumi.String(ns),
				Labels:    pulumi.ToStringMap(locals.Labels),
			}),
			RoleRef: kubernetesrbacv1.RoleRefArgs{
				ApiGroup: pulumi.String("rbac.authorization.k8s.io"),
				Kind:     pulumi.String("Role"),
				Name:     pulumi.String(locals.FlagReaderName),
			},
			Subjects: kubernetesrbacv1.SubjectArray{
				kubernetesrbacv1.SubjectArgs{
					Kind:      pulumi.String("ServiceAccount"),
					Name:      pulumi.String(locals.ServiceAccountName),
					Namespace: pulumi.String(locals.Namespace),
				},
			},
		}, append(opts, pulumi.DependsOn([]pulumi.Resource{role}))...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create flag-reader role binding in %s", ns)
		}
		created = append(created, role, binding)
	}
	return created, nil
}
