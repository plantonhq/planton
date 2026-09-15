package verify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// OpenBaoVerifier proves an OpenBao deployment the way a customer would
// bootstrap it — THE SEAL LIFECYCLE IS THE PROOF:
//
//  1. Fresh pods run but report NotReady (the readiness probe is
//     `bao status`, which fails for sealed servers) — asserted, because
//     that unusual steady state is the component's designed behavior.
//  2. The API reports 501 uninitialized (the documented sys/health
//     status-code contract), then the verifier performs the real
//     bootstrap: `sys/init` (1 share — a lab shape; the keys and root
//     token exist only in this process and are never printed) and
//     `sys/unseal` on EVERY pod (Raft standbys unseal independently;
//     joins happen through the module-synthesized retry_join).
//  3. Readiness then FLIPS — pod readiness tracking seal status is the
//     other half of the lifecycle assertion.
//  4. A KV v2 round-trip (mount → write → read a run-unique marker)
//     proves the unsealed server actually serves secrets.
//
// Dev mode skips init/unseal (auto-unsealed, root token "root") and
// proves the KV round-trip on the built-in `secret/` mount.
//
// The behavioral-raft scenario (recognized by name) additionally deletes
// pod 0 after the write, waits for its replacement, unseals it AGAIN
// (restart = sealed, the Shamir-mode truth), and re-reads the marker —
// Raft data surviving pod replacement plus the re-seal reality is the
// durability proof.
//
// AUTO-UNSEAL changes the bootstrap, not the proof: `sys/init` on an
// auto-unseal seal refuses `secret_shares` and takes `recovery_shares`
// instead (the server's own rule), there is no unseal call — the seal
// unseals the server right after init — and a replaced pod comes back
// Ready on its own. The verifier follows whichever seal the manifest
// declares.
//
// Two more proofs ride on this verifier, each switched on by its
// scenario's name and documented on its method:
//
//   - THE BACKUP PROOF (with-backup, openbao_backup.go): after the
//     lifecycle, the four-command login recipe runs EXACTLY as the backup
//     job prints it, a run is created from the backup CronJob, the store is
//     listed from inside the cluster with the job's own identity, and
//     retention is proven against a seeded stale object.
//   - THE RESTORE PROOF (*-backup-restore, openbao_backup.go): the target's
//     CronJob is asserted suspended, the target is initialized and its root
//     token handed to the Secret the spec names, the restore Job completes,
//     and the restored state is read with the SOURCE's token: the marker
//     written before the snapshot is present, the one written after is
//     absent, the target's own init token is dead, and a replaced pod comes
//     back unsealed with no human.
//
// FIXTURES (manifests named `fixture-*.yaml`, deployed as prerequisites of
// a scenario) are verified for PRESENCE only — Services and Running pods —
// because a fresh OpenBao is sealed by design and its initialization
// belongs to the lane's seed script, which runs after the chain deploys.
// The one fixture that does more is the transit key holder: verifying it
// means MAKING it a key holder (see proveTransitKeyHolder).
type OpenBaoVerifier struct {
	Namespace string
	Name      string
	// Mode: dev | standalone | ha (drives which lifecycle arms run).
	Mode     string
	Replicas int
	// Behavioral enables the pod-replacement durability arm.
	Behavioral bool
	// AutoUnseal is set when the manifest declares an auto_unseal seal
	// arm: init uses recovery shares, no unseal calls are made, and a
	// replaced pod is expected Ready without help.
	AutoUnseal bool
	// Fixture marks a prerequisite manifest: presence only, no init.
	Fixture bool
	// TransitKeyHolder marks the dev-mode fixture that serves the transit
	// seal of the vaults behind it in the chain.
	TransitKeyHolder bool
	// BackupProof switches on THE BACKUP PROOF.
	BackupProof bool
	// RestoreProof switches on THE RESTORE PROOF; RootTokenSecretName and
	// RootTokenSecretKey are the spec's `restore.root_token`, where the
	// verifier places the target's initial root token after init.
	RestoreProof        bool
	RootTokenSecretName string
	RootTokenSecretKey  string
}

