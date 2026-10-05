package module

import (
	"github.com/pkg/errors"
	kubernetesclustersecretstorev1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesclustersecretstore/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/externalsecretsstore"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one External Secrets Operator ClusterSecretStore plus
// the credential Secrets its backend needs.
//
// The CR spec is rendered by the shared externalsecretsstore builder — the
// SAME builder the KubernetesSecretStore module uses, because upstream
// ClusterSecretStore and SecretStore share an identical spec and the two
// Planton kinds share the ExternalSecretsStoreConfig proto. One builder
// means the two kinds can never drift. (The typed crd2pulumi args are two
// disjoint generated type trees for the same schema — using them here would
// force two divergent copies of this whole mapping. ESO's validating
// webhook checks the applied spec strictly, and the kind-cluster E2E lanes
// exercise the machinery live, so shape errors still fail loudly.)
//
// Credential Secrets land in the spec's secrets namespace — cluster-scoped
// stores read their referenced Secrets from an EXPLICIT namespace, and the
// builder stamps that namespace into every secret reference.
//
// Neither engine waits for the store to reach Ready: readiness depends on
// external reachability (the cloud secrets API, Vault) that is not part of
// applying the resource — the same never-block-on-a-controller posture as
// the cert-manager issuers. Terraform equivalent: kubectl_manifest without
// a wait_for block.
func Resources(ctx *pulumi.Context, iacInput *kubernetesclustersecretstorev1alpha1.KubernetesClusterSecretStoreIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	result, err := externalsecretsstore.BuildSpec(
		locals.StoreName, locals.SecretsNamespace, true, iacInput.Target.Spec.Config)
	if err != nil {
		return errors.Wrap(err, "failed to build cluster secret store spec")
	}

	// The cluster kind's namespace fence renders alongside the shared
	// config: which namespaces' ExternalSecrets may sync from this store.
	if conditions := buildConditions(locals.Spec.Conditions); len(conditions) > 0 {
		result.Spec["conditions"] = conditions
	}

	// Credential Secrets first; the CR depends on them so ESO never
	// observes a store whose secretRefs dangle.
	var secretResources []pulumi.Resource
	for _, credential := range result.Secrets {
		createdSecret, err := corev1.NewSecret(ctx, credential.Name,
			&corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(credential.Name),
					Namespace: pulumi.String(locals.SecretsNamespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
				},
				StringData: pulumi.ToSecret(pulumi.ToStringMap(credential.Data)).(pulumi.StringMapOutput),
			},
			pulumi.Provider(kubernetesProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create credential secret %s", credential.Name)
		}
		secretResources = append(secretResources, createdSecret)
	}

	opts := []pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}
	if len(secretResources) > 0 {
		opts = append(opts, pulumi.DependsOn(secretResources))
	}

	_, err = apiextensions.NewCustomResource(ctx, locals.StoreName,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("external-secrets.io/v1"),
			Kind:       pulumi.String("ClusterSecretStore"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.StoreName),
				Labels: pulumi.ToStringMap(locals.Labels),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": result.Spec,
			},
		},
		opts...)
	if err != nil {
		return errors.Wrap(err, "failed to create cluster secret store")
	}

	ctx.Export(OpStoreName, pulumi.String(locals.StoreName))
	ctx.Export(OpSecretsNamespace, pulumi.String(locals.SecretsNamespace))

	return nil
}

// buildConditions renders the namespace fence into the CRD's conditions
// shape. Terraform twin: the conditions list in locals.tf — keep in
// lockstep.
func buildConditions(conditions []*kubernetesclustersecretstorev1alpha1.KubernetesClusterSecretStoreCondition) []interface{} {
	rendered := make([]interface{}, 0, len(conditions))
	for _, condition := range conditions {
		entry := map[string]interface{}{}
		if len(condition.GetNamespaces()) > 0 {
			namespaces := make([]interface{}, 0, len(condition.GetNamespaces()))
			for _, namespace := range condition.GetNamespaces() {
				namespaces = append(namespaces, namespace)
			}
			entry["namespaces"] = namespaces
		}
		if len(condition.GetNamespaceLabelSelector()) > 0 {
			matchLabels := map[string]interface{}{}
			for k, v := range condition.GetNamespaceLabelSelector() {
				matchLabels[k] = v
			}
			entry["namespaceSelector"] = map[string]interface{}{"matchLabels": matchLabels}
		}
		if len(condition.GetNamespaceRegexes()) > 0 {
			regexes := make([]interface{}, 0, len(condition.GetNamespaceRegexes()))
			for _, regex := range condition.GetNamespaceRegexes() {
				regexes = append(regexes, regex)
			}
			entry["namespaceRegexes"] = regexes
		}
		if len(entry) > 0 {
			rendered = append(rendered, entry)
		}
	}
	return rendered
}
