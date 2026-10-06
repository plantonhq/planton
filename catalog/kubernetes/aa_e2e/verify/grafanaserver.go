package verify

import (
	"bytes"
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

// GrafanaVerifier checks a standalone Grafana to the point a person could
// sign in and see data sources: the Deployment available, the Service
// present, /api/health reporting a working database, an AUTHENTICATED
// API round-trip as the admin credentials (read from the chart-owned or
// referenced Secret — which also proves the credential wiring end to
// end), and every declared datasource actually provisioned.
//
// The behavioral-sso scenario (recognized by name) signs in end to end
// through a stand-in for Google's token endpoint (grafana_signin.go).
//
// A manifest declaring spec.agent_reader proves agent teammates' read-only
// way in (grafana_agent_reader.go): the module's Job, its exact Role, the
// token Secret, a Viewer that is refused writes, one token, and -- in a
// second act -- the replaced or disabled token refused.
//
// The behavioral-persistence scenario (recognized by name) additionally
// CREATES a dashboard through the API, deletes the pod, waits for a
// REPLACEMENT pod (a new UID — status flapping back Ready on the dying
// pod is not recovery), and re-reads the dashboard: state surviving pod
// loss through the PVC is the proof.
type GrafanaVerifier struct {
	Namespace string
	Name      string
	// AdminSecretName is the Secret carrying the admin credentials
	// (chart-owned `<name>` for the generate arm; the referenced Secret
	// for the existing arm).
	AdminSecretName string
	// AdminUserKey / AdminPasswordKey are the Secret keys (the chart's
	// admin-user / admin-password unless the manifest overrode them).
	AdminUserKey     string
	AdminPasswordKey string
	// Datasources lists the declared datasource names that must be
	// provisioned.
	Datasources []string
	// Persistence switches on the dashboard-survives-pod-loss proof.
	Persistence bool
	// SignIn switches on the end-to-end Google sign-in proof against a
	// stand-in token endpoint (grafana_signin.go).
	SignIn bool
	// AgentReader is the manifest's spec.agent_reader with its defaults
	// applied; nil when the block is not declared. Declared, it switches
	// on the agent teammates' read-only path proof (grafana_agent_reader.go).
	AgentReader *grafanaAgentReader
}

func (v *GrafanaVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] grafana %q in namespace %q\n", v.Name, v.Namespace)

	if err := v.waitDeploymentAvailable(ctx, kubeconfig, 8*time.Minute); err != nil {
		return errors.Wrap(err, "the grafana deployment never became available")
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "service", v.Name, v.Namespace); err != nil {
		return errors.Wrap(err, "grafana service not found")
	}
	return v.proveApiRoundTrip(ctx, kubeconfig)
}

func (v *GrafanaVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.Name, v.Namespace)
}

