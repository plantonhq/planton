package containerenv

import (
	kubernetes "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
)

// BuildSecretData collects literal secret values from the ContainerEnv spec
// into a map suitable for creating a Kubernetes Secret's stringData field.
// Returns nil if there are no literal secret values.
func BuildSecretData(env *kubernetes.ContainerEnv) map[string]string {
	if env == nil {
		return nil
	}

	data := make(map[string]string)

	for _, s := range env.Secrets {
		if src, ok := s.Source.(*kubernetes.SecretEnvVar_Value); ok && src.Value != "" {
			data[s.Name] = src.Value
		}
	}

	if len(data) == 0 {
		return nil
	}

	return data
}
