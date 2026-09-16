//go:build e2e

package e2e

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/platformversion"
	"github.com/plantonhq/planton/operator/internal/resources"
	"github.com/plantonhq/planton/operator/test/fixtures"
	"github.com/plantonhq/planton/operator/test/utils"
)

// The vault lanes' vocabulary and seams. Everything a lane says about a
// platform it says through the operator's own names (resources.*) and the
// operator's own types (v1.PlantonPlatform), so a rename in the operator is
// a compile error here and never a lane that passes against the wrong
// object. Everything a lane says TO the vault it says through the token the
// operator minted for the control plane, from inside the vault's own pod --
// the exact credential, policy, and address the platform uses.

const (
	// platformName is the name every lane gives its platform; the lanes are
	// told apart by namespace.
	platformName = "planton"

	// laneKeysSecret is the Secret an adopter names in spec.vault.initSecretName
	// -- pre-created empty by the lane, the GitOps shape, filled by the
	// operator. laneKeysSecretCopy is the copy "kept outside the cluster",
	// which the lane keeps as a second Secret in the same namespace: what the
	// bad-day runbook recreates the real one from.
	laneKeysSecret     = "lane-vault-keys"
	laneKeysSecretCopy = "lane-vault-keys-kept-copy"

	// laneSealCredentials carries VAULT_TOKEN for the transit seal.
	laneSealCredentials = "lane-seal-creds"

	// laneBackupCredentials carries the store's ACCESS_KEY_ID and
	// SECRET_ACCESS_KEY for the backup and the restore.
	laneBackupCredentials = "lane-backup-keys"

	// oidcSigningKey is the OIDC issuer's Transit key the control plane
	// creates in the vault the moment it is live -- the platform's real
	// signing key, and the lane's proof that the control plane started
	// against the token the operator minted. rsa-2048 by the platform's
	// default.
	oidcSigningKey = "planton-oidc-signing"

	// consoleImageTagEnv lets a run pin the console's image tag when the
	// console's published line lags the platform version the lane declares
	// (the console publishes on its own release cadence). Empty means the
	// declared version, the operator's default.
	consoleImageTagEnv = "E2E_CONSOLE_IMAGE_TAG"
)

// Budgets for the waits a real platform boot needs on a Kind cluster: the
// databases and the vault come up in a few minutes once images are cached;
// the whole platform, the identity server included, in ten to fifteen.
const (
	componentBudget = 25 * time.Minute
	restoreBudget   = 30 * time.Minute
	backupBudget    = 15 * time.Minute
	sentenceBudget  = 6 * time.Minute
)

// kubectl runs kubectl with the given arguments and returns its output.
func kubectl(args ...string) (string, error) {
	return utils.Run(exec.Command("kubectl", args...))
}

// kubectlApply applies a manifest from stdin.
func kubectlApply(manifest []byte) (string, error) {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(string(manifest))
	return utils.Run(cmd)
}

// mustApply applies a manifest and fails the spec when the API refuses it.
func mustApply(manifest []byte, what string) {
	out, err := kubectlApply(manifest)
	Expect(err).NotTo(HaveOccurred(), "applying %s: %s", what, out)
}

// applyFixture applies one of the embedded lab fixtures.
func applyFixture(manifest []byte, what string) {
	By("applying the " + what + " fixture")
	mustApply(manifest, what)
}

// waitAvailable waits for a Deployment to report Available.
func waitAvailable(ns, name string, budget time.Duration) {
	_, err := kubectl("wait", "--for=condition=Available", "deployment/"+name, "-n", ns, "--timeout="+budget.String())
	Expect(err).NotTo(HaveOccurred(), "deployment %s/%s never became Available", ns, name)
}

// waitJobComplete waits for a Job to complete.
func waitJobComplete(ns, name string, budget time.Duration) {
	_, err := kubectl("wait", "--for=condition=Complete", "job/"+name, "-n", ns, "--timeout="+budget.String())
	Expect(err).NotTo(HaveOccurred(), "job %s/%s never completed", ns, name)
}

// createNamespace creates a namespace the lane owns. The platform is declared
// INTO it and never creates it, which is the layer at which an adopter's
// keys Secret survives a platform destroy: the operator never deletes that
// Secret and deleting the PlantonPlatform leaves it standing, while a
// namespace the declaring resource owns would take every Secret with it.
func createNamespace(ns string) {
	_, err := kubectl("create", "namespace", ns)
	Expect(err).NotTo(HaveOccurred(), "creating namespace %s", ns)
}

