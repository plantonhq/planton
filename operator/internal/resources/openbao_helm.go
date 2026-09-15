package resources

import "fmt"

const (
	OpenBAOHelmChartVersion = "0.25.6"
	OpenBAOPort             = 8200

	// OpenBAOInitSecretUnsealKeysKey is the data key for unseal keys in the
	// operator-created init Secret. Stored as JSON array of strings.
	OpenBAOInitSecretUnsealKeysKey = "unseal-keys"

	// OpenBAOInitSecretRootTokenKey is the data key for the root token in the
	// operator-created init Secret.
	OpenBAOInitSecretRootTokenKey = "root-token"

	// OpenBAO Shamir's secret sharing parameters for auto-init.
	OpenBAOSecretShares    = 5
	OpenBAOSecretThreshold = 3
)

// OpenBAOHelmValues builds the Helm values map for rendering the official
// OpenBAO Helm chart in standalone mode with file storage backend.
//
// The chart deploys a single OpenBAO server with the UI enabled and TLS
// disabled (suitable for in-cluster use behind a reverse proxy or for
// development). The listener is configured to accept connections on all
// interfaces.
//
// Reference: Planton KubernetesOpenBao module (openbao/openbao).
//
// storageClass pins the data volume's StorageClass; empty means the key is
// OMITTED so the cluster default provisions.
func OpenBAOHelmValues(crName, storageSize, storageClass string) map[string]any {
	standaloneConfig := `ui = true

listener "tcp" {
  tls_disable = 1
  address = "[::]:8200"
  cluster_address = "[::]:8201"
}

storage "file" {
  path = "/openbao/data"
}
`
	dataStorage := map[string]any{
		"enabled": true,
		"size":    storageSize,
	}
	if storageClass != "" {
		dataStorage["storageClass"] = storageClass
	}
	return map[string]any{
		"fullnameOverride": openbaoReleaseName(crName),
		"global": map[string]any{
			"enabled":    true,
			"tlsDisable": true,
		},
		"server": map[string]any{
			// Deployed on every default install, so it schedules honestly:
			// explicit requests (the chart ships none) sized from observed
			// idle usage -- a single-tenant vault serving one control plane
			// is a small, steady workload. No CPU limit (requests-only, the
			// house pattern); the memory limit guards the node.
			"resources": map[string]any{
				"requests": map[string]any{
					"cpu":    "50m",
					"memory": "128Mi",
				},
				"limits": map[string]any{
					"memory": "512Mi",
				},
			},
			// The auth-delegator ClusterRoleBinding grants tokenreview/
			// subjectaccessreview permissions the operator's own RBAC does
			// not carry -- and Planton does not use Kubernetes auth for
			// OpenBAO anyway (token auth from the init Secret only).
			"authDelegator": map[string]any{
				"enabled": false,
			},
			"standalone": map[string]any{
				"enabled": true,
				"config":  standaloneConfig,
			},
			"ha": map[string]any{
				"enabled": false,
			},
			"dataStorage": dataStorage,
		},
		"ui": map[string]any{
			"enabled": true,
		},
		"injector": map[string]any{
			"enabled": false,
		},
	}
}

// openbaoReleaseName returns the Helm release name: "{crName}-openbao".
func openbaoReleaseName(crName string) string {
	return fmt.Sprintf("%s-openbao", crName)
}

// OpenBAOInitSecretName returns the name of the Secret that stores unseal keys
// and root token after auto-initialization: "{crName}-openbao-init".
func OpenBAOInitSecretName(crName string) string {
	return fmt.Sprintf("%s-openbao-init", crName)
}

// OpenBAOInitSecretAnnotation marks the init Secret as self-describing: with
// the vault deployed by default, every install carries this Secret, and the
// key material inside is the most security-sensitive object the operator
// creates. A person who finds it must not have to guess what it is, why it
// exists, or what deleting it would cost.
const OpenBAOInitSecretAnnotation = "planton.ai/openbao-init"

// OpenBAOInitSecretNote renders the annotation's plain-language explanation:
// what the keys unlock, why the operator holds them, and the alternative for
// teams that want to own the Secret themselves.
func OpenBAOInitSecretNote(crName string) string {
	return fmt.Sprintf(
		"Unseal keys and root token for the bundled secrets manager (OpenBAO release %s-openbao). "+
			"The operator uses the keys to unseal the vault after every pod restart and hands the token "+
			"to the control plane, which stores platform secrets here. Deleting this Secret leaves an "+
			"initialized-but-locked vault only these keys can open. Teams that prefer to own this Secret "+
			"name it in spec.vault.initSecretName: the operator writes the keys there and never deletes it.",
		crName)
}

// OpenBAOServiceHost returns the in-cluster DNS hostname for the OpenBAO
// Service: "{crName}-openbao.{namespace}.svc.cluster.local".
func OpenBAOServiceHost(crName, namespace string) string {
	return fmt.Sprintf("%s-openbao.%s.svc.cluster.local", crName, namespace)
}

// OpenBAOAPIAddr returns the full API address for OpenBAO:
// "http://{crName}-openbao.{namespace}.svc.cluster.local:8200".
func OpenBAOAPIAddr(crName, namespace string) string {
	return fmt.Sprintf("http://%s:%d", OpenBAOServiceHost(crName, namespace), OpenBAOPort)
}
