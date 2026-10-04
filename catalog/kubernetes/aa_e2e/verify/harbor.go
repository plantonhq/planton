package verify

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/pkg/errors"
)

// HarborVerifier checks a Harbor registry to the point a customer could
// push and pull OCI artifacts against it: every stateless component
// rolled out (core, portal, registry, jobservice, nginx), the front-door
// Service present, and THE REGISTRY PROOF on every lane — login as the
// module-generated admin, create a PRIVATE project, OCI push/pull
// round-trip, asserting BOTH the granted path and THE AUTH GATE at two
// surfaces: an unauthenticated API call to an authenticated-only route
// answers 401, and an ANONYMOUS pull of the pushed private artifact is
// rejected by the registry itself.
//
// Gate-endpoint truth (verified live and in the server source at the
// pin): Harbor's visibility model deliberately serves ANONYMOUS reads
// of public-project metadata — GET /api/v2.0/projects has no
// authentication requirement and answers 200 with the public subset,
// so it can never be an auth gate. GET /api/v2.0/users/current calls
// RequireAuthenticated and is the honest 401 probe.
//
// The behavioral-durability scenario (recognized by name) additionally
// DELETES the registry pod after the push, waits for a UID-verified
// replacement, re-establishes the port-forward, and pulls the artifact
// back byte-identical — registry blobs surviving pod loss through the
// filesystem PVC. The tunnel targets the nginx front-door Service, so
// the registry deletion does not kill it — the re-tunnel is deliberate
// defense against the dead-tunnel class (a mid-proof nginx restart
// would silently break the old tunnel, and a fresh one costs nothing).
//
// externalUrl in the scenario MUST match the verifier's port-forward
// address (http://127.0.0.1:18080): Harbor embeds it in the token-service
// URL returned to every OCI client, so push/pull FAIL AUTH when the
// dialed address and externalUrl disagree.
type HarborVerifier struct {
	Namespace  string
	Name       string
	Durability bool
}

// harborLocalPort is the workstation side of the port-forward. Kept in
// lockstep with every scenario's externalUrl.
const harborLocalPort = "18080"

func (v *HarborVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] harbor %q in namespace %q\n", v.Name, v.Namespace)

	// Stateless components self-ready; core's first-boot schema migration
	// is the long pole (startup probe budgets 60 minutes upstream — the
	// Helm wait already absorbed the normal case).
	for _, component := range []string{"core", "portal", "registry", "jobservice", "nginx"} {
		if err := kubectlRolloutStatus(ctx, kubeconfig, "deployment/"+v.Name+"-"+component, v.Namespace, 15*time.Minute); err != nil {
			return errors.Wrapf(err, "the %s deployment never rolled out", component)
		}
	}
	if err := KubectlResourceExists(ctx, kubeconfig, "service", v.Name, v.Namespace); err != nil {
		return errors.Wrap(err, "the front-door Service not found")
	}

	password, err := v.adminPassword(ctx, kubeconfig)
	if err != nil {
		return err
	}

	return v.proveRegistryRoundTrip(ctx, kubeconfig, password)
}

func (v *HarborVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.Name+"-core", v.Namespace)
}