// initResponse is the sys/init reply — held in memory only; the keys and
// root token are run-scoped bootstrap material and are NEVER logged. An
// auto-unseal init returns recovery keys instead of unseal keys; the
// verifier keeps neither past the run.
type initResponse struct {
	Keys         []string `json:"keys"`
	RecoveryKeys []string `json:"recovery_keys"`
	RootToken    string   `json:"root_token"`
}

func (v *OpenBaoVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] openbao %q in namespace %q (mode %s, replicas %d)\n", v.Name, v.Namespace, v.Mode, v.Replicas)

	// Every mode: the chart's Services exist.
	if err := KubectlResourceExists(ctx, kubeconfig, "service", v.Name, v.Namespace); err != nil {
		return errors.Wrap(err, "openbao client service not found")
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "service", v.Name+"-internal", v.Namespace); err != nil {
		return errors.Wrap(err, "openbao internal (headless) service not found")
	}
	if v.Mode == "ha" {
		if err := KubectlResourceExists(ctx, kubeconfig, "service", v.Name+"-active", v.Namespace); err != nil {
			return errors.Wrap(err, "openbao active service not found")
		}
	}

	// Pods reach RUNNING (never gate on Ready here — sealed servers are
	// NotReady by design until the bootstrap below).
	if err := v.waitForRunningPods(ctx, kubeconfig, 10*time.Minute); err != nil {
		return err
	}

	if v.Fixture && v.Mode != "dev" {
		// A prerequisite vault: present and running is the whole
		// verification. It is sealed by design; the lane's seed script
		// initializes it once the chain is up.
		fmt.Printf("  [verify] fixture %q is present and running; its initialization belongs to the seed script\n", v.Name)
		return nil
	}

	if v.Mode == "dev" {
		// Dev mode auto-initializes and auto-unseals with root token
		// "root"; the round-trip on the built-in secret/ mount is the
		// whole proof.
		if err := v.proveKvRoundTrip(ctx, kubeconfig, "root", "", "secret", false); err != nil {
			return err
		}
		if v.TransitKeyHolder {
			return v.proveTransitKeyHolder(ctx, kubeconfig, "root")
		}
		return nil
	}

	if v.RestoreProof {
		// A restore target is not bootstrapped into its own state: it is
		// initialized so the restore Job can run, and then it carries the
		// SOURCE's state. The whole lifecycle is the restore proof.
		return v.proveRestore(ctx, kubeconfig)
	}

	// ----------------------- the seal lifecycle ------------------------
	rootToken, unsealKey, err := v.bootstrap(ctx, kubeconfig)
	if err != nil {
		return err
	}

	// Readiness flips only AFTER unseal — the probe IS the seal status
	// (for an auto-unseal seal, after the seal's own unseal that follows
	// init).
	if err := v.waitForReadyPods(ctx, kubeconfig, 5*time.Minute); err != nil {
		return errors.Wrap(err, "pods never became Ready after unseal (the readiness-tracks-seal-status contract)")
	}
	fmt.Printf("  [verify] SEAL LIFECYCLE: all %d pods flipped to Ready after unseal\n", v.Replicas)

	if err := v.proveKvRoundTrip(ctx, kubeconfig, rootToken, unsealKey, "e2e-proof", true); err != nil {
		return err
	}
	if v.BackupProof {
		return v.proveBackup(ctx, kubeconfig, rootToken)
	}
	return nil
}

func (v *OpenBaoVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "statefulset", v.Name, v.Namespace)
}

