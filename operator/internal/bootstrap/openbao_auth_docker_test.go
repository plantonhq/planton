//go:build requires_docker

package bootstrap

// The login, proven against a real server with no cluster: a real OpenBao
// on PostgreSQL storage (the operator's five-variable connection) with the
// exact access arrangement the operator writes -- the kubernetes auth
// method, both policies, the operator's role, the control plane's token
// role -- and a stand-in Kubernetes API in the test process that answers
// TokenReview for one ServiceAccount token and refuses another. The proof:
// the bound identity signs in and the unbound one is refused with the
// server's own words; every read-back matches what was written (the
// read-compare-write the steady state depends on); the operator's session
// mints the control plane's token through the role without sudo, and that
// token carries exactly the control plane's policy -- it can use both
// engines and is refused at sys/ and auth/, while the operator's session
// is refused at the engines; a minted token outlives the session that
// minted it (orphan), is renewed and looked up by accessor, and is gone
// after a revoke by accessor. The real cluster's TokenReview is the kind
// lane's to prove; this proves every mechanic the operator runs.
//
// Run with `go test -tags=requires_docker ./internal/bootstrap/ -run
// TestOpenBAO_KubernetesAuth`.

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/plantonhq/planton/operator/internal/resources"
)

const (
	proofPlatform          = "planton"
	proofOperatorNamespace = "planton-operator-system"
	proofOperatorAccount   = "planton-operator"
	proofReviewerJWT       = "reviewer-token"
)

// fakeServiceAccountJWT mints an unsigned token with the claims the auth
// method reads (a supported alg in the header; the signature is not checked
// when no public keys are configured, TokenReview is the verdict).
func fakeServiceAccountJWT(namespace, name, uid string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(
		`{"sub":"system:serviceaccount:%s:%s","kubernetes.io/serviceaccount/namespace":%q,"kubernetes.io/serviceaccount/service-account.name":%q,"kubernetes.io/serviceaccount/service-account.uid":%q}`,
		namespace, name, namespace, name, uid)))
	return header + "." + payload + ".c2lnbmF0dXJl"
}

// throwawayCAPEM is a self-signed certificate the auth method accepts as a
// CA when its local files are disabled; the stand-in API speaks http, so it
// is never consulted.
func throwawayCAPEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "stand-in api"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

// startTokenReviewStandIn serves the one Kubernetes API the auth method
// calls, on every interface so the container reaches it through
// host.docker.internal. It authenticates exactly one token.
func startTokenReviewStandIn(t *testing.T, acceptedJWT, namespace, name, uid string) (addr string, reviews *atomic.Int32) {
	t.Helper()
	reviews = &atomic.Int32{}
	mux := http.NewServeMux()
	mux.HandleFunc("/apis/authentication.k8s.io/v1/tokenreviews", func(w http.ResponseWriter, r *http.Request) {
		reviews.Add(1)
		if r.Header.Get("Authorization") != "Bearer "+proofReviewerJWT {
			http.Error(w, "the vault must review with the configured reviewer token", http.StatusUnauthorized)
			return
		}
		var review struct {
			Spec struct {
				Token string `json:"token"`
			} `json:"spec"`
		}
		_ = json.NewDecoder(r.Body).Decode(&review)
		status := map[string]any{"authenticated": false, "error": "token not recognized by the stand-in API"}
		if review.Spec.Token == acceptedJWT {
			status = map[string]any{"authenticated": true, "user": map[string]any{
				"username": fmt.Sprintf("system:serviceaccount:%s:%s", namespace, name), "uid": uid,
			}}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"apiVersion": "authentication.k8s.io/v1", "kind": "TokenReview", "status": status})
	})
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatalf("listening for the stand-in API: %v", err)
	}
	srv := &httptest.Server{Listener: listener, Config: &http.Server{Handler: mux}}
	srv.Start()
	t.Cleanup(srv.Close)
	port := listener.Addr().(*net.TCPAddr).Port
	return fmt.Sprintf("http://host.docker.internal:%d", port), reviews
}

