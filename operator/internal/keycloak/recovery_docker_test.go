//go:build requires_docker

package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/plantonhq/planton/operator/internal/resources"
)

// The restored-realm seam, end to end on the real product: a Keycloak whose
// master realm exists with admin/A (the source platform's password); the
// server STOPPED; Keycloak's own recovery command run against the same data
// creating the recovery admin (every node stopped, as the command requires);
// the server started again; ReestablishAdmin giving admin the password this
// install generated (B) and removing the recovery admin. Afterwards admin/B
// signs in, admin/A does not, and the recovery admin is gone.
//
// Its own container and its own data volume, not the suite's shared server:
// the recovery command needs the server down, and the suite's other tests
// need it up.
func TestRecovery_ReestablishAdminOnARestoredRealm(t *testing.T) {
	const (
		sourcePassword   = "source-platform-password"
		installPassword  = "this-install-password"
		recoveryPassword = "recovery-password"
	)
	image := resources.IdentityDefaultImageRepo + ":" + resources.IdentityDefaultImageTag
	stamp := fmt.Sprintf("%d", time.Now().UnixNano())
	volume := "planton-recovery-data-" + stamp
	container := "planton-recovery-kc-" + stamp

	run := func(args ...string) string {
		t.Helper()
		out, err := exec.Command("docker", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("docker %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return strings.TrimSpace(string(out))
	}
	defer func() {
		_ = exec.Command("docker", "rm", "-f", container).Run()
		_ = exec.Command("docker", "volume", "rm", "-f", volume).Run()
	}()

	// The source platform: a master realm bootstrapped with admin/A.
	run("volume", "create", volume)
	run("run", "-d", "--name", container,
		"-v", volume+":/opt/keycloak/data",
		"-e", "KC_BOOTSTRAP_ADMIN_USERNAME="+resources.IdentityBootstrapAdminUsername,
		"-e", "KC_BOOTSTRAP_ADMIN_PASSWORD="+sourcePassword,
		"-e", "KC_HTTP_RELATIVE_PATH="+resources.IdentityPathPrefix,
		"-p", "127.0.0.1:0:8080",
		image, "start-dev")
	serverRoot := func() string {
		portOut := run("port", container, "8080/tcp")
		return "http://" + strings.Split(portOut, "\n")[0] + resources.IdentityPathPrefix
	}
	if err := waitForKeycloak(serverRoot(), 3*time.Minute); err != nil {
		t.Fatalf("source keycloak never became ready: %v\n%s", err, run("logs", "--tail", "40", container))
	}

	// The restore: the realm comes back with admin/A while the new install
	// knows only B. Every node stopped, then the recovery command on the
	// same data creates the recovery admin.
	run("stop", container)
	recoveryOut := run("run", "--rm",
		"-v", volume+":/opt/keycloak/data",
		"-e", "KC_BOOTSTRAP_ADMIN_PASSWORD="+recoveryPassword,
		image, "bootstrap-admin", "user",
		"--username", resources.IdentityRecoveryAdminUsername,
		"--password:env", "KC_BOOTSTRAP_ADMIN_PASSWORD")
	if !strings.Contains(strings.ToLower(recoveryOut), "created") && !strings.Contains(strings.ToLower(recoveryOut), "temporary") {
		t.Logf("recovery command output (no creation word seen, continuing to the proof):\n%s", recoveryOut)
	}
	run("start", container)
	root := serverRoot()
	if err := waitForKeycloak(root, 3*time.Minute); err != nil {
		t.Fatalf("restored keycloak never became ready: %v\n%s", err, run("logs", "--tail", "40", container))
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	ctx := context.Background()

	// Before: this install's password is refused -- the seam itself.
	before := NewAdminClient(httpClient, root)
	if err := before.Authenticate(ctx, resources.IdentityBootstrapAdminUsername, installPassword); !IsCredentialRefused(err) {
		t.Fatalf("the restored realm must refuse this install's password before recovery, got %v", err)
	}

	if err := ReestablishAdmin(ctx, ReestablishAdminInput{
		HTTPClient:       httpClient,
		ServerRoot:       root,
		RecoveryUsername: resources.IdentityRecoveryAdminUsername,
		RecoveryPassword: recoveryPassword,
		AdminUsername:    resources.IdentityBootstrapAdminUsername,
		AdminPassword:    installPassword,
	}); err != nil {
		t.Fatalf("re-establishing the admin: %v", err)
	}

	// After: this install's password signs in, the source's does not, and
	// the recovery admin is gone.
	admin := NewAdminClient(httpClient, root)
	if err := admin.Authenticate(ctx, resources.IdentityBootstrapAdminUsername, installPassword); err != nil {
		t.Fatalf("admin must sign in with this install's password: %v", err)
	}
	if err := NewAdminClient(httpClient, root).Authenticate(ctx, resources.IdentityBootstrapAdminUsername, sourcePassword); !IsCredentialRefused(err) {
		t.Errorf("the source platform's password must be refused now, got %v", err)
	}
	if _, found, err := admin.FindUserByUsername(ctx, MasterRealm, resources.IdentityRecoveryAdminUsername); err != nil || found {
		t.Errorf("the recovery admin must be gone (found=%v, err=%v)", found, err)
	}
	if err := NewAdminClient(httpClient, root).Authenticate(ctx, resources.IdentityRecoveryAdminUsername, recoveryPassword); !IsCredentialRefused(err) {
		t.Errorf("the recovery credential must be dead, got %v", err)
	}
}
