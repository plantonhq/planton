package verify

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ArgoCdVerifier checks an Argo CD install to the point a customer could
// run their GitOps delivery on it: the API/UI server, application
// controller (a StatefulSet at the chart's default distribution mode) and
// repo server rolled out, the APPLICATION-generated initial admin Secret
// present (the credential handle the outputs promise), a session opened
// through Argo CD's own session API as that admin, and an authenticated
// applications-API round-trip.
//
// The behavioral-gitops scenario (recognized by name) additionally runs
// THE GITOPS PROOF: an Application declared through the product API
// pointing at a public Git repository, automated sync driving the repo's
// manifests into the destination namespace (asserted on the synced
// workload itself, not just the reported status), then the server pod
// DELETED and — after a REPLACEMENT pod (a new UID) — a fresh login and
// the same Application still reporting Synced/Healthy: config and state
// survive the control plane's pods because they live in CRs and the
// cluster, which is the product's whole promise.
//
// Destroy asserts the designed CRD keep posture: the release's workloads
// are gone while applications.argoproj.io (and its siblings) REMAIN —
// deleting them would cascade to every Application in the cluster.
type ArgoCdVerifier struct {
	Namespace string
	Name      string
	// AdminEnabled gates the login proof — the initial-admin Secret only
	// exists while the local admin user is on (the spec default).
	AdminEnabled bool
	// ServerInsecure follows spec.server.insecure: the proof talks plain
	// HTTP to the Service's port 80 instead of the self-signed HTTPS
	// listener on 443 (both map to the same server port; the scheme is
	// the contract under proof).
	ServerInsecure bool
	// CrdsKeep follows spec.crds.keep (default true). Kept lanes assert
	// the CRDs SURVIVE destroy — then remove them, because the chart
	// TEMPLATES its CRDs and the kept copies carry this release's Helm
	// ownership metadata, which would fail any later lane's install on
	// the shared cluster. keep=false lanes assert the cascade truth (the
	// CRDs are gone with the release).
	CrdsKeep bool
	// GitOpsProof switches on the Application sync + pod-replacement arm.
	GitOpsProof bool
}

// The verifier-owned proof Application. The repo is Argo CD's own public
// example collection — the canonical first-sync target its docs teach
// (internet-at-pod: the repo server clones it live; flagged on the queue
// entry's watch-list).
const (
	argocdProofAppName   = "e2e-proof-app"
	argocdProofRepoURL   = "https://github.com/argoproj/argocd-example-apps"
	argocdProofRepoPath  = "guestbook"
	argocdProofWorkload  = "guestbook-ui"
	argocdCrdApplication = "applications.argoproj.io"
)

func (v *ArgoCdVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] argocd %q in namespace %q\n", v.Name, v.Namespace)

	// The control plane: server + repo server Deployments, the
	// application controller StatefulSet (the chart's default
	// distribution mode renders a StatefulSet).
	if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/"+v.Name+"-server", v.Namespace, 10*time.Minute); err != nil {
		return errors.Wrap(err, "the argocd server deployment never rolled out")
	}
	if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/"+v.Name+"-repo-server", v.Namespace, 10*time.Minute); err != nil {
		return errors.Wrap(err, "the argocd repo-server deployment never rolled out")
	}
	if err := waitStatefulSetReady(ctx, kubeconfig, v.Name+"-application-controller", v.Namespace, 10*time.Minute); err != nil {
		return errors.Wrap(err, "the application controller statefulset never became ready")
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "service", v.Name+"-server", v.Namespace); err != nil {
		return errors.Wrap(err, "argocd server service not found")
	}

	if !v.AdminEnabled {
		fmt.Printf("  [verify] admin user disabled by the manifest — control-plane rollout is the assertion\n")
		return nil
	}

	return v.proveApiAndGitOps(ctx, kubeconfig)
}