func (v *GrafanaVerifier) waitDeploymentAvailable(ctx context.Context, kubeconfig string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var last string
	for time.Now().Before(deadline) {
		ready, _ := kubectlGetJSONPath(ctx, kubeconfig, "deployment", v.Name, v.Namespace, "{.status.availableReplicas}")
		last = ready
		if ready != "" && ready != "0" {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("grafana deployment never reported available replicas (last %q)", last)
}

// adminCredentials reads the admin username and password from the
// credentials Secret — the read itself is part of the proof (a Secret
// that does not exist, or carries the wrong keys, fails here rather than
// at a customer's first login).
func (v *GrafanaVerifier) adminCredentials(ctx context.Context, kubeconfig string) (string, string, error) {
	userKey := v.AdminUserKey
	if userKey == "" {
		userKey = "admin-user"
	}
	passwordKey := v.AdminPasswordKey
	if passwordKey == "" {
		passwordKey = "admin-password"
	}
	userB64, err := kubectlGetJSONPath(ctx, kubeconfig, "secret", v.AdminSecretName, v.Namespace,
		fmt.Sprintf("{.data.%s}", strings.ReplaceAll(userKey, ".", "\\.")))
	if err != nil {
		return "", "", errors.Wrapf(err, "reading secret %q key %q", v.AdminSecretName, userKey)
	}
	passwordB64, err := kubectlGetJSONPath(ctx, kubeconfig, "secret", v.AdminSecretName, v.Namespace,
		fmt.Sprintf("{.data.%s}", strings.ReplaceAll(passwordKey, ".", "\\.")))
	if err != nil {
		return "", "", errors.Wrapf(err, "reading secret %q key %q", v.AdminSecretName, passwordKey)
	}
	user, err := base64.StdEncoding.DecodeString(userB64)
	if err != nil {
		return "", "", err
	}
	password, err := base64.StdEncoding.DecodeString(passwordB64)
	if err != nil {
		return "", "", err
	}
	return strings.TrimSpace(string(user)), strings.TrimSpace(string(password)), nil
}

// proveApiRoundTrip drives Grafana's API over a port-forward: /api/health
// must report a working database, the admin credentials must authenticate,
// and every declared datasource must be provisioned. The persistence
// variant creates a dashboard, kills the pod, waits for a REPLACEMENT and
// re-reads the dashboard.
func (v *GrafanaVerifier) proveApiRoundTrip(ctx context.Context, kubeconfig string) error {
	const localPort = "13000"

	pfCtx, cancel := context.WithCancel(ctx)
	pf := exec.CommandContext(pfCtx, "kubectl", "--kubeconfig", kubeconfig,
		"port-forward", "svc/"+v.Name, localPort+":80", "-n", v.Namespace)
	var pfOut strings.Builder
	pf.Stdout = &pfOut
	pf.Stderr = &pfOut
	if err := pf.Start(); err != nil {
		cancel()
		return errors.Wrap(err, "starting port-forward to the grafana service")
	}
	// ONE deferred func, cancel FIRST — Wait blocks forever on a
	// port-forward that is never told to exit.
	defer func() {
		cancel()
		_ = pf.Wait()
	}()

	base := "http://127.0.0.1:" + localPort

	user, password, err := v.adminCredentials(ctx, kubeconfig)
	if err != nil {
		return err
	}
	fmt.Printf("  [verify] CREDENTIALS: admin credentials read from secret %q\n", v.AdminSecretName)

	// /api/health reports the embedded/external database state — "ok" is
	// the whole point of the check (a Grafana with a broken database
	// serves logins into a wall).
	if out, err := v.request(ctx, http.MethodGet, base+"/api/health", "", "", "", 5*time.Minute); err != nil {
		return errors.Wrapf(err, "grafana /api/health never reported healthy: %s", out)
	}
	fmt.Printf("  [verify] HEALTH: /api/health reports a working database\n")

	// An authenticated call proves the admin credential wiring: the org
	// endpoint requires admin and answers 200 only for valid basic auth.
	if out, err := v.request(ctx, http.MethodGet, base+"/api/org", user, password, "", 2*time.Minute); err != nil {
		return errors.Wrapf(err, "the admin credentials failed to authenticate: %s", out)
	}
	fmt.Printf("  [verify] AUTH: the admin credentials authenticated against the API\n")

	// Every declared datasource must be provisioned — code-provisioned
	// datasources are the composition contract this kind exports.
	if len(v.Datasources) > 0 {
		out, err := v.request(ctx, http.MethodGet, base+"/api/datasources", user, password, "", 2*time.Minute)
		if err != nil {
			return errors.Wrapf(err, "listing datasources: %s", out)
		}
		var datasources []struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal([]byte(out), &datasources); err != nil {
			return errors.Wrapf(err, "parsing the datasources response: %s", firstLines(out, 3))
		}
		present := map[string]bool{}
		for _, ds := range datasources {
			present[ds.Name] = true
		}
		for _, want := range v.Datasources {
			if !present[want] {
				return errors.Errorf("declared datasource %q was not provisioned (got: %v)", want, present)
			}
		}
		fmt.Printf("  [verify] DATASOURCES: all %d declared datasources provisioned\n", len(v.Datasources))
	}

	if v.SignIn {
		if err := v.proveSignIn(ctx, kubeconfig, base, user, password); err != nil {
			return err
		}
	}

	if v.AgentReader != nil {
		if err := v.proveAgentReader(ctx, kubeconfig, base, user, password); err != nil {
			return err
		}
	}

	if !v.Persistence {
		return nil
	}

	// The persistence proof: state written through the API must survive
	// a pod replacement via the PVC.
	dashboardTitle := fmt.Sprintf("e2e-proof-%d", time.Now().Unix())
	createBody := fmt.Sprintf(`{"dashboard": {"title": "%s", "panels": []}, "overwrite": false}`, dashboardTitle)
	out, err := v.request(ctx, http.MethodPost, base+"/api/dashboards/db", user, password, createBody, 2*time.Minute)
	if err != nil {
		return errors.Wrapf(err, "creating the proof dashboard: %s", out)
	}
	var created struct {
		Uid string `json:"uid"`
	}
	if err := json.Unmarshal([]byte(out), &created); err != nil || created.Uid == "" {
		return errors.Errorf("the dashboard create response carried no uid: %s", firstLines(out, 3))
	}
	fmt.Printf("  [verify] PERSISTENCE: dashboard %q created (uid %s)\n", dashboardTitle, created.Uid)

	if err := v.deleteAndAwaitReplacement(ctx, kubeconfig); err != nil {
		return err
	}

	// Same tunnel, new pod: the port-forward may need re-establishing; the
	// retry loop inside request covers the window.
	out, err = v.request(ctx, http.MethodGet, base+"/api/dashboards/uid/"+created.Uid, user, password, "", 6*time.Minute)
	if err != nil {
		// The old tunnel died with the pod — re-establish once and retry.
		cancel()
		_ = pf.Wait()
		pfCtx2, cancel2 := context.WithCancel(ctx)
		pf2 := exec.CommandContext(pfCtx2, "kubectl", "--kubeconfig", kubeconfig,
			"port-forward", "svc/"+v.Name, localPort+":80", "-n", v.Namespace)
		pf2.Stdout = &pfOut
		pf2.Stderr = &pfOut
		if startErr := pf2.Start(); startErr != nil {
			cancel2()
			return errors.Wrap(startErr, "re-establishing the port-forward after the pod kill")
		}
		defer func() {
			cancel2()
			_ = pf2.Wait()
		}()
		out, err = v.request(ctx, http.MethodGet, base+"/api/dashboards/uid/"+created.Uid, user, password, "", 4*time.Minute)
		if err != nil {
			return errors.Wrapf(err, "the proof dashboard did not survive the pod replacement: %s", out)
		}
	}
	if !strings.Contains(out, dashboardTitle) {
		return errors.Errorf("the re-read dashboard does not carry the proof title: %s", firstLines(out, 3))
	}
	fmt.Printf("  [verify] PERSISTENCE: dashboard %q intact AFTER pod replacement — state survived on the PVC\n", dashboardTitle)
	return nil
}

// deleteAndAwaitReplacement kills the Grafana pod and waits until a pod
// with a NEW UID is Ready — StatefulSet/Deployment status can flap Ready
// against the dying pod, so a new UID is the only honest recovery signal.
func (v *GrafanaVerifier) deleteAndAwaitReplacement(ctx context.Context, kubeconfig string) error {
	podJsonPath := "{.items[0].metadata.uid}"
	selector := "app.kubernetes.io/instance=" + v.Name

	uidOut, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"get", "pods", "-n", v.Namespace, "-l", selector, "-o", "jsonpath="+podJsonPath).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "reading the grafana pod uid: %s", string(uidOut))
	}
	oldUid := strings.TrimSpace(string(uidOut))

	fmt.Printf("  [verify] PERSISTENCE: deleting the grafana pod (uid %s)\n", oldUid)
	if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"delete", "pod", "-n", v.Namespace, "-l", selector, "--wait=false").CombinedOutput(); err != nil {
		return errors.Wrapf(err, "deleting the grafana pod: %s", string(out))
	}

	deadline := time.Now().Add(8 * time.Minute)
	for time.Now().Before(deadline) {
		uidNow, _ := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "pods", "-n", v.Namespace, "-l", selector,
			"--field-selector", "status.phase=Running",
			"-o", "jsonpath="+podJsonPath).CombinedOutput()
		newUid := strings.TrimSpace(string(uidNow))
		if newUid != "" && newUid != oldUid {
			ready, _ := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
				"get", "pods", "-n", v.Namespace, "-l", selector,
				"-o", "jsonpath={.items[0].status.conditions[?(@.type=='Ready')].status}").CombinedOutput()
			if strings.TrimSpace(string(ready)) == "True" {
				fmt.Printf("  [verify] PERSISTENCE: replacement pod (uid %s) is Ready\n", newUid)
				return nil
			}
		}
		time.Sleep(10 * time.Second)
	}
	return errors.New("no replacement grafana pod became Ready after the deletion")
}

// request performs one JSON request with optional basic auth, retrying
// across the warm-up window (body-read inside the loop — a response dying
// mid-stream retries rather than escaping). Non-2xx responses are errors.
func (v *GrafanaVerifier) request(ctx context.Context, method, url, user, password, body string, budget time.Duration) (string, error) {
	deadline := time.Now().Add(budget)
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
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		if user != "" {
			req.SetBasicAuth(user, password)
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			buf := new(bytes.Buffer)
			_, readErr := buf.ReadFrom(resp.Body)
			resp.Body.Close()
			lastOut = buf.String()
			if readErr == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return lastOut, nil
			}
			if readErr != nil {
				lastErr = readErr
			} else {
				lastErr = errors.Errorf("HTTP %d", resp.StatusCode)
			}
		} else {
			lastErr = err
		}
		time.Sleep(10 * time.Second)
	}
	return lastOut, lastErr
}
