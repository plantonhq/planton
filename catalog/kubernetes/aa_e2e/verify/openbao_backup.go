package verify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// The backup and restore proofs of the OpenBao verifier. Both ride on the
// module-owned objects the kind renders from a `backup` block — the
// `<name>-backup` ServiceAccount, policy, role, and CronJob, the
// `<name>-restore-<hash>` Job — and on the seed script of the restore lanes
// (the kind's `e2e/scenarios/backup-restore-seed.setup.sh`), which owns the
// names below that the two sides must agree on.
const (
	// openBaoBackupSuffix is the module's suffix for everything the backup
	// job is: its ServiceAccount, its policy, its auth role, its CronJob.
	openBaoBackupSuffix = "-backup"
	// openBaoRestoreJobPrefix precedes the hash in the restore Job's name.
	openBaoRestoreJobPrefix = "-restore-"
	// openBaoBackupAuthMount is the Kubernetes auth mount the recipe
	// enables (the spec default the module renders when `auth.mount_path`
	// is unset).
	openBaoBackupAuthMount = "kubernetes"
	// openBaoSeedKvMount and the two markers are what the seed script
	// writes into the SOURCE: marker-a BEFORE the snapshot, marker-b AFTER.
	// A restored target must carry exactly the first.
	openBaoSeedKvMount = "e2e-dr"
	openBaoSeedMarkerA = "marker-a"
	openBaoSeedMarkerB = "marker-b"
	// openBaoSourceRootTokenSecret is where the seed script keeps the
	// SOURCE's initial root token (key `token`) — the token the restored
	// state carries, and so the only one that can read it back.
	openBaoSourceRootTokenSecret = "e2e-bao-src-root-token"
	// openBaoTransitMount and openBaoTransitKey are the transit engine and
	// wrapping key the key-holder fixture serves; the satellite fixtures
	// and scenarios declare the same two names on their `transit` seal.
	openBaoTransitMount = "transit"
	openBaoTransitKey   = "openbao-e2e"
	// openBaoServerContainer is the chart's server container name, where
	// the `bao` CLI runs with BAO_ADDR already pointed at the local listener.
	openBaoServerContainer = "openbao"
	// openBaoUploadContainer is the CronJob's rclone container — the one
	// whose environment carries the store remote and paths, and therefore
	// the one an identity probe reuses.
	openBaoUploadContainer = "upload"
)

// proveTransitKeyHolder makes the dev-mode fixture a transit key holder and
// proves it serves as one: the `transit/` engine is enabled, the wrapping
// key exists, and an encrypt call answers with ciphertext. This is part of
// the fixture's verification on purpose: the satellite that follows it in
// the chain test-encrypts against this key at STARTUP and refuses to start
// until the engine exists, so a key holder that is merely running has not
// been proven to be one. Idempotent: an already-enabled mount and an
// existing key are the same success.
func (v *OpenBaoVerifier) proveTransitKeyHolder(ctx context.Context, kubeconfig, token string) error {
	pod0 := v.Name + "-0"
	if out, err := v.baoExec(ctx, kubeconfig, pod0, token, nil, "secrets", "enable", "-path="+openBaoTransitMount, "transit"); err != nil {
		if !strings.Contains(out, "path is already in use") {
			return errors.Wrapf(err, "enabling the transit engine on the key holder: %s", firstLines(out, 3))
		}
	}
	if out, err := v.baoExec(ctx, kubeconfig, pod0, token, nil, "write", "-f", openBaoTransitMount+"/keys/"+openBaoTransitKey); err != nil {
		return errors.Wrapf(err, "creating the transit wrapping key: %s", firstLines(out, 3))
	}
	plaintext := base64.StdEncoding.EncodeToString([]byte("key-holder-proof"))
	out, err := v.baoExec(ctx, kubeconfig, pod0, token, nil, "write", "-field=ciphertext", openBaoTransitMount+"/encrypt/"+openBaoTransitKey, "plaintext="+plaintext)
	if err != nil {
		return errors.Wrapf(err, "encrypting through the transit key: %s", firstLines(out, 3))
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "vault:v") {
		return errors.Errorf("the transit encrypt call returned no ciphertext: %s", firstLines(out, 2))
	}
	fmt.Printf("  [verify] KEY HOLDER: %s/ enabled, key %q present, encrypt answers ciphertext — the satellites' seal is ready\n", openBaoTransitMount, openBaoTransitKey)
	return nil
}

