package verify

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// LokiVerifier checks a Grafana Loki install to the point a customer could
// ship logs to it and read them back: the Loki workload ready, the gateway
// Service present, and a LIVE push→query round-trip — a log line pushed
// through the gateway's Loki push API and then returned by a LogQL query
// (a log store that cannot store and return a line is not a log store).
//
// The behavioral-durability scenario (recognized by name) additionally
// DELETES the Loki pod after the push, waits for a REPLACEMENT pod (a new
// UID — status flapping Ready on the dying pod is not recovery), and
// re-queries the line: logs surviving pod loss through the PersistentVolume
// is the proof.
type LokiVerifier struct {
	Namespace string
	Name      string
	// CanaryOff marks a spec with canary_enabled false.
	CanaryOff bool
	// GatewayEnabled marks whether the nginx gateway front door is
	// deployed (the exported endpoints and this proof route through it).
	GatewayEnabled bool
	// TenantUser / TenantPassword carry the gateway basic-auth identity
	// for multi-tenant installs. With tenants declared the chart guards
	// the WHOLE gateway server with auth_basic (only the bare health
	// route is exempt) and injects X-Scope-OrgID from the authenticated
	// username, so the push→query proof MUST authenticate — and doing so
	// proves the full tenant path: htpasswd materialized from the
	// declared bcrypt hash, nginx auth, OrgID injection, tenant-scoped
	// storage and query. Empty = single-tenant posture (no auth).
	TenantUser     string
	TenantPassword string
	// Durability switches on the log-survives-pod-loss proof.
	Durability bool
	// R2 switches on the object-store proof: after the push, Loki flushes
	// its chunks and the proof lists this run's objects in the real R2
	// test bucket (r2_objects.go) before the pod-loss proof reads the
	// line back.
	R2 bool
}

func (v *LokiVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] loki %q in namespace %q\n", v.Name, v.Namespace)

	// Monolithic mode runs a single StatefulSet named after the release
	// (fullnameOverride). Wait for it before touching the gateway.
	if err := waitStatefulSetReady(ctx, kubeconfig, v.Name, v.Namespace, 10*time.Minute); err != nil {
		return errors.Wrap(err, "the loki statefulset never became ready")
	}
	// An install with the canary off is the proof by itself (the chart
	// refuses to render it while its Helm test is on); the canary's
	// DaemonSet must then be absent.
	if v.CanaryOff {
		if err := KubectlResourceAbsent(ctx, kubeconfig, "daemonset", v.Name+"-canary", v.Namespace); err != nil {
			return errors.Wrap(err, "the canary is off in the spec but its DaemonSet exists")
		}
		fmt.Printf("  [verify] CANARY OFF: installed, and no canary DaemonSet\n")
	}
	if !v.GatewayEnabled {
		return errors.New("the log round-trip requires the gateway (disabled in this scenario) — enable it or address the internal services directly")
	}
	gatewaySvc := v.Name + "-gateway"
	if err := KubectlResourceExists(ctx, kubeconfig, "service", gatewaySvc, v.Namespace); err != nil {
		return errors.Wrap(err, "loki gateway service not found")
	}
	return v.proveRoundTrip(ctx, kubeconfig, gatewaySvc)
}

func (v *LokiVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "statefulset", v.Name, v.Namespace)
}

