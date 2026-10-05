package module

import (
	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	kubernetesrbacv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/rbac/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// sourcesSecret holds flagd's SourceConfig array (it carries authorization
// headers) under the data key `sources`, read into FLAGD_SOURCES.
// Terraform twin: kubernetes_secret_v1.sources.
func sourcesSecret(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, deps []pulumi.ResourceOption) (*kubernetescorev1.Secret, error) {
	secret, err := kubernetescorev1.NewSecret(ctx, locals.SourcesSecretName, &kubernetescorev1.SecretArgs{
		Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
			Name:      pulumi.String(locals.SourcesSecretName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		}),
		Type:       pulumi.String("Opaque"),
		StringData: pulumi.StringMap{"sources": pulumi.String(locals.Config.SourcesJSON)},
	}, append([]pulumi.ResourceOption{
		pulumi.Provider(provider),
		pulumi.AdditionalSecretOutputs([]string{"data", "stringData"}),
	}, deps...)...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sources secret")
	}
	return secret, nil
}

// serviceAccount creates flagd's ServiceAccount unless an existing one is
// named. Terraform twin: kubernetes_service_account_v1.flagd with count.
func serviceAccount(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, deps []pulumi.ResourceOption) (*kubernetescorev1.ServiceAccount, error) {
	if !locals.CreateServiceAccount {
		return nil, nil
	}
	sa, err := kubernetescorev1.NewServiceAccount(ctx, locals.Name, &kubernetescorev1.ServiceAccountArgs{
		Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
			Name:        pulumi.String(locals.Name),
			Namespace:   pulumi.String(locals.Namespace),
			Labels:      pulumi.ToStringMap(locals.Labels),
			Annotations: pulumi.ToStringMap(locals.Spec.GetServiceAccount().GetAnnotations()),
		}),
	}, append([]pulumi.ResourceOption{pulumi.Provider(provider)}, deps...)...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create service account")
	}
	return sa, nil
}

// featureFlagReaderRbac grants get/list/watch on the OpenFeature Operator's
// FeatureFlag resources in each namespace a `feature_flag` source reads (flagd
// watches them through an informer). Terraform twin: kubernetes_role_v1 and
// kubernetes_role_binding_v1 with for_each.
func featureFlagReaderRbac(ctx *pulumi.Context, locals *Locals, provider pulumi.ProviderResource, deps []pulumi.ResourceOption) ([]pulumi.Resource, error) {
	opts := append([]pulumi.ResourceOption{pulumi.Provider(provider)}, deps...)
	var created []pulumi.Resource
	for _, ns := range locals.Config.FeatureFlagNamespaces {
		role, err := kubernetesrbacv1.NewRole(ctx, locals.FlagReaderName+"-"+ns, &kubernetesrbacv1.RoleArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name: pulumi.String(locals.FlagReaderName), Namespace: pulumi.String(ns), Labels: pulumi.ToStringMap(locals.Labels),
			}),
			Rules: kubernetesrbacv1.PolicyRuleArray{kubernetesrbacv1.PolicyRuleArgs{
				ApiGroups: pulumi.StringArray{pulumi.String("core.openfeature.dev")},
				Resources: pulumi.StringArray{pulumi.String("featureflags")},
				Verbs:     pulumi.StringArray{pulumi.String("get"), pulumi.String("list"), pulumi.String("watch")},
			}},
		}, opts...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create feature-flag reader role in %s", ns)
		}
		binding, err := kubernetesrbacv1.NewRoleBinding(ctx, locals.FlagReaderName+"-"+ns, &kubernetesrbacv1.RoleBindingArgs{
			Metadata: kubernetesmeta.ObjectMetaPtrInput(&kubernetesmeta.ObjectMetaArgs{
				Name: pulumi.String(locals.FlagReaderName), Namespace: pulumi.String(ns), Labels: pulumi.ToStringMap(locals.Labels),
			}),
			RoleRef: kubernetesrbacv1.RoleRefArgs{
				ApiGroup: pulumi.String("rbac.authorization.k8s.io"), Kind: pulumi.String("Role"), Name: pulumi.String(locals.FlagReaderName),
			},
			Subjects: kubernetesrbacv1.SubjectArray{kubernetesrbacv1.SubjectArgs{
				Kind: pulumi.String("ServiceAccount"), Name: pulumi.String(locals.ServiceAccountName), Namespace: pulumi.String(locals.Namespace),
			}},
		}, append(opts, pulumi.DependsOn([]pulumi.Resource{role}))...)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create feature-flag reader role binding in %s", ns)
		}
		created = append(created, role, binding)
	}
	return created, nil
}