// proveBackup is THE BACKUP PROOF, after the seal lifecycle and the KV
// round-trip:
//
//  1. The four-command login recipe runs EXACTLY as the backup job prints
//     it — the same policy text, mount, role, and bindings — through the
//     server's own `bao` CLI. The recipe is the one step the module cannot
//     take (OpenBao is sealed at deploy), so the GUIDE's text is what is
//     proven, not a paraphrase of it.
//  2. One run is created from the backup CronJob and waited to completion;
//     its logs are printed on failure (they name the next step).
//  3. The store is listed FROM INSIDE THE CLUSTER, with the job's own
//     identity and remote configuration (see identityJob), and a snapshot
//     object under the prefix is asserted — independent of the run's own
//     claim of success.
//  4. Retention: an object dated 2001 is seeded under the prefix with the
//     same identity, a second run is created, and the listing must show the
//     stale object gone and the fresh snapshots kept. Counts are tolerant:
//     the scenario keeps a customer-shaped hourly schedule that may fire
//     mid-proof.
//
// Everything the proof creates that no engine owns (the two runs, the
// probe Jobs) is deleted before it returns.
func (v *OpenBaoVerifier) proveBackup(ctx context.Context, kubeconfig, rootToken string) error {
	backupName := v.Name + openBaoBackupSuffix
	pod0 := v.Name + "-0"

	// ---- 1. the login recipe, verbatim ----------------------------------
	fmt.Printf("  [verify] BACKUP PROOF: running the login recipe (policy %s, auth/%s, role %s -> %s/%s)\n",
		backupName, openBaoBackupAuthMount, backupName, v.Namespace, backupName)
	policy := "path \"sys/storage/raft/snapshot\" { capabilities = [\"read\"] }\n"
	if out, err := v.baoExec(ctx, kubeconfig, pod0, rootToken, strings.NewReader(policy), "policy", "write", backupName, "-"); err != nil {
		return errors.Wrapf(err, "recipe step 1 (policy write): %s", firstLines(out, 3))
	}
	if out, err := v.baoExec(ctx, kubeconfig, pod0, rootToken, nil, "auth", "enable", "-path="+openBaoBackupAuthMount, "kubernetes"); err != nil {
		if !strings.Contains(out, "path is already in use") {
			return errors.Wrapf(err, "recipe step 2 (auth enable): %s", firstLines(out, 3))
		}
	}
	if out, err := v.baoExec(ctx, kubeconfig, pod0, rootToken, nil, "write", "auth/"+openBaoBackupAuthMount+"/config",
		"kubernetes_host=https://kubernetes.default.svc:443"); err != nil {
		return errors.Wrapf(err, "recipe step 3 (auth config): %s", firstLines(out, 3))
	}
	if out, err := v.baoExec(ctx, kubeconfig, pod0, rootToken, nil, "write", "auth/"+openBaoBackupAuthMount+"/role/"+backupName,
		"bound_service_account_names="+backupName,
		"bound_service_account_namespaces="+v.Namespace,
		"token_policies="+backupName, "token_ttl=1h"); err != nil {
		return errors.Wrapf(err, "recipe step 4 (role): %s", firstLines(out, 3))
	}
	fmt.Printf("  [verify] BACKUP PROOF: recipe applied\n")

	// ---- 2. one run from the CronJob --------------------------------------
	run1 := backupName + "-e2e-1"
	run2 := backupName + "-e2e-2"
	defer func() {
		_ = kubectlDeleteResource(ctx, kubeconfig, "job", run1, v.Namespace)
		_ = kubectlDeleteResource(ctx, kubeconfig, "job", run2, v.Namespace)
	}()
	if err := v.runBackupOnce(ctx, kubeconfig, backupName, run1); err != nil {
		return err
	}

	// ---- 3. the object, seen with the job's own identity -------------------
	listing, err := v.identityJob(ctx, kubeconfig, "probe-1", `rclone lsf "$STORE_PATH" --files-only`)
	if err != nil {
		return errors.Wrap(err, "listing the store with the backup job's identity")
	}
	snapshots := snapshotObjects(listing)
	if len(snapshots) == 0 {
		return errors.Errorf("no .snap object under the backup prefix after a completed run; the store listing was:\n%s", firstLines(listing, 10))
	}
	fmt.Printf("  [verify] BACKUP PROOF: %d snapshot object(s) under the prefix, listed from inside the cluster with the job's identity (%s)\n", len(snapshots), snapshots[0])

	// ---- 4. retention ------------------------------------------------------
	stale := v.Name + "-20010101T000000Z.snap"
	seeded, err := v.identityJob(ctx, kubeconfig, "seed-stale",
		`rclone touch --timestamp 2001-01-01T00:00:00 "$STORE_PATH/`+stale+`" && rclone lsf "$STORE_PATH" --files-only`)
	if err != nil {
		return errors.Wrap(err, "seeding a stale snapshot object")
	}
	if !containsLine(seeded, stale) {
		return errors.Errorf("the seeded stale object %s is not in the listing:\n%s", stale, firstLines(seeded, 10))
	}
	fmt.Printf("  [verify] BACKUP PROOF: stale object %s seeded (dated 2001)\n", stale)

	if err := v.runBackupOnce(ctx, kubeconfig, backupName, run2); err != nil {
		return err
	}
	after, err := v.identityJob(ctx, kubeconfig, "probe-2", `rclone lsf "$STORE_PATH" --files-only`)
	if err != nil {
		return errors.Wrap(err, "listing the store after the second run")
	}
	if containsLine(after, stale) {
		return errors.Errorf("retention did not prune the stale object %s (retention_days on this scenario is 1); the listing was:\n%s", stale, firstLines(after, 10))
	}
	kept := snapshotObjects(after)
	if len(kept) < 2 {
		return errors.Errorf("expected at least the two fresh snapshots to survive pruning, found %d:\n%s", len(kept), firstLines(after, 10))
	}
	fmt.Printf("  [verify] BACKUP PROOF: retention pruned the stale object and kept %d fresh snapshot(s)\n", len(kept))
	return nil
}