func (v *ArgoCdVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.Name+"-server", v.Namespace); err != nil {
		return err
	}

	if !v.CrdsKeep {
		// The cascade truth: with keep off, the CRDs leave with the
		// release (lane hygiene on the shared cluster — kept TEMPLATED
		// CRDs would fail the next lane's install on Helm ownership
		// validation).
		if err := KubectlResourceAbsent(ctx, kubeconfig, "crd", argocdCrdApplication, ""); err != nil {
			return errors.Wrap(err, "crds.keep=false must remove the argoproj CRDs with the release, but applications.argoproj.io remains")
		}
		fmt.Printf("  [verify] DESTROY: release workloads gone; argoproj CRDs removed (crds.keep=false as declared)\n")
		return nil
	}

	// The designed keep posture (the kind's default): destroying the
	// release must LEAVE the Application CRD behind (removing it would
	// cascade-delete every Application in the cluster).
	if err := KubectlResourceExists(ctx, kubeconfig, "crd", argocdCrdApplication, ""); err != nil {
		return errors.Wrap(err, "the argoproj CRDs must SURVIVE destroy (the crds.keep posture) but applications.argoproj.io is gone")
	}
	fmt.Printf("  [verify] DESTROY: release workloads gone; applications.argoproj.io KEPT (the designed crds.keep posture)\n")

	// Post-assertion cleanup: the kept copies carry THIS release's Helm
	// ownership metadata (the chart templates its CRDs), which would fail
	// any later lane's install — the shared cluster returns to zero once
	// the posture is proven.
	for _, crd := range []string{
		"applications.argoproj.io",
		"appprojects.argoproj.io",
		"applicationsets.argoproj.io",
	} {
		delCmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"delete", "crd", crd, "--ignore-not-found")
		if out, err := delCmd.CombinedOutput(); err != nil {
			return errors.Errorf("cleaning up the kept CRD %s after the posture proof: %v: %s", crd, err, firstLines(string(out), 2))
		}
	}
	fmt.Printf("  [verify] CLEANUP: kept argoproj CRDs removed after the posture proof (shared-cluster lane hygiene)\n")
	return nil
}

// proveApiAndGitOps opens a session as the generated admin and, on the
// gitops lane, drives the full sync + pod-replacement proof.
func (v *ArgoCdVerifier) proveApiAndGitOps(ctx context.Context, kubeconfig string) error {
	// The APPLICATION-generated credential (fixed name, key `password`).
	password, err := v.initialAdminPassword(ctx, kubeconfig)
	if err != nil {
		return err
	}
	fmt.Printf("  [verify] SECRET: argocd-initial-admin-secret present (the application-generated credential)\n")

	const apiPort = "18086"
	servicePort, scheme := "443", "https"
	if v.ServerInsecure {
		servicePort, scheme = "80", "http"
	}
	apiCancel, err := startPortForward(ctx, kubeconfig, "svc/"+v.Name+"-server", v.Namespace, apiPort+":"+servicePort)
	if err != nil {
		return errors.Wrap(err, "starting port-forward to the argocd server")
	}
	defer apiCancel()
	apiBase := scheme + "://127.0.0.1:" + apiPort

	token, err := v.login(ctx, apiBase, password)
	if err != nil {
		return errors.Wrap(err, "opening a session as admin through /api/v1/session")
	}
	fmt.Printf("  [verify] LOGIN: session opened as admin through the product API\n")

	// The authenticated API round-trip (list applications).
	if _, err := argocdHTTPS(ctx, http.MethodGet, apiBase+"/api/v1/applications", token, "", 2*time.Minute); err != nil {
		return errors.Wrap(err, "the authenticated applications API round-trip failed")
	}
	fmt.Printf("  [verify] API: authenticated applications round-trip OK\n")

	if !v.GitOpsProof {
		return nil
	}

	// ---- THE GITOPS PROOF -------------------------------------------------
	// Declare an Application through the product API and let automated
	// sync drive the repo's manifests into the destination namespace.
	appBody := fmt.Sprintf(`{
      "metadata": {"name": %q},
      "spec": {
        "project": "default",
        "source": {"repoURL": %q, "path": %q, "targetRevision": "HEAD"},
        "destination": {"server": "https://kubernetes.default.svc", "namespace": %q},
        "syncPolicy": {"automated": {"prune": true}}
      }
    }`, argocdProofAppName, argocdProofRepoURL, argocdProofRepoPath, v.Namespace)
	if _, err := argocdHTTPS(ctx, http.MethodPost, apiBase+"/api/v1/applications", token, appBody, 2*time.Minute); err != nil {
		return errors.Wrap(err, "declaring the proof Application through the API")
	}
	fmt.Printf("  [verify] DECLARE: Application %q created through the API (repo %s)\n", argocdProofAppName, argocdProofRepoURL)

	// Best-effort cleanup so DESTROY meets zero orphans even on failure.
	defer v.cleanupProofApp(ctx, apiBase, kubeconfig, password)

	if err := v.awaitSyncedHealthy(ctx, apiBase, token, 8*time.Minute); err != nil {
		return err
	}

	// The synced workload itself — the proof is the cluster state, never
	// the reported status alone.
	if err := KubectlResourceExists(ctx, kubeconfig, "deployment", argocdProofWorkload, v.Namespace); err != nil {
		return errors.Wrap(err, "the synced guestbook workload never appeared in the destination namespace")
	}
	fmt.Printf("  [verify] SYNC: %s deployed by automated sync — GitOps loop proven live\n", argocdProofWorkload)

	// ---- the state proof: server pod replacement ---------------------------
	if err := deletePodAwaitReplacement(ctx, kubeconfig, v.Namespace,
		"app.kubernetes.io/instance="+v.Name+",app.kubernetes.io/component=server", 10*time.Minute); err != nil {
		return errors.Wrap(err, "the argocd server pod did not recover after deletion")
	}
	apiCancel()
	apiCancel2, err := startPortForward(ctx, kubeconfig, "svc/"+v.Name+"-server", v.Namespace, apiPort+":"+servicePort)
	if err != nil {
		return errors.Wrap(err, "re-establishing the API port-forward after the pod kill")
	}
	defer apiCancel2()

	token, err = v.login(ctx, apiBase, password)
	if err != nil {
		return errors.Wrap(err, "re-authenticating after the pod replacement")
	}
	if err := v.awaitSyncedHealthy(ctx, apiBase, token, 4*time.Minute); err != nil {
		return errors.Wrap(err, "the Application should still report Synced/Healthy after the server pod replacement")
	}
	fmt.Printf("  [verify] STATE: re-login + Application Synced/Healthy AFTER server pod replacement — state lives in CRs, not pods\n")
	return nil
}

