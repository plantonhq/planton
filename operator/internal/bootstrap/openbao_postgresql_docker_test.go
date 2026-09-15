//go:build requires_docker

package bootstrap

// The storage decision, proven against real servers: an OpenBao initialized
// on PostgreSQL storage keeps its whole state -- init, unseal keys' effect,
// mounts -- in the database, not the process. The proof is a second OpenBao
// started against the same database that finds itself initialized and
// sealed, opens with the first one's keys, and already has the mounts. That
// is exactly what a platform restore does to the bundled vault.
//
// Container management is hand-rolled docker CLI, like the Keycloak suite:
// the whole need is a few `docker run`s on one network plus readiness polls.
// Run with `go test -tags=requires_docker ./internal/bootstrap/ -run
// TestOpenBAO_PostgreSQLStorage`.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

const (
	pgTestImage      = "postgres:16"
	openbaoTestImage = "quay.io/openbao/openbao:2.6.1"
	pgTestRole       = "openbao"
	pgTestPassword   = "storage-proof-password"
	pgTestDatabase   = "openbao"
)

// The same stanza the operator renders (openbao_helm.go), as the JSON the
// image's entrypoint accepts through BAO_LOCAL_CONFIG; the URL rides the
// environment variable the backend reads first, exactly as the chart projects
// it. disable_mlock because the container has no IPC_LOCK.
const openbaoTestConfig = `{
  "ui": false,
  "disable_mlock": true,
  "listener": {"tcp": {"address": "[::]:8200", "tls_disable": 1}},
  "storage": {"postgresql": {"max_parallel": "32"}}
}`

func TestOpenBAO_PostgreSQLStorage(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not on PATH")
	}
	network := fmt.Sprintf("openbao-pg-proof-%d", os.Getpid())
	mustDocker(t, "network", "create", network)
	t.Cleanup(func() { _ = exec.Command("docker", "network", "rm", network).Run() })

	// PostgreSQL with the vault's own role and database -- the shape the
	// database component declares through CloudNativePG in the operator.
	pg := "openbao-proof-pg"
	_ = exec.Command("docker", "rm", "-f", pg).Run()
	mustDocker(t, "run", "-d", "--rm", "--name", pg, "--network", network,
		"-e", "POSTGRES_PASSWORD=superuser-password", pgTestImage)
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", pg).Run() })
	waitFor(t, 60*time.Second, "postgres ready", func() bool {
		return exec.Command("docker", "exec", pg, "pg_isready", "-U", "postgres").Run() == nil
	})
	mustDocker(t, "exec", pg, "psql", "-U", "postgres", "-v", "ON_ERROR_STOP=1", "-c",
		fmt.Sprintf("CREATE ROLE %s LOGIN PASSWORD '%s'; CREATE DATABASE %s OWNER %s;", pgTestRole, pgTestPassword, pgTestDatabase, pgTestRole))

	connURL := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", pgTestRole, pgTestPassword, pg, pgTestDatabase)
	ctx := context.Background()

	// First server: born uninitialized against an empty database.
	first := startOpenBAO(t, network, "openbao-proof-1", connURL)
	health, err := CheckOpenBAOHealth(ctx, http.DefaultClient, first)
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	if health.Initialized {
		t.Fatal("a fresh database must yield an uninitialized vault")
	}
	initResult, err := InitializeOpenBAO(ctx, http.DefaultClient, first, 5, 3)
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := UnsealOpenBAO(ctx, http.DefaultClient, first, initResult.UnsealKeys); err != nil {
		t.Fatalf("unseal: %v", err)
	}
	if err := EnsureOpenBAOMounts(ctx, http.DefaultClient, first, initResult.RootToken); err != nil {
		t.Fatalf("mounts: %v", err)
	}
	health, _ = CheckOpenBAOHealth(ctx, http.DefaultClient, first)
	if !health.Initialized || health.Sealed {
		t.Fatalf("after init and unseal: %+v", health)
	}

	// The process dies; the database keeps the vault. A second server against
	// the same database is the restored vault: initialized, sealed, waiting
	// for the keys the first one produced.
	mustDocker(t, "rm", "-f", "openbao-proof-1")
	second := startOpenBAO(t, network, "openbao-proof-2", connURL)
	health, err = CheckOpenBAOHealth(ctx, http.DefaultClient, second)
	if err != nil {
		t.Fatalf("health (second): %v", err)
	}
	if !health.Initialized || !health.Sealed {
		t.Fatalf("a server on the same database must come up initialized and sealed, got %+v", health)
	}
	if err := UnsealOpenBAO(ctx, http.DefaultClient, second, initResult.UnsealKeys); err != nil {
		t.Fatalf("unseal (second) with the first's keys: %v", err)
	}
	mounts, err := listMounts(ctx, http.DefaultClient, second, initResult.RootToken)
	if err != nil {
		t.Fatalf("list mounts (second): %v", err)
	}
	for _, path := range []string{"secret/", "transit/"} {
		if _, ok := mounts[path]; !ok {
			t.Errorf("the restored vault lacks the %s mount the first one created; mounts=%v", path, mountKeys(mounts))
		}
	}
	// The table the backend created for itself is the archive's payload.
	out := mustDocker(t, "exec", pg, "psql", "-U", pgTestRole, "-d", pgTestDatabase, "-tAc",
		"SELECT count(*) FROM openbao_kv_store")
	if strings.TrimSpace(out) == "0" || strings.TrimSpace(out) == "" {
		t.Errorf("openbao_kv_store carries no rows after init; got %q", out)
	}
}

// startOpenBAO runs one OpenBao server container against the connection URL
// and returns its API address on the host once the health endpoint answers.
func startOpenBAO(t *testing.T, network, name, connURL string) string {
	t.Helper()
	_ = exec.Command("docker", "rm", "-f", name).Run()
	mustDocker(t, "run", "-d", "--rm", "--name", name, "--network", network,
		"-p", "0:8200",
		"-e", "BAO_LOCAL_CONFIG="+openbaoTestConfig,
		"-e", "BAO_PG_CONNECTION_URL="+connURL,
		openbaoTestImage, "server")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", name).Run() })
	port := strings.TrimSpace(mustDocker(t, "port", name, "8200/tcp"))
	// docker port prints "0.0.0.0:PORT" (and may print an IPv6 line too).
	port = strings.Split(port, "\n")[0]
	port = port[strings.LastIndex(port, ":")+1:]
	apiAddr := "http://127.0.0.1:" + port
	waitFor(t, 60*time.Second, name+" health", func() bool {
		_, err := CheckOpenBAOHealth(context.Background(), http.DefaultClient, apiAddr)
		return err == nil
	})
	return apiAddr
}

func mustDocker(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("docker %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func waitFor(t *testing.T, timeout time.Duration, what string, ready func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ready() {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func mountKeys(mounts map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(mounts))
	for k := range mounts {
		keys = append(keys, k)
	}
	return keys
}
