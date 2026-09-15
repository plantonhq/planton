//go:build requires_docker

package bootstrap

// The seal decision, proven against real servers with no cloud account: a
// dev-mode OpenBao stands in for the adopter's key service with a transit
// key; a second OpenBao on PostgreSQL storage declares a transit seal against
// it and takes its token from the environment, exactly as the operator
// projects it. The proof: the server refuses a built-in-seal init and takes
// a recovery quorum instead; after init it opens ITSELF with no unseal call;
// killed and started again on the same database and the same seal, it comes
// up initialized and open on its own -- what a platform restore does to a
// cloud-sealed vault. And a server whose seal token is wrong exits at start
// with the one log line the operator's status points a person at.
//
// Run with `go test -tags=requires_docker ./internal/bootstrap/ -run
// TestOpenBAO_TransitSeal`.

import (
	"bytes"
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
	keyHolderRootToken = "holder-root"
	transitKeyName     = "planton-unseal"
)

// The operator's config with a transit seal stanza appended (openbao_seal.go
// renders the HCL; this is its JSON form for the image's entrypoint).
func transitSealConfig(holderAddr string) string {
	return fmt.Sprintf(`{
  "ui": false,
  "disable_mlock": true,
  "listener": {"tcp": {"address": "[::]:8200", "tls_disable": 1}},
  "storage": {"postgresql": {"max_parallel": "32"}},
  "seal": {"transit": {"address": %q, "key_name": %q, "mount_path": "transit/"}}
}`, holderAddr, transitKeyName)
}

func TestOpenBAO_TransitSealOpensARestoredVault(t *testing.T) {
	requireDocker(t)
	network := fmt.Sprintf("openbao-seal-proof-%d", os.Getpid())
	mustDocker(t, "network", "create", network)
	t.Cleanup(func() { _ = exec.Command("docker", "network", "rm", network).Run() })

	pg := startPostgreSQL(t, network, "openbao-seal-pg")
	holder := startKeyHolder(t, network, "openbao-seal-holder")
	ctx := context.Background()

	env := append(storageEnv(pg), "VAULT_TOKEN="+keyHolderRootToken)
	config := transitSealConfig("http://" + holder + ":8200")

	// First server: uninitialized, and it refuses the built-in seal's init
	// shape because a seal is configured.
	first := startOpenBAO(t, network, "openbao-seal-1", config, env...)
	health, err := CheckOpenBAOHealth(ctx, http.DefaultClient, first)
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	if health.Initialized {
		t.Fatal("a fresh database must yield an uninitialized vault")
	}
	if _, err := InitializeOpenBAO(ctx, http.DefaultClient, first, ShamirInit(5, 3)); err == nil || !strings.Contains(err.Error(), "not applicable to seal type") {
		t.Fatalf("a sealed server must refuse secret_shares, got %v", err)
	}
	initResult, err := InitializeOpenBAO(ctx, http.DefaultClient, first, RecoveryInit(5, 3))
	if err != nil {
		t.Fatalf("init with recovery shares: %v", err)
	}
	if len(initResult.RecoveryKeys) != 5 || len(initResult.UnsealKeys) != 0 || initResult.RootToken == "" {
		t.Fatalf("a cloud-seal init returns recovery keys, no unseal keys, and a root token; got %d/%d %q", len(initResult.RecoveryKeys), len(initResult.UnsealKeys), initResult.RootToken)
	}

	// The server opens itself: no unseal call, just its own loop.
	open, err := WaitUntilUnsealed(ctx, http.DefaultClient, first, 20*time.Second)
	if err != nil || !open {
		t.Fatalf("the server must open itself after init under a transit seal, got open=%v err=%v", open, err)
	}
	if err := EnsureOpenBAOMounts(ctx, http.DefaultClient, first, initResult.RootToken); err != nil {
		t.Fatalf("mounts: %v", err)
	}

	// The process dies; the database keeps the vault; the key holder keeps
	// the key. A new server on the same storage and the same seal is the
	// restored vault -- and it opens on its own.
	mustDocker(t, "rm", "-f", "openbao-seal-1")
	second := startOpenBAO(t, network, "openbao-seal-2", config, env...)
	open, err = WaitUntilUnsealed(ctx, http.DefaultClient, second, 20*time.Second)
	if err != nil || !open {
		t.Fatalf("a restored cloud-sealed vault must open itself, got open=%v err=%v", open, err)
	}
	mounts, err := listMounts(ctx, http.DefaultClient, second, initResult.RootToken)
	if err != nil {
		t.Fatalf("list mounts (restored): %v", err)
	}
	for _, path := range []string{"secret/", "transit/"} {
		if _, ok := mounts[path]; !ok {
			t.Errorf("the restored vault lacks the %s mount, mounts=%v", path, mountKeys(mounts))
		}
	}
}

