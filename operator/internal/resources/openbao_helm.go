package resources

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
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
	// file. The URL names the role, the host, and the database and carries
	// NO password: the backend honors the standard PostgreSQL environment,
	// so the password rides OpenBAOStoragePasswordEnv, projected from the
	// role's credential Secret -- the same projection every consumer of the
	// platform's database uses for its own password. The config, which the
	// chart renders into a ConfigMap, never sees either.
	OpenBAOStorageURLEnv = "BAO_PG_CONNECTION_URL"

	// OpenBAOStoragePasswordEnv is libpq's password variable; the vault's
	// PostgreSQL driver reads it when the connection URL carries none.
	OpenBAOStoragePasswordEnv = "PGPASSWORD"

	// openbaoStorageMaxParallel caps the storage backend's connection pool
	// (OpenBao sets the pool's open-connection limit to max_parallel). The
	// backend's default is 128; the platform's one PostgreSQL cluster allows
	// 300 connections for everyone -- the control plane's pool per logical
	// database, Temporal's four services, the identity server, OpenFGA --
	// and 100 has saturated once (see postgresqlMaxConnections). A
	// single-tenant vault serving one control plane is ample at 32.
	openbaoStorageMaxParallel = "32"

	// The init Secret's data keys. Which key set a Secret carries says which
	// seal the vault was initialized under, and the words are deliberately
	// different: unseal keys reconstruct the root key and open a built-in-
	// seal vault; recovery keys never open anything -- under a cloud seal the
	// key does that -- they authorize the quorum operations (generate-root,
	// rekey) that are the vault's break-glass. A person who finds
	// "unseal-keys" beside a KMS seal would try to unseal with them.
	OpenBAOInitSecretUnsealKeysKey   = "unseal-keys"   // JSON array of strings; built-in seal
	OpenBAOInitSecretRecoveryKeysKey = "recovery-keys" // JSON array of strings; cloud seal
	OpenBAOInitSecretRootTokenKey    = "root-token"

	// The share count and threshold for the keys /sys/init returns -- the
	// unseal keys under the built-in seal, the recovery keys under a cloud
	// seal (the server forces the barrier itself to one share there and
	// takes these as the recovery quorum instead).
	OpenBAOSecretShares    = 5
	OpenBAOSecretThreshold = 3
)

// OpenBAOHelmOptions is everything the chart's values are rendered from.
type OpenBAOHelmOptions struct {
	CRName    string
	Namespace string

	// StoragePasswordSecretName is the vault role's credential Secret (the
	// database component's, a kubernetes.io/basic-auth Secret); its
	// BasicAuthPasswordKey is projected as OpenBAOStoragePasswordEnv.
	StoragePasswordSecretName string

	// Seal is the declared seal, or nil for the built-in key shares.
	Seal *OpenBAOSealOptions

	// ServiceAccountAnnotations go on the vault's ServiceAccount -- the
	// keyless seal identity's seam (GKE Workload Identity, IRSA, Azure
	// Workload Identity). Already merged by the declaring module; rendered
	// as given.
	ServiceAccountAnnotations map[string]string
}

// OpenBAOHelmValues builds the Helm values map for rendering the official
// OpenBAO Helm chart in standalone mode on the platform's PostgreSQL.
//
// The chart deploys a single OpenBAO server with the UI enabled and TLS
// disabled (in-cluster use behind the platform's own door). The vault has no
// volume: its storage backend is the platform's database, so the platform's
// one archive carries every secret with the records, and a restored database
// brings back a vault that finds itself initialized. The chart's native seams
// carry every half -- `server.standalone.config` is the HCL the chart
// renders into a ConfigMap, so it names the backend, its sizing, and the
// seal's non-credential parameters and NOTHING credential-bearing; public
// identifiers ride `server.extraEnvironmentVars`; every credential (the
// storage role's password, a seal's key material) reaches the process as a
// variable projected from a Secret through `server.extraSecretEnvironmentVars`.
//
// Reference: Planton KubernetesOpenBao module (openbao/openbao) for the
// config rendering; OpenBao's physical/postgresql for the backend's contract.
func OpenBAOHelmValues(opts OpenBAOHelmOptions) map[string]any {
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
	// The seal stanza, when a seal is declared, is the last block: the seal
	// type and the key it names, never a credential (see OpenBAOSealOptions).
	if stanza := opts.Seal.Stanza(); stanza != "" {
		standaloneConfig += "\n" + stanza
	}

	server := map[string]any{
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
		"extraEnvironmentVars":       openbaoPlainEnv(opts),
		"extraSecretEnvironmentVars": openbaoSecretEnv(opts),
	}
	if len(opts.ServiceAccountAnnotations) > 0 {
		server["serviceAccount"] = map[string]any{
			"annotations": stringMapToAny(opts.ServiceAccountAnnotations),
		}
	}

	return map[string]any{
		"fullnameOverride": openbaoReleaseName(opts.CRName),
		"global": map[string]any{
			"enabled":    true,
			"tlsDisable": true,
		},
		"server": server,
		"ui": map[string]any{
			"enabled": true,
		},
		"injector": map[string]any{
			"enabled": false,
		},
	}
}