// bootstrap initializes through pod 0 and, on a Shamir seal, unseals every
// pod. Returns the run-scoped root token and unseal key (empty under
// auto-unseal, where no unseal key exists).
func (v *OpenBaoVerifier) bootstrap(ctx context.Context, kubeconfig string) (string, string, error) {
	pod0 := v.Name + "-0"

	initResp, err := v.initThroughPod(ctx, kubeconfig, pod0)
	if err != nil {
		return "", "", err
	}

	if v.AutoUnseal {
		// The seal unseals the server itself right after init; readiness
		// flipping is the proof, asserted by the caller.
		fmt.Printf("  [verify] SEAL LIFECYCLE: auto-unseal — no unseal step; the seal unseals the server on its own\n")
		return initResp.RootToken, "", nil
	}

	// Unseal every pod: Shamir unseal keys are per-CLUSTER but the
	// unseal OPERATION is per-POD (each server holds its own barrier).
	// Raft peers join through the module-synthesized retry_join once
	// the leader (pod 0) is unsealed.
	for i := 0; i < v.Replicas; i++ {
		pod := fmt.Sprintf("%s-%d", v.Name, i)
		if err := v.unsealPod(ctx, kubeconfig, pod, initResp.Keys[0], 6*time.Minute); err != nil {
			return "", "", errors.Wrapf(err, "unsealing pod %s", pod)
		}
		fmt.Printf("  [verify] SEAL LIFECYCLE: pod %s unsealed\n", pod)
	}

	return initResp.RootToken, initResp.Keys[0], nil
}

// initThroughPod asserts the uninitialized state (501 from sys/health —
// the server's documented status-code contract) and performs sys/init.
// The init body follows the seal: a Shamir seal takes `secret_shares`; an
// auto-unseal seal REFUSES that parameter ("not applicable to seal type",
// the server's own check) and takes `recovery_shares` — the recovery keys
// exist for the day the seal itself is lost, and the barrier key never
// leaves the seal.
func (v *OpenBaoVerifier) initThroughPod(ctx context.Context, kubeconfig string, pod string) (*initResponse, error) {
	var out *initResponse
	err := v.withPodPortForward(ctx, kubeconfig, pod, func(base string) error {
		// The uninitialized assertion: a fresh server answers
		// sys/health with 501.
		status, _, err := v.httpOnce(ctx, http.MethodGet, base+"/v1/sys/health", "", "", 4*time.Minute, 501)
		if err != nil {
			return errors.Wrap(err, "asserting the uninitialized (501) health state")
		}
		fmt.Printf("  [verify] SEAL LIFECYCLE: sys/health returned %d — uninitialized, as a fresh install must be\n", status)

		// Initialize with a single share — the lab shape; real
		// operators pick shares/threshold to their custody model.
		initBody := `{"secret_shares": 1, "secret_threshold": 1}`
		if v.AutoUnseal {
			initBody = `{"recovery_shares": 1, "recovery_threshold": 1}`
		}
		_, body, err := v.httpOnce(ctx, http.MethodPost, base+"/v1/sys/init", initBody, "", 2*time.Minute, 200)
		if err != nil {
			return errors.Wrap(err, "sys/init failed")
		}
		var resp initResponse
		if err := json.Unmarshal([]byte(body), &resp); err != nil {
			return errors.Wrap(err, "parsing the init response")
		}
		if resp.RootToken == "" {
			return errors.New("init returned no root token")
		}
		if v.AutoUnseal {
			if len(resp.RecoveryKeys) == 0 {
				return errors.New("auto-unseal init returned no recovery keys")
			}
			fmt.Printf("  [verify] SEAL LIFECYCLE: initialized (1 recovery share) — the recovery key and root token are held in-process only\n")
		} else {
			if len(resp.Keys) == 0 {
				return errors.New("init returned no unseal keys")
			}
			fmt.Printf("  [verify] SEAL LIFECYCLE: initialized (1 key share) — keys held in-process only\n")
		}
		out = &resp
		return nil
	})
	return out, err
}

// unsealPod submits the unseal key to one pod until it reports unsealed.
func (v *OpenBaoVerifier) unsealPod(ctx context.Context, kubeconfig, pod, key string, budget time.Duration) error {
	return v.withPodPortForward(ctx, kubeconfig, pod, func(base string) error {
		deadline := time.Now().Add(budget)
		var lastErr error
		for time.Now().Before(deadline) {
			_, body, err := v.httpOnce(ctx, http.MethodPut, base+"/v1/sys/unseal",
				fmt.Sprintf(`{"key": %q}`, key), "", 30*time.Second, 200)
			if err == nil {
				var status struct {
					Sealed bool `json:"sealed"`
				}
				if jsonErr := json.Unmarshal([]byte(body), &status); jsonErr == nil && !status.Sealed {
					return nil
				}
				lastErr = errors.New("server still reports sealed after the unseal call")
			} else {
				lastErr = err
			}
			time.Sleep(5 * time.Second)
		}
		return lastErr
	})
}

