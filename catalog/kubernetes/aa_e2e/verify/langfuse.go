package verify

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// LangfuseVerifier checks a Langfuse install to the point a customer could
// point a Langfuse SDK at it:
//
//   - both workloads rolled out (the web server is Ready only after its
//     Postgres and ClickHouse migrations ran);
//   - the web server's own health and readiness contracts answer;
//   - THE AUTH GATE: an anonymous public-API read is refused (401);
//   - THE INGESTION PROOF on every lane: the seeded project key pair is read
//     from the instance Secret, a span is pushed as OTLP/JSON to the
//     `otel_endpoint` output's path with Basic auth built from the key pair
//     (Langfuse v4's trace ingestion door - its legacy ingestion API takes
//     only scores), and the span is read back by trace ID through the v4
//     observations API - a round trip through the web server, blob storage,
//     the Redis queue, the worker and ClickHouse, so every composed store is
//     proven wired at once.
//
// The full-surface scenario (recognized by name) adds THE SCORE PROOF: a
// score on the ingested trace is accepted by the public ingestion API; and
// asserts the seeded admin's generated password exists. The
// behavioral-state and reinstall scenarios carry a deterministic STATE
// TRACE: the first verify ingests it, every later verify (the upgrade act,
// the reinstall) finds it again - the upgrade act (a `.to.yaml` manifest)
// requires it.
//
// Destroy leaves one object by the chart's own design: the `<name>-app`
// Secret is marked helm.sh/resource-policy: keep (empty here - every
// credential it would hold is supplied by the module). The verifier
// asserts the workloads, the web Service and the module's instance Secret
// are gone; the kept Secret leaves with the namespace.
type LangfuseVerifier struct {
	Namespace string
	Name      string
	WebPort   string
	// OtlpProof switches the score arm and the admin-password check on.
	OtlpProof bool
	// StateTrace carries the deterministic state trace; RequireState
	// demands it already exists (the upgrade act).
	StateTrace   bool
	RequireState bool
}

const langfuseLocalPort = "13000"

func (v *LangfuseVerifier) baseUrl() string { return "http://127.0.0.1:" + langfuseLocalPort }

