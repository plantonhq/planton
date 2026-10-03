package module

import (
	"strconv"

	"github.com/pkg/errors"
	kuberneteskafkauserv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kuberneteskafkauser/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	kubernetesmeta "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources renders one kafka.strimzi.io/v1 KafkaUser as an UNTYPED
// CustomResource — no generated package is shipped for the Kafka family
// (crd2pulumi cannot carry the family CRDs' free-typed config objects; the
// same ruling as the KubernetesKafka module).
//
// The CR deploys into the KAFKA CLUSTER'S OWN namespace and binds to the
// cluster through the strimzi.io/cluster label — the cluster's user
// operator watches only there (the spec comments carry the placement
// contract). No namespace resource is created here, deliberately: the
// namespace belongs to the KubernetesKafka resource's lifecycle.
//
// The user operator GENERATES the credentials (a Secret named after the
// user) — this module never touches secret material, which is why nothing
// here is sensitive: the declaration is public shape, the credential is
// operator-born.
//
// The spec body is the exact twin of the Terraform module's
// local.user_manifest. No await machinery: reconciliation belongs to the
// user operator, not to applying the resource.
func Resources(ctx *pulumi.Context, iacInput *kuberneteskafkauserv1alpha1.KubernetesKafkaUserIacInput) error {
	target := iacInput.Target
	spec := target.Spec

	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		iacInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	labels := map[string]string{
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceName: target.Metadata.Name,
		kuberneteslabelkeys.ResourceKind: catalogkind.CatalogKind_KubernetesKafkaUser.String(),
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
	// The Strimzi binding label rides ON TOP of the identity labels —
	// without it the user operator never picks the resource up.
	labels["strimzi.io/cluster"] = spec.KafkaCluster.GetValue()

	specBody := map[string]interface{}{}
	if auth := spec.GetAuthentication(); auth != nil {
		specBody["authentication"] = map[string]interface{}{
			"type": auth.GetType(),
		}
	}
	if authorization := authorizationBody(spec.GetAuthorization()); authorization != nil {
		specBody["authorization"] = authorization
	}
	if quotas := quotasBody(spec.GetQuotas()); quotas != nil {
		specBody["quotas"] = quotas
	}

	if _, err := apiextensions.NewCustomResource(ctx, target.Metadata.Name,
		&apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("kafka.strimzi.io/v1"),
			Kind:       pulumi.String("KafkaUser"),
			Metadata: &kubernetesmeta.ObjectMetaArgs{
				Name:      pulumi.String(target.Metadata.Name),
				Namespace: pulumi.String(spec.Namespace.GetValue()),
				Labels:    pulumi.ToStringMap(labels),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": specBody,
			},
		}, pulumi.Provider(kubernetesProvider)); err != nil {
		return errors.Wrap(err, "failed to create kafka user")
	}

	// tls-external users authenticate with certificates issued OUTSIDE
	// the cluster — the user operator generates no Secret for them, so
	// the handle is honestly empty. Twin of TF's conditional.
	secretName := target.Metadata.Name
	if spec.GetAuthentication().GetType() == "tls-external" || spec.GetAuthentication() == nil {
		secretName = ""
	}

	ctx.Export(OpNamespace, pulumi.String(spec.Namespace.GetValue()))
	ctx.Export(OpUsername, pulumi.String(target.Metadata.Name))
	ctx.Export(OpSecretName, pulumi.String(secretName))

	return nil
}

// authorizationBody renders the simple-authorizer ACL set — twin of TF's
// authorization locals.
func authorizationBody(authorization *kuberneteskafkauserv1alpha1.KubernetesKafkaUserAuthorization) map[string]interface{} {
	if authorization == nil {
		return nil
	}

	acls := make([]interface{}, 0, len(authorization.GetAcls()))
	for _, acl := range authorization.GetAcls() {
		resource := map[string]interface{}{
			"type": acl.GetResource().GetType(),
		}
		if acl.GetResource().GetName() != "" {
			resource["name"] = acl.GetResource().GetName()
		}
		if acl.GetResource().PatternType != nil && acl.GetResource().GetPatternType() != "" {
			resource["patternType"] = acl.GetResource().GetPatternType()
		}

		aclBody := map[string]interface{}{
			"resource":   resource,
			"operations": stringSliceToInterface(acl.GetOperations()),
		}
		if acl.GetHost() != "" {
			aclBody["host"] = acl.GetHost()
		}
		acls = append(acls, aclBody)
	}

	authType := authorization.GetType()
	if authType == "" {
		authType = "simple"
	}

	return map[string]interface{}{
		"type": authType,
		"acls": acls,
	}
}

func quotasBody(quotas *kuberneteskafkauserv1alpha1.KubernetesKafkaUserQuotas) map[string]interface{} {
	if quotas == nil {
		return nil
	}
	body := map[string]interface{}{}
	if quotas.ProducerByteRate != nil {
		body["producerByteRate"] = int(quotas.GetProducerByteRate())
	}
	if quotas.ConsumerByteRate != nil {
		body["consumerByteRate"] = int(quotas.GetConsumerByteRate())
	}
	if quotas.RequestPercentage != nil {
		body["requestPercentage"] = int(quotas.GetRequestPercentage())
	}
	if quotas.ControllerMutationRate != nil {
		body["controllerMutationRate"] = quotas.GetControllerMutationRate()
	}
	if len(body) == 0 {
		return nil
	}
	return body
}

func stringSliceToInterface(in []string) []interface{} {
	out := make([]interface{}, 0, len(in))
	for _, v := range in {
		out = append(out, v)
	}
	return out
}