// proveKvRoundTrip mounts a KV v2 engine (unless the built-in dev mount
// is used), writes a run-unique marker and reads it back. The behavioral
// arm replaces pod 0 in the middle and re-unseals the replacement —
// restart = sealed is the Shamir-mode reality customers must know, and
// re-unsealing with the run's key completes the lifecycle story.
func (v *OpenBaoVerifier) proveKvRoundTrip(ctx context.Context, kubeconfig, token, unsealKey, mount string, createMount bool) error {
	marker := fmt.Sprintf("bao-proof-%d", time.Now().Unix())
	pod0 := v.Name + "-0"

	// Writes must land on the ACTIVE node; pod 0 is the leader right
	// after bootstrap (it was initialized and unsealed first).
	err := v.withPodPortForward(ctx, kubeconfig, pod0, func(base string) error {
		if createMount {
			if _, _, err := v.httpOnce(ctx, http.MethodPost, base+"/v1/sys/mounts/"+mount,
				`{"type": "kv", "options": {"version": "2"}}`, token, 2*time.Minute, 204, 200, 400); err != nil {
				// 400 tolerated: "path is already in use" on a re-run —
				// the write below still decides the proof.
				return errors.Wrap(err, "mounting the proof KV engine")
			}
		}
		if _, _, err := v.httpOnce(ctx, http.MethodPost, base+"/v1/"+mount+"/data/e2e-marker",
			fmt.Sprintf(`{"data": {"value": %q}}`, marker), token, 2*time.Minute, 200, 204); err != nil {
			return errors.Wrap(err, "writing the proof secret")
		}
		fmt.Printf("  [verify] KV: marker written to %s/data/e2e-marker\n", mount)
		return nil
	})
	if err != nil {
		return err
	}

	if v.Behavioral {
		if err := v.replacePod0(ctx, kubeconfig, unsealKey); err != nil {
			return err
		}
	}

	// Read the marker back — through the ACTIVE service in HA (the
	// leader may have moved after a behavioral pod loss), through pod 0
	// otherwise.
	readTarget := pod0
	viaService := false
	if v.Mode == "ha" {
		viaService = true
	}
	readOnce := func(base string) error {
		_, body, err := v.httpOnce(ctx, http.MethodGet, base+"/v1/"+mount+"/data/e2e-marker", "", token, 6*time.Minute, 200)
		if err != nil {
			return errors.Wrap(err, "reading the proof secret back")
		}
		var resp struct {
			Data struct {
				Data map[string]string `json:"data"`
			} `json:"data"`
		}
		if err := json.Unmarshal([]byte(body), &resp); err != nil {
			return errors.Wrap(err, "parsing the read response")
		}
		if resp.Data.Data["value"] != marker {
			return errors.Errorf("marker mismatch: wrote %q, read %q", marker, resp.Data.Data["value"])
		}
		return nil
	}
	if viaService {
		err = v.withServicePortForward(ctx, kubeconfig, v.Name+"-active", readOnce)
	} else {
		err = v.withPodPortForward(ctx, kubeconfig, readTarget, readOnce)
	}
	if err != nil {
		return err
	}
	if v.Behavioral {
		fmt.Printf("  [verify] DURABILITY: marker read back AFTER pod replacement — Raft data survived\n")
	} else {
		fmt.Printf("  [verify] KV: marker read back — the server serves secrets\n")
	}
	return nil
}