// runBackupOnce creates one Job from the backup CronJob — the same command
// the recipe tells an operator to run — and waits for it to complete,
// printing both containers' logs when it does not (they name the next step).
func (v *OpenBaoVerifier) runBackupOnce(ctx context.Context, kubeconfig, cronJob, jobName string) error {
	_ = kubectlDeleteResource(ctx, kubeconfig, "job", jobName, v.Namespace)
	if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "-n", v.Namespace,
		"create", "job", "--from=cronjob/"+cronJob, jobName).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "creating a run from cronjob/%s: %s", cronJob, firstLines(string(out), 3))
	}
	fmt.Printf("  [verify] BACKUP PROOF: run %s created from cronjob/%s\n", jobName, cronJob)
	succeeded, err := waitForJobFinished(ctx, kubeconfig, v.Namespace, jobName, 10*time.Minute)
	if err != nil {
		return err
	}
	if !succeeded {
		return errors.Errorf("backup run %s failed; its logs:\n%s", jobName, jobLogs(ctx, kubeconfig, v.Namespace, jobName, ""))
	}
	fmt.Printf("  [verify] BACKUP PROOF: run %s completed\n", jobName)
	return nil
}

// identityJob runs one shell command WITH THE BACKUP JOB'S OWN IDENTITY: it
// takes the backup CronJob's pod template as the cluster holds it — the
// `<name>-backup` ServiceAccount and its cloud-identity annotation, the
// rclone remote configuration, the credentials Secret, the store paths —
// drops the snapshot init container, and replaces the rclone container's
// command. The result is an independent read of the store through exactly
// the credentials or Workload Identity the product declared, on every arm
// alike (a keyless GCS arm included), never through whatever the machine
// running the verifier happens to be logged in as. Both engines render the
// same CronJob, so one helper serves both. Returns the container's log; the
// Job is deleted before returning.
func (v *OpenBaoVerifier) identityJob(ctx context.Context, kubeconfig, suffix, command string) (string, error) {
	cronJob := v.Name + openBaoBackupSuffix
	raw, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "-n", v.Namespace,
		"get", "cronjob", cronJob, "-o", "json").Output()
	if err != nil {
		return "", errors.Wrapf(err, "reading cronjob %s", cronJob)
	}
	var cj map[string]interface{}
	if err := json.Unmarshal(raw, &cj); err != nil {
		return "", errors.Wrapf(err, "parsing cronjob %s", cronJob)
	}
	jobSpec, ok := nestedMap(cj, "spec", "jobTemplate", "spec")
	if !ok {
		return "", errors.Errorf("cronjob %s has no jobTemplate.spec", cronJob)
	}
	podSpec, ok := nestedMap(jobSpec, "template", "spec")
	if !ok {
		return "", errors.Errorf("cronjob %s has no pod template", cronJob)
	}
	delete(podSpec, "initContainers")
	containers, _ := podSpec["containers"].([]interface{})
	var upload map[string]interface{}
	for _, c := range containers {
		if m, ok := c.(map[string]interface{}); ok && m["name"] == openBaoUploadContainer {
			upload = m
		}
	}
	if upload == nil {
		return "", errors.Errorf("cronjob %s has no %q container to borrow the store identity from", cronJob, openBaoUploadContainer)
	}
	upload["command"] = []interface{}{"sh", "-c", command}
	delete(upload, "args")
	podSpec["containers"] = []interface{}{upload}
	// A probe is one attempt: its failure is the answer, not something to
	// retry into a different one.
	jobSpec["backoffLimit"] = 0
	delete(jobSpec, "activeDeadlineSeconds")

	jobName := cronJob + "-" + suffix
	job := map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata":   map[string]interface{}{"name": jobName, "namespace": v.Namespace},
		"spec":       jobSpec,
	}
	manifest, err := json.Marshal(job)
	if err != nil {
		return "", errors.Wrap(err, "building the identity probe Job")
	}
	_ = kubectlDeleteResource(ctx, kubeconfig, "job", jobName, v.Namespace)
	if err := kubectlApplyStdin(ctx, kubeconfig, string(manifest)); err != nil {
		return "", errors.Wrapf(err, "creating the identity probe Job %s", jobName)
	}
	defer func() { _ = kubectlDeleteResource(ctx, kubeconfig, "job", jobName, v.Namespace) }()

	succeeded, err := waitForJobFinished(ctx, kubeconfig, v.Namespace, jobName, 5*time.Minute)
	logs := jobLogs(ctx, kubeconfig, v.Namespace, jobName, openBaoUploadContainer)
	if err != nil {
		return logs, err
	}
	if !succeeded {
		return logs, errors.Errorf("identity probe %s failed (%s); its log:\n%s", jobName, command, firstLines(logs, 10))
	}
	return logs, nil
}