// openbaoPlainEnv is the non-secret environment: the storage URL (no
// password in it) and the seal's public identifiers. A map, as the chart
// takes it; Helm ranges maps in key order, so the render is deterministic.
func openbaoPlainEnv(opts OpenBAOHelmOptions) map[string]any {
	env := map[string]any{
		OpenBAOStorageURLEnv: OpenBAOStorageURL(opts.CRName, opts.Namespace),
	}
	for k, v := range opts.Seal.PlainEnv() {
		env[k] = v
	}
	return env
}

// openbaoSecretEnv is every credential the process needs, each projected
// from its Secret as the variable of the same name: the storage password
// first, then the seal's, in a fixed order so the rendered StatefulSet never
// diffs between passes.
func openbaoSecretEnv(opts OpenBAOHelmOptions) []any {
	entries := []OpenBAOSecretEnv{{
		EnvName:    OpenBAOStoragePasswordEnv,
		SecretName: opts.StoragePasswordSecretName,
		SecretKey:  BasicAuthPasswordKey,
	}}
	seal := opts.Seal.SecretEnv()
	sort.Slice(seal, func(i, j int) bool { return seal[i].EnvName < seal[j].EnvName })
	entries = append(entries, seal...)

	out := make([]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]any{
			"envName":    e.EnvName,
			"secretName": e.SecretName,
			"secretKey":  e.SecretKey,
		})
	}
	return out
}

func stringMapToAny(in map[string]string) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// openbaoReleaseName returns the Helm release name: "{crName}-openbao".
func openbaoReleaseName(crName string) string {
	return fmt.Sprintf("%s-openbao", crName)
}

// OpenBAOInitSecretName returns the operator-owned init Secret's name,
// "{crName}-openbao-init" -- the Secret the keys and root token go to when
// the declaration names none of its own.
func OpenBAOInitSecretName(crName string) string {
	return fmt.Sprintf("%s-openbao-init", crName)
}

// OpenBAOStorageURL composes the vault's PostgreSQL connection URL: its own
// role and database on the platform's cluster, through the primary's -rw
// Service, with NO password -- that rides OpenBAOStoragePasswordEnv.
// sslmode=disable is the posture every platform consumer uses on this
// in-cluster, single-namespace link (OpenFGA's URL says the same).
func OpenBAOStorageURL(crName, namespace string) string {
	return fmt.Sprintf("postgres://%s@%s:%d/%s?sslmode=disable",
		url.User(PostgreSQLVaultRole).String(),
		PostgreSQLHost(crName, namespace), PostgreSQLPort, DBOpenBAO)
}

// OpenBAOInitSecretAnnotation marks the init Secret as self-describing: with
// the vault deployed by default, every install carries this Secret, and the
// key material inside is the most security-sensitive object the operator
// touches. A person who finds it must not have to guess what it is, why it
// exists, what deleting it would cost, or whether the operator will.
const OpenBAOInitSecretAnnotation = "planton.ai/openbao-init"

// OpenBAOInitSecretNote renders the annotation's plain-language explanation
// for the seal the vault runs under and for who owns the Secret: what each
// key does, what the operator uses the Secret for, what deleting it costs,
// and what keeps it.
func OpenBAOInitSecretNote(crName string, seal *OpenBAOSealOptions, adoptersOwn bool) string {
	release := openbaoReleaseName(crName)
	var b strings.Builder
	if seal.Word() == OpenBAOSealShamir {
		fmt.Fprintf(&b, "Unseal keys and root token for the bundled secrets manager (OpenBAO release %s). ", release)
		b.WriteString("The vault is sealed with these key shares: the operator unseals it with them after every pod restart and after a restore, ")
		b.WriteString("and the root token is what the operator and the control plane sign in with. ")
		b.WriteString("Without this Secret an initialized vault stays locked and nothing can open it -- not the operator, not a restore. ")
	} else {
		fmt.Fprintf(&b, "Recovery keys and root token for the bundled secrets manager (OpenBAO release %s), sealed by %s. ", release, seal.Human())
		b.WriteString("The vault opens itself from that key on every start, including after a restore; these keys never unseal anything. ")
		b.WriteString("They authorize the vault's break-glass (generating a new root token, rekeying), and the root token is what the operator and the control plane sign in with. ")
		b.WriteString("Without this Secret the vault still opens, but its break-glass is gone. ")
	}
	if adoptersOwn {
		b.WriteString("You own this Secret (spec.vault.initSecretName): the operator wrote into it once and never deletes it, and deleting the PlantonPlatform leaves it standing. ")
		b.WriteString("A namespace the declaration owns is deleted with the declaration and takes every Secret in it, so keep a copy of this Secret outside the cluster -- ")
		b.WriteString("it is the one object a lost cluster takes with it that no archive brings back.")
	} else {
		b.WriteString("The operator owns this Secret and it is deleted with the platform. ")
		b.WriteString("Teams that want the keys to outlive the platform name a Secret they own in spec.vault.initSecretName and keep a copy of it outside the cluster.")
	}
	return b.String()
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