// A server whose seal cannot be configured never serves: the transit wrapper
// test-encrypts on the key when the seal is configured, so a wrong token is
// a process exit with "Error configuring seal" -- the line the operator's
// crash-loop sentence points at.
func TestOpenBAO_TransitSealWithWrongTokenExitsAtStart(t *testing.T) {
	requireDocker(t)
	network := fmt.Sprintf("openbao-seal-bad-%d", os.Getpid())
	mustDocker(t, "network", "create", network)
	t.Cleanup(func() { _ = exec.Command("docker", "network", "rm", network).Run() })

	pg := startPostgreSQL(t, network, "openbao-seal-bad-pg")
	holder := startKeyHolder(t, network, "openbao-seal-bad-holder")

	name := "openbao-seal-bad"
	_ = exec.Command("docker", "rm", "-f", name).Run()
	args := []string{"run", "-d", "--name", name, "--network", network, "-e", "BAO_LOCAL_CONFIG=" + transitSealConfig("http://"+holder+":8200")}
	for _, e := range append(storageEnv(pg), "VAULT_TOKEN=not-the-token") {
		args = append(args, "-e", e)
	}
	args = append(args, openbaoTestImage, "server")
	mustDocker(t, args...)
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", name).Run() })

	waitFor(t, 60*time.Second, "the misconfigured server to exit", func() bool {
		out, err := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", name).Output()
		return err == nil && strings.TrimSpace(string(out)) == "false"
	})
	logs, _ := exec.Command("docker", "logs", name).CombinedOutput()
	if !strings.Contains(string(logs), "Error configuring seal") {
		t.Fatalf("the server must exit at seal configuration with the line the status names; log:\n%s", logs)
	}
}

// startKeyHolder runs a dev-mode OpenBao as the stand-in for a central key
// service, with the transit engine mounted and the unseal key created, and
// returns its container name (its hostname on the network).
func startKeyHolder(t *testing.T, network, name string) string {
	t.Helper()
	_ = exec.Command("docker", "rm", "-f", name).Run()
	mustDocker(t, "run", "-d", "--rm", "--name", name, "--network", network, "-p", "0:8200",
		"-e", "BAO_DEV_ROOT_TOKEN_ID="+keyHolderRootToken,
		"-e", "BAO_DEV_LISTEN_ADDRESS=0.0.0.0:8200",
		openbaoTestImage, "server", "-dev")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", name).Run() })
	port := strings.TrimSpace(mustDocker(t, "port", name, "8200/tcp"))
	port = strings.Split(port, "\n")[0]
	port = port[strings.LastIndex(port, ":")+1:]
	apiAddr := "http://127.0.0.1:" + port
	waitFor(t, 60*time.Second, name+" health", func() bool {
		h, err := CheckOpenBAOHealth(context.Background(), http.DefaultClient, apiAddr)
		return err == nil && h.Initialized && !h.Sealed
	})
	ctx := context.Background()
	if err := enableMount(ctx, http.DefaultClient, apiAddr, keyHolderRootToken, "transit", map[string]any{"type": "transit"}); err != nil {
		t.Fatalf("mounting transit on the key holder: %v", err)
	}
	body, _ := json.Marshal(map[string]any{"type": "aes256-gcm96"})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, apiAddr+"/v1/transit/keys/"+transitKeyName, bytes.NewReader(body))
	req.Header.Set("X-Vault-Token", keyHolderRootToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("creating the transit key: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("creating the transit key returned %d", resp.StatusCode)
	}
	return name
}