// deleteNamespace removes a lane's namespace and everything in it.
func deleteNamespace(ns string) {
	_, _ = kubectl("delete", "namespace", ns, "--ignore-not-found", "--wait=false")
}

// literalSecret creates or rewrites an Opaque Secret from literal pairs.
func literalSecret(ns, name string, data map[string]string) {
	args := make([]string, 0, 9+len(data))
	args = append(args, "create", "secret", "generic", name, "-n", ns, "--dry-run=client", "-o", "yaml")
	for k, v := range data {
		args = append(args, fmt.Sprintf("--from-literal=%s=%s", k, v))
	}
	rendered, err := kubectl(args...)
	Expect(err).NotTo(HaveOccurred(), "rendering Secret %s/%s", ns, name)
	mustApply([]byte(rendered), "Secret "+ns+"/"+name)
}

// emptySecret pre-creates an Opaque Secret with no data: the shape a GitOps
// tool leaves for the operator to fill.
func emptySecret(ns, name string) {
	manifest := fmt.Sprintf("apiVersion: v1\nkind: Secret\nmetadata:\n  name: %s\n  namespace: %s\ntype: Opaque\n",
		name, ns)
	mustApply([]byte(manifest), "empty Secret "+ns+"/"+name)
}

// secretJSONPath reads one jsonpath off a Secret; "" when the Secret is gone.
func secretJSONPath(ns, name, jsonPath string) string {
	out, err := kubectl("get", "secret", name, "-n", ns, "-o", "jsonpath="+jsonPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// secretExists reports whether a Secret exists.
func secretExists(ns, name string) bool {
	_, err := kubectl("get", "secret", name, "-n", ns)
	return err == nil
}

// secretValue reads and decodes one data key of a Secret.
func secretValue(ns, name, key string) string {
	encoded := secretJSONPath(ns, name, fmt.Sprintf("{.data.%s}", strings.ReplaceAll(key, ".", "\\.")))
	if encoded == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	Expect(err).NotTo(HaveOccurred(), "decoding Secret %s/%s key %s", ns, name, key)
	return string(decoded)
}

// secretDataKeys lists a Secret's data keys.
func secretDataKeys(ns, name string) []string {
	out := secretJSONPath(ns, name, "{range $k, $v := .data}{$k}{\"\\n\"}{end}")
	return utils.GetNonEmptyLines(out)
}

// copySecret duplicates a Secret's data and type under another name in the
// same namespace, with none of the source's metadata: the lane's stand-in
// for "a copy kept outside the cluster".
func copySecret(ns, from, to string) {
	out, err := kubectl("get", "secret", from, "-n", ns, "-o", "json")
	Expect(err).NotTo(HaveOccurred(), "reading Secret %s/%s", ns, from)
	var source struct {
		Type string            `json:"type"`
		Data map[string]string `json:"data"`
	}
	Expect(json.Unmarshal([]byte(out), &source)).To(Succeed())
	copied, err := json.Marshal(map[string]any{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata":   map[string]any{"name": to, "namespace": ns},
		"type":       source.Type,
		"data":       source.Data,
	})
	Expect(err).NotTo(HaveOccurred())
	mustApply(copied, "Secret copy "+ns+"/"+to)
}

// deleteSecret removes a Secret and waits for it to be gone.
func deleteSecret(ns, name string) {
	_, err := kubectl("delete", "secret", name, "-n", ns, "--ignore-not-found", "--wait=true")
	Expect(err).NotTo(HaveOccurred(), "deleting Secret %s/%s", ns, name)
}

// platformOptions is what varies between the lanes' declarations; the
// declaration itself is one typed object built by platformDeclaration.
type platformOptions struct {
	namespace string
	// backupPrefix is the archive prefix under the lane's bucket. Every
	// declaration of a backup uses its own: Barman refuses a path that
	// already holds another system's WAL, so a restored platform archives
	// under a new prefix beside the source's.
	backupPrefix string
	// initSecretName names the adopter-owned keys Secret; "" keeps the
	// operator-owned default.
	initSecretName string
	// transitKey, when set, declares a transit seal against the lane's key
	// holder with that key name (the lane varies it to prove the seal is
	// pinned).
	transitKey string
	// recoverFromServer, when set, declares recoverFrom against the source
	// archive at recoverFromPrefix, filed under that server name.
	recoverFromServer string
	recoverFromPrefix string
}

// platformDeclaration renders a lane's platform as the operator's own type,
// so every field name is the compiler's to check. The shape: the floor
// version (the newest contract this manager honors), the control plane at
// its published image so the real consumer starts against the minted token,
// the runner opted out (a platform that proves its vault does not need to
// deploy infrastructure), a small database, a backup to the lane's store,
// and the vault's seal and keys Secret as the lane declares them.
func platformDeclaration(o platformOptions) []byte {
	store := func(prefix string) v1.ObjectStoreSpec {
		return v1.ObjectStoreSpec{
			DestinationPath: fmt.Sprintf("s3://%s/%s", fixtures.MinIOBucket, prefix),
			S3: &v1.S3ObjectStoreSpec{
				EndpointURL:           fixtures.MinIOEndpoint,
				Region:                "us-east-1",
				CredentialsSecretName: laneBackupCredentials,
			},
		}
	}
	runnerOff := false
	platform := &v1.PlantonPlatform{
		TypeMeta:   metav1.TypeMeta{APIVersion: v1.GroupVersion.String(), Kind: "PlantonPlatform"},
		ObjectMeta: metav1.ObjectMeta{Name: platformName, Namespace: o.namespace},
		Spec: v1.PlantonPlatformSpec{
			Version: platformversion.MinimumSupported,
			Runner:  &v1.RunnerSpec{Enabled: &runnerOff},
			Database: &v1.DatabaseSpec{
				PostgreSQL: &v1.PostgreSQLSpec{
					StorageSize: resource.MustParse("2Gi"),
					Backup:      &v1.PostgreSQLBackupSpec{ObjectStore: store(o.backupPrefix), RetentionPolicy: "7d"},
				},
			},
			Vault: &v1.OpenBAOSpec{InitSecretName: o.initSecretName},
		},
	}
	if tag := os.Getenv(consoleImageTagEnv); tag != "" {
		platform.Spec.Console = &v1.ConsoleSpec{Image: &v1.ImageSpec{Tag: tag}}
	}
	if o.transitKey != "" {
		platform.Spec.Vault.AutoUnseal = &v1.OpenBAOAutoUnsealSpec{Transit: &v1.OpenBAOTransitSealSpec{
			Address:               fixtures.KeyHolderAddress,
			KeyName:               o.transitKey,
			MountPath:             fixtures.KeyHolderTransitMount + "/",
			CredentialsSecretName: laneSealCredentials,
		}}
	}
	if o.recoverFromServer != "" {
		platform.Spec.Database.PostgreSQL.RecoverFrom = &v1.PostgreSQLRecoverFromSpec{
			ObjectStore: store(o.recoverFromPrefix),
			ServerName:  o.recoverFromServer,
		}
	}
	rendered, err := yaml.Marshal(platform)
	Expect(err).NotTo(HaveOccurred(), "rendering the platform declaration")
	return rendered
}

// platformJSONPath reads one jsonpath off the lane's platform; "" when it
// cannot be read.
func platformJSONPath(ns, jsonPath string) string {
	out, err := kubectl("get", "plantonplatform", platformName, "-n", ns, "-o", "jsonpath="+jsonPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// componentStatus is one component's slot in the platform's status, read
// live. The vault's slot is status.components.openBao.
type componentStatus struct {
	Phase   string
	Reason  string
	Message string
	Object  string
}

func readComponent(ns, component string) componentStatus {
	return componentStatus{
		Phase:   platformJSONPath(ns, fmt.Sprintf("{.status.components.%s.phase}", component)),
		Reason:  platformJSONPath(ns, fmt.Sprintf("{.status.components.%s.reason}", component)),
		Message: platformJSONPath(ns, fmt.Sprintf("{.status.components.%s.message}", component)),
		Object:  platformJSONPath(ns, fmt.Sprintf("{.status.components.%s.object.name}", component)),
	}
}

const (
	vaultComponent        = "openBao"
	postgresqlComponent   = "postgresql"
	controlPlaneComponent = "controlPlane"
)

// waitComponentReady waits for a component's phase to reach Ready, printing
// its phase and message on every change so a long boot reads as progress.
func waitComponentReady(ns, component string, budget time.Duration) {
	var last componentStatus
	Eventually(func(g Gomega) {
		current := readComponent(ns, component)
		if current != last {
			_, _ = fmt.Fprintf(GinkgoWriter, "  %s/%s %s: %s %s -- %s\n",
				ns, platformName, component, current.Phase, current.Reason, current.Message)
			last = current
		}
		g.Expect(current.Phase).To(Equal(string(v1.ComponentPhaseReady)),
			"%s is %s (%s): %s", component, current.Phase, current.Reason, current.Message)
	}, budget, 15*time.Second).Should(Succeed())
}

// waitVaultSays waits for the vault component to carry the given reason (""
// for any) and every fragment in its message: how a lane proves a refusal
// or a hazard is explained in the operator's words.
func waitVaultSays(ns, reason string, budget time.Duration, fragments ...string) componentStatus {
	var got componentStatus
	Eventually(func(g Gomega) {
		got = readComponent(ns, vaultComponent)
		if reason != "" {
			g.Expect(got.Reason).To(Equal(reason), "the vault reads %s (%s): %s", got.Phase, got.Reason, got.Message)
		}
		for _, fragment := range fragments {
			g.Expect(got.Message).To(ContainSubstring(fragment), "the vault says: %s", got.Message)
		}
	}, budget, 10*time.Second).Should(Succeed())
	return got
}

// controlPlaneAccessorAnnotation reads the accessor the control-plane
// Deployment's pod template carries -- what rolls the control plane when
// the operator re-issues its token.
func controlPlaneAccessorAnnotation(ns string) (string, error) {
	key := strings.ReplaceAll(resources.ControlPlaneVaultTokenAccessorAnnotation, ".", "\\.")
	out, err := kubectl("get", "deployment", resources.ControlPlaneDeploymentName(platformName), "-n", ns,
		"-o", "jsonpath={.spec.template.metadata.annotations."+key+"}")
	return strings.TrimSpace(out), err
}

// deploymentJSONPath reads one jsonpath off the control-plane Deployment.
func controlPlaneDeploymentJSONPath(ns, jsonPath string) string {
	out, err := kubectl("get", "deployment", resources.ControlPlaneDeploymentName(platformName), "-n", ns,
		"-o", "jsonpath="+jsonPath)
	Expect(err).NotTo(HaveOccurred(), "reading the control-plane Deployment")
	return strings.TrimSpace(out)
}

// managerLogsSince returns the manager's log after the given line offset:
// what the operator wrote during one lane, not the ones before it.
func managerLogsSince(offset int) string {
	lines := strings.Split(managerLogs(), "\n")
	if offset > len(lines) {
		offset = len(lines)
	}
	return strings.Join(lines[offset:], "\n")
}

// managerLogLines is the manager's log length now, for managerLogsSince.
func managerLogLines() int { return len(strings.Split(managerLogs(), "\n")) }

// applyDeclaration renders and applies a lane's platform declaration.
func applyDeclaration(o platformOptions, what string) {
	mustApply(platformDeclaration(o), what)
}

// backupJSONPath reads one jsonpath under status.backup.
func backupJSONPath(ns, path string) string {
	return platformJSONPath(ns, "{.status.backup"+path+"}")
}

// waitBackupState waits for status.backup.state.
func waitBackupState(ns string, want v1.BackupState, budget time.Duration) {
	Eventually(func(g Gomega) {
		state := backupJSONPath(ns, ".state")
		_, _ = fmt.Fprintf(GinkgoWriter, "  %s/%s backup=%s :: %s\n", ns, platformName, state, backupJSONPath(ns, ".message"))
		g.Expect(state).To(Equal(string(want)))
	}, budget, 15*time.Second).Should(Succeed())
}

// deletePlatform deletes the lane's platform and waits for its owned
// workloads to drain (the operator has no finalizers; garbage collection
// does the work), then removes any volume claim the database left behind
// so the restored cluster starts from the archive and never from a disk.
func deletePlatform(ns string) {
	_, err := kubectl("delete", "plantonplatform", platformName, "-n", ns, "--wait=true", "--timeout=5m")
	Expect(err).NotTo(HaveOccurred(), "deleting the platform in %s", ns)
	Eventually(func(g Gomega) {
		out, _ := kubectl("get", "pods", "-n", ns, "-o", "name")
		g.Expect(strings.TrimSpace(out)).To(BeEmpty(), "pods still draining in %s: %s", ns, out)
	}, 5*time.Minute, 10*time.Second).Should(Succeed())
	_, _ = kubectl("delete", "pvc", "--all", "-n", ns, "--wait=true", "--timeout=3m")
}

// vaultPod is the vault server's pod: the chart's fullnameOverride makes the
// StatefulSet the release name, so its one pod is "<release>-0".
func vaultPod() string { return resources.OpenBAOServiceAccountName(platformName) + "-0" }

// vaultExec runs the bao CLI inside the vault's own pod with the given token
// -- the pod already carries BAO_ADDR for its own probes -- and returns the
// command's output. Speaking to the vault from inside its pod with the
// control plane's token is the lane acting as the control plane: same
// address, same credential, same policy.
func vaultExec(ns, token string, args ...string) (string, error) {
	full := make([]string, 0, 10+len(args))
	full = append(full, "exec", "-n", ns, vaultPod(), "-c", "openbao", "--", "env", "BAO_TOKEN="+token, "bao")
	full = append(full, args...)
	return kubectl(full...)
}

// mustVaultExec is vaultExec that fails the spec on error.
func mustVaultExec(ns, token string, args ...string) string {
	out, err := vaultExec(ns, token, args...)
	Expect(err).NotTo(HaveOccurred(), "bao %s: %s", strings.Join(args, " "), out)
	return out
}

// controlPlaneToken reads the token the operator minted for the control
// plane, and its accessor, from the Secret the control plane reads.
func controlPlaneToken(ns string) (token, accessor string) {
	name := resources.OpenBAOTokenSecretName(platformName)
	return secretValue(ns, name, resources.OpenBAOTokenSecretTokenKey),
		secretValue(ns, name, resources.OpenBAOTokenSecretAccessorKey)
}

// baoJSON runs a bao command with -format=json and returns its "data".
func baoJSON(ns, token string, args ...string) map[string]any {
	out := mustVaultExec(ns, token, append(args, "-format=json")...)
	var parsed struct {
		Data map[string]any `json:"data"`
	}
	Expect(json.Unmarshal([]byte(out), &parsed)).To(Succeed(), "parsing bao output: %s", out)
	return parsed.Data
}

// A marker is a secret the lane writes the way the control plane writes
// one: KV v2 under the platform's secret/ mount.
func writeMarker(ns, token, name string) {
	mustVaultExec(ns, token, "kv", "put", "-mount=secret", "lane/"+name, "value="+name)
}

func readMarker(ns, token, name string) (string, error) {
	return vaultExec(ns, token, "kv", "get", "-mount=secret", "-field=value", "lane/"+name)
}

// signingKeyPublic returns the public material of the OIDC signing key's
// current version, or "" while the key does not exist yet.
func signingKeyPublic(ns, token string) string {
	out, err := vaultExec(ns, token, "read", "-format=json", "transit/keys/"+oidcSigningKey)
	if err != nil {
		return ""
	}
	var parsed struct {
		Data struct {
			LatestVersion int                       `json:"latest_version"`
			Keys          map[string]map[string]any `json:"keys"`
		} `json:"data"`
	}
	if json.Unmarshal([]byte(out), &parsed) != nil {
		return ""
	}
	current := parsed.Data.Keys[fmt.Sprint(parsed.Data.LatestVersion)]
	public, _ := current["public_key"].(string)
	return public
}

// signWithSigningKey signs a payload with the OIDC signing key and returns
// the signature.
func signWithSigningKey(ns, token, payload string) string {
	input := "input=" + base64.StdEncoding.EncodeToString([]byte(payload))
	data := baoJSON(ns, token, "write", "transit/sign/"+oidcSigningKey, input)
	signature, _ := data["signature"].(string)
	Expect(signature).NotTo(BeEmpty(), "the signing key must sign")
	return signature
}

// verifyWithSigningKey reports whether the signature verifies on the key as
// it is now -- after a restore, the proof that it is the same key.
func verifyWithSigningKey(ns, token, payload, signature string) bool {
	input := "input=" + base64.StdEncoding.EncodeToString([]byte(payload))
	data := baoJSON(ns, token, "write", "transit/verify/"+oidcSigningKey, input, "signature="+signature)
	valid, _ := data["valid"].(bool)
	return valid
}

// pgSQL runs one statement on the platform's database, on its primary pod.
func pgSQL(ns, database, sql string) string {
	primary := resources.PostgreSQLClusterName(platformName) + "-1"
	out, err := kubectl("exec", "-n", ns, primary, "-c", "postgres", "--",
		"psql", "-U", "postgres", "-d", database, "-tAc", sql)
	Expect(err).NotTo(HaveOccurred(), "psql: %s", out)
	return strings.TrimSpace(out)
}

// takeBaseBackup asks CloudNativePG for a base backup through the plugin
// and waits for it to complete: the point in time before which marker A was
// written and after which marker B is, so a restore that carries both has
// replayed the archive past the base.
func takeBaseBackup(ns, name string) {
	manifest := fmt.Sprintf(`apiVersion: postgresql.cnpg.io/v1
kind: Backup
metadata:
  name: %s
  namespace: %s
spec:
  cluster:
    name: %s
  method: plugin
  pluginConfiguration:
    name: %s
`, name, ns, resources.PostgreSQLClusterName(platformName), resources.BarmanCloudPluginName)
	mustApply([]byte(manifest), "the base backup")
	Eventually(func(g Gomega) {
		phase, _ := kubectl("get", "backup.postgresql.cnpg.io", name, "-n", ns, "-o", "jsonpath={.status.phase}")
		g.Expect(strings.TrimSpace(phase)).NotTo(Equal("failed"), "the base backup failed")
		g.Expect(strings.TrimSpace(phase)).To(Equal("completed"))
	}, backupBudget, 15*time.Second).Should(Succeed())
}

// switchWAL forces the current WAL segment out to the archive, so a marker
// written after the base backup is in the archive before the destroy.
func switchWAL(ns string) {
	pgSQL(ns, "postgres", "SELECT pg_switch_wal()")
	// The archiver ships the closed segment within its next cycle.
	time.Sleep(30 * time.Second)
}

// keyHolderExec runs the bao CLI on the key holder as its root.
func keyHolderExec(args ...string) (string, error) {
	full := append([]string{"exec", "-n", fixtures.KeyHolderNamespace, "deploy/key-holder", "--", "env",
		"BAO_ADDR=http://127.0.0.1:8200", "BAO_TOKEN=" + fixtures.KeyHolderRootToken, "bao"}, args...)
	return kubectl(full...)
}

// makeKeyHolder turns the dev-mode OpenBao into a key service: the transit
// engine mounted and the wrapping key created. Both idempotent.
func makeKeyHolder() {
	By("making the key holder a key service: the transit engine and the wrapping key")
	out, err := keyHolderExec("secrets", "enable", "-path="+fixtures.KeyHolderTransitMount, "transit")
	if err != nil && !strings.Contains(out, "path is already in use") {
		Fail(fmt.Sprintf("mounting transit on the key holder: %s", out))
	}
	out, err = keyHolderExec("write", "-f", fixtures.KeyHolderTransitMount+"/keys/"+fixtures.KeyHolderTransitKey)
	Expect(err).NotTo(HaveOccurred(), "creating the wrapping key: %s", out)
}

// dumpPlatformOnFailure adds the platform's status and the vault pod's log
// to the report when the current spec failed.
func dumpPlatformOnFailure(ns string) {
	if !CurrentSpecReport().Failed() {
		return
	}
	if out, err := kubectl("get", "plantonplatform", platformName, "-n", ns, "-o", "yaml"); err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Platform %s/%s:\n%s\n", ns, platformName, out)
	}
	if out, err := kubectl("get", "pods", "-n", ns, "-o", "wide"); err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Pods in %s:\n%s\n", ns, out)
	}
	if out, err := kubectl("logs", vaultPod(), "-n", ns, "-c", "openbao", "--tail=200"); err == nil {
		_, _ = fmt.Fprintf(GinkgoWriter, "Vault log (%s):\n%s\n", vaultPod(), out)
	}
}
