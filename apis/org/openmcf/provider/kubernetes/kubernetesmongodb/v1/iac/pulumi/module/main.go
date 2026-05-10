package module

import (
	"github.com/pkg/errors"
	kubernetesmongodbv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesmongodb/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	kubernetesmetav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *kubernetesmongodbv1.KubernetesMongodbStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// Conditionally create namespace based on create_namespace flag
	createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create namespace")
	}

	// Build conditional namespace dependency (Pulumi equivalent of Terraform depends_on).
	// When create_namespace is false, createdNamespace is nil and namespaceDeps is empty.
	var namespaceDeps []pulumi.ResourceOption
	if createdNamespace != nil {
		namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
	}

	//create new random secret to set as password
	createdRandomString, err := random.NewRandomPassword(ctx,
		locals.PasswordSecretName,
		&random.RandomPasswordArgs{
			Length:     pulumi.Int(12),
			Special:    pulumi.Bool(true),
			Numeric:    pulumi.Bool(true),
			Upper:      pulumi.Bool(true),
			Lower:      pulumi.Bool(true),
			MinSpecial: pulumi.Int(3),
			MinNumeric: pulumi.Int(2),
			MinUpper:   pulumi.Int(2),
			MinLower:   pulumi.Int(2),
		})
	if err != nil {
		return errors.Wrap(err, "failed to generate random password value")
	}

	// Create kubernetes secret to store generated password
	// Percona operator expects plaintext passwords in StringData (Kubernetes auto-encodes)
	secretOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
	createdPasswordSecret, err := kubernetescorev1.NewSecret(ctx,
		locals.PasswordSecretName,
		&kubernetescorev1.SecretArgs{
			Metadata: &kubernetesmetav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.PasswordSecretName),
				Namespace: pulumi.String(locals.Namespace),
			},
			StringData: pulumi.StringMap{
				vars.MongodbRootPasswordKey: createdRandomString.Result,
			},
		}, secretOpts...)
	if err != nil {
		return errors.Wrap(err, "failed to create password secret")
	}

	// Create MongoDB using Percona operator CRD
	if err := mongodb(ctx, locals, kubernetesProvider, createdPasswordSecret, namespaceDeps); err != nil {
		return errors.Wrap(err, "failed to create MongoDB PerconaServerMongoDB resource")
	}

	//create service of type load-balancer if ingress is enabled.
	if locals.KubernetesMongodb.Spec.Ingress != nil && locals.KubernetesMongodb.Spec.Ingress.Enabled {
		lbOpts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
		_, err := kubernetescorev1.NewService(ctx,
			locals.ExternalLbServiceName,
			&kubernetescorev1.ServiceArgs{
				Metadata: &kubernetesmetav1.ObjectMetaArgs{
					Name:      pulumi.String(locals.ExternalLbServiceName),
					Namespace: pulumi.String(locals.Namespace),
					Labels:    pulumi.ToStringMap(locals.Labels),
					Annotations: pulumi.StringMap{
						"external-dns.alpha.kubernetes.io/hostname": pulumi.String(locals.IngressExternalHostname),
					},
				},
				Spec: &kubernetescorev1.ServiceSpecArgs{
					Type: pulumi.String("LoadBalancer"), // Service type is LoadBalancer
					Ports: kubernetescorev1.ServicePortArray{
						&kubernetescorev1.ServicePortArgs{
							Name:       pulumi.String("tcp-mongodb"),
							Port:       pulumi.Int(vars.MongoDbPort),
							Protocol:   pulumi.String("TCP"),
							TargetPort: pulumi.String("mongodb"),
						},
					},
					Selector: pulumi.ToStringMap(locals.MongodbPodSelectorLabels),
				},
			}, lbOpts...)
		if err != nil {
			return errors.Wrapf(err, "failed to create external load balancer service")
		}
	}

	return nil
}
