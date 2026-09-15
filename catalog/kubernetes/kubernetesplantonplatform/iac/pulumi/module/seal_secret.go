package module

import (
	kubernetesplantonplatformv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonplatform/v1alpha1"
)

// The environment-variable names the vault's seal wrappers read their
// credentials from — the cloud SDKs' standard variables; transit follows the
// wrapper's token variable. They are the KEYS of the seal-credentials Secret
// this module materializes, so the operator can hand every key straight to
// the vault's process as the variable of the same name; nothing
// credential-bearing ever enters the vault's config ConfigMap.
const (
	sealEnvAwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	sealEnvAzureClientSecret  = "AZURE_CLIENT_SECRET"
	sealEnvTransitToken       = "VAULT_TOKEN"
)

// The annotation GKE Workload Identity reads on a ServiceAccount. The GCP
// seal arm's declared identity writes it; an explicit entry in
// vault.service_account_annotations wins on conflict.
const gkeWorkloadIdentityAnnotation = "iam.gke.io/gcp-service-account"

// sealCredentialsSecret is the one Secret a declared seal arm needs, or nil
// for a keyless posture (workload identity, an instance profile, a managed
// identity) — no Secret exists, and the CR names none, so the operator
// leaves the vault to its ambient identity. The GCP arm is keyless by
// construction: its identity is a ServiceAccount annotation, never a key.
func sealCredentialsSecret(locals *Locals) *materializedSecret {
	data := sealSecretData(locals.Spec.GetVault())
	if data == nil {
		return nil
	}
	return &materializedSecret{Name: locals.SealCredentialsSecretName, Data: data}
}

// sealSecretData is the credential material a declared seal arm carries,
// keyed by the environment-variable name it must reach the vault as — or
// nil when the arm declares none. Public identifiers (an AWS access key id,
// an Azure client id) are not credentials and ride the CR as plain fields.
func sealSecretData(vault *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVault) map[string]string {
	seal := vault.GetAutoUnseal()
	if seal == nil {
		return nil
	}
	data := map[string]string{}
	switch {
	case seal.GetAwsKms() != nil:
		if v := seal.GetAwsKms().GetSecretAccessKey(); v != "" {
			data[sealEnvAwsSecretAccessKey] = v
		}
	case seal.GetAzureKeyVault() != nil:
		if v := seal.GetAzureKeyVault().GetClientSecret(); v != "" {
			data[sealEnvAzureClientSecret] = v
		}
	case seal.GetTransit() != nil:
		if v := seal.GetTransit().GetToken(); v != "" {
			// The transit seal wrapper authenticates to the central
			// instance with the standard token variable. The vault's own
			// process never self-authenticates, so the variable is inert
			// beyond the seal wrapper.
			data[sealEnvTransitToken] = v
		}
	}
	if len(data) == 0 {
		return nil
	}
	return data
}

// vaultServiceAccountAnnotations is the annotation map the CR carries for the
// vault's ServiceAccount: the GCP seal arm's declared workload identity
// contributes the GKE annotation (the field promises exactly this — the
// annotation follows the identity by reference), and every explicit
// vault.service_account_annotations entry is laid over it, so an explicit
// value wins on conflict. Nil when there is nothing to say, so the CR omits
// the key.
func vaultServiceAccountAnnotations(vault *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVault) map[string]string {
	out := map[string]string{}
	if email := vault.GetAutoUnseal().GetGcpKms().GetWorkloadIdentityServiceAccount().GetValue(); email != "" {
		out[gkeWorkloadIdentityAnnotation] = email
	}
	for k, v := range vault.GetServiceAccountAnnotations() {
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
