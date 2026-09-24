package verify

import (
	"context"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Neo4jVerifier checks a Neo4j server to the point clients can rely on
// it: the StatefulSet's pod ready, the main Service present, and — on
// every lane with declared credentials — a LIVE Cypher write/read
// round-trip through cypher-shell inside the pod (a graph database that
// cannot persist a node is not a database).
//
// The behavioral-persistence scenario (recognized by name) additionally
// DELETES the pod after the write and re-reads the node once the pod
// returns — data surviving pod loss through the PVC is the proof.
type Neo4jVerifier struct {
	Namespace string
	Name      string
	// AuthSecretName holds the NEO4J_AUTH credential ("neo4j/<password>"):
	// the module-materialized `<name>-auth` (declared or module-generated
	// password) or the referenced existing Secret. The module always names
	// one, so the Cypher proof always runs; the empty arm is kept only as a
	// guard against a scenario the resolver could not read.
	AuthSecretName string
	Persistence    bool
}

func (v *Neo4jVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] neo4j server %q in namespace %q\n", v.Name, v.Namespace)

	if err := v.waitForPodReady(ctx, kubeconfig, 10*time.Minute); err != nil {
		return err
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "service", v.Name, v.Namespace); err != nil {
		return errors.Wrap(err, "neo4j service not found")
	}
	if v.AuthSecretName == "" {
		return errors.New("no credentials Secret resolved for the neo4j server — the module materializes <name>-auth for every arm but existing_secret, so an empty name is a verifier defect, not a scenario shape")
	}
	return v.proveCypherRoundTrip(ctx, kubeconfig)
}

func (v *Neo4jVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "statefulset", v.Name, v.Namespace)
}

// waitForPodReady polls the StatefulSet (the chart names it after the
// server) until its single replica is ready.
func (v *Neo4jVerifier) waitForPodReady(ctx context.Context, kubeconfig string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastReady string
	for time.Now().Before(deadline) {
		ready, _ := kubectlGetJSONPath(ctx, kubeconfig, "statefulset", v.Name, v.Namespace, "{.status.readyReplicas}")
		lastReady = ready
		if ready == "1" {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("neo4j statefulset never became ready (last readyReplicas %q)", lastReady)
}

// password reads the NEO4J_AUTH value ("neo4j/<password>") and strips
// the username prefix — the chart's credential contract.
func (v *Neo4jVerifier) password(ctx context.Context, kubeconfig string) (string, error) {
	authB64, err := kubectlGetJSONPath(ctx, kubeconfig, "secret", v.AuthSecretName, v.Namespace, "{.data.NEO4J_AUTH}")
	if err != nil {
		return "", errors.Wrapf(err, "reading secret %q NEO4J_AUTH", v.AuthSecretName)
	}
	auth, err := base64.StdEncoding.DecodeString(authB64)
	if err != nil {
		return "", err
	}
	parts := strings.SplitN(strings.TrimSpace(string(auth)), "/", 2)
	if len(parts) != 2 {
		return "", errors.Errorf("secret %q NEO4J_AUTH is not in the neo4j/<password> form", v.AuthSecretName)
	}
	return parts[1], nil
}

// proveCypherRoundTrip MERGEs a run-unique marker node and reads it
// back through cypher-shell; the persistence variant kills the pod
// between write and read.
func (v *Neo4jVerifier) proveCypherRoundTrip(ctx context.Context, kubeconfig string) error {
	password, err := v.password(ctx, kubeconfig)
	if err != nil {
		return err
	}
	pod := v.Name + "-0"
	marker := fmt.Sprintf("e2e-marker-%d", time.Now().Unix())

	write := fmt.Sprintf("MERGE (n:E2EMarker {id: '%s'}) RETURN n.id", marker)
	if out, err := v.cypher(ctx, kubeconfig, pod, password, write, 4*time.Minute); err != nil {
		return errors.Wrapf(err, "the Cypher write never succeeded: %s", out)
	}
	fmt.Printf("  [verify] CYPHER: marker node %q written\n", marker)

	if v.Persistence {
		fmt.Printf("  [verify] PERSISTENCE: deleting pod %q\n", pod)
		if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"delete", "pod", pod, "-n", v.Namespace, "--wait=false").CombinedOutput(); err != nil {
			return errors.Wrapf(err, "deleting the neo4j pod: %s", string(out))
		}
		if err := v.waitForPodReady(ctx, kubeconfig, 10*time.Minute); err != nil {
			return errors.Wrap(err, "the pod never returned after deletion")
		}
	}

	read := fmt.Sprintf("MATCH (n:E2EMarker {id: '%s'}) RETURN n.id", marker)
	out, err := v.cypher(ctx, kubeconfig, pod, password, read, 4*time.Minute)
	if err != nil {
		return errors.Wrapf(err, "the Cypher read never succeeded: %s", out)
	}
	if !strings.Contains(out, marker) {
		return errors.Errorf("the marker node did not survive: %s", firstLines(out, 3))
	}
	if v.Persistence {
		fmt.Printf("  [verify] PERSISTENCE: marker read back AFTER pod loss — data survived on the PVC\n")
	} else {
		fmt.Printf("  [verify] CYPHER: marker read back\n")
	}
	return nil
}

// cypher retries cypher-shell inside the pod until it succeeds (the bolt
// listener answers probes before auth is fully warm on first boot).
func (v *Neo4jVerifier) cypher(ctx context.Context, kubeconfig, pod, password, statement string, budget time.Duration) (string, error) {
	deadline := time.Now().Add(budget)
	var lastOut string
	var lastErr error
	for time.Now().Before(deadline) {
		out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"exec", pod, "-n", v.Namespace, "--",
			"cypher-shell", "-u", "neo4j", "-p", password, statement).CombinedOutput()
		lastOut = string(out)
		if err == nil {
			return lastOut, nil
		}
		lastErr = err
		time.Sleep(10 * time.Second)
	}
	return lastOut, lastErr
}