func (v *LangfuseVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] langfuse %q in namespace %q\n", v.Name, v.Namespace)
	for _, component := range []string{"web", "worker"} {
		if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/"+v.Name+"-"+component, v.Namespace, 15*time.Minute); err != nil {
			return errors.Wrapf(err, "the langfuse %s deployment never rolled out", component)
		}
	}

	cancel, err := openServiceTunnel(ctx, kubeconfig, v.Namespace, v.Name+"-web", langfuseLocalPort, v.WebPort)
	if err != nil {
		return errors.Wrap(err, "opening the tunnel to the langfuse web service")
	}
	defer cancel()

	client := &http.Client{Timeout: 30 * time.Second}
	for _, path := range []string{"/api/public/health", "/api/public/ready"} {
		if err := v.awaitStatus(ctx, client, http.MethodGet, path, "", http.StatusOK, 3*time.Minute); err != nil {
			return errors.Wrapf(err, "%s never answered 200", path)
		}
	}
	fmt.Printf("  [verify] HEALTH: /api/public/health and /api/public/ready answer 200\n")

	if err := v.awaitStatus(ctx, client, http.MethodGet, "/api/public/projects", "", http.StatusUnauthorized, time.Minute); err != nil {
		return errors.Wrap(err, "THE AUTH GATE: an anonymous public-API read was not refused")
	}
	fmt.Printf("  [verify] AUTH GATE: an anonymous public-API read is refused (401)\n")

	secretName := v.Name + "-secrets"
	publicKey, err := readSecretKey(ctx, kubeconfig, v.Namespace, secretName, "init-project-public-key")
	if err != nil {
		return errors.Wrap(err, "reading the seeded project public key from the instance secret")
	}
	secretKey, err := readSecretKey(ctx, kubeconfig, v.Namespace, secretName, "init-project-secret-key")
	if err != nil {
		return errors.Wrap(err, "reading the seeded project secret key from the instance secret")
	}
	auth := "Basic " + basicAuth(publicKey, secretKey)

	traceId := randomHex(16)
	if err := v.pushOtlp(ctx, client, auth, traceId); err != nil {
		return err
	}
	if err := v.awaitObservation(ctx, client, auth, traceId, 5*time.Minute); err != nil {
		return errors.Wrap(err, "THE INGESTION PROOF: the OTLP span never came back through the observations API")
	}
	fmt.Printf("  [verify] INGESTION: span pushed over OTLP with the seeded key pair and read back as trace %s - web, blob storage, queue, worker and ClickHouse are wired\n", traceId)

	if v.StateTrace {
		// OTLP trace IDs are 32 hex characters: the state trace's ID is
		// derived from the resource name, so every act computes the same one.
		digest := sha256.Sum256([]byte("e2e-state-" + v.Name))
		stateTraceId := hex.EncodeToString(digest[:16])
		found, err := v.observationExists(ctx, client, auth, stateTraceId)
		if err != nil {
			return err
		}
		switch {
		case found:
			fmt.Printf("  [verify] STATE: trace %s from an earlier act is still readable with the same key pair\n", stateTraceId)
		case v.RequireState:
			return errors.Errorf("THE STATE PROOF: trace %s ingested before the upgrade is gone", stateTraceId)
		default:
			if err := v.pushOtlp(ctx, client, auth, stateTraceId); err != nil {
				return err
			}
			if err := v.awaitObservation(ctx, client, auth, stateTraceId, 5*time.Minute); err != nil {
				return errors.Wrap(err, "the state trace never came back")
			}
			fmt.Printf("  [verify] STATE: trace %s ingested for the next act\n", stateTraceId)
		}
	}

	if v.OtlpProof {
		if err := v.ingestScore(ctx, client, auth, traceId); err != nil {
			return errors.Wrap(err, "THE SCORE PROOF")
		}
		fmt.Printf("  [verify] SCORE: a score on trace %s accepted by /api/public/ingestion\n", traceId)

		password, err := readSecretKey(ctx, kubeconfig, v.Namespace, secretName, "init-user-password")
		if err != nil || password == "" {
			return errors.Errorf("the seeded admin's generated password is missing from %s (key init-user-password)", secretName)
		}
		fmt.Printf("  [verify] SEEDED ADMIN: a generated password is in %s\n", secretName)
	}
	return nil
}

func (v *LangfuseVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	for _, check := range []struct{ kind, name string }{
		{"deployment", v.Name + "-web"},
		{"deployment", v.Name + "-worker"},
		{"service", v.Name + "-web"},
		{"secret", v.Name + "-secrets"},
	} {
		if err := KubectlResourceAbsent(ctx, kubeconfig, check.kind, check.name, v.Namespace); err != nil {
			return err
		}
	}
	fmt.Printf("  [verify] DESTROY: both workloads, the web service and the instance secret are gone (the chart's kept <name>-app secret leaves with the namespace)\n")
	return nil
}

// ingestScore posts a score-create on traceId through the public ingestion
// API, the one event type it accepts in Langfuse v4 (207 with no per-event
// error).
func (v *LangfuseVerifier) ingestScore(ctx context.Context, client *http.Client, auth, traceId string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	batch := map[string]interface{}{"batch": []interface{}{
		map[string]interface{}{
			"id": randomHex(16), "timestamp": now, "type": "score-create",
			"body": map[string]interface{}{
				"id": randomHex(16), "traceId": traceId, "name": "e2e-score", "value": 1,
			},
		},
	}}
	body, _ := json.Marshal(batch)
	status, response, err := v.do(ctx, client, http.MethodPost, "/api/public/ingestion", auth, "application/json", body)
	if err != nil {
		return errors.Wrap(err, "posting to the ingestion API")
	}
	if status != http.StatusMultiStatus && status != http.StatusOK {
		return errors.Errorf("the ingestion API answered %d: %s", status, firstLines(response, 3))
	}
	var parsed struct {
		Errors []interface{} `json:"errors"`
	}
	_ = json.Unmarshal([]byte(response), &parsed)
	if len(parsed.Errors) > 0 {
		return errors.Errorf("the ingestion API rejected events: %s", firstLines(response, 3))
	}
	return nil
}