// proveRoundTrip pushes a uniquely-labelled log line through the gateway's
// Loki push API and queries it back through LogQL. On the durability lane
// it kills the Loki pod between push and query.
func (v *LokiVerifier) proveRoundTrip(ctx context.Context, kubeconfig, gatewaySvc string) error {
	const localPort = "13100"

	pfCancel, err := startPortForward(ctx, kubeconfig, "svc/"+gatewaySvc, v.Namespace, localPort+":80")
	if err != nil {
		return errors.Wrap(err, "starting port-forward to the loki gateway")
	}
	defer pfCancel()

	base := "http://127.0.0.1:" + localPort
	marker := fmt.Sprintf("e2e-loki-proof-%d", time.Now().UnixNano())
	nowNs := fmt.Sprintf("%d", time.Now().UnixNano())

	pushBody := fmt.Sprintf(
		`{"streams":[{"stream":{"job":"e2e-proof"},"values":[["%s","%s"]]}]}`,
		nowNs, marker)

	// Loki's distributor can 5xx briefly during warm-up — the retry loop
	// covers the window; a 204 is success.
	if _, err := httpRoundTripAuth(ctx, http.MethodPost, base+"/loki/api/v1/push",
		"application/json", pushBody, 5*time.Minute, v.TenantUser, v.TenantPassword); err != nil {
		return errors.Wrap(err, "pushing the proof log line to loki")
	}
	as := ""
	if v.TenantUser != "" {
		as = fmt.Sprintf(" as tenant %q", v.TenantUser)
	}
	fmt.Printf("  [verify] PUSH: proof log line %q accepted by the gateway%s\n", marker, as)

	if v.R2 {
		keys, err := v.proveChunksInR2(ctx, kubeconfig)
		if err != nil {
			return err
		}
		defer removeR2Objects(ctx, keys)
	}

	if v.Durability {
		if err := deletePodAwaitReplacement(ctx, kubeconfig, v.Namespace,
			"app.kubernetes.io/instance="+v.Name, 8*time.Minute); err != nil {
			return errors.Wrap(err, "loki pod did not recover after deletion")
		}
		// The old tunnel died with the gateway pod's peer; re-establish.
		pfCancel()
		pfCancel2, err := startPortForward(ctx, kubeconfig, "svc/"+gatewaySvc, v.Namespace, localPort+":80")
		if err != nil {
			return errors.Wrap(err, "re-establishing the port-forward after the pod kill")
		}
		defer pfCancel2()
	}

	// LogQL query over the last hour; the marker must come back.
	query := `{job="e2e-proof"}`
	start := fmt.Sprintf("%d", time.Now().Add(-1*time.Hour).UnixNano())
	end := fmt.Sprintf("%d", time.Now().Add(1*time.Minute).UnixNano())
	deadline := time.Now().Add(6 * time.Minute)
	var lastBody string
	for time.Now().Before(deadline) {
		url := fmt.Sprintf("%s/loki/api/v1/query_range?query=%s&start=%s&end=%s&limit=100",
			base, urlQueryEscape(query), start, end)
		body, err := httpRoundTripAuth(ctx, http.MethodGet, url, "", "", 1*time.Minute, v.TenantUser, v.TenantPassword)
		if err == nil && strings.Contains(body, marker) {
			verb := "QUERY"
			if v.Durability {
				verb = "DURABILITY"
			}
			fmt.Printf("  [verify] %s: proof log line returned by LogQL%s%s\n", verb, as,
				map[bool]string{true: " AFTER pod replacement — logs survived " + map[bool]string{true: "in the R2 bucket", false: "on the PVC"}[v.R2], false: ""}[v.Durability])
			return nil
		}
		lastBody = body
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("the proof log line was never returned by LogQL: %s", firstLines(lastBody, 3))
}

// proveChunksInR2 asks Loki to flush its in-memory chunks (POST /flush on
// the Loki HTTP port) and requires chunk objects written by this run under
// the single-tenant prefix `fake/` in the R2 test bucket.
func (v *LokiVerifier) proveChunksInR2(ctx context.Context, kubeconfig string) ([]string, error) {
	const flushPort = "13101"
	since := time.Now()
	cancel, err := startPortForward(ctx, kubeconfig, "svc/"+v.Name, v.Namespace, flushPort+":3100")
	if err != nil {
		return nil, errors.Wrap(err, "R2: starting port-forward to the loki HTTP port")
	}
	defer cancel()
	if _, err := httpRoundTrip(ctx, http.MethodPost, "http://127.0.0.1:"+flushPort+"/flush", "", "", 2*time.Minute); err != nil {
		return nil, errors.Wrap(err, "R2: asking loki to flush its chunks")
	}
	keys, err := awaitR2Objects(ctx, "fake/", since, 5*time.Minute)
	if err != nil {
		return nil, errors.Wrap(err, "R2: loki flushed but no chunk reached the bucket; check the <name>-r2-credentials Secret, the LOKI_S3_* variables and the composed endpoint")
	}
	fmt.Printf("  [verify] R2: loki's chunks landed in the R2 bucket (%s)\n", summarizeKeys(keys))
	return keys, nil
}

// httpRoundTrip performs one JSON request retrying across a warm-up window;
// non-2xx is an error, and the body is read inside the loop so a response
// dying mid-stream retries rather than escaping.
func httpRoundTrip(ctx context.Context, method, url, contentType, body string, budget time.Duration) (string, error) {
	return httpRoundTripAuth(ctx, method, url, contentType, body, budget, "", "")
}

// httpRoundTripAuth is httpRoundTrip with optional basic-auth credentials
// (empty user = anonymous) — the front door for auth-gated gateways.
func httpRoundTripAuth(ctx context.Context, method, url, contentType, body string, budget time.Duration, user, password string) (string, error) {
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
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
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
				lastErr = errors.Errorf("HTTP %d: %s", resp.StatusCode, firstLines(lastOut, 1))
			}
		} else {
			lastErr = err
		}
		time.Sleep(10 * time.Second)
	}
	return lastOut, lastErr
}

