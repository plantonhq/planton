package resources

import (
	"fmt"
	"net/url"
)

const (
	// OpenBAOHelmChartVersion is the vendored openbao-helm chart (the tgz under
	// manifests/openbao-chart, pulled by `make pull-chart-openbao`). 0.28.6
	// pairs with OpenBao v2.6.1 -- the same pin the catalog's standalone vault
	// kind runs, so one upstream clone and one set of chart facts serve both.
	OpenBAOHelmChartVersion = "0.28.6"
	OpenBAOPort             = 8200

	// OpenBAOStorageURLEnv is the environment variable OpenBao's PostgreSQL
	// storage backend reads its connection URL from, ahead of the config
	// file. The URL carries the vault role's password, so it reaches the
	// server ONLY this way -- through a Secret the chart projects as this
	// variable -- and never through the config, which the chart renders into
	// a ConfigMap.
	OpenBAOStorageURLEnv = "BAO_PG_CONNECTION_URL"

	// openbaoStorageMaxParallel caps the storage backend's connection pool
	// (OpenBao sets the pool's open-connection limit to max_parallel). The
	// backend's default is 128; the platform's one PostgreSQL cluster allows
	// 300 connections for everyone -- the control plane's pool per logical
	// database, Temporal's four services, the identity server, OpenFGA --
	// and 100 has saturated once (see postgresqlMaxConnections). A
	// single-tenant vault serving one control plane is ample at 32.
	openbaoStorageMaxParallel = "32"

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
// OpenBAO Helm chart in standalone mode on the platform's PostgreSQL.
//
// The chart deploys a single OpenBAO server with the UI enabled and TLS
// disabled (in-cluster use behind the platform's own door). The vault has no
// volume: its storage backend is the platform's database, so the platform's
// one archive carries every secret with the records, and a restored database
// brings back a vault that finds itself initialized. The chart's native seams
// carry both halves -- `server.standalone.config` is the HCL the chart
// renders into a ConfigMap, so the stanza names the backend and its sizing
// and NOTHING credential-bearing; the connection URL (which carries the
// role's password) reaches the process as an environment variable projected
// from storageSecretName through `server.extraSecretEnvironmentVars`, the
// variable the backend reads ahead of the config file.
//
// Reference: Planton KubernetesOpenBao module (openbao/openbao) for the
// config rendering; OpenBao's physical/postgresql for the backend's contract.
func OpenBAOHelmValues(crName, storageSecretName string) map[string]any {
	// max_parallel bounds the pool (see openbaoStorageMaxParallel). The
	// backend creates its own table on first start and gives up after ONE
	// failed connect (max_connect_retries default 1) -- so the component
	// waits for the role and database to exist before this renders, and a
	// transient outage is a pod restart, never a stuck server.
	standaloneConfig := `ui = true

listener "tcp" {
  tls_disable = 1
  address = "[::]:8200"
  cluster_address = "[::]:8201"
}

storage "postgresql" {
  max_parallel = "` + openbaoStorageMaxParallel + `"
}
`
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
			// No volume: the chart guards the claim template and the data
			// mount on this one switch, and the vault's data lives in the
			// platform's database.
			"dataStorage": map[string]any{
				"enabled": false,
			},
			// The connection URL, password included, as the variable the
			// backend reads -- projected from the operator-owned storage
			// Secret, never written into the config above.
			"extraSecretEnvironmentVars": []any{
				map[string]any{
					"envName":    OpenBAOStorageURLEnv,
					"secretName": storageSecretName,
					"secretKey":  OpenBAOStorageURLEnv,
				},
			},
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

// OpenBAOStorageSecretName returns the operator-owned Secret that carries the
// vault's storage connection URL under the OpenBAOStorageURLEnv key:
// "{crName}-openbao-storage". Derived every pass from the vault role's
// credential Secret (the database component's), so it can never drift from
// it; the process reads it at start, like every consumer's database password,
// so a rotation of the role's password would need the pod rolled.
func OpenBAOStorageSecretName(crName string) string {
	return fmt.Sprintf("%s-openbao-storage", crName)
}

// OpenBAOStorageURL composes the vault's PostgreSQL connection URL: its own
// role and database on the platform's cluster, through the primary's -rw
// Service. sslmode=disable is the posture every platform consumer uses on
// this in-cluster, single-namespace link (OpenFGA's URL says the same); the
// password is URL-escaped even though the generator emits URL-safe base64,
// so a hand-set Secret with other characters still parses.
func OpenBAOStorageURL(crName, namespace, password string) string {
	return fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
		url.UserPassword(PostgreSQLVaultRole, password).String(),
		PostgreSQLHost(crName, namespace), PostgreSQLPort, DBOpenBAO)
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