func TestOpenBAO_KubernetesAuthLoginAndControlPlaneToken(t *testing.T) {
	requireDocker(t)
	network := fmt.Sprintf("openbao-auth-proof-%d", os.Getpid())
	mustDocker(t, "network", "create", network)
	t.Cleanup(func() { _ = exec.Command("docker", "network", "rm", network).Run() })

	pg := startPostgreSQL(t, network, "openbao-auth-pg")
	ctx := context.Background()
	operatorJWT := fakeServiceAccountJWT(proofOperatorNamespace, proofOperatorAccount, "uid-operator")
	strangerJWT := fakeServiceAccountJWT("somewhere-else", "someone-else", "uid-stranger")
	apiServer, reviews := startTokenReviewStandIn(t, operatorJWT, proofOperatorNamespace, proofOperatorAccount, "uid-operator")

	addr := startOpenBAO(t, network, "openbao-auth-1", openbaoTestConfig, storageEnv(pg)...)
	initResult, err := InitializeOpenBAO(ctx, http.DefaultClient, addr, ShamirInit(1, 1))
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := UnsealOpenBAO(ctx, http.DefaultClient, addr, initResult.UnsealKeys); err != nil {
		t.Fatalf("unseal: %v", err)
	}
	root := initResult.RootToken
	if err := EnsureOpenBAOMounts(ctx, http.DefaultClient, addr, root); err != nil {
		t.Fatalf("mounts: %v", err)
	}

	// ── the arrangement, written with root exactly as the operator writes it ──
	operatorRole := resources.OpenBAOOperatorRoleName(proofPlatform)
	controlPlaneRole := resources.OpenBAOControlPlaneRoleName(proofPlatform)
	authPath := resources.OpenBAOKubernetesAuthPath

	if err := EnableKubernetesAuth(ctx, http.DefaultClient, addr, root, authPath); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if err := EnableKubernetesAuth(ctx, http.DefaultClient, addr, root, authPath); err != nil {
		t.Fatalf("enabling twice must be idempotent: %v", err)
	}
	// Outside a pod the method has no local CA file to read -- and it reads
	// it the moment the config is written -- so the proof disables the local
	// files and hands it a CA of its own (unused over http). In a pod the
	// operator sets neither: the files are there, and not pinning them is
	// what keeps a restored vault working on another cluster.
	wantConfig := KubernetesAuthConfig{Host: apiServer, TokenReviewerJWT: proofReviewerJWT, DisableLocalCAJWT: true, CACert: throwawayCAPEM(t)}
	if err := WriteKubernetesAuthConfig(ctx, http.DefaultClient, addr, root, authPath, wantConfig); err != nil {
		t.Fatalf("config: %v", err)
	}
	for name, policy := range map[string]string{
		operatorRole:     resources.OpenBAOOperatorPolicy(proofPlatform),
		controlPlaneRole: resources.OpenBAOControlPlanePolicy(),
	} {
		if err := WritePolicy(ctx, http.DefaultClient, addr, root, name, policy); err != nil {
			t.Fatalf("policy %s: %v", name, err)
		}
	}
	wantRole := KubernetesAuthRole{
		ServiceAccountNames:      []string{proofOperatorAccount},
		ServiceAccountNamespaces: []string{proofOperatorNamespace},
		Policies:                 []string{operatorRole},
		TTL:                      resources.OpenBAOOperatorSessionTTL,
		MaxTTL:                   resources.OpenBAOOperatorSessionTTL,
	}
	if err := WriteKubernetesAuthRole(ctx, http.DefaultClient, addr, root, authPath, operatorRole, wantRole); err != nil {
		t.Fatalf("role: %v", err)
	}
	wantTokenRole := TokenRole{AllowedPolicies: []string{controlPlaneRole}, Orphan: true, Renewable: true, Period: resources.OpenBAOControlPlaneTokenPeriod, TokenType: "service"}
	if err := WriteTokenRole(ctx, http.DefaultClient, addr, root, controlPlaneRole, wantTokenRole); err != nil {
		t.Fatalf("token role: %v", err)
	}

	// ── read-backs match the writes (what read-compare-write relies on) ──
	gotConfig, found, err := ReadKubernetesAuthConfig(ctx, http.DefaultClient, addr, root, authPath)
	if err != nil || !found || gotConfig.Host != apiServer {
		t.Fatalf("config read-back: %+v %v %v", gotConfig, found, err)
	}
	gotRole, found, err := ReadKubernetesAuthRole(ctx, http.DefaultClient, addr, root, authPath, operatorRole)
	if err != nil || !found {
		t.Fatalf("role read-back: %v %v", found, err)
	}
	if gotRole.TTL != wantRole.TTL || gotRole.MaxTTL != wantRole.MaxTTL || strings.Join(gotRole.Policies, ",") != operatorRole ||
		strings.Join(gotRole.ServiceAccountNames, ",") != proofOperatorAccount || strings.Join(gotRole.ServiceAccountNamespaces, ",") != proofOperatorNamespace {
		t.Errorf("role read-back differs from the write: got %+v, want %+v", gotRole, wantRole)
	}
	gotTokenRole, found, err := ReadTokenRole(ctx, http.DefaultClient, addr, root, controlPlaneRole)
	if err != nil || !found {
		t.Fatalf("token role read-back: %v %v", found, err)
	}
	if gotTokenRole.Period != wantTokenRole.Period || !gotTokenRole.Orphan || !gotTokenRole.Renewable || gotTokenRole.TokenType != "service" || strings.Join(gotTokenRole.AllowedPolicies, ",") != controlPlaneRole {
		t.Errorf("token role read-back differs from the write: got %+v, want %+v", gotTokenRole, wantTokenRole)
	}
	for name, want := range map[string]string{operatorRole: resources.OpenBAOOperatorPolicy(proofPlatform), controlPlaneRole: resources.OpenBAOControlPlanePolicy()} {
		got, found, err := ReadPolicy(ctx, http.DefaultClient, addr, root, name)
		if err != nil || !found || got != want {
			t.Errorf("policy %s read-back differs from the write (found %v, err %v):\n%s", name, found, err, got)
		}
	}
	if _, found, err := ReadPolicy(ctx, http.DefaultClient, addr, root, "never-written"); err != nil || found {
		t.Errorf("a policy that does not exist reads as not found, got found=%v err=%v", found, err)
	}
	if _, found, err := ReadKubernetesAuthRole(ctx, http.DefaultClient, addr, root, authPath, "never-written"); err != nil || found {
		t.Errorf("a role that does not exist reads as not found, got found=%v err=%v", found, err)
	}

	// ── the login: the bound identity in, the stranger refused ──
	session, err := KubernetesLogin(ctx, http.DefaultClient, addr, authPath, operatorRole, operatorJWT)
	if err != nil {
		t.Fatalf("the operator's login must succeed: %v", err)
	}
	if reviews.Load() == 0 {
		t.Fatal("the vault must have reviewed the token with the API server")
	}
	if _, err := KubernetesLogin(ctx, http.DefaultClient, addr, authPath, operatorRole, strangerJWT); err == nil {
		t.Fatal("an identity the role does not bind must be refused")
	} else if apiErr, ok := AsAPIError(err); !ok || !apiErr.IsPermissionDenied() {
		t.Fatalf("the refusal must be the server's 403 with its reason, got %v", err)
	}

	// ── the operator's session: provisioning yes, the engines no ──
	if err := EnsureOpenBAOMounts(ctx, http.DefaultClient, addr, session.ClientToken); err != nil {
		t.Errorf("the operator's policy must allow the mounts check: %v", err)
	}
	if _, err := vaultCall(ctx, http.DefaultClient, addr, http.MethodGet, "secret/data/planton/probe", session.ClientToken, nil); err == nil {
		t.Error("the operator's session must not read the platform's secrets")
	} else if apiErr, ok := AsAPIError(err); !ok || !apiErr.IsPermissionDenied() {
		t.Errorf("expected a 403 at the engine for the operator, got %v", err)
	}

	// ── the control plane's token, minted on the session without sudo ──
	minted, err := CreateTokenWithRole(ctx, http.DefaultClient, addr, session.ClientToken, controlPlaneRole, CreateTokenOptions{
		Policies: []string{controlPlaneRole}, DisplayName: controlPlaneRole, Meta: map[string]string{"platform": proofPlatform},
	})
	if err != nil {
		t.Fatalf("minting through the role on the operator's session: %v", err)
	}
	info, live, err := LookupAccessor(ctx, http.DefaultClient, addr, session.ClientToken, minted.Accessor)
	if err != nil || !live {
		t.Fatalf("lookup by accessor: live=%v err=%v", live, err)
	}
	if !info.Renewable || info.TTL <= 6*24*time.Hour || info.TTL > resources.OpenBAOControlPlaneTokenPeriod {
		t.Errorf("the minted token is renewable with the role's period as its life, got %+v", info)
	}
	if strings.Join(info.Policies, ",") != controlPlaneRole+",default" && strings.Join(info.Policies, ",") != "default,"+controlPlaneRole {
		t.Errorf("the minted token carries exactly the control plane's policy (plus default), got %v", info.Policies)
	}
	if err := RenewAccessor(ctx, http.DefaultClient, addr, session.ClientToken, minted.Accessor); err != nil {
		t.Errorf("renew by accessor on the operator's session: %v", err)
	}

	// The control plane's token: both engines usable, sys/ and auth/ refused.
	cp := minted.ClientToken
	if _, err := vaultCall(ctx, http.DefaultClient, addr, http.MethodPost, "secret/data/planton/org/probe", cp, map[string]any{"data": map[string]string{"k": "v"}}); err != nil {
		t.Errorf("the control plane writes its secrets: %v", err)
	}
	if env, err := vaultCall(ctx, http.DefaultClient, addr, http.MethodGet, "secret/data/planton/org/probe", cp, nil); err != nil {
		t.Errorf("the control plane reads its secrets: %v", err)
	} else if data, _ := dataOf(env)["data"].(map[string]any); data["k"] != "v" {
		t.Errorf("read back %v", env)
	}
	if _, err := vaultCall(ctx, http.DefaultClient, addr, http.MethodPost, "transit/keys/planton-oidc-signing", cp, map[string]any{"type": "rsa-2048"}); err != nil {
		t.Errorf("the control plane creates its signing key: %v", err)
	}
	if _, err := vaultCall(ctx, http.DefaultClient, addr, http.MethodPost, "transit/sign/planton-oidc-signing", cp, map[string]any{"input": base64.StdEncoding.EncodeToString([]byte("hello")), "signature_algorithm": "pkcs1v15", "hash_algorithm": "sha2-256"}); err != nil {
		t.Errorf("the control plane signs with it: %v", err)
	}
	for _, path := range []string{"sys/policies/acl/" + controlPlaneRole, "auth/token/roles/" + controlPlaneRole, "sys/mounts", "auth/" + authPath + "/config"} {
		if _, err := vaultCall(ctx, http.DefaultClient, addr, http.MethodGet, path, cp, nil); err == nil {
			t.Errorf("the control plane's token must be refused at %s", path)
		} else if apiErr, ok := AsAPIError(err); !ok || !apiErr.IsPermissionDenied() {
			t.Errorf("expected a 403 at %s, got %v", path, err)
		}
	}

	// ── orphan: the minted token outlives the session that minted it ──
	if err := RevokeSelf(ctx, http.DefaultClient, addr, session.ClientToken); err != nil {
		t.Fatalf("revoke-self: %v", err)
	}
	if err := EnsureOpenBAOMounts(ctx, http.DefaultClient, addr, session.ClientToken); err == nil {
		t.Error("a revoked session must be refused")
	}
	if _, live, err := LookupAccessor(ctx, http.DefaultClient, addr, root, minted.Accessor); err != nil || !live {
		t.Fatalf("the control plane's token must outlive the operator's session (orphan), got live=%v err=%v", live, err)
	}
	if _, err := vaultCall(ctx, http.DefaultClient, addr, http.MethodGet, "secret/data/planton/org/probe", cp, nil); err != nil {
		t.Errorf("the control plane keeps working after the operator's session ends: %v", err)
	}

	// ── a second session, a revoke by accessor, and the lookup that says so ──
	session2, err := KubernetesLogin(ctx, http.DefaultClient, addr, authPath, operatorRole, operatorJWT)
	if err != nil {
		t.Fatalf("second login: %v", err)
	}
	if err := RevokeAccessor(ctx, http.DefaultClient, addr, session2.ClientToken, minted.Accessor); err != nil {
		t.Fatalf("revoke by accessor: %v", err)
	}
	if _, live, err := LookupAccessor(ctx, http.DefaultClient, addr, session2.ClientToken, minted.Accessor); err != nil || live {
		t.Errorf("a revoked accessor must read as not live (what makes the operator mint anew), got live=%v err=%v", live, err)
	}
	if err := RevokeAccessor(ctx, http.DefaultClient, addr, session2.ClientToken, minted.Accessor); err != nil {
		t.Errorf("revoking an accessor twice is not an error: %v", err)
	}
	if _, err := vaultCall(ctx, http.DefaultClient, addr, http.MethodGet, "secret/data/planton/org/probe", cp, nil); err == nil {
		t.Error("a revoked control plane token must be refused")
	}
	_ = RevokeSelf(ctx, http.DefaultClient, addr, session2.ClientToken)

	// ── restarted on the same storage: the arrangement came back with the data ──
	mustDocker(t, "rm", "-f", "openbao-auth-1")
	addr2 := startOpenBAO(t, network, "openbao-auth-2", openbaoTestConfig, storageEnv(pg)...)
	if err := UnsealOpenBAO(ctx, http.DefaultClient, addr2, initResult.UnsealKeys); err != nil {
		t.Fatalf("unseal the restarted server: %v", err)
	}
	session3, err := KubernetesLogin(ctx, http.DefaultClient, addr2, authPath, operatorRole, operatorJWT)
	if err != nil {
		t.Fatalf("the login must work on a server restarted from the same storage -- the arrangement is data: %v", err)
	}
	if got, found, err := ReadPolicy(ctx, http.DefaultClient, addr2, session3.ClientToken, controlPlaneRole); err != nil || !found || got != resources.OpenBAOControlPlanePolicy() {
		t.Errorf("the policies come back with the storage (found %v, err %v)", found, err)
	}
	_ = RevokeSelf(ctx, http.DefaultClient, addr2, session3.ClientToken)
}