// proveRestore is THE RESTORE PROOF. The target was deployed with the
// source's backup store and a `restore` block; the seed script has already
// initialized the source, run the recipe, written marker A, taken the
// snapshot, and written marker B. Here:
//
//  1. The target's backup CronJob is asserted SUSPENDED — a vault in
//     restore mode takes no snapshots of its own (it would snapshot an
//     empty vault into the shared prefix, prune the source's, and let
//     `latest` pick itself). This is that rule's live proof.
//  2. The target is initialized (recovery shares — a declared restore
//     requires auto-unseal, so no unseal step) and its initial root token
//     is placed in the Secret the spec names; the waiting restore Job
//     picks it up.
//  3. The restore Job completes; its logs are printed when it does not.
//  4. The restored state is read through the `-active` Service with the
//     SOURCE's root token: marker A present, marker B absent, and the
//     target's own init token answers 403 — the state, and every token in
//     it, is the source's now.
//  5. Pod 0 is replaced and must come back Ready with no human: the seal
//     that made the restore a single call also makes the restart silent.
//
// The Secret the verifier created is deleted before returning; the Job is
// the engine's and stays (a vanished Job would be recreated on the next
// apply and restore AGAIN).
func (v *OpenBaoVerifier) proveRestore(ctx context.Context, kubeconfig string) error {
	if !v.AutoUnseal {
		return errors.New("a declared restore requires an auto_unseal seal on the target (the spec rule refuses anything else) — this scenario declares none")
	}
	if v.RootTokenSecretName == "" || v.RootTokenSecretKey == "" {
		return errors.New("the scenario's restore.root_token names no Secret and key — the verifier has nowhere to place the target's initial root token")
	}
	backupName := v.Name + openBaoBackupSuffix

	// ---- 1. restore mode suspends backups ----------------------------------
	suspend, err := kubectlGetJSONPath(ctx, kubeconfig, "cronjob", backupName, v.Namespace, "{.spec.suspend}")
	if err != nil {
		return errors.Wrapf(err, "reading cronjob %s", backupName)
	}
	if strings.TrimSpace(suspend) != "true" {
		return errors.Errorf("cronjob %s is not suspended while restore is declared (spec.suspend=%q) — a target in restore mode would snapshot an empty vault into the source's prefix", backupName, strings.TrimSpace(suspend))
	}
	fmt.Printf("  [verify] RESTORE PROOF: cronjob %s is suspended while restore is declared\n", backupName)

	// ---- 2. init the target, hand over the token ----------------------------
	initToken, _, err := v.bootstrap(ctx, kubeconfig)
	if err != nil {
		return err
	}
	if err := v.waitForReadyPods(ctx, kubeconfig, 5*time.Minute); err != nil {
		return errors.Wrap(err, "the target never became Ready after init — the auto-unseal seal did not unseal it")
	}
	_ = kubectlDeleteResource(ctx, kubeconfig, "secret", v.RootTokenSecretName, v.Namespace)
	if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "-n", v.Namespace,
		"create", "secret", "generic", v.RootTokenSecretName, "--from-literal="+v.RootTokenSecretKey+"="+initToken).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "creating the root-token Secret %s: %s", v.RootTokenSecretName, firstLines(string(out), 2))
	}
	defer func() { _ = kubectlDeleteResource(ctx, kubeconfig, "secret", v.RootTokenSecretName, v.Namespace) }()
	fmt.Printf("  [verify] RESTORE PROOF: target initialized and unsealed by its seal; initial root token placed in Secret %s/%s\n", v.Namespace, v.RootTokenSecretName)

	// ---- 3. the restore Job completes ---------------------------------------
	jobName, err := v.findRestoreJob(ctx, kubeconfig, 2*time.Minute)
	if err != nil {
		return err
	}
	succeeded, err := waitForJobFinished(ctx, kubeconfig, v.Namespace, jobName, 15*time.Minute)
	if err != nil {
		return errors.Wrapf(err, "waiting for the restore Job %s", jobName)
	}
	if !succeeded {
		return errors.Errorf("restore Job %s failed; its logs:\n%s", jobName, jobLogs(ctx, kubeconfig, v.Namespace, jobName, ""))
	}
	fmt.Printf("  [verify] RESTORE PROOF: Job %s completed\n", jobName)

	// ---- 4. the restored state, read with the source's token ----------------
	sourceToken, err := readSecretKey(ctx, kubeconfig, v.Namespace, openBaoSourceRootTokenSecret, "token")
	if err != nil {
		return errors.Wrapf(err, "reading the source's root token from Secret %s (the seed script keeps it there)", openBaoSourceRootTokenSecret)
	}
	readMarkers := func(base string) error {
		_, body, err := v.httpOnce(ctx, http.MethodGet, base+"/v1/"+openBaoSeedKvMount+"/data/"+openBaoSeedMarkerA, "", sourceToken, 6*time.Minute, 200)
		if err != nil {
			return errors.Wrap(err, "reading marker A (written BEFORE the snapshot) from the restored vault")
		}
		var resp struct {
			Data struct {
				Data map[string]string `json:"data"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(body), &resp); err != nil {
			return errors.Wrap(err, "parsing marker A")
		}
		if !strings.HasPrefix(resp.Data.Data["value"], openBaoSeedMarkerA) {
			return errors.Errorf("marker A came back as %q", resp.Data.Data["value"])
		}
		fmt.Printf("  [verify] RESTORE PROOF: %s present in the restored vault — the snapshot's state came back\n", openBaoSeedMarkerA)

		if _, _, err := v.httpOnce(ctx, http.MethodGet, base+"/v1/"+openBaoSeedKvMount+"/data/"+openBaoSeedMarkerB, "", sourceToken, 30*time.Second, 404); err != nil {
			return errors.Wrapf(err, "marker B (written AFTER the snapshot) must be absent from the restored vault")
		}
		fmt.Printf("  [verify] RESTORE PROOF: %s absent — nothing past the snapshot leaked in\n", openBaoSeedMarkerB)

		if _, _, err := v.httpOnce(ctx, http.MethodGet, base+"/v1/auth/token/lookup-self", "", initToken, 30*time.Second, 403); err != nil {
			return errors.Wrap(err, "the target's own initial root token must be dead after the restore (the state, and every token in it, is the source's)")
		}
		fmt.Printf("  [verify] RESTORE PROOF: the target's initial root token no longer exists — exactly as the Job's closing message says\n")
		return nil
	}
	if err := v.withServicePortForward(ctx, kubeconfig, v.Name+"-active", readMarkers); err != nil {
		return err
	}

	// ---- 5. the seal survives a restart alone -------------------------------
	if err := v.replacePod0(ctx, kubeconfig, ""); err != nil {
		return err
	}
	err = v.withServicePortForward(ctx, kubeconfig, v.Name+"-active", func(base string) error {
		_, _, err := v.httpOnce(ctx, http.MethodGet, base+"/v1/"+openBaoSeedKvMount+"/data/"+openBaoSeedMarkerA, "", sourceToken, 6*time.Minute, 200)
		return errors.Wrap(err, "re-reading marker A after the pod replacement")
	})
	if err != nil {
		return err
	}
	fmt.Printf("  [verify] RESTORE PROOF: marker A read back after pod replacement — the restored state is on disk and the seal unseals it alone\n")
	return nil
}

// findRestoreJob locates the module's restore Job by its name shape,
// `<name>-restore-<hash>` (the hash is of the restore declaration; the
// verifier does not recompute it). Polls because the engine creates the Job
// without awaiting it.
func (v *OpenBaoVerifier) findRestoreJob(ctx context.Context, kubeconfig string, budget time.Duration) (string, error) {
	prefix := v.Name + openBaoRestoreJobPrefix
	deadline := time.Now().Add(budget)
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "-n", v.Namespace,
			"get", "jobs", "-o", "jsonpath={.items[*].metadata.name}").Output()
		if err == nil {
			for _, name := range strings.Fields(string(out)) {
				if strings.HasPrefix(name, prefix) {
					return name, nil
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
	return "", errors.Errorf("no Job named %s* in namespace %s — the module renders one when restore is declared", prefix, v.Namespace)
}

// baoExec runs one `bao` command inside the server pod (the chart points
// BAO_ADDR at the local listener there), with the token as an environment
// variable for that one process and optional stdin. Returns the combined
// output; the token is never printed.
func (v *OpenBaoVerifier) baoExec(ctx context.Context, kubeconfig, pod, token string, stdin *strings.Reader, args ...string) (string, error) {
	kargs := []string{"--kubeconfig", kubeconfig, "-n", v.Namespace, "exec"}
	if stdin != nil {
		kargs = append(kargs, "-i")
	}
	kargs = append(kargs, pod, "-c", openBaoServerContainer, "--", "env", "BAO_TOKEN="+token, "bao")
	kargs = append(kargs, args...)
	cmd := exec.CommandContext(ctx, "kubectl", kargs...)
	if stdin != nil {
		cmd.Stdin = stdin
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// waitForJobFinished polls a Job until it succeeds or fails within the
// budget. A Job still waiting on the token Secret sits in neither state,
// which is why the budget is the caller's.
func waitForJobFinished(ctx context.Context, kubeconfig, namespace, name string, budget time.Duration) (bool, error) {
	deadline := time.Now().Add(budget)
	var last string
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "-n", namespace,
			"get", "job", name, "-o", "jsonpath={.status.succeeded} {.status.failed}").Output()
		if err == nil {
			last = strings.TrimSpace(string(out))
			fields := strings.Fields(last)
			if len(fields) >= 1 && fields[0] != "" && fields[0] != "0" {
				return true, nil
			}
			if len(fields) == 2 && fields[1] != "" && fields[1] != "0" {
				return false, nil
			}
		}
		time.Sleep(10 * time.Second)
	}
	return false, errors.Errorf("job %s/%s neither succeeded nor failed within %s (last status %q); its logs:\n%s",
		namespace, name, budget, last, jobLogs(ctx, kubeconfig, namespace, name, ""))
}

// jobLogs returns a Job's pod logs — one container, or every container when
// container is empty — for the failure messages that must name the cause.
func jobLogs(ctx context.Context, kubeconfig, namespace, name, container string) string {
	args := []string{"--kubeconfig", kubeconfig, "-n", namespace, "logs", "job/" + name}
	if container != "" {
		args = append(args, "-c", container)
	} else {
		args = append(args, "--all-containers=true")
	}
	out, err := exec.CommandContext(ctx, "kubectl", args...).CombinedOutput()
	if err != nil && len(out) == 0 {
		return fmt.Sprintf("(logs unavailable: %v)", err)
	}
	return string(out)
}

// readSecretKey reads and decodes one key of a Secret.
func readSecretKey(ctx context.Context, kubeconfig, namespace, name, key string) (string, error) {
	b64, err := kubectlGetJSONPath(ctx, kubeconfig, "secret", name, namespace, "{.data."+key+"}")
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return "", errors.Wrapf(err, "decoding key %q of secret %s/%s", key, namespace, name)
	}
	return strings.TrimSpace(string(raw)), nil
}

// snapshotObjects returns the `.snap` names in an `rclone lsf` listing.
func snapshotObjects(listing string) []string {
	var out []string
	for _, line := range strings.Split(listing, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ".snap") {
			out = append(out, line)
		}
	}
	return out
}

// containsLine reports whether a listing has an exact line.
func containsLine(listing, want string) bool {
	for _, line := range strings.Split(listing, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

// nestedMap walks map keys, returning the map at the end of the path.
func nestedMap(m map[string]interface{}, keys ...string) (map[string]interface{}, bool) {
	cur := m
	for _, k := range keys {
		next, ok := cur[k].(map[string]interface{})
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}
