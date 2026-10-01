package verify

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// TempoVerifier checks a Grafana Tempo install to the point a customer
// could ship traces to it and read them back: the Tempo StatefulSet ready,
// the Service present, and a LIVE round-trip — a span pushed over OTLP/HTTP
// and then the trace retrieved by ID through the query API (a trace store
// that cannot store and return a trace is not a trace store).
//
// The behavioral-persistence scenario (recognized by name) additionally
// DELETES the Tempo pod after the push, waits for a REPLACEMENT pod (a new
// UID — status flapping Ready on the dying pod is not recovery), and
// re-queries the trace: traces surviving pod loss through the
// PersistentVolume is the proof.
type TempoVerifier struct {
	Namespace string
	Name      string
	// Persistence switches on the trace-survives-pod-loss proof.
	Persistence bool
	// R2 switches on the object-store proof: after the push, Tempo is
	// asked to shut down its ingester, which writes every block to the
	// backend, and the proof lists this run's objects in the real R2 test
	// bucket (r2_objects.go); the pod-loss proof then reads the trace back
	// from R2 alone.
	R2 bool
}

func (v *TempoVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] tempo %q in namespace %q\n", v.Name, v.Namespace)

	if err := waitStatefulSetReady(ctx, kubeconfig, v.Name, v.Namespace, 10*time.Minute); err != nil {
		return errors.Wrap(err, "the tempo statefulset never became ready")
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "service", v.Name, v.Namespace); err != nil {
		return errors.Wrap(err, "tempo service not found")
	}
	return v.proveRoundTrip(ctx, kubeconfig)
}

func (v *TempoVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "statefulset", v.Name, v.Namespace)
}

// proveRoundTrip pushes one span over OTLP/HTTP (port 4318) and retrieves
// the trace by ID through the query API (port 3200). On the persistence
// lane it kills the Tempo pod between push and query.
func (v *TempoVerifier) proveRoundTrip(ctx context.Context, kubeconfig string) error {
	const httpPort = "13200"
	const otlpPort = "14318"

	pfCancel, err := startPortForward(ctx, kubeconfig, "svc/"+v.Name, v.Namespace,
		httpPort+":3200")
	if err != nil {
		return errors.Wrap(err, "starting port-forward to the tempo query API")
	}
	defer pfCancel()

	otlpCancel, err := startPortForward(ctx, kubeconfig, "svc/"+v.Name, v.Namespace,
		otlpPort+":4318")
	if err != nil {
		return errors.Wrap(err, "starting port-forward to the tempo OTLP receiver")
	}
	defer otlpCancel()

	queryBase := "http://127.0.0.1:" + httpPort
	otlpBase := "http://127.0.0.1:" + otlpPort

	// A deterministic 32-hex trace ID and 16-hex span ID.
	traceId := fmt.Sprintf("%032x", time.Now().UnixNano())
	spanId := fmt.Sprintf("%016x", time.Now().UnixNano())
	startNs := time.Now().UnixNano()
	endNs := startNs + int64(time.Millisecond)

	// OTLP/HTTP JSON trace payload — one resource span with one span.
	pushBody := fmt.Sprintf(`{
      "resourceSpans": [{
        "resource": {"attributes": [{"key": "service.name", "value": {"stringValue": "e2e-proof"}}]},
        "scopeSpans": [{
          "spans": [{
            "traceId": "%s",
            "spanId": "%s",
            "name": "e2e-proof-span",
            "kind": 1,
            "startTimeUnixNano": "%d",
            "endTimeUnixNano": "%d"
          }]
        }]
      }]
    }`, traceId, spanId, startNs, endNs)

	if _, err := httpRoundTrip(ctx, http.MethodPost, otlpBase+"/v1/traces",
		"application/json", pushBody, 5*time.Minute); err != nil {
		return errors.Wrap(err, "pushing the proof span to tempo over OTLP")
	}
	fmt.Printf("  [verify] PUSH: proof span (trace %s) accepted over OTLP\n", traceId)

	if v.R2 {
		since := time.Now()
		// /shutdown flushes all in-memory traces and the WAL to the
		// backend (Tempo's documented pre-stop call); /flush alone only
		// reaches the WAL, and a block waits max_block_duration before
		// leaving it.
		if _, err := httpRoundTrip(ctx, http.MethodPost, queryBase+"/shutdown", "", "", 2*time.Minute); err != nil {
			return errors.Wrap(err, "R2: asking tempo to flush its blocks to the backend")
		}
		keys, err := awaitR2Objects(ctx, "single-tenant/", since, 6*time.Minute)
		if err != nil {
			return errors.Wrap(err, "R2: tempo flushed but no block reached the bucket; check the <name>-r2-credentials Secret, the TEMPO_S3_* variables and the composed endpoint")
		}
		fmt.Printf("  [verify] R2: tempo's blocks landed in the R2 bucket (%s)\n", summarizeKeys(keys))
		defer removeR2Objects(ctx, keys)
	}

	if v.Persistence {
		if err := deletePodAwaitReplacement(ctx, kubeconfig, v.Namespace,
			"app.kubernetes.io/instance="+v.Name, 8*time.Minute); err != nil {
			return errors.Wrap(err, "tempo pod did not recover after deletion")
		}
		pfCancel()
		pfCancel2, err := startPortForward(ctx, kubeconfig, "svc/"+v.Name, v.Namespace, httpPort+":3200")
		if err != nil {
			return errors.Wrap(err, "re-establishing the query port-forward after the pod kill")
		}
		defer pfCancel2()
	}

	// Retrieve the trace by ID. Tempo flushes through the ingester before
	// the trace is queryable, so the retry budget is generous.
	//
	// The query API returns OTLP JSON, where trace/span IDs are protobuf
	// BYTES fields — protojson renders them BASE64, never hex. The success
	// check must therefore match the base64 form of the span ID: a hex
	// substring check holds the successful response and still reports
	// failure (caught live — every lane's "failure" body carried the
	// retrieved trace).
	spanIdBytes, err := hex.DecodeString(spanId)
	if err != nil {
		return errors.Wrap(err, "decoding the proof span id")
	}
	spanIdB64 := base64.StdEncoding.EncodeToString(spanIdBytes)

	deadline := time.Now().Add(6 * time.Minute)
	var lastBody string
	for time.Now().Before(deadline) {
		body, err := httpRoundTrip(ctx, http.MethodGet, queryBase+"/api/traces/"+traceId, "", "", 1*time.Minute)
		if err == nil && strings.Contains(body, spanIdB64) {
			verb := "QUERY"
			if v.Persistence {
				verb = "PERSISTENCE"
			}
			fmt.Printf("  [verify] %s: trace %s retrieved by ID%s\n", verb, traceId,
				map[bool]string{true: " AFTER pod replacement — traces survived " + map[bool]string{true: "in the R2 bucket", false: "on the PVC"}[v.R2], false: ""}[v.Persistence])
			return nil
		}
		lastBody = body
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("the proof trace was never retrievable by ID: %s", firstLines(lastBody, 3))
}