// initialAdminPassword reads the application-generated credential.
func (v *ArgoCdVerifier) initialAdminPassword(ctx context.Context, kubeconfig string) (string, error) {
	deadline := time.Now().Add(3 * time.Minute)
	var lastErr error
	for time.Now().Before(deadline) {
		b64, err := kubectlGetJSONPath(ctx, kubeconfig, "secret", "argocd-initial-admin-secret", v.Namespace, "{.data.password}")
		if err == nil && b64 != "" {
			raw, decErr := base64.StdEncoding.DecodeString(b64)
			if decErr == nil {
				return strings.TrimSpace(string(raw)), nil
			}
			lastErr = decErr
		} else if err != nil {
			lastErr = err
		}
		time.Sleep(5 * time.Second)
	}
	return "", errors.Wrap(lastErr, "the argocd-initial-admin-secret never appeared (the application creates it at first start)")
}

// login opens a session through POST /api/v1/session (route verified in
// the app source at v3.4.5) and returns the bearer token.
func (v *ArgoCdVerifier) login(ctx context.Context, apiBase, password string) (string, error) {
	body := fmt.Sprintf(`{"username": "admin", "password": %q}`, password)
	deadline := time.Now().Add(3 * time.Minute)
	var lastErr error
	for time.Now().Before(deadline) {
		out, err := argocdHTTPS(ctx, http.MethodPost, apiBase+"/api/v1/session", "", body, 30*time.Second)
		if err == nil {
			var parsed map[string]interface{}
			if json.Unmarshal([]byte(out), &parsed) == nil {
				if token, _ := parsed["token"].(string); token != "" {
					return token, nil
				}
			}
			lastErr = errors.Errorf("session response carried no token: %s", firstLines(out, 2))
		} else {
			lastErr = err
		}
		time.Sleep(5 * time.Second)
	}
	return "", lastErr
}

