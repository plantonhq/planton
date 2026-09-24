//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
	"github.com/plantonhq/planton/operator/test/fixtures"
	"github.com/plantonhq/planton/operator/test/utils"
)

// The vault lanes prove, on a real cluster, the promise the platform makes
// about its bundled secrets manager: declare one backup, and on the bad day
// declare the platform again from its archive and every secret comes back
// -- with nobody carrying keys. Two teams, two lifecycles, each install ->
// seed -> destroy -> restore:
//
//   - a team with no cloud key, on the built-in seal, naming a Secret they
//     own for the vault's keys (and the runbook of recreating that Secret
//     from a copy they kept, rehearsed inside the lane);
//   - a team with a key service, on a transit seal against an in-cluster
//     key holder (the only auto-unseal a Kind cluster can run, and the
//     shape a cloud-sealed platform takes: the restored vault opens itself).
//
// The lane speaks to the vault as the control plane does -- with the token
// the operator minted, from inside the vault's pod -- and reads the platform
// the way a person does: status.components, status.backup, the Secrets,
// the operator's log. It gates on the vault's own facts and on per-component
// phases, never on the platform's overall phase: the console's published
// image line can lag the version this manager requires (E2E_CONSOLE_IMAGE_TAG
// pins it when it does), and the vault's promise does not depend on it. The
// control plane runs its real image, and the OIDC issuer's signing key it
// creates in the vault the moment it is live is both the proof that it
// started against the minted token and the platform's real root of trust: a
// signature made before the disaster verifying on the restored key IS "a
// keyless connection keeps verifying on the restored signing key".
//
// Selected by label: `make test-e2e-vault-restore` runs these alone with the
// budget two platform lifecycles need; `make test-e2e` excludes them.

const (
	// The lines the operator writes -- and the one it must never write --
	// while the lanes run, asserted on the manager's log.
	repairLogLine = "repairing the access arrangement with the root token"
	unsealLogLine = "OpenBAO unsealed from stored keys"
	mintedLogLine = "Minted the control plane's vault token"

	// The fragments of the operator's sentences the lanes assert; each one is
	// the operator's own text (component/openbao_backup_status.go,
	// openbao_init_secret.go, openbao_seal.go), read by fragment so a
	// rewording that keeps the meaning does not break the lane.
	backupRefusedByAPI        = "set vault.initSecretName to a Secret you own"
	sealedWithoutKeysFragment = "so the operator holds no keys to open it"
	keptCopyFragment          = "Recreate that Secret from the copy you kept"
	sealChangedFragment       = "cannot be changed on a running platform"
	sealStartCheckFragment    = "the seal check the server makes at start"
	sealStartLogLine          = "Error configuring seal"
	breakGlassGoneFragment    = "break-glass (its root token and recovery keys) is gone"
	signaturePayload          = "a token minted before the bad day"
	// The role's period in seconds -- what `token lookup` reports as the
	// minted token's creation_ttl.
	operatorTokenPeriodInSecond = float64(resources.OpenBAOControlPlaneTokenPeriod / time.Second)

	// The jsonpaths of the init Secret's self-description.
	sealAnnotationPath = "{.metadata.annotations.planton\\.ai/openbao-seal}"
	noteAnnotationPath = "{.metadata.annotations.planton\\.ai/openbao-init}"
	managedByLabelPath = "{.metadata.labels.app\\.kubernetes\\.io/managed-by}"
)

