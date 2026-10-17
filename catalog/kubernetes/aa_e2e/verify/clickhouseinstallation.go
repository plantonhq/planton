package verify

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ClickHouseInstallationVerifier checks an operator-managed ClickHouse
// cluster to the point clients can rely on it: the
// ClickHouseInstallation reaches status Completed with every declared
// host reconciled, the operator's naming contract holds (the
// clickhouse-<name> client Service), and — on every lane — a LIVE SQL
// round-trip (CREATE/INSERT/SELECT) through the HTTP interface as a
// declared user (a database that cannot store and return a row is not
// a database).
//
// The behavioral-durability scenario (recognized by name) uses a
// ReplicatedMergeTree table across the shard's replicas, DELETES a
// replica pod, proves the rows are served DURING the outage, and
// re-proves them after the pod returns.
type ClickHouseInstallationVerifier struct {
	Namespace string
	ChiName   string
	// ClusterName is the CHI's logical cluster (a child-name segment
	// and the ON CLUSTER target).
	ClusterName string
	// TotalHosts is shards × replicas — the .status.hosts target.
	TotalHosts int
	// Username/Password are the first declared user's credentials; the
	// SQL round-trip runs as that user (the built-in default user is
	// network-fenced to the cluster's own pods by the operator). An
	// empty Password means the scenario declares the user without one,
	// and the verifier reads the generated password from the auth Secret.
	Username string
	Password string
	// ManagedKeeper asserts the module-managed Keeper and its client
	// Service when the scenario's coordination calls for one.
	ManagedKeeper bool
	Durability    bool
}

func (v *ClickHouseInstallationVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] clickhouse installation %q in namespace %q\n", v.ChiName, v.Namespace)

	// First boot pulls images, renders per-host config and rolls every
	// host StatefulSet — poll the operator's own status.
	if err := v.waitForCompleted(ctx, kubeconfig, 15*time.Minute); err != nil {
		return err
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "service", "clickhouse-"+v.ChiName, v.Namespace); err != nil {
		return errors.Wrap(err, "cluster client service not found")
	}
	if v.ManagedKeeper {
		keeperName := v.ChiName + "-keeper"
		if err := KubectlResourceExists(ctx, kubeconfig, "clickhousekeeperinstallation", keeperName, v.Namespace); err != nil {
			return errors.Wrap(err, "managed keeper CR not found")
		}
		if err := KubectlResourceExists(ctx, kubeconfig, "service", "keeper-"+keeperName, v.Namespace); err != nil {
			return errors.Wrap(err, "managed keeper client service not found")
		}
	}
	if v.Username == "" {
		fmt.Printf("  [verify] no declared user in the scenario — skipping the SQL round-trip\n")
		return nil
	}
	if v.Password == "" {
		// A user declared without a password authenticates with the one
		// the module generated, read from where every client reads it:
		// the auth Secret, keyed by the user's name. The round-trip below
		// then proves the generated credential end to end.
		authSecret := v.ChiName + "-clickhouse-auth"
		generated, err := readSecretKey(ctx, kubeconfig, v.Namespace, authSecret, v.Username)
		if err != nil {
			return errors.Wrapf(err, "reading the generated password of user %q from secret %s", v.Username, authSecret)
		}
		if generated == "" {
			return errors.Errorf("secret %s carries no generated password under key %q", authSecret, v.Username)
		}
		v.Password = generated
		fmt.Printf("  [verify] GENERATED PASSWORD: user %q authenticates with the module-generated password from %s\n", v.Username, authSecret)
	}
	return v.proveSqlRoundTrip(ctx, kubeconfig)
}

func (v *ClickHouseInstallationVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "clickhouseinstallation", v.ChiName, v.Namespace)
}

