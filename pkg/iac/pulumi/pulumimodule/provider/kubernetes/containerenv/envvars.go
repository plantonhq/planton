package containerenv

import (
	kubernetes "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// BuildEnvVars converts a ContainerEnv spec into Pulumi EnvVar inputs ready for a container.
// It prepends standard pod-identity env vars (HOSTNAME, K8S_POD_ID), then appends
// user-defined variables and secrets in their declared list order.
// envSecretName is the name of the auto-created Kubernetes Secret that holds literal secret values.
func BuildEnvVars(env *kubernetes.ContainerEnv, envSecretName string) []corev1.EnvVarInput {
	inputs := make([]corev1.EnvVarInput, 0)

	inputs = appendPodIdentityVars(inputs)

	if env == nil {
		return inputs
	}

	inputs = appendVariables(inputs, env.Variables)
	inputs = appendSecrets(inputs, env.Secrets, envSecretName)

	return inputs
}

func appendPodIdentityVars(inputs []corev1.EnvVarInput) []corev1.EnvVarInput {
	inputs = append(inputs, corev1.EnvVarInput(corev1.EnvVarArgs{
		Name: pulumi.String("HOSTNAME"),
		ValueFrom: &corev1.EnvVarSourceArgs{
			FieldRef: &corev1.ObjectFieldSelectorArgs{
				FieldPath: pulumi.String("status.podIP"),
			},
		},
	}))

	inputs = append(inputs, corev1.EnvVarInput(corev1.EnvVarArgs{
		Name: pulumi.String("K8S_POD_ID"),
		ValueFrom: &corev1.EnvVarSourceArgs{
			FieldRef: &corev1.ObjectFieldSelectorArgs{
				ApiVersion: pulumi.String("v1"),
				FieldPath:  pulumi.String("metadata.name"),
			},
		},
	}))

	return inputs
}

func appendVariables(inputs []corev1.EnvVarInput, vars []*kubernetes.EnvVar) []corev1.EnvVarInput {
	for _, v := range vars {
		args := corev1.EnvVarArgs{
			Name: pulumi.String(v.Name),
		}

		switch src := v.Source.(type) {
		case *kubernetes.EnvVar_Value:
			args.Value = pulumi.String(src.Value)

		case *kubernetes.EnvVar_ValueFrom:
			// Orchestrator resolves ValueFromRef into a literal before IaC runs.
			// If the value is already resolved (non-empty), use it directly.
			// Otherwise skip -- the orchestrator hasn't resolved it yet.
			continue

		case *kubernetes.EnvVar_ConfigMapKeyRef:
			ref := src.ConfigMapKeyRef
			args.ValueFrom = &corev1.EnvVarSourceArgs{
				ConfigMapKeyRef: &corev1.ConfigMapKeySelectorArgs{
					Name:     pulumi.String(ref.Name),
					Key:      pulumi.String(ref.Key),
					Optional: pulumi.BoolPtr(ref.Optional),
				},
			}

		case *kubernetes.EnvVar_FieldRef:
			ref := src.FieldRef
			fieldRefArgs := &corev1.ObjectFieldSelectorArgs{
				FieldPath: pulumi.String(ref.FieldPath),
			}
			if ref.ApiVersion != "" {
				fieldRefArgs.ApiVersion = pulumi.String(ref.ApiVersion)
			}
			args.ValueFrom = &corev1.EnvVarSourceArgs{
				FieldRef: fieldRefArgs,
			}

		case *kubernetes.EnvVar_ResourceFieldRef:
			ref := src.ResourceFieldRef
			resourceFieldArgs := &corev1.ResourceFieldSelectorArgs{
				Resource: pulumi.String(ref.Resource),
			}
			if ref.ContainerName != "" {
				resourceFieldArgs.ContainerName = pulumi.String(ref.ContainerName)
			}
			if ref.Divisor != "" {
				resourceFieldArgs.Divisor = pulumi.String(ref.Divisor)
			}
			args.ValueFrom = &corev1.EnvVarSourceArgs{
				ResourceFieldRef: resourceFieldArgs,
			}

		default:
			continue
		}

		inputs = append(inputs, corev1.EnvVarInput(args))
	}

	return inputs
}

func appendSecrets(inputs []corev1.EnvVarInput, secrets []*kubernetes.SecretEnvVar, envSecretName string) []corev1.EnvVarInput {
	for _, s := range secrets {
		args := corev1.EnvVarArgs{
			Name: pulumi.String(s.Name),
		}

		switch src := s.Source.(type) {
		case *kubernetes.SecretEnvVar_SecretRef:
			ref := src.SecretRef
			args.ValueFrom = &corev1.EnvVarSourceArgs{
				SecretKeyRef: &corev1.SecretKeySelectorArgs{
					Name:     pulumi.String(ref.Name),
					Key:      pulumi.String(ref.Key),
					Optional: pulumi.BoolPtr(ref.Optional),
				},
			}

		case *kubernetes.SecretEnvVar_Value:
			if src.Value == "" {
				continue
			}
			args.ValueFrom = &corev1.EnvVarSourceArgs{
				SecretKeyRef: &corev1.SecretKeySelectorArgs{
					Name: pulumi.String(envSecretName),
					Key:  pulumi.String(s.Name),
				},
			}

		case *kubernetes.SecretEnvVar_ValueFrom:
			// Orchestrator resolves ValueFromRef into a literal before IaC runs.
			continue

		default:
			continue
		}

		inputs = append(inputs, corev1.EnvVarInput(args))
	}

	return inputs
}