// adminPassword reads the module-generated admin password from the
// exported `<name>-admin-auth` Secret (key HARBOR_ADMIN_PASSWORD).
func (v *HarborVerifier) adminPassword(ctx context.Context, kubeconfig string) (string, error) {
	secretName := v.Name + "-admin-auth"
	b64, err := kubectlGetJSONPath(ctx, kubeconfig, "secret", secretName, v.Namespace, "{.data.HARBOR_ADMIN_PASSWORD}")
	if err != nil {
		return "", errors.Wrapf(err, "reading secret %q HARBOR_ADMIN_PASSWORD", secretName)
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	password := strings.TrimSpace(string(raw))
	if password == "" {
		return "", errors.New("admin password Secret was empty")
	}
	// The chart's publicly documented default must never ship.
	if password == "Harbor12345" {
		return "", errors.New("admin password is the chart's public default Harbor12345 — the module-generated Secret was not wired")
	}
	return password, nil
}

// proveRegistryRoundTrip drives the Harbor API + OCI distribution over a
// port-forward to the front-door Service: assert the auth gate, create a
// run-unique project, push a random image, pull it back. The durability
// variant deletes the registry pod between push and pull and opens a
// fresh tunnel after replacement.
func (v *HarborVerifier) proveRegistryRoundTrip(ctx context.Context, kubeconfig, password string) error {
	cancel, err := v.startPortForward(ctx, kubeconfig)
	if err != nil {
		return err
	}

	base := "http://127.0.0.1:" + harborLocalPort
	client := &http.Client{Timeout: 30 * time.Second}

	// THE API AUTH GATE: an authenticated-only route without credentials
	// must answer 401 — proving no anonymous identity exists. See the
	// type comment for why the projects listing can never be this gate.
	if err := v.waitForAuthGate(ctx, client, base, 4*time.Minute); err != nil {
		cancel()
		return err
	}
	fmt.Printf("  [verify] AUTH GATE: unauthenticated /users/current rejected with 401\n")

	// PRIVATE by design: the registry-surface gate below proves an
	// anonymous pull of this project's artifact is rejected — a public
	// project would be anonymously pullable BY DESIGN and prove nothing.
	project := fmt.Sprintf("e2e%d", time.Now().Unix()%1000000)
	if err := v.createProject(ctx, client, base, password, project, 4*time.Minute); err != nil {
		cancel()
		return err
	}
	fmt.Printf("  [verify] PROJECT: private %q created as admin\n", project)

	repo := "proof"
	tag := fmt.Sprintf("t%d", time.Now().Unix())
	refStr := fmt.Sprintf("127.0.0.1:%s/%s/%s:%s", harborLocalPort, project, repo, tag)
	digest, err := v.pushImage(ctx, refStr, password)
	if err != nil {
		cancel()
		return err
	}
	fmt.Printf("  [verify] OCI: pushed %s (digest %s)\n", refStr, digest)

	// THE REGISTRY AUTH GATE: the OCI surface itself must reject an
	// anonymous pull of the private artifact — the guard a customer's
	// registry actually depends on, distinct from the portal API gate.
	if err := v.assertAnonymousPullRejected(ctx, refStr); err != nil {
		cancel()
		return err
	}
	fmt.Printf("  [verify] AUTH GATE: anonymous pull of the private artifact rejected by the registry\n")

	if v.Durability {
		// Close the tunnel across the replacement window and reopen fresh
		// after — see the type comment on the dead-tunnel defense.
		cancel()
		if err := deletePodAwaitReplacement(ctx, kubeconfig, v.Namespace,
			"app.kubernetes.io/instance="+v.Name+",app.kubernetes.io/component=registry", 10*time.Minute); err != nil {
			return errors.Wrap(err, "the registry pod did not recover after deletion")
		}
		// Fresh tunnel — the previous one died with the pod.
		cancel, err = v.startPortForward(ctx, kubeconfig)
		if err != nil {
			return errors.Wrap(err, "re-establishing the port-forward after registry replacement")
		}
	}
	defer cancel()

	gotDigest, err := v.pullImage(ctx, refStr, password)
	if err != nil {
		return err
	}
	if gotDigest != digest {
		return errors.Errorf("pulled digest %s does not match pushed digest %s", gotDigest, digest)
	}
	if v.Durability {
		fmt.Printf("  [verify] STATE: artifact survived registry pod replacement AND pulled byte-identical — blobs live on the PVC\n")
	} else {
		fmt.Printf("  [verify] OCI: pulled %s (digest match)\n", refStr)
	}
	return nil
}

// startPortForward opens the tunnel to the front-door Service and only
// returns once the LOCAL port accepts a TCP connection — kubectl binds
// the listener asynchronously, so "process started" is not "tunnel up"
// (caught live: a pull dialed the fresh post-replacement tunnel before
// it listened and read connection-refused as a durability failure; the
// first tunnel only ever worked because the auth-gate retry loop
// happened to absorb the same race).
func (v *HarborVerifier) startPortForward(ctx context.Context, kubeconfig string) (context.CancelFunc, error) {
	pfCtx, cancel := context.WithCancel(ctx)
	pf := exec.CommandContext(pfCtx, "kubectl", "--kubeconfig", kubeconfig,
		"port-forward", "svc/"+v.Name, harborLocalPort+":80", "-n", v.Namespace)
	var pfOut strings.Builder
	pf.Stdout = &pfOut
	pf.Stderr = &pfOut
	if err := pf.Start(); err != nil {
		cancel()
		return nil, errors.Wrap(err, "starting port-forward to the front-door Service")
	}
	// Cancel FIRST — Wait blocks forever on a port-forward that is never
	// told to exit.
	done := make(chan struct{})
	go func() {
		_ = pf.Wait()
		close(done)
	}()
	cancelAndReap := func() {
		cancel()
		<-done
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:"+harborLocalPort, 2*time.Second)
		if err == nil {
			_ = conn.Close()
			return cancelAndReap, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	cancelAndReap()
	return nil, errors.Errorf("the port-forward never started listening on 127.0.0.1:%s within 30s; kubectl output: %s", harborLocalPort, pfOut.String())
}

// waitForAuthGate retries until Harbor answers 401 to an unauthenticated
// current-user read (core may still be finishing first-boot migrations
// when the tunnel first answers). Route truth at the pin: the handler
// calls RequireAuthenticated before anything else, so anonymous == 401
// unconditionally — unlike the projects listing, which serves the
// public subset to anonymous callers by design.
func (v *HarborVerifier) waitForAuthGate(ctx context.Context, client *http.Client, base string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var last error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/v2.0/users/current", nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			last = err
			time.Sleep(5 * time.Second)
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized {
			return nil
		}
		last = errors.Errorf("expected 401, got %d", resp.StatusCode)
		time.Sleep(5 * time.Second)
	}
	return errors.Wrap(last, "the unauthenticated request was NOT rejected (the auth gate)")
}

func (v *HarborVerifier) createProject(ctx context.Context, client *http.Client, base, password, project string, budget time.Duration) error {
	// public:false is load-bearing: the registry-surface auth gate pulls
	// this project anonymously and MUST be rejected.
	body := fmt.Sprintf(`{"project_name":%q,"public":false}`, project)
	deadline := time.Now().Add(budget)
	var last error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/v2.0/projects", strings.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("admin", password)
		resp, err := client.Do(req)
		if err != nil {
			last = err
			time.Sleep(5 * time.Second)
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		// 201 created; 409 already exists (retry-safe).
		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
			return nil
		}
		last = errors.Errorf("create project: status %d: %s", resp.StatusCode, string(respBody))
		time.Sleep(5 * time.Second)
	}
	return errors.Wrap(last, "creating the proof project")
}

func (v *HarborVerifier) pushImage(ctx context.Context, refStr, password string) (string, error) {
	ref, err := name.ParseReference(refStr, name.Insecure)
	if err != nil {
		return "", errors.Wrap(err, "parsing push reference")
	}
	img, err := random.Image(1024, 1)
	if err != nil {
		return "", errors.Wrap(err, "building the proof image")
	}
	auth := &authn.Basic{Username: "admin", Password: password}
	if err := remote.Write(ref, img, remote.WithAuth(auth), remote.WithContext(ctx)); err != nil {
		return "", errors.Wrap(err, "the OCI push never succeeded")
	}
	digest, err := img.Digest()
	if err != nil {
		return "", err
	}
	return digest.String(), nil
}

// assertAnonymousPullRejected proves the OCI surface guards the private
// artifact: an anonymous pull must fail with an authorization rejection
// (Harbor's token service withholds the pull scope, and the registry
// answers 401 UNAUTHORIZED / 403 DENIED). A succeeding pull means the
// project visibility or the token-service auth is broken; any OTHER
// failure (network, tunnel) is reported as its own error, never
// mistaken for the gate holding.
func (v *HarborVerifier) assertAnonymousPullRejected(ctx context.Context, refStr string) error {
	ref, err := name.ParseReference(refStr, name.Insecure)
	if err != nil {
		return errors.Wrap(err, "parsing anonymous pull reference")
	}
	_, err = remote.Image(ref, remote.WithAuth(authn.Anonymous), remote.WithContext(ctx))
	if err == nil {
		return errors.New("an ANONYMOUS pull of the private artifact SUCCEEDED — the registry auth gate is not holding")
	}
	var terr *transport.Error
	if errors.As(err, &terr) && (terr.StatusCode == http.StatusUnauthorized || terr.StatusCode == http.StatusForbidden) {
		return nil
	}
	return errors.Wrap(err, "the anonymous pull failed for a reason OTHER than an auth rejection")
}

func (v *HarborVerifier) pullImage(ctx context.Context, refStr, password string) (string, error) {
	ref, err := name.ParseReference(refStr, name.Insecure)
	if err != nil {
		return "", errors.Wrap(err, "parsing pull reference")
	}
	auth := &authn.Basic{Username: "admin", Password: password}
	img, err := remote.Image(ref, remote.WithAuth(auth), remote.WithContext(ctx))
	if err != nil {
		return "", errors.Wrap(err, "the OCI pull never succeeded")
	}
	digest, err := img.Digest()
	if err != nil {
		return "", err
	}
	// Touch a config read so the pull is not a no-op against a stub.
	if _, err := img.ConfigName(); err != nil {
		return "", errors.Wrap(err, "reading the pulled image config")
	}
	return digest.String(), nil
}