// waitForCompleted polls the CHI status until Completed with every
// declared host counted.
func (v *ClickHouseInstallationVerifier) waitForCompleted(ctx context.Context, kubeconfig string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastStatus, lastHosts string
	for time.Now().Before(deadline) {
		status, _ := kubectlGetJSONPath(ctx, kubeconfig, "clickhouseinstallation", v.ChiName, v.Namespace, "{.status.status}")
		hosts, _ := kubectlGetJSONPath(ctx, kubeconfig, "clickhouseinstallation", v.ChiName, v.Namespace, "{.status.hosts}")
		lastStatus, lastHosts = status, hosts
		if status == "Completed" && hosts == fmt.Sprintf("%d", v.TotalHosts) {
			fmt.Printf("  [verify] installation Completed with %s/%d hosts\n", hosts, v.TotalHosts)
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("installation never reached Completed with %d hosts (last status %q, hosts %q)",
		v.TotalHosts, lastStatus, lastHosts)
}

// proveSqlRoundTrip drives the HTTP interface over a port-forward:
// create a database and table, insert a run-unique marker row, and
// select it back. The durability variant uses ReplicatedMergeTree and
// re-proves the select through a live replica loss.
func (v *ClickHouseInstallationVerifier) proveSqlRoundTrip(ctx context.Context, kubeconfig string) error {
	const localPort = "18123"
	service := "clickhouse-" + v.ChiName

	// A kubectl port-forward through a Service binds ONE backing pod at
	// establishment and dies with it — the durability proof DELETES a
	// pod, so the tunnel must be re-establishable (otherwise
	// connection-refused means the dead tunnel, not the cluster).
	startTunnel := func() (stop func(), err error) {
		pfCtx, cancel := context.WithCancel(ctx)
		pf := exec.CommandContext(pfCtx, "kubectl", "--kubeconfig", kubeconfig,
			"port-forward", "svc/"+service, localPort+":8123", "-n", v.Namespace)
		var pfOut strings.Builder
		pf.Stdout = &pfOut
		pf.Stderr = &pfOut
		if err := pf.Start(); err != nil {
			cancel()
			return nil, errors.Wrap(err, "starting port-forward to the cluster service")
		}
		return func() {
			cancel()
			_ = pf.Wait()
		}, nil
	}
	stopTunnel, err := startTunnel()
	if err != nil {
		return err
	}
	defer func() { stopTunnel() }()

	client := &http.Client{Timeout: 30 * time.Second}
	base := "http://127.0.0.1:" + localPort

	marker := fmt.Sprintf("e2e-marker-%d", time.Now().Unix())

	// The durability substrate is a replica-synced table; the plain
	// lanes prove the same round-trip on a local MergeTree.
	onCluster := ""
	engine := "MergeTree"
	if v.Durability {
		onCluster = fmt.Sprintf(" ON CLUSTER '%s'", v.ClusterName)
		engine = "ReplicatedMergeTree"
		// remote_servers propagation LAGS the CHI reaching Completed:
		// the operator publishes the cluster definition through mounted
		// config, and ON CLUSTER DDL initiated before the initiator
		// sees the full definition silently executes on the visible
		// subset only (verified live: the DDL-queue task listed one of
		// two replicas and returned success). The tunnel binds one
		// backing pod, so this gate checks the view of exactly the
		// host that will initiate the DDL below.
		if err := v.awaitClusterView(ctx, client, base, 4*time.Minute); err != nil {
			return errors.Wrap(err, "the initiator's cluster view never converged to every declared host")
		}
	}
	statements := []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS e2e%s", onCluster),
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS e2e.marker%s (id UInt32, marker String) ENGINE = %s ORDER BY id", onCluster, engine),
		fmt.Sprintf("INSERT INTO e2e.marker VALUES (1, '%s')", marker),
	}
	for _, stmt := range statements {
		if err := v.query(ctx, client, base, stmt, 4*time.Minute); err != nil {
			return errors.Wrapf(err, "executing %q", firstLines(stmt, 1))
		}
	}
	if err := v.selectMarker(ctx, client, base, marker, 2*time.Minute); err != nil {
		return errors.Wrap(err, "the inserted row never came back from SELECT")
	}
	fmt.Printf("  [verify] SQL ROUND-TRIP: marker %q inserted and selected as user %q\n", marker, v.Username)

	if !v.Durability {
		return nil
	}

	// THE durability proof: kill the shard's first replica, the rows
	// must stay served from the surviving replica.
	victim := fmt.Sprintf("chi-%s-%s-0-0-0", v.ChiName, v.ClusterName)
	fmt.Printf("  [verify] DURABILITY: deleting replica pod %q\n", victim)
	if out, err := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"delete", "pod", victim, "-n", v.Namespace, "--wait=false").CombinedOutput(); err != nil {
		return errors.Wrapf(err, "deleting replica pod: %s", string(out))
	}
	// The deleted pod may be the tunnel's backing pod — re-establish so
	// a refused connection means the CLUSTER, never the dead tunnel.
	stopTunnel()
	if stopTunnel, err = startTunnel(); err != nil {
		return err
	}
	if err := v.selectMarker(ctx, client, base, marker, 4*time.Minute); err != nil {
		return errors.Wrap(err, "the row was NOT served during the replica outage")
	}
	fmt.Printf("  [verify] DURABILITY: marker served DURING the replica outage\n")

	if err := kubectlWait(ctx, kubeconfig, "pod", victim, v.Namespace,
		"condition=Ready", 8*time.Minute); err != nil {
		return errors.Wrap(err, "the deleted replica never returned to Ready")
	}
	stopTunnel()
	if stopTunnel, err = startTunnel(); err != nil {
		return err
	}
	if err := v.selectMarker(ctx, client, base, marker, 2*time.Minute); err != nil {
		return errors.Wrap(err, "the row was lost after the replica recovered")
	}
	fmt.Printf("  [verify] DURABILITY: replica recovered and marker re-read\n")
	return nil
}

