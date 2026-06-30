package module

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	kubernetessecretv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetessecret/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds all derived configuration and state for the module
type Locals struct {
	// Context for Pulumi operations
	Ctx *pulumi.Context

	// Stack input containing the target resource
	StackInput *kubernetessecretv1.KubernetesSecretStackInput

	// Target secret resource
	Target *kubernetessecretv1.KubernetesSecret

	// Spec from the target
	Spec *kubernetessecretv1.KubernetesSecretSpec

	// Secret name
	SecretName string

	// Secret namespace
	SecretNamespace string

	// Combined labels (spec labels + standard labels)
	Labels map[string]string

	// Combined annotations (spec annotations)
	Annotations map[string]string

	// Whether the secret is immutable
	Immutable bool

	// Kubernetes secret type string (e.g., "Opaque", "kubernetes.io/tls")
	SecretType string

	// Secret data map (stringData keys to values)
	SecretData map[string]string
}

// initializeLocals creates and populates the Locals struct
func initializeLocals(ctx *pulumi.Context, stackInput *kubernetessecretv1.KubernetesSecretStackInput) (*Locals, error) {
	locals := &Locals{
		Ctx:        ctx,
		StackInput: stackInput,
		Target:     stackInput.Target,
		Spec:       stackInput.Target.Spec,
	}

	locals.SecretName = stackInput.Target.Spec.Name
	locals.SecretNamespace = stackInput.Target.Spec.GetNamespace()
	locals.Immutable = stackInput.Target.Spec.Immutable

	// Build labels
	locals.Labels = buildLabels(locals)

	// Build annotations
	locals.Annotations = buildAnnotations(locals)

	// Compute secret type and data from the oneof variant
	secretType, secretData, err := computeSecretTypeAndData(locals.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to compute secret type and data: %w", err)
	}
	locals.SecretType = secretType
	locals.SecretData = secretData

	return locals, nil
}

// buildLabels combines spec labels with standard labels
func buildLabels(locals *Locals) map[string]string {
	labels := make(map[string]string)

	// Add standard labels
	labels["managed-by"] = "planton"
	labels["resource"] = locals.Target.Metadata.Name
	labels["resource-kind"] = "KubernetesSecret"

	// Add spec labels
	for k, v := range locals.Spec.Labels {
		labels[k] = v
	}

	return labels
}

// buildAnnotations combines spec annotations
func buildAnnotations(locals *Locals) map[string]string {
	annotations := make(map[string]string)

	// Add spec annotations
	for k, v := range locals.Spec.Annotations {
		annotations[k] = v
	}

	return annotations
}

// computeSecretTypeAndData determines the Kubernetes secret type and constructs
// the stringData map based on which oneof variant is set in the spec.
func computeSecretTypeAndData(spec *kubernetessecretv1.KubernetesSecretSpec) (string, map[string]string, error) {
	switch data := spec.SecretData.(type) {
	case *kubernetessecretv1.KubernetesSecretSpec_Opaque:
		return "Opaque", data.Opaque.Data, nil

	case *kubernetessecretv1.KubernetesSecretSpec_Tls:
		return "kubernetes.io/tls", map[string]string{
			"tls.crt": data.Tls.TlsCrt,
			"tls.key": data.Tls.TlsKey,
		}, nil

	case *kubernetessecretv1.KubernetesSecretSpec_DockerConfigJson:
		dockerConfigJSON, err := buildDockerConfigJSON(data.DockerConfigJson)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build docker config json: %w", err)
		}
		return "kubernetes.io/dockerconfigjson", map[string]string{
			".dockerconfigjson": dockerConfigJSON,
		}, nil

	case *kubernetessecretv1.KubernetesSecretSpec_BasicAuth:
		return "kubernetes.io/basic-auth", map[string]string{
			"username": data.BasicAuth.Username,
			"password": data.BasicAuth.Password,
		}, nil

	case *kubernetessecretv1.KubernetesSecretSpec_SshAuth:
		return "kubernetes.io/ssh-auth", map[string]string{
			"ssh-privatekey": data.SshAuth.SshPrivateKey,
		}, nil

	default:
		return "", nil, fmt.Errorf("no secret data variant set in spec")
	}
}

// dockerConfigAuth represents the auth entry for a single registry
type dockerConfigAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth"`
}

// dockerConfigJSON represents the top-level .dockerconfigjson structure
type dockerConfigJSON struct {
	Auths map[string]dockerConfigAuth `json:"auths"`
}

// buildDockerConfigJSON constructs the .dockerconfigjson JSON string from structured fields
func buildDockerConfigJSON(data *kubernetessecretv1.KubernetesSecretDockerConfigJsonData) (string, error) {
	auth := base64.StdEncoding.EncodeToString(
		[]byte(data.Username + ":" + data.Password),
	)

	config := dockerConfigJSON{
		Auths: map[string]dockerConfigAuth{
			data.RegistryServer: {
				Username: data.Username,
				Password: data.Password,
				Email:    data.Email,
				Auth:     auth,
			},
		},
	}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal docker config json: %w", err)
	}

	return string(jsonBytes), nil
}
