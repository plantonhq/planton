package containerenv

import (
	kubernetes "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// BuildEnvFrom converts ContainerEnv.EnvFrom entries into Pulumi EnvFromSource inputs.
// Returns nil if there are no envFrom entries.
func BuildEnvFrom(env *kubernetes.ContainerEnv) corev1.EnvFromSourceArray {
	if env == nil || len(env.EnvFrom) == 0 {
		return nil
	}

	result := make(corev1.EnvFromSourceArray, 0, len(env.EnvFrom))

	for _, entry := range env.EnvFrom {
		args := &corev1.EnvFromSourceArgs{}

		if entry.Prefix != "" {
			args.Prefix = pulumi.String(entry.Prefix)
		}

		switch src := entry.Source.(type) {
		case *kubernetes.EnvFromSource_ConfigMapRef:
			ref := src.ConfigMapRef
			args.ConfigMapRef = &corev1.ConfigMapEnvSourceArgs{
				Name:     pulumi.String(ref.Name),
				Optional: pulumi.BoolPtr(ref.Optional),
			}

		case *kubernetes.EnvFromSource_SecretRef:
			ref := src.SecretRef
			args.SecretRef = &corev1.SecretEnvSourceArgs{
				Name:     pulumi.String(ref.Name),
				Optional: pulumi.BoolPtr(ref.Optional),
			}

		default:
			continue
		}

		result = append(result, args)
	}

	if len(result) == 0 {
		return nil
	}

	return result
}
