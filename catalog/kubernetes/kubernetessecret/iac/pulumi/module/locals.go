package module

import (
	"fmt"

	kubernetessecretv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetessecret/v1alpha1"
	"github.com/plantonhq/planton/pkg/kubernetes/dockerconfigjson"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds all derived configuration and state for the module
type Locals struct {
	// Context for Pulumi operations
	Ctx *pulumi.Context

	// IaC input containing the target resource
	IacInput *kubernetessecretv1alpha1.KubernetesSecretIacInput

	// Target secret resource
	Target *kubernetessecretv1alpha1.KubernetesSecret

	// Spec from the target
	Spec *kubernetessecretv1alpha1.KubernetesSecretSpec

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

	// Secret data map (stringData keys to plain-string values)
	SecretData map[string]string

	// Secret binary data map (data keys to base64-encoded values)
	SecretBinaryData map[string]string
}

// initializeLocals creates and populates the Locals struct
func initializeLocals(ctx *pulumi.Context, iacInput *kubernetessecretv1alpha1.KubernetesSecretIacInput) (*Locals, error) {
	locals := &Locals{
		Ctx:      ctx,
		IacInput: iacInput,
		Target:   iacInput.Target,
		Spec:     iacInput.Target.Spec,
	}

	locals.SecretName = iacInput.Target.Spec.Name

	// Namespace is a value-or-ref; by module execution time any reference has been
	// resolved into the literal value. Fall back to "default" when unset, mirroring
	// kubectl behavior without a namespace flag.
	locals.SecretNamespace = iacInput.Target.Spec.Namespace.GetValue()
	if locals.SecretNamespace == "" {
		locals.SecretNamespace = "default"
	}

	locals.Immutable = iacInput.Target.Spec.Immutable

	// Build labels
	locals.Labels = buildLabels(locals)

	// Build annotations
	locals.Annotations = buildAnnotations(locals)

	// Compute secret type and data from the oneof variant
	variant, err := computeSecretTypeAndData(locals.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to compute secret type and data: %w", err)
	}
	locals.SecretType = variant.secretType
	locals.SecretData = variant.stringData
	locals.SecretBinaryData = variant.binaryData

	// Some secret types mandate annotations (service-account-token binds to its
	// ServiceAccount via annotation); merge them into the user-provided ones.
	for k, v := range variant.annotations {
		locals.Annotations[k] = v
	}

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

// secretVariant holds everything derived from the oneof secret_data selection:
// the Kubernetes secret type plus the data and annotations that type mandates.
type secretVariant struct {
	secretType string

	// stringData carries plain-string values (Kubernetes stringData semantics)
	stringData map[string]string

	// binaryData carries base64-encoded values (Kubernetes data semantics)
	binaryData map[string]string

	// annotations mandated by the secret type (e.g., the service-account binding)
	annotations map[string]string
}

// computeSecretTypeAndData determines the Kubernetes secret type and constructs
// the data and annotation maps based on which oneof variant is set in the spec.
func computeSecretTypeAndData(spec *kubernetessecretv1alpha1.KubernetesSecretSpec) (*secretVariant, error) {
	switch data := spec.SecretData.(type) {
	case *kubernetessecretv1alpha1.KubernetesSecretSpec_Opaque:
		return &secretVariant{
			secretType: "Opaque",
			stringData: data.Opaque.Data,
			binaryData: data.Opaque.BinaryData,
		}, nil

	case *kubernetessecretv1alpha1.KubernetesSecretSpec_Tls:
		return &secretVariant{
			secretType: "kubernetes.io/tls",
			stringData: map[string]string{
				"tls.crt": data.Tls.TlsCrt,
				"tls.key": data.Tls.TlsKey,
			},
		}, nil

	case *kubernetessecretv1alpha1.KubernetesSecretSpec_DockerConfigJson:
		// One registry login, rendered by the catalog's single docker-config encoder
		// (the same one the workload kinds use for `pod.image_registries`), so the
		// document shape is defined once. The password arrives already resolved from
		// its `$secret/` reference.
		d := data.DockerConfigJson
		dockerConfigJSON, err := dockerconfigjson.Encode([]dockerconfigjson.Auth{{
			Server:   d.GetRegistryServer(),
			Username: d.GetUsername(),
			Password: d.GetPassword(),
			Email:    d.GetEmail(),
		}})
		if err != nil {
			return nil, fmt.Errorf("failed to build docker config json: %w", err)
		}
		return &secretVariant{
			secretType: "kubernetes.io/dockerconfigjson",
			stringData: map[string]string{
				".dockerconfigjson": dockerConfigJSON,
			},
		}, nil

	case *kubernetessecretv1alpha1.KubernetesSecretSpec_BasicAuth:
		return &secretVariant{
			secretType: "kubernetes.io/basic-auth",
			stringData: map[string]string{
				"username": data.BasicAuth.Username,
				"password": data.BasicAuth.Password,
			},
		}, nil

	case *kubernetessecretv1alpha1.KubernetesSecretSpec_SshAuth:
		return &secretVariant{
			secretType: "kubernetes.io/ssh-auth",
			stringData: map[string]string{
				"ssh-privatekey": data.SshAuth.SshPrivateKey,
			},
		}, nil

	case *kubernetessecretv1alpha1.KubernetesSecretSpec_ServiceAccountToken:
		// No data is set: the cluster's token controller populates token/ca.crt/namespace
		// after creation. The annotation binds the token to its ServiceAccount.
		return &secretVariant{
			secretType: "kubernetes.io/service-account-token",
			annotations: map[string]string{
				"kubernetes.io/service-account.name": data.ServiceAccountToken.ServiceAccountName.GetValue(),
			},
		}, nil

	default:
		return nil, fmt.Errorf("no secret data variant set in spec")
	}
}