// replacePod0 deletes pod 0 and proves the restart truth of the declared
// seal. On Shamir the replacement comes back SEALED and is unsealed again
// with the run's key (auto-unseal arms exist to remove exactly this step);
// under auto-unseal the replacement must come back Ready with no help —
// the seal doing its one job, asserted rather than assumed.
func (v *OpenBaoVerifier) replacePod0(ctx context.Context, kubeconfig, unsealKey string) error {
	pod0 := v.Name + "-0"

	// Capture the doomed pod's uid first — a terminating pod keeps
	// reporting phase Running for its whole grace period, so a
	// phase-keyed wait returns instantly against the DYING pod and a
	// re-unseal would be burned as a no-op on the old (still unsealed)
	// server. The replacement is only real once a pod with a NEW uid is
	// Running.
	oldUid, err := podUid(ctx, kubeconfig, v.Namespace, pod0)
	if err != nil {
		return err
	}
	fmt.Printf("  [verify] DURABILITY: deleting pod %q\n", pod0)
	if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"delete", "pod", pod0, "-n", v.Namespace, "--wait=false").CombinedOutput(); err != nil {
		return errors.Wrapf(err, "deleting pod 0: %s", string(out))
	}

	if v.AutoUnseal {
		// Ready with a new uid and no unseal call: the seal unsealed it.
		if err := waitForPodReplaced(ctx, kubeconfig, v.Namespace, pod0, oldUid, true, 10*time.Minute); err != nil {
			return errors.Wrap(err, "the replacement pod never became Ready on its own — the auto-unseal seal did not unseal it (check the seal's key and the server's identity)")
		}
		fmt.Printf("  [verify] DURABILITY: replacement pod came back Ready with no unseal step — the auto-unseal seal unsealed it\n")
		return nil
	}

	// A replaced Shamir-mode pod is sealed by design: Running, NOT Ready.
	if err := waitForPodReplaced(ctx, kubeconfig, v.Namespace, pod0, oldUid, false, 10*time.Minute); err != nil {
		return errors.Wrap(err, "the replacement pod never reached Running")
	}
	if err := v.waitForRunningPods(ctx, kubeconfig, 10*time.Minute); err != nil {
		return errors.Wrap(err, "pod 0 never returned after deletion")
	}
	fmt.Printf("  [verify] DURABILITY: replacement pod is running and SEALED (the Shamir restart truth) — re-unsealing\n")
	if err := v.unsealPod(ctx, kubeconfig, pod0, unsealKey, 6*time.Minute); err != nil {
		return errors.Wrap(err, "re-unsealing the replacement pod")
	}
	fmt.Printf("  [verify] DURABILITY: replacement pod re-unsealed and rejoining the Raft cluster\n")
	return nil
}

// ------------------------------ plumbing --------------------------------