// startPortForward launches a kubectl port-forward and returns a cancel
// func that tears it down (cancel FIRST, then Wait — Wait blocks forever on
// a port-forward never told to exit).
func startPortForward(ctx context.Context, kubeconfig, target, namespace, ports string) (func(), error) {
	pfCtx, cancel := context.WithCancel(ctx)
	pf := exec.CommandContext(pfCtx, "kubectl", "--kubeconfig", kubeconfig,
		"port-forward", target, ports, "-n", namespace)
	var out strings.Builder
	pf.Stdout = &out
	pf.Stderr = &out
	if err := pf.Start(); err != nil {
		cancel()
		return nil, err
	}
	// Give the tunnel a moment to bind before the first request.
	time.Sleep(3 * time.Second)
	return func() {
		cancel()
		_ = pf.Wait()
	}, nil
}

// deletePodAwaitReplacement deletes the workload's pod(s) by selector and
// waits until a pod with a NEW UID is Ready — a new UID is the only honest
// recovery signal (status can flap Ready against the dying pod). Every
// pre-delete pod's uid goes into the "old" set, and each poll reads uid,
// phase and readiness for ALL matched pods in ONE query: during a
// Deployment-managed replacement the selector transiently matches both
// the dying pod and its successor, so two separate queries indexing
// .items[0] can each land on a DIFFERENT pod and never agree.
func deletePodAwaitReplacement(ctx context.Context, kubeconfig, namespace, selector string, budget time.Duration) error {
	uidOut, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"get", "pods", "-n", namespace, "-l", selector, "-o", "jsonpath={.items[*].metadata.uid}").CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "reading the pod uids: %s", string(uidOut))
	}
	oldUids := map[string]bool{}
	for _, uid := range strings.Fields(string(uidOut)) {
		oldUids[uid] = true
	}
	if len(oldUids) == 0 {
		return errors.Errorf("no pod matched selector %q before the deletion", selector)
	}

	fmt.Printf("  [verify] DURABILITY: deleting %d pod(s) matching %q\n", len(oldUids), selector)
	if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"delete", "pod", "-n", namespace, "-l", selector, "--wait=false").CombinedOutput(); err != nil {
		return errors.Wrapf(err, "deleting the pod: %s", string(out))
	}

	deadline := time.Now().Add(budget)
	for time.Now().Before(deadline) {
		out, _ := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "pods", "-n", namespace, "-l", selector,
			"-o", `jsonpath={range .items[*]}{.metadata.uid} {.status.phase} {.status.conditions[?(@.type=='Ready')].status}{"\n"}{end}`).CombinedOutput()
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 3 && !oldUids[fields[0]] && fields[1] == "Running" && fields[2] == "True" {
				fmt.Printf("  [verify] DURABILITY: replacement pod (uid %s) is Ready\n", fields[0])
				return nil
			}
		}
		time.Sleep(10 * time.Second)
	}
	return errors.New("no replacement pod became Ready after the deletion")
}

// urlQueryEscape percent-escapes a LogQL/TraceQL query for a URL query
// parameter.
func urlQueryEscape(q string) string {
	var b strings.Builder
	for _, r := range q {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.', r == '~':
			b.WriteRune(r)
		default:
			b.WriteString(fmt.Sprintf("%%%02X", r))
		}
	}
	return b.String()
}

// waitStatefulSetReady blocks until the named StatefulSet reports at least
// one ready replica matching its desired count.
func waitStatefulSetReady(ctx context.Context, kubeconfig, name, namespace string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var last string
	for time.Now().Before(deadline) {
		desired, _ := kubectlGetJSONPath(ctx, kubeconfig, "statefulset", name, namespace, "{.status.replicas}")
		ready, _ := kubectlGetJSONPath(ctx, kubeconfig, "statefulset", name, namespace, "{.status.readyReplicas}")
		last = fmt.Sprintf("ready=%q desired=%q", ready, desired)
		if ready != "" && ready != "0" && ready == desired {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("statefulset %q never reached its ready replica count (last %s)", name, last)
}