// awaitSyncedHealthy polls the Application until sync=Synced AND
// health=Healthy.
func (v *ArgoCdVerifier) awaitSyncedHealthy(ctx context.Context, apiBase, token string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastState string
	for time.Now().Before(deadline) {
		out, err := argocdHTTPS(ctx, http.MethodGet, apiBase+"/api/v1/applications/"+argocdProofAppName, token, "", 30*time.Second)
		if err == nil {
			sync, health := argocdAppState(out)
			lastState = fmt.Sprintf("sync=%s health=%s", sync, health)
			if sync == "Synced" && health == "Healthy" {
				fmt.Printf("  [verify] APP: %s reports Synced/Healthy\n", argocdProofAppName)
				return nil
			}
		} else {
			lastState = err.Error()
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("the proof Application never reached Synced/Healthy: %s", lastState)
}

// cleanupProofApp deletes the proof Application with cascade (pruning the
// synced workload) so destroy meets the zero-orphan bar. Best-effort: a
// fresh login guards against expired tokens after pod replacements.
func (v *ArgoCdVerifier) cleanupProofApp(ctx context.Context, apiBase, kubeconfig, password string) {
	token, err := v.login(ctx, apiBase, password)
	if err != nil {
		return
	}
	_, _ = argocdHTTPS(ctx, http.MethodDelete,
		apiBase+"/api/v1/applications/"+argocdProofAppName+"?cascade=true", token, "", 2*time.Minute)
	// Wait for the cascade to prune the synced workload.
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", argocdProofWorkload, v.Namespace); err == nil {
			return
		}
		time.Sleep(5 * time.Second)
	}
}

// argocdAdminEnabled reads the spec's admin toggle (default true — the
// proto default; both manifest key forms tolerated).
func argocdAdminEnabled(spec map[string]interface{}) bool {
	for _, key := range []string{"admin_enabled", "adminEnabled"} {
		if raw, ok := spec[key]; ok {
			if enabled, ok := raw.(bool); ok {
				return enabled
			}
		}
	}
	return true
}

// argocdServerInsecure reads spec.server.insecure (default false — the
// self-signed HTTPS listener).
func argocdServerInsecure(spec map[string]interface{}) bool {
	server, _ := spec["server"].(map[string]interface{})
	if server == nil {
		return false
	}
	if raw, ok := server["insecure"].(bool); ok {
		return raw
	}
	return false
}

// argocdCrdsKeep reads spec.crds.keep (default true — the component's
// keep posture).
func argocdCrdsKeep(spec map[string]interface{}) bool {
	crds, _ := spec["crds"].(map[string]interface{})
	if crds == nil {
		return true
	}
	if raw, ok := crds["keep"].(bool); ok {
		return raw
	}
	return true
}

// argocdAppState digs sync/health out of an Application API response.
func argocdAppState(body string) (string, string) {
	var parsed struct {
		Status struct {
			Sync struct {
				Status string `json:"status"`
			} `json:"sync"`
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
		} `json:"status"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return "", ""
	}
	return parsed.Status.Sync.Status, parsed.Status.Health.Status
}

// argocdHTTPS is a single HTTPS request against the server's self-signed
// listener (the chart default — no server.insecure needed for the proof),
// with optional bearer auth, retried within the budget.
func argocdHTTPS(ctx context.Context, method, url, token, body string, budget time.Duration) (string, error) {
	client := &http.Client{Transport: &http.Transport{
		// The server's default listener serves a self-signed certificate;
		// composed exposure normally terminates TLS in front of it. The
		// proof talks to the pod directly, so verification is skipped —
		// against localhost only, never a production posture.
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	deadline := time.Now().Add(budget)
	var lastOut string
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
		if err != nil {
			return "", err
		}
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := client.Do(req)
		if err == nil {
			raw, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return string(raw), nil
			}
			lastOut = string(raw)
			lastErr = errors.Errorf("HTTP %d: %s", resp.StatusCode, firstLines(string(raw), 2))
		} else {
			lastErr = err
		}
		time.Sleep(5 * time.Second)
	}
	if lastErr == nil {
		lastErr = errors.New("request never succeeded within the budget")
	}
	return lastOut, lastErr
}

// writeVerifierTempManifest writes a verifier-owned CR to a temp file for
// kubectl apply (shared by the Argo verifiers).
func writeVerifierTempManifest(prefix, content string) (string, error) {
	f, err := os.CreateTemp("", prefix+"-*.yaml")
	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(content); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return filepath.Clean(f.Name()), nil
}

// kubectlApplyFile applies a manifest file with the lane kubeconfig.
func kubectlApplyFile(ctx context.Context, kubeconfig, path string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "apply", "-f", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.Errorf("kubectl apply -f %s: %v: %s", path, err, firstLines(string(out), 3))
	}
	return nil
}