// pushOtlp posts one span as OTLP/JSON.
func (v *LangfuseVerifier) pushOtlp(ctx context.Context, client *http.Client, auth, traceId string) error {
	now := time.Now().UnixNano()
	payload := map[string]interface{}{"resourceSpans": []interface{}{map[string]interface{}{
		"resource": map[string]interface{}{"attributes": []interface{}{
			map[string]interface{}{"key": "service.name", "value": map[string]interface{}{"stringValue": "e2e-otlp"}},
		}},
		"scopeSpans": []interface{}{map[string]interface{}{
			"scope": map[string]interface{}{"name": "e2e"},
			"spans": []interface{}{map[string]interface{}{
				"traceId":           traceId,
				"spanId":            randomHex(8),
				"name":              "e2e-otlp-span",
				"kind":              1,
				"startTimeUnixNano": fmt.Sprintf("%d", now-int64(time.Millisecond)),
				"endTimeUnixNano":   fmt.Sprintf("%d", now),
			}},
		}},
	}}}
	body, _ := json.Marshal(payload)
	status, response, err := v.do(ctx, client, http.MethodPost, "/api/public/otel/v1/traces", auth, "application/json", body)
	if err != nil {
		return errors.Wrap(err, "posting to the OTLP endpoint")
	}
	if status < 200 || status > 299 {
		return errors.Errorf("the OTLP endpoint answered %d: %s", status, firstLines(response, 3))
	}
	return nil
}

func (v *LangfuseVerifier) observationExists(ctx context.Context, client *http.Client, auth, traceId string) (bool, error) {
	status, response, err := v.do(ctx, client, http.MethodGet,
		"/api/public/v2/observations?traceId="+url.QueryEscape(traceId)+"&limit=10", auth, "", nil)
	if err != nil {
		return false, errors.Wrap(err, "querying the observations API")
	}
	if status != http.StatusOK {
		return false, errors.Errorf("the observations API answered %d: %s", status, firstLines(response, 3))
	}
	var parsed struct {
		Data []struct {
			TraceId string `json:"traceId"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return false, errors.Wrapf(err, "parsing the observations API response: %s", firstLines(response, 3))
	}
	for _, o := range parsed.Data {
		if o.TraceId == traceId {
			return true, nil
		}
	}
	return false, nil
}

// awaitObservation polls until an observation of traceId is readable - the
// worker loads ingested events asynchronously.
func (v *LangfuseVerifier) awaitObservation(ctx context.Context, client *http.Client, auth, traceId string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastErr error
	for time.Now().Before(deadline) {
		found, err := v.observationExists(ctx, client, auth, traceId)
		if found {
			return nil
		}
		lastErr = err
		if lastErr == nil {
			lastErr = errors.Errorf("no observation of trace %s yet", traceId)
		}
		time.Sleep(5 * time.Second)
	}
	return lastErr
}

func (v *LangfuseVerifier) awaitStatus(ctx context.Context, client *http.Client, method, path, auth string, want int, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastErr error
	for time.Now().Before(deadline) {
		status, response, err := v.do(ctx, client, method, path, auth, "", nil)
		if err == nil && status == want {
			return nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = errors.Errorf("HTTP %d (want %d): %s", status, want, firstLines(response, 2))
		}
		time.Sleep(5 * time.Second)
	}
	return lastErr
}

func (v *LangfuseVerifier) do(ctx context.Context, client *http.Client, method, path, auth, contentType string, body []byte) (int, string, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, v.baseUrl()+path, reader)
	if err != nil {
		return 0, "", err
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out), nil
}

// langfuseWebPort is the port the web Service exposes: externalPort when
// set, the container port otherwise (the chart's `externalPort | default
// port`).
func langfuseWebPort(spec map[string]interface{}) string {
	if web, ok := spec["web"].(map[string]interface{}); ok {
		if service, ok := web["service"].(map[string]interface{}); ok {
			if port, ok := specInt(service["externalPort"]); ok {
				return fmt.Sprintf("%d", port)
			}
			if port, ok := specInt(service["port"]); ok {
				return fmt.Sprintf("%d", port)
			}
		}
	}
	return "3000"
}

func basicAuth(user, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
}

func randomHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