// query POSTs one SQL statement, retrying until success or the budget
// expires (first statements race host config propagation and, on
// replicated lanes, distributed-DDL readiness).
func (v *ClickHouseInstallationVerifier) query(ctx context.Context, client *http.Client, base, sql string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/", strings.NewReader(sql))
		if err != nil {
			return err
		}
		req.Header.Set("X-ClickHouse-User", v.Username)
		req.Header.Set("X-ClickHouse-Key", v.Password)
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Second)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		lastErr = errors.Errorf("HTTP %d: %s", resp.StatusCode, firstLines(string(body), 2))
		time.Sleep(5 * time.Second)
	}
	return lastErr
}

// awaitClusterView retries until the connected host's own view of the
// declared cluster (system.clusters) carries every declared host —
// the precondition for distributed DDL to reach all of them.
func (v *ClickHouseInstallationVerifier) awaitClusterView(ctx context.Context, client *http.Client, base string, budget time.Duration) error {
	sql := fmt.Sprintf("SELECT count() FROM system.clusters WHERE cluster = '%s'", v.ClusterName)
	deadline := time.Now().Add(budget)
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			base+"/?query="+url.QueryEscape(sql), nil)
		if err != nil {
			return err
		}
		req.Header.Set("X-ClickHouse-User", v.Username)
		req.Header.Set("X-ClickHouse-Key", v.Password)
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Second)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if got := strings.TrimSpace(string(body)); got == fmt.Sprintf("%d", v.TotalHosts) {
				fmt.Printf("  [verify] cluster view converged: %s/%d hosts visible to the DDL initiator\n", got, v.TotalHosts)
				return nil
			} else {
				lastErr = errors.Errorf("cluster %q shows %s of %d declared hosts", v.ClusterName, got, v.TotalHosts)
			}
		} else {
			lastErr = errors.Errorf("HTTP %d: %s", resp.StatusCode, firstLines(string(body), 2))
		}
		time.Sleep(5 * time.Second)
	}
	return lastErr
}

// selectMarker retries until the marker row comes back.
func (v *ClickHouseInstallationVerifier) selectMarker(ctx context.Context, client *http.Client, base, marker string, budget time.Duration) error {
	sql := "SELECT marker FROM e2e.marker WHERE id = 1"
	deadline := time.Now().Add(budget)
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			base+"/?query="+url.QueryEscape(sql), nil)
		if err != nil {
			return err
		}
		req.Header.Set("X-ClickHouse-User", v.Username)
		req.Header.Set("X-ClickHouse-Key", v.Password)
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(5 * time.Second)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK && strings.Contains(string(body), marker) {
			return nil
		}
		lastErr = errors.Errorf("select HTTP %d: %s", resp.StatusCode, firstLines(string(body), 2))
		time.Sleep(5 * time.Second)
	}
	return lastErr
}

// clickhouseScenarioShape pulls the verifier's inputs out of a
// KubernetesClickHouse scenario manifest: the logical cluster name,
// total host count, first declared user, and whether the module
// deploys a managed Keeper (coordination unset with a multi-host
// topology, or an explicit managed_keeper).
func clickhouseScenarioShape(spec map[string]interface{}) (clusterName string, totalHosts int, username, password string, managedKeeper bool) {
	clusterName = "main"
	if s, ok := spec["cluster_name"].(string); ok && s != "" {
		clusterName = s
	}
	shards, replicas := 1, 1
	if n, ok := specInt(spec["shards"]); ok {
		shards = n
	}
	if n, ok := specInt(spec["replicas"]); ok {
		replicas = n
	}
	totalHosts = shards * replicas

	if users, ok := spec["users"].([]interface{}); ok && len(users) > 0 {
		if user, ok := users[0].(map[string]interface{}); ok {
			username, _ = user["name"].(string)
			if pw, ok := user["password"].(map[string]interface{}); ok {
				password, _ = pw["value"].(string)
			}
		}
	}

	coordination, hasCoordination := spec["coordination"].(map[string]interface{})
	if !hasCoordination {
		managedKeeper = shards > 1 || replicas > 1
		return
	}
	coordinationType, _ := coordination["type"].(string)
	managedKeeper = coordinationType == "managed_keeper" ||
		(coordinationType == "" && (shards > 1 || replicas > 1))
	return
}

// specInt reads a YAML-decoded scalar that may arrive as int or float.
func specInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}