var _ = Describe("The bundled vault comes back from the archive", Ordered, Label("vault-restore"), func() {
	var runID string
	// currentLane is the namespace the running context owns, for the failure
	// dump.
	var currentLane string

	BeforeAll(func() {
		runID = time.Now().UTC().Format("20060102150405")
		deployManager()

		applyFixture(fixtures.MinIO(), "in-cluster object store")
		waitAvailable(fixtures.MinIONamespace, "minio", 5*time.Minute)
		waitJobComplete(fixtures.MinIONamespace, fixtures.MinIOBucketJob, 5*time.Minute)

		applyFixture(fixtures.KeyHolder(), "key holder")
		waitAvailable(fixtures.KeyHolderNamespace, "key-holder", 5*time.Minute)
		makeKeyHolder()
	})

	AfterAll(func() {
		undeployManager()
		deleteNamespace(fixtures.MinIONamespace)
		deleteNamespace(fixtures.KeyHolderNamespace)
	})

	AfterEach(func() {
		if CurrentSpecReport().Failed() {
			if pod, err := utils.Run(execKubectlManagerPod()); err == nil {
				dumpManagerOnFailure(strings.TrimSpace(pod))
			}
			if currentLane != "" {
				dumpPlatformOnFailure(currentLane)
			}
		}
	})

	// ── Lane A ───────────────────────────────────────────────────────────

	Context("a team with no cloud key: the built-in seal and a Secret they own", Ordered, func() {
		const ns = "planton-lane-shamir"
		var (
			sourceServer      string
			accessorBefore    string
			signingKeyBefore  string
			signature         string
			platformUID       string
			logOffsetAtLaneUp int
		)

		BeforeAll(func() {
			currentLane = ns
			createNamespace(ns)
			literalSecret(ns, laneBackupCredentials, map[string]string{
				resources.ObjectStoreKeyAccessKeyID:     fixtures.MinIOAccessKey,
				resources.ObjectStoreKeySecretAccessKey: fixtures.MinIOSecretKey,
			})
			// The GitOps shape: the Secret exists, empty, before the platform
			// does; the operator fills it and never owns it.
			emptySecret(ns, laneKeysSecret)
			logOffsetAtLaneUp = managerLogLines()
		})

		AfterAll(func() {
			_, _ = kubectl("delete", "plantonplatform", platformName, "-n", ns, "--ignore-not-found", "--wait=false")
			deleteNamespace(ns)
			currentLane = ""
		})

		It("refuses, at the API, a backup whose vault keys would die with the platform", func() {
			// The one contradiction the contract forbids: a backup declared
			// over a vault whose keys live in the operator's own Secret. The
			// definition's rule names both remedies in the refusal.
			declaration := platformDeclaration(platformOptions{namespace: ns, backupPrefix: runID + "/refused"})
			out, err := kubectlApply(declaration)
			Expect(err).To(HaveOccurred(), "a backup with neither a seal nor a keys Secret must be refused")
			Expect(out).To(ContainSubstring(backupRefusedByAPI))
		})

		It("installs: the vault stores in the platform's database and the operator fills the Secret the team owns", func() {
			applyDeclaration(platformOptions{
				namespace: ns, backupPrefix: runID + "/shamir", initSecretName: laneKeysSecret,
			}, "the platform")
			platformUID = platformJSONPath(ns, "{.metadata.uid}")
			Expect(platformUID).NotTo(BeEmpty())

			waitComponentReady(ns, postgresqlComponent, componentBudget)
			waitComponentReady(ns, vaultComponent, componentBudget)

			By("the vault has no volume of its own: the database is its storage")
			out, _ := kubectl("get", "pvc", "-n", ns, "-o", "name")
			for _, claim := range utils.GetNonEmptyLines(out) {
				Expect(claim).NotTo(ContainSubstring("openbao"), "the vault must not own a volume claim")
			}

			By("the team's Secret is filled in place, owned by nobody, stamped with its seal")
			Expect(secretDataKeys(ns, laneKeysSecret)).To(ConsistOf(
				resources.OpenBAOInitSecretUnsealKeysKey, resources.OpenBAOInitSecretRootTokenKey))
			var shares []string
			unsealKeys := secretValue(ns, laneKeysSecret, resources.OpenBAOInitSecretUnsealKeysKey)
			Expect(json.Unmarshal([]byte(unsealKeys), &shares)).To(Succeed())
			Expect(shares).To(HaveLen(resources.OpenBAOSecretShares))
			Expect(secretValue(ns, laneKeysSecret, resources.OpenBAOInitSecretRootTokenKey)).NotTo(BeEmpty())
			Expect(secretJSONPath(ns, laneKeysSecret, "{.metadata.ownerReferences}")).To(BeEmpty(),
				"an adopter-owned keys Secret carries no owner reference")
			Expect(secretJSONPath(ns, laneKeysSecret, sealAnnotationPath)).To(Equal(resources.OpenBAOSealShamir))
			Expect(secretJSONPath(ns, laneKeysSecret, noteAnnotationPath)).To(ContainSubstring("spec.vault.initSecretName"))
			Expect(secretJSONPath(ns, laneKeysSecret, managedByLabelPath)).To(Equal(resources.ManagedByLabel))

			By("status.backup.vault tells the truth for this posture")
			Expect(backupJSONPath(ns, ".vault.covered")).To(Equal("true"))
			Expect(backupJSONPath(ns, ".vault.seal")).To(Equal(resources.OpenBAOSealShamir))
			Expect(backupJSONPath(ns, ".vault.initSecretName")).To(Equal(laneKeysSecret))
		})

		It("signed in as itself through the cluster's own identity check, never with the root token", func() {
			By("the vault's ServiceAccount holds the cluster's token-review role through the platform's satellite")
			binding := resources.OpenBAOAuthDelegatorClusterRoleBindingName(ns, platformName)
			out, err := kubectl("get", "clusterrolebinding", binding, "-o", "jsonpath={.roleRef.name} {.subjects[0].kind} "+
				"{.subjects[0].name} {.subjects[0].namespace} {.metadata.labels.planton\\.ai/platform-uid}")
			Expect(err).NotTo(HaveOccurred(), "the auth-delegator binding %s must exist", binding)
			Expect(strings.Fields(out)).To(Equal([]string{"system:auth-delegator", "ServiceAccount",
				resources.OpenBAOServiceAccountName(platformName), ns, platformUID}))

			By("the operator's first sign-in was accepted: the real TokenReview, no repair")
			logs := managerLogs()
			Expect(logs).NotTo(ContainSubstring(repairLogLine))
			Expect(logs).To(ContainSubstring(mintedLogLine))
		})

		It("minted the control plane a token of its own, scoped to the platform's engines", func() {
			token, accessor := controlPlaneToken(ns)
			Expect(token).NotTo(BeEmpty())
			Expect(accessor).NotTo(BeEmpty())
			accessorBefore = accessor

			By("the token is what the role promises: orphan, renewable, the role's seven-day period, the control plane's policy")
			// A token minted through a token role carries no period of its
			// own: the server keeps the period on the role and reads it there
			// at every renewal, so a lookup names the role and shows the
			// period as the token's creation TTL, never as a `period` field.
			self := baoJSON(ns, token, "token", "lookup")
			Expect(self["orphan"]).To(BeTrue())
			Expect(self["renewable"]).To(BeTrue())
			Expect(self["role"]).To(Equal(resources.OpenBAOControlPlaneRoleName(platformName)))
			Expect(self["creation_ttl"]).To(BeNumerically("==", operatorTokenPeriodInSecond))
			Expect(self["policies"]).To(ContainElement(resources.OpenBAOControlPlaneRoleName(platformName)))
			Expect(self["policies"]).NotTo(ContainElement("root"))

			By("it reaches the platform's engines and nothing else")
			writeMarker(ns, token, "reach")
			value, err := readMarker(ns, token, "reach")
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.TrimSpace(value)).To(Equal("reach"))
			out, err := vaultExec(ns, token, "policy", "list")
			Expect(err).To(HaveOccurred(), "the control plane's token must not administer the vault")
			Expect(out).To(ContainSubstring("permission denied"))

			By("the control plane reads that Secret and rolls on its accessor; the init Secret is not in its environment")
			// The control plane's Deployment is rendered only once its own
			// dependencies are Ready -- the identity server among them,
			// minutes after the vault -- so it is waited for, never read cold.
			Eventually(func(g Gomega) {
				annotated, err := controlPlaneAccessorAnnotation(ns)
				g.Expect(err).NotTo(HaveOccurred(), "the control-plane Deployment is not rendered yet")
				g.Expect(annotated).To(Equal(accessor))
			}, componentBudget, 15*time.Second).Should(Succeed())
			tokenSource := controlPlaneDeploymentJSONPath(ns,
				`{.spec.template.spec.containers[0].env[?(@.name=="VAULT_TOKEN")].valueFrom.secretKeyRef.name}`)
			Expect(tokenSource).To(Equal(resources.OpenBAOTokenSecretName(platformName)))
			Expect(controlPlaneDeploymentJSONPath(ns, "{.spec.template}")).NotTo(ContainSubstring(laneKeysSecret),
				"no consumer reads the init Secret")
		})

		It("the real control plane started against that token: the OIDC issuer's signing key appears in the vault", func() {
			token, _ := controlPlaneToken(ns)
			Eventually(func() string {
				signingKeyBefore = signingKeyPublic(ns, token)
				return signingKeyBefore
			}, componentBudget, 15*time.Second).ShouldNot(BeEmpty(),
				"the control plane creates %s the moment it is live", oidcSigningKey)
			signature = signWithSigningKey(ns, token, signaturePayload)
			Expect(verifyWithSigningKey(ns, token, signaturePayload, signature)).To(BeTrue())
		})

		It("archives: a secret before the base backup and one after it", func() {
			token, _ := controlPlaneToken(ns)
			waitBackupState(ns, v1.BackupStateHealthy, componentBudget)
			writeMarker(ns, token, "marker-a")
			takeBaseBackup(ns, "lane-seed-"+runID)
			writeMarker(ns, token, "marker-b")
			switchWAL(ns)
			sourceServer = backupJSONPath(ns, ".serverName")
			Expect(sourceServer).NotTo(BeEmpty(), "the source platform records the archive's server name")
		})

		It("the bad day: the platform is deleted and the Secret the team owns stands", func() {
			// The copy "kept outside the cluster": the runbook's precondition,
			// taken while the platform is still up.
			copySecret(ns, laneKeysSecret, laneKeysSecretCopy)
			deletePlatform(ns)

			By("the team's Secret survived; the operator's own Secrets did not")
			Expect(secretExists(ns, laneKeysSecret)).To(BeTrue(), "the adopter-owned keys Secret must outlive the platform")
			Expect(secretDataKeys(ns, laneKeysSecret)).To(ConsistOf(
				resources.OpenBAOInitSecretUnsealKeysKey, resources.OpenBAOInitSecretRootTokenKey))
			tokenSecret := resources.OpenBAOTokenSecretName(platformName)
			Eventually(func() bool { return secretExists(ns, tokenSecret) }, 3*time.Minute, 5*time.Second).
				Should(BeFalse(), "the control plane's token Secret is the platform's and goes with it")
		})

		It("declared again from the archive without the Secret: refused, with the Secret and the archive named", func() {
			// The team forgot the first step of the runbook. The database
			// comes back from the archive, the vault finds itself initialized
			// and sealed, and the operator refuses in words instead of
			// leaving a sealed vault and a Ready column.
			deleteSecret(ns, laneKeysSecret)
			applyDeclaration(platformOptions{
				namespace: ns, backupPrefix: runID + "/shamir-restored", initSecretName: laneKeysSecret,
				recoverFromServer: sourceServer, recoverFromPrefix: runID + "/shamir",
			}, "the restore declaration")
			waitComponentReady(ns, postgresqlComponent, restoreBudget)
			refused := waitVaultSays(ns, string(v1.ComponentReasonConfigurationRefused), sentenceBudget,
				"came back from archive "+sourceServer, laneKeysSecret, sealedWithoutKeysFragment, keptCopyFragment)
			Expect(refused.Object).To(Equal(laneKeysSecret))
		})

		It("the Secret recreated from the kept copy: the vault unseals and every secret is back", func() {
			copySecret(ns, laneKeysSecretCopy, laneKeysSecret)
			waitComponentReady(ns, vaultComponent, componentBudget)

			By("the operator unsealed the restored vault from the team's Secret")
			Expect(managerLogsSince(logOffsetAtLaneUp)).To(ContainSubstring(unsealLogLine))
			Expect(managerLogs()).NotTo(ContainSubstring(repairLogLine))

			By("the control plane's token was minted again; the old accessor names nothing")
			token, accessor := controlPlaneToken(ns)
			Expect(accessor).NotTo(BeEmpty())
			Expect(accessor).NotTo(Equal(accessorBefore), "a restored platform gets a fresh token")

			By("both markers came back: the base backup AND the archive after it")
			for _, marker := range []string{"marker-a", "marker-b"} {
				value, err := readMarker(ns, token, marker)
				Expect(err).NotTo(HaveOccurred(), "marker %s must be readable after the restore", marker)
				Expect(strings.TrimSpace(value)).To(Equal(marker))
			}

			By("the signing key is the same key: the old signature verifies, the public material is identical")
			Expect(signingKeyPublic(ns, token)).To(Equal(signingKeyBefore))
			Expect(verifyWithSigningKey(ns, token, signaturePayload, signature)).To(BeTrue())

			By("the status says where the platform came from and what opens its vault")
			Expect(backupJSONPath(ns, ".restoredFrom")).To(Equal(sourceServer))
			Expect(backupJSONPath(ns, ".vault.covered")).To(Equal("true"))
			Expect(backupJSONPath(ns, ".vault.seal")).To(Equal(resources.OpenBAOSealShamir))
			Expect(backupJSONPath(ns, ".vault.initSecretName")).To(Equal(laneKeysSecret))
			waitBackupState(ns, v1.BackupStateHealthy, componentBudget)
			Expect(backupJSONPath(ns, ".serverName")).NotTo(Equal(sourceServer),
				"a restored platform archives under its own name")
		})

		It("a lost token is minted again and the control plane rolls to it", func() {
			_, before := controlPlaneToken(ns)
			deleteSecret(ns, resources.OpenBAOTokenSecretName(platformName))
			var after string
			Eventually(func(g Gomega) {
				_, after = controlPlaneToken(ns)
				g.Expect(after).NotTo(BeEmpty())
				g.Expect(after).NotTo(Equal(before))
			}, 3*time.Minute, 5*time.Second).Should(Succeed())
			Eventually(func(g Gomega) {
				annotated, err := controlPlaneAccessorAnnotation(ns)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(annotated).To(Equal(after), "the control plane rolls on the new accessor")
			}, 3*time.Minute, 5*time.Second).Should(Succeed())
		})
	})

	// ── Lane B ───────────────────────────────────────────────────────────

	Context("a team with a key service: the transit seal", Ordered, func() {
		const ns = "planton-lane-transit"
		var (
			sourceServer      string
			accessorBefore    string
			signingKeyBefore  string
			signature         string
			logOffsetAtLaneUp int
			// transitPlatform is the lane's declaration; built in BeforeAll,
			// never at tree construction, because runID does not exist until
			// the suite's BeforeAll runs -- a prefix built earlier would read
			// "/transit" and the restore would look for the archive under a
			// run id the source never wrote to.
			transitPlatform platformOptions
		)
		fingerprint := fmt.Sprintf("%s %s/%s/%s", resources.OpenBAOSealTransit,
			fixtures.KeyHolderAddress, fixtures.KeyHolderTransitMount, fixtures.KeyHolderTransitKey)

		BeforeAll(func() {
			currentLane = ns
			transitPlatform = platformOptions{
				namespace: ns, backupPrefix: runID + "/transit", initSecretName: laneKeysSecret,
				transitKey: fixtures.KeyHolderTransitKey,
			}
			createNamespace(ns)
			literalSecret(ns, laneBackupCredentials, map[string]string{
				resources.ObjectStoreKeyAccessKeyID:     fixtures.MinIOAccessKey,
				resources.ObjectStoreKeySecretAccessKey: fixtures.MinIOSecretKey,
			})
			logOffsetAtLaneUp = managerLogLines()
		})

		AfterAll(func() {
			_, _ = kubectl("delete", "plantonplatform", platformName, "-n", ns, "--ignore-not-found", "--wait=false")
			deleteNamespace(ns)
			currentLane = ""
		})

		It("a seal the vault cannot configure crash-loops, and the status names the check and the log line", func() {
			// The credentials Secret exists with its key, so the preflight
			// passes -- only the server can find out the token is wrong, and
			// it does so at start. The classifier names the crash-loop; the
			// seal hint names the check.
			literalSecret(ns, laneSealCredentials, map[string]string{"VAULT_TOKEN": "not-the-holder-token"})
			applyDeclaration(transitPlatform, "the platform")
			waitComponentReady(ns, postgresqlComponent, componentBudget)
			waitVaultSays(ns, string(v1.ComponentReasonCrashLooping), 10*time.Minute, sealStartCheckFragment, sealStartLogLine)
		})

		It("with the right token the vault initializes under a recovery quorum and opens itself", func() {
			literalSecret(ns, laneSealCredentials, map[string]string{"VAULT_TOKEN": fixtures.KeyHolderRootToken})
			// A container's Secret-backed environment is resolved when the
			// container is created; deleting the pod makes the corrected
			// token reach the server on the StatefulSet's next pod rather
			// than on the kubelet's next backoff.
			_, _ = kubectl("delete", "pod", vaultPod(), "-n", ns, "--ignore-not-found", "--wait=false")
			waitComponentReady(ns, vaultComponent, componentBudget)

			By("the keys Secret the team named was created for them: recovery keys, the root token, the seal recorded")
			Expect(secretDataKeys(ns, laneKeysSecret)).To(ConsistOf(
				resources.OpenBAOInitSecretRecoveryKeysKey, resources.OpenBAOInitSecretRootTokenKey))
			var recovery []string
			recoveryKeys := secretValue(ns, laneKeysSecret, resources.OpenBAOInitSecretRecoveryKeysKey)
			Expect(json.Unmarshal([]byte(recoveryKeys), &recovery)).To(Succeed())
			Expect(recovery).To(HaveLen(resources.OpenBAOSecretShares))
			Expect(secretJSONPath(ns, laneKeysSecret, "{.metadata.ownerReferences}")).To(BeEmpty())
			Expect(secretJSONPath(ns, laneKeysSecret, sealAnnotationPath)).To(Equal(fingerprint))

			By("the operator never unsealed anything: the seal opened the vault")
			Expect(managerLogsSince(logOffsetAtLaneUp)).NotTo(ContainSubstring(unsealLogLine))
			Expect(managerLogs()).NotTo(ContainSubstring(repairLogLine))
			Expect(backupJSONPath(ns, ".vault.seal")).To(Equal(resources.OpenBAOSealTransit))
		})

		It("a changed seal is refused before anything renders", func() {
			changed := transitPlatform
			changed.transitKey = "another-key"
			applyDeclaration(changed, "the changed declaration")
			refused := waitVaultSays(ns, string(v1.ComponentReasonConfigurationRefused), sentenceBudget,
				fingerprint, sealChangedFragment)
			Expect(refused.Object).To(Equal(laneKeysSecret))
			applyDeclaration(transitPlatform, "the restored declaration")
			waitComponentReady(ns, vaultComponent, sentenceBudget)
		})

		It("the real control plane's signing key appears; a secret before the base backup and one after", func() {
			token, accessor := controlPlaneToken(ns)
			Expect(accessor).NotTo(BeEmpty())
			accessorBefore = accessor
			Eventually(func() string {
				signingKeyBefore = signingKeyPublic(ns, token)
				return signingKeyBefore
			}, componentBudget, 15*time.Second).ShouldNot(BeEmpty())
			signature = signWithSigningKey(ns, token, signaturePayload)

			waitBackupState(ns, v1.BackupStateHealthy, componentBudget)
			writeMarker(ns, token, "marker-a")
			takeBaseBackup(ns, "lane-seed-"+runID)
			writeMarker(ns, token, "marker-b")
			switchWAL(ns)
			sourceServer = backupJSONPath(ns, ".serverName")
			Expect(sourceServer).NotTo(BeEmpty())
		})

		It("the bad day: the platform is deleted; the keys Secret and the key service stand", func() {
			copySecret(ns, laneKeysSecret, laneKeysSecretCopy)
			deletePlatform(ns)
			Expect(secretExists(ns, laneKeysSecret)).To(BeTrue())
		})

		It("declared again from the archive: the restored vault opens itself and every secret is back", func() {
			restore := transitPlatform
			restore.backupPrefix = runID + "/transit-restored"
			restore.recoverFromServer = sourceServer
			restore.recoverFromPrefix = runID + "/transit"
			applyDeclaration(restore, "the restore declaration")
			waitComponentReady(ns, postgresqlComponent, restoreBudget)
			waitComponentReady(ns, vaultComponent, componentBudget)

			By("no unseal call, no root token: the seal opened it and the operator signed in as itself")
			Expect(managerLogsSince(logOffsetAtLaneUp)).NotTo(ContainSubstring(unsealLogLine))
			Expect(managerLogs()).NotTo(ContainSubstring(repairLogLine))

			token, accessor := controlPlaneToken(ns)
			Expect(accessor).NotTo(Equal(accessorBefore))
			for _, marker := range []string{"marker-a", "marker-b"} {
				value, err := readMarker(ns, token, marker)
				Expect(err).NotTo(HaveOccurred(), "marker %s must be readable after the restore", marker)
				Expect(strings.TrimSpace(value)).To(Equal(marker))
			}
			Expect(signingKeyPublic(ns, token)).To(Equal(signingKeyBefore))
			Expect(verifyWithSigningKey(ns, token, signaturePayload, signature)).To(BeTrue())
			Expect(backupJSONPath(ns, ".restoredFrom")).To(Equal(sourceServer))
			Expect(backupJSONPath(ns, ".vault.seal")).To(Equal(resources.OpenBAOSealTransit))
			Expect(backupJSONPath(ns, ".vault.covered")).To(Equal("true"))
		})

		It("with the keys Secret gone the platform works and its sentence names what is gone; back, it is healthy", func() {
			// Under a cloud seal nothing running needs the Secret: it is the
			// break-glass. The operator says so on every pass instead of
			// refusing a working platform.
			deleteSecret(ns, laneKeysSecret)
			hazard := waitVaultSays(ns, "", sentenceBudget,
				"came back from archive "+sourceServer, laneKeysSecret, breakGlassGoneFragment)
			Expect(hazard.Phase).To(Equal(string(v1.ComponentPhaseReady)))

			copySecret(ns, laneKeysSecretCopy, laneKeysSecret)
			Eventually(func() string { return readComponent(ns, vaultComponent).Message }, sentenceBudget, 10*time.Second).
				Should(Equal("OpenBAO healthy"))
		})
	})
})

// execKubectlManagerPod is the manager pod lookup as a command, for the
// failure dump, where a Gomega assertion would mask the spec's own failure.
func execKubectlManagerPod() *exec.Cmd {
	return exec.Command("kubectl", "get", "pods", "-l", "control-plane=controller-manager", "-n", namespace,
		"-o", "jsonpath={.items[0].metadata.name}")
}