func (v *OpenBaoVerifier) waitForRunningPods(ctx context.Context, kubeconfig string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var last string
	for time.Now().Before(deadline) {
		running := 0
		for i := 0; i < v.Replicas; i++ {
			phase, _ := kubectlGetJSONPath(ctx, kubeconfig, "pod",
				fmt.Sprintf("%s-%d", v.Name, i), v.Namespace, "{.status.phase}")
			if phase == "Running" {
				running++
			}
			last = phase
		}
		if running == v.Replicas {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("openbao pods never all reached Running (last phase %q)", last)
}

func (v *OpenBaoVerifier) waitForReadyPods(ctx context.Context, kubeconfig string, budget time.Duration) error {
	want := fmt.Sprintf("%d", v.Replicas)
	deadline := time.Now().Add(budget)
	var lastReady string
	for time.Now().Before(deadline) {
		ready, _ := kubectlGetJSONPath(ctx, kubeconfig, "statefulset", v.Name, v.Namespace, "{.status.readyReplicas}")
		lastReady = ready
		if ready == want {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return errors.Errorf("statefulset never reached %s ready replicas (last %q)", want, lastReady)
}

// withPodPortForward runs fn with a fresh tunnel to ONE pod — sealed pods
// accept connections (the chart publishes not-ready addresses), and
// per-pod tunnels are what unseal requires. Fresh per call: a tunnel dies
// silently with its pod (the caught-live port-forward class).
func (v *OpenBaoVerifier) withPodPortForward(ctx context.Context, kubeconfig, pod string, fn func(base string) error) error {
	return v.withPortForward(ctx, kubeconfig, "pod/"+pod, fn)
}

func (v *OpenBaoVerifier) withServicePortForward(ctx context.Context, kubeconfig, service string, fn func(base string) error) error {
	return v.withPortForward(ctx, kubeconfig, "svc/"+service, fn)
}

func (v *OpenBaoVerifier) withPortForward(ctx context.Context, kubeconfig, target string, fn func(base string) error) error {
	const localPort = "18200"

	pfCtx, cancel := context.WithCancel(ctx)
	pf := exec.CommandContext(pfCtx, "kubectl", "--kubeconfig", kubeconfig,
		"port-forward", target, localPort+":8200", "-n", v.Namespace)
	var pfOut strings.Builder
	pf.Stdout = &pfOut
	pf.Stderr = &pfOut
	if err := pf.Start(); err != nil {
		cancel()
		return errors.Wrapf(err, "starting port-forward to %s", target)
	}
	// ONE deferred func, cancel FIRST — Wait blocks forever on a
	// port-forward never told to exit.
	defer func() {
		cancel()
		_ = pf.Wait()
	}()

	return fn("http://127.0.0.1:" + localPort)
}

// openBaoScenarioShape pulls the verifier's inputs out of a
// KubernetesOpenBao scenario manifest: which server mode the oneof
// declares (absent = standalone, the chart default) and the replica
// count (only HA runs more than one). Manifests may spell fields in
// either protojson case; the keys read here are case-neutral.
func openBaoScenarioShape(spec map[string]interface{}) (mode string, replicas int) {
	mode, replicas = "standalone", 1
	server := specNestedMap(spec, "server")
	if server == nil {
		return
	}
	if _, ok := server["dev"]; ok {
		return "dev", 1
	}
	if ha, ok := server["ha"].(map[string]interface{}); ok {
		replicas = 3
		if n, ok := specInt(ha["replicas"]); ok {
			replicas = n
		}
		return "ha", replicas
	}
	return
}

// openBaoAutoUnseal reports whether the manifest declares an auto_unseal
// seal arm (any of the four): the bootstrap then uses recovery shares and
// makes no unseal calls.
func openBaoAutoUnseal(spec map[string]interface{}) bool {
	seal := specNestedMap(spec, "autoUnseal", "auto_unseal")
	return len(seal) > 0
}

// openBaoRestoreRootToken reads `restore.root_token` (name, key) — the
// Secret the verifier fills with the target's initial root token so the
// restore Job can proceed. Empty when the manifest declares no restore.
func openBaoRestoreRootToken(spec map[string]interface{}) (name, key string) {
	restore := specNestedMap(spec, "restore")
	if restore == nil {
		return "", ""
	}
	token := specNestedMap(restore, "rootToken", "root_token")
	if token == nil {
		return "", ""
	}
	name, _ = token["name"].(string)
	key, _ = token["key"].(string)
	return name, key
}

// httpOnce performs one JSON request with retries across the tunnel
// warm-up, succeeding on any of wantStatuses. Returns (status, body).
// The X-Vault-Token header carries the token (OpenBao serves the
// Vault-compatible API surface).
func (v *OpenBaoVerifier) httpOnce(ctx context.Context, method, url, body, token string, budget time.Duration, wantStatuses ...int) (int, string, error) {
	deadline := time.Now().Add(budget)
	var lastStatus int
	var lastOut string
	var lastErr error
	for time.Now().Before(deadline) {
		var reader *bytes.Reader
		if body != "" {
			reader = bytes.NewReader([]byte(body))
		} else {
			reader = bytes.NewReader(nil)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, reader)
		if err != nil {
			return 0, "", err
		}
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("X-Vault-Token", token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(resp.Body)
			resp.Body.Close()
			lastStatus = resp.StatusCode
			lastOut = buf.String()
			for _, want := range wantStatuses {
				if resp.StatusCode == want {
					return resp.StatusCode, lastOut, nil
				}
			}
			lastErr = errors.Errorf("HTTP %d (wanted one of %v)", resp.StatusCode, wantStatuses)
		} else {
			lastErr = err
		}
		time.Sleep(5 * time.Second)
	}
	return lastStatus, lastOut, errors.Wrapf(lastErr, "last body: %s", firstLines(lastOut, 3))
}
