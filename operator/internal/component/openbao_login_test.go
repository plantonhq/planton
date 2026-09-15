package component

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// The login and the control plane's token, every arm on the fake vault:
// what initialization writes with the root token and what it then proves on
// the operator's own session; what a steady pass does and does not write;
// when the control plane's token is renewed, when it is minted again, and
// what a person reads when the vault refuses the operator.

// ── the fake's access surface ─────────────────────────────────────────────

// serveAccess wires the auth method, policies, token roles, and token
// operations onto the fake's mux.
func (f *fakeVault) serveAccess(t *testing.T, mux *http.ServeMux) {
	t.Helper()

	// Enabling the method needs sudo: only root may.
	mux.HandleFunc("/v1/sys/auth/", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		token := r.Header.Get("X-Vault-Token")
		f.authEnableTokens = append(f.authEnableTokens, token)
		if !f.isRoot(token) {
			writeVaultError(w, http.StatusForbidden, "permission denied")
			return
		}
		if f.authEnabled {
			writeVaultError(w, http.StatusBadRequest, "path is already in use at kubernetes/")
			return
		}
		f.authEnabled = true
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/v1/auth/kubernetes/config", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if !f.authEnabled {
			writeVaultError(w, http.StatusNotFound, "no handler for route \"auth/kubernetes/config\"")
			return
		}
		if !f.isLive(r.Header.Get("X-Vault-Token")) {
			writeVaultError(w, http.StatusForbidden, "permission denied")
			return
		}
		switch r.Method {
		case http.MethodGet:
			if len(f.authConfig) == 0 {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errors":[]}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": f.authConfig})
		default:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			f.authConfig = body
			f.configWrites = append(f.configWrites, r.Header.Get("X-Vault-Token"))
			w.WriteHeader(http.StatusNoContent)
		}
	})

	mux.HandleFunc("/v1/auth/kubernetes/role/", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if !f.authEnabled {
			writeVaultError(w, http.StatusNotFound, "no handler for route")
			return
		}
		if !f.isLive(r.Header.Get("X-Vault-Token")) {
			writeVaultError(w, http.StatusForbidden, "permission denied")
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/v1/auth/kubernetes/role/")
		switch r.Method {
		case http.MethodGet:
			role, ok := f.authRoles[name]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errors":[]}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": role})
		default:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			// The server reports durations back in seconds.
			for _, k := range []string{"token_ttl", "token_max_ttl"} {
				if s, ok := body[k].(string); ok {
					d, _ := time.ParseDuration(s)
					body[k] = float64(d / time.Second)
				}
			}
			f.authRoles[name] = body
			f.authRoleWrites = append(f.authRoleWrites, r.Header.Get("X-Vault-Token"))
			w.WriteHeader(http.StatusNoContent)
		}
	})

	mux.HandleFunc("/v1/auth/kubernetes/login", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.logins++
		if !f.authEnabled {
			writeVaultError(w, http.StatusNotFound, "no handler for route \"auth/kubernetes/login\"")
			return
		}
		if f.loginFails {
			writeVaultError(w, http.StatusInternalServerError, "lookup failed: tokenreviews.authentication.k8s.io is forbidden")
			return
		}
		var body struct {
			Role string `json:"role"`
			JWT  string `json:"jwt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		role, ok := f.authRoles[body.Role]
		if !ok {
			writeVaultError(w, http.StatusBadRequest, fmt.Sprintf("invalid role name %q", body.Role))
			return
		}
		id, err := identityFromServiceAccountToken(body.JWT)
		if err != nil {
			writeVaultError(w, http.StatusForbidden, err.Error())
			return
		}
		names, _ := role["bound_service_account_names"].([]any)
		namespaces, _ := role["bound_service_account_namespaces"].([]any)
		if !containsAny(names, id.ServiceAccount) || !containsAny(namespaces, id.Namespace) {
			writeVaultError(w, http.StatusForbidden, "service account name not authorized")
			return
		}
		policies, _ := role["token_policies"].([]any)
		tok := f.issue("s.login", stringsOfAny(policies), 600)
		_ = json.NewEncoder(w).Encode(map[string]any{"auth": map[string]any{"client_token": f.clientTokenOf(tok), "accessor": tok.accessor}})
	})

	mux.HandleFunc("/v1/sys/policies/acl/", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if !f.isLive(r.Header.Get("X-Vault-Token")) {
			writeVaultError(w, http.StatusForbidden, "permission denied")
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/v1/sys/policies/acl/")
		switch r.Method {
		case http.MethodGet:
			policy, ok := f.policies[name]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errors":[]}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"name": name, "policy": policy}})
		default:
			var body struct {
				Policy string `json:"policy"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			f.policies[name] = body.Policy
			f.policyWrites = append(f.policyWrites, r.Header.Get("X-Vault-Token"))
			w.WriteHeader(http.StatusNoContent)
		}
	})

	mux.HandleFunc("/v1/auth/token/roles/", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if !f.isLive(r.Header.Get("X-Vault-Token")) {
			writeVaultError(w, http.StatusForbidden, "permission denied")
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/v1/auth/token/roles/")
		switch r.Method {
		case http.MethodGet:
			role, ok := f.tokenRoles[name]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"errors":[]}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": role})
		default:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if s, ok := body["token_period"].(string); ok {
				d, _ := time.ParseDuration(s)
				body["token_period"] = float64(d / time.Second)
			}
			f.tokenRoles[name] = body
			f.tokenRoleWrites = append(f.tokenRoleWrites, r.Header.Get("X-Vault-Token"))
			w.WriteHeader(http.StatusNoContent)
		}
	})

	mux.HandleFunc("/v1/auth/token/create/", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if !f.isLive(r.Header.Get("X-Vault-Token")) {
			writeVaultError(w, http.StatusForbidden, "permission denied")
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/v1/auth/token/create/")
		role, ok := f.tokenRoles[name]
		if !ok {
			writeVaultError(w, http.StatusBadRequest, fmt.Sprintf("unknown role %s", name))
			return
		}
		var body struct {
			Policies []string `json:"policies"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		allowed, _ := role["allowed_policies"].([]any)
		for _, p := range body.Policies {
			if !containsAny(allowed, p) {
				writeVaultError(w, http.StatusBadRequest, fmt.Sprintf("policy %q not allowed by role", p))
				return
			}
		}
		period, _ := role["token_period"].(float64)
		tok := f.issue("s.cp", body.Policies, int(period))
		f.mintedThrough = append(f.mintedThrough, r.Header.Get("X-Vault-Token"))
		_ = json.NewEncoder(w).Encode(map[string]any{"auth": map[string]any{"client_token": f.clientTokenOf(tok), "accessor": tok.accessor}})
	})

	byAccessor := func(w http.ResponseWriter, r *http.Request) (*fakeToken, bool) {
		if !f.isLive(r.Header.Get("X-Vault-Token")) {
			writeVaultError(w, http.StatusForbidden, "permission denied")
			return nil, false
		}
		var body struct {
			Accessor string `json:"accessor"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		client, ok := f.accessors[body.Accessor]
		tok := f.tokens[client]
		if !ok || tok == nil || tok.revoked {
			writeVaultError(w, http.StatusBadRequest, "invalid accessor")
			return nil, false
		}
		return tok, true
	}
	mux.HandleFunc("/v1/auth/token/lookup-accessor", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		tok, ok := byAccessor(w, r)
		if !ok {
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"ttl": tok.ttl, "renewable": true, "policies": tok.policies}})
	})
	mux.HandleFunc("/v1/auth/token/renew-accessor", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		tok, ok := byAccessor(w, r)
		if !ok {
			return
		}
		tok.ttl = int(resources.OpenBAOControlPlaneTokenPeriod / time.Second)
		f.renewals++
		_ = json.NewEncoder(w).Encode(map[string]any{"auth": map[string]any{"client_token": f.clientTokenOf(tok), "accessor": tok.accessor}})
	})
	mux.HandleFunc("/v1/auth/token/revoke-accessor", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		tok, ok := byAccessor(w, r)
		if !ok {
			return
		}
		tok.revoked = true
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/v1/auth/token/revoke-self", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		token := r.Header.Get("X-Vault-Token")
		tok, ok := f.tokens[token]
		if !ok || tok.revoked {
			writeVaultError(w, http.StatusForbidden, "permission denied")
			return
		}
		tok.revoked = true
		f.revokedSelf = append(f.revokedSelf, token)
		w.WriteHeader(http.StatusNoContent)
	})
}

func containsAny(items []any, want string) bool {
	for _, item := range items {
		if s, ok := item.(string); ok && s == want {
			return true
		}
	}
	return false
}

func stringsOfAny(items []any) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// ── the operator's identity in tests ──────────────────────────────────────

// fakeServiceAccountJWT mints an unsigned token whose subject is the given
// ServiceAccount -- the claims the component and the fake read.
func fakeServiceAccountJWT(namespace, name string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(
		`{"sub":"system:serviceaccount:%s:%s","kubernetes.io/serviceaccount/namespace":%q,"kubernetes.io/serviceaccount/service-account.name":%q,"kubernetes.io/serviceaccount/service-account.uid":"uid-%s"}`,
		namespace, name, namespace, name, name)))
	return header + "." + payload + ".signature"
}

func testOperatorIdentity() operatorIdentity {
	return operatorIdentity{Namespace: "planton-operator-system", ServiceAccount: "planton-operator", JWT: fakeServiceAccountJWT("planton-operator-system", "planton-operator")}
}

// withOperatorIdentity points the component's identity seam at a fixed
// identity for the test's duration.
func withOperatorIdentity(t *testing.T, id operatorIdentity) {
	t.Helper()
	previous := readOperatorIdentity
	readOperatorIdentity = func() (*operatorIdentity, error) { return &id, nil }
	t.Cleanup(func() { readOperatorIdentity = previous })
}

func tokenSecretOf(t *testing.T, c client.Client) (*corev1.Secret, bool) {
	t.Helper()
	var s corev1.Secret
	err := c.Get(context.Background(), types.NamespacedName{Name: "planton-openbao-token", Namespace: "planton"}, &s)
	if err != nil {
		return nil, false
	}
	return &s, true
}

// ── initialization writes with root, then proves the session ──────────────

// Initialization enables the method and writes the arrangement with the root
// token; then the operator signs in as itself, and everything after -- the
// engines checked, the token minted -- is done on that session, which is
// revoked at the end. The control plane's token Secret exists, owner-
// referenced, with both keys and the note.
func TestEnsureInitialized_Fresh_ConfiguresAccessWithRootAndMintsOnSession(t *testing.T) {
	vault := newFakeVault(false, true, false)
	srv := vault.serve(t)
	planton := vaultTestPlatform(nil)
	c := vaultFakeClient(t)

	if res := runEnsure(t, planton, c, srv); !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}

	if len(vault.authEnableTokens) != 1 || vault.authEnableTokens[0] != "root-from-init" {
		t.Errorf("the auth method is enabled once, with the root token, got %v", vault.authEnableTokens)
	}
	if !vault.authEnabled || vault.authConfig["kubernetes_host"] != resources.OpenBAOKubernetesHost {
		t.Errorf("the method must be configured with the in-cluster host only, got %v", vault.authConfig)
	}
	if _, pinned := vault.authConfig["kubernetes_ca_cert"]; pinned {
		t.Error("the config must never pin a cluster's CA: a restored vault must work on another cluster")
	}
	if vault.policies["planton-operator"] != resources.OpenBAOOperatorPolicy("planton") || vault.policies["planton-control-plane"] != resources.OpenBAOControlPlanePolicy() {
		t.Error("both policies must be written as the resources package renders them")
	}
	role := vault.authRoles["planton-operator"]
	if !containsAny(role["bound_service_account_names"].([]any), "planton-operator") || !containsAny(role["bound_service_account_namespaces"].([]any), "planton-operator-system") {
		t.Errorf("the operator's role must bind the operator's own identity, got %v", role)
	}
	tokenRole := vault.tokenRoles["planton-control-plane"]
	if tokenRole["orphan"] != true || tokenRole["renewable"] != true || tokenRole["token_period"] != float64(7*24*3600) {
		t.Errorf("the control plane's token role must be orphan, renewable, seven-day periodic, got %v", tokenRole)
	}
	if vault.logins != 1 {
		t.Errorf("the operator signs in exactly once per pass, got %d logins", vault.logins)
	}
	if len(vault.mintedThrough) != 1 || strings.HasPrefix(vault.mintedThrough[0], "root") {
		t.Errorf("the control plane's token is minted on the operator's session, never with root, got %v", vault.mintedThrough)
	}
	if len(vault.revokedSelf) != 1 {
		t.Errorf("the session is revoked when the pass ends, got %v", vault.revokedSelf)
	}
	for _, tok := range vault.mountTokens {
		if tok != "root-from-init" && !strings.HasPrefix(tok, "s.login") {
			t.Errorf("the engines are ensured with root at init or the session after, never anything else, got %q", tok)
		}
	}

	s, ok := tokenSecretOf(t, c)
	if !ok {
		t.Fatal("the control plane's token Secret must exist after initialization")
	}
	if len(s.Data[resources.OpenBAOTokenSecretTokenKey]) == 0 || len(s.Data[resources.OpenBAOTokenSecretAccessorKey]) == 0 {
		t.Errorf("the token Secret carries the token and its accessor, got keys %v", dataKeys(s))
	}
	if string(s.Data[resources.OpenBAOTokenSecretTokenKey]) == "root-from-init" {
		t.Error("the control plane must never be handed the root token")
	}
	if len(s.OwnerReferences) != 1 || s.OwnerReferences[0].Name != "planton" {
		t.Errorf("the token Secret is the operator's own, owner-referenced, got %v", s.OwnerReferences)
	}
	mustContain(t, s.Annotations[resources.OpenBAOTokenSecretAnnotation], "minted by the operator", "Nothing here is the vault's root token")
	minted := vault.tokens[string(s.Data[resources.OpenBAOTokenSecretTokenKey])]
	if minted == nil || len(minted.policies) != 1 || minted.policies[0] != "planton-control-plane" {
		t.Errorf("the control plane's token carries exactly its own policy, got %+v", minted)
	}
}

// ── the steady state ───────────────────────────────────────────────────────

// A steady pass on an open vault with everything in place: the operator
// signs in, reads the arrangement, writes nothing, looks the control
// plane's token up and leaves it alone (its life is long), and revokes its
// session. No root token is presented anywhere.
func TestEnsureInitialized_OpenVault_SteadyPassWritesNothing(t *testing.T) {
	id := testOperatorIdentity()
	vault := newFakeVault(true, false, false).withAccessConfigured("planton", id)
	vault.mounts["secret/"], vault.mounts["transit/"] = true, true
	minted := vault.issue("s.cp", []string{"planton-control-plane"}, int(resources.OpenBAOControlPlaneTokenPeriod/time.Second))
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{InitSecretName: "my-vault-keys"})
	kept := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-vault-keys", Namespace: "planton"},
		Data:       map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("kept-root")},
	}
	tokenSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao-token", Namespace: "planton"},
		Data:       map[string][]byte{resources.OpenBAOTokenSecretTokenKey: []byte(vault.clientTokenOf(minted)), resources.OpenBAOTokenSecretAccessorKey: []byte(minted.accessor)},
	}
	c := vaultFakeClient(t, kept, tokenSecret)

	if res := runEnsure(t, planton, c, srv); !res.Ready || res.Message != "OpenBAO healthy" {
		t.Fatalf("expected a plain Ready, got %+v", res)
	}
	if len(vault.policyWrites)+len(vault.authRoleWrites)+len(vault.tokenRoleWrites)+len(vault.configWrites) != 0 {
		t.Errorf("a steady pass with no drift writes nothing to the arrangement, got policies %v roles %v token roles %v config %v", vault.policyWrites, vault.authRoleWrites, vault.tokenRoleWrites, vault.configWrites)
	}
	if len(vault.mintedThrough) != 0 || vault.renewals != 0 {
		t.Errorf("a token with most of its life left is neither minted nor renewed, got minted %v renewals %d", vault.mintedThrough, vault.renewals)
	}
	for _, tok := range vault.mountTokens {
		if tok == "kept-root" {
			t.Error("the steady state never presents the root token")
		}
	}
	if len(vault.revokedSelf) != 1 {
		t.Errorf("the session is revoked at the end of the pass, got %v", vault.revokedSelf)
	}
	s, _ := tokenSecretOf(t, c)
	if string(s.Data[resources.OpenBAOTokenSecretAccessorKey]) != minted.accessor {
		t.Error("the token Secret is untouched when the token is fine")
	}
}

// Drift is written back: a policy whose text changed (an operator upgrade)
// and a role bound to the wrong identity are rewritten on the session,
// without the root token.
func TestEnsureInitialized_OpenVault_DriftIsRewrittenOnSession(t *testing.T) {
	id := testOperatorIdentity()
	vault := newFakeVault(true, false, false).withAccessConfigured("planton", id)
	vault.mounts["secret/"], vault.mounts["transit/"] = true, true
	vault.policies["planton-control-plane"] = "# an older operator's text\n"
	vault.tokenRoles["planton-control-plane"]["token_period"] = float64(3600)
	srv := vault.serve(t)
	planton := vaultTestPlatform(nil)
	c := vaultFakeClient(t, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao-init", Namespace: "planton"},
		Data:       map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("root-from-init")},
	})

	if res := runEnsure(t, planton, c, srv); !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}
	if vault.policies["planton-control-plane"] != resources.OpenBAOControlPlanePolicy() {
		t.Error("a drifted policy is written back to the current text")
	}
	if vault.tokenRoles["planton-control-plane"]["token_period"] != float64(7*24*3600) {
		t.Error("a drifted token role is written back")
	}
	if len(vault.policyWrites) != 1 || !strings.HasPrefix(vault.policyWrites[0], "s.login") {
		t.Errorf("exactly the drifted policy is rewritten, on the session, got %v", vault.policyWrites)
	}
	if len(vault.authRoleWrites) != 0 || len(vault.configWrites) != 0 {
		t.Errorf("what did not drift is not written, got roles %v config %v", vault.authRoleWrites, vault.configWrites)
	}
}

// ── the control plane's token over time ───────────────────────────────────

// Renewed only when its remaining life has fallen below half the period; the
// Secret and the accessor are unchanged by a renewal.
func TestEnsureInitialized_ControlPlaneToken_RenewedWhenLow(t *testing.T) {
	id := testOperatorIdentity()
	vault := newFakeVault(true, false, false).withAccessConfigured("planton", id)
	vault.mounts["secret/"], vault.mounts["transit/"] = true, true
	minted := vault.issue("s.cp", []string{"planton-control-plane"}, 3600) // an hour left of seven days
	srv := vault.serve(t)
	planton := vaultTestPlatform(nil)
	c := vaultFakeClient(t,
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao-init", Namespace: "planton"}, Data: map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("root-from-init")}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao-token", Namespace: "planton"}, Data: map[string][]byte{resources.OpenBAOTokenSecretTokenKey: []byte(vault.clientTokenOf(minted)), resources.OpenBAOTokenSecretAccessorKey: []byte(minted.accessor)}},
	)

	if res := runEnsure(t, planton, c, srv); !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}
	if vault.renewals != 1 || len(vault.mintedThrough) != 0 {
		t.Errorf("a token with little life left is renewed, not re-minted, got renewals %d minted %v", vault.renewals, vault.mintedThrough)
	}
	s, _ := tokenSecretOf(t, c)
	if string(s.Data[resources.OpenBAOTokenSecretAccessorKey]) != minted.accessor {
		t.Error("a renewal keeps the token and its accessor")
	}
}

// Minted anew when the Secret is missing (the restored case: the data came
// back, the Secret did not) and when the accessor names no live token (a
// revoked token, an operator away longer than the period). A re-mint
// rewrites the Secret in place -- the one credential Secret that is not
// create-once -- and the accessor changes, which is what rolls the control
// plane.
func TestEnsureInitialized_ControlPlaneToken_ReMintedWhenMissingOrInvalid(t *testing.T) {
	t.Run("Secret missing after a restore", func(t *testing.T) {
		id := testOperatorIdentity()
		vault := newFakeVault(true, false, true).withAccessConfigured("planton", id)
		vault.mounts["secret/"], vault.mounts["transit/"] = true, true
		srv := vault.serve(t)
		planton := vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: transitSeal(), InitSecretName: "my-vault-keys"})
		planton.Status.Backup = &v1.BackupStatus{RestoredFrom: "planton-postgres-deadbeef"}
		c := vaultFakeClient(t, transitCredentials(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "my-vault-keys", Namespace: "planton"},
			Data:       map[string][]byte{resources.OpenBAOInitSecretRecoveryKeysKey: []byte(`["r1"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("kept-root")},
		})

		if res := runEnsure(t, planton, c, srv); !res.Ready {
			t.Fatalf("expected Ready, got %+v", res)
		}
		if len(vault.mintedThrough) != 1 {
			t.Errorf("a missing Secret means a fresh token, got minted %v", vault.mintedThrough)
		}
		s, ok := tokenSecretOf(t, c)
		if !ok || len(s.Data[resources.OpenBAOTokenSecretAccessorKey]) == 0 {
			t.Fatal("the token Secret is created with the new token")
		}
		if len(s.OwnerReferences) != 1 {
			t.Error("the re-created token Secret is the operator's own")
		}
	})

	t.Run("token revoked", func(t *testing.T) {
		id := testOperatorIdentity()
		vault := newFakeVault(true, false, false).withAccessConfigured("planton", id)
		vault.mounts["secret/"], vault.mounts["transit/"] = true, true
		old := vault.issue("s.cp", []string{"planton-control-plane"}, 3600)
		old.revoked = true
		srv := vault.serve(t)
		planton := vaultTestPlatform(nil)
		c := vaultFakeClient(t,
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao-init", Namespace: "planton"}, Data: map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("root-from-init")}},
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao-token", Namespace: "planton", Labels: map[string]string{"team": "platform"}}, Data: map[string][]byte{resources.OpenBAOTokenSecretTokenKey: []byte(vault.clientTokenOf(old)), resources.OpenBAOTokenSecretAccessorKey: []byte(old.accessor)}},
		)

		if res := runEnsure(t, planton, c, srv); !res.Ready {
			t.Fatalf("expected Ready, got %+v", res)
		}
		if len(vault.mintedThrough) != 1 || vault.renewals != 0 {
			t.Errorf("a revoked token is re-minted, never renewed, got minted %v renewals %d", vault.mintedThrough, vault.renewals)
		}
		s, _ := tokenSecretOf(t, c)
		if string(s.Data[resources.OpenBAOTokenSecretAccessorKey]) == old.accessor {
			t.Error("the Secret must carry the new token's accessor -- that change is what rolls the control plane")
		}
		if string(s.Data[resources.OpenBAOTokenSecretTokenKey]) == vault.clientTokenOf(old) {
			t.Error("the Secret must carry the new token")
		}
		if s.Labels["team"] != "platform" {
			t.Error("a rewrite keeps what else the Secret carried")
		}
		mustContain(t, s.Annotations[resources.OpenBAOTokenSecretAnnotation], "mints a new one")
	})
}

// ── the sign-in refused ────────────────────────────────────────────────────

// The vault does not know the operator's identity (it was initialized by an
// operator installed under another name) and the init Secret holds the root
// token: the arrangement is repaired with it -- once -- and the pass
// proceeds on the operator's own session.
func TestEnsureInitialized_LoginRefused_RepairedWithRootToken(t *testing.T) {
	previous := operatorIdentity{Namespace: "old-system", ServiceAccount: "old-operator"}
	vault := newFakeVault(true, false, false).withAccessConfigured("planton", previous)
	vault.mounts["secret/"], vault.mounts["transit/"] = true, true
	vault.roots["kept-root"] = true
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{InitSecretName: "my-vault-keys"})
	c := vaultFakeClient(t, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-vault-keys", Namespace: "planton"},
		Data:       map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("kept-root")},
	})

	if res := runEnsure(t, planton, c, srv); !res.Ready {
		t.Fatalf("expected Ready after repair, got %+v", res)
	}
	role := vault.authRoles["planton-operator"]
	if !containsAny(role["bound_service_account_names"].([]any), "planton-operator") || !containsAny(role["bound_service_account_namespaces"].([]any), "planton-operator-system") {
		t.Errorf("the repair rebinds the role to the current identity, got %v", role)
	}
	if len(vault.authRoleWrites) == 0 || vault.authRoleWrites[0] != "kept-root" {
		t.Errorf("the repair is written with the root token, got %v", vault.authRoleWrites)
	}
	if vault.logins != 2 {
		t.Errorf("one refused login, one accepted after repair, got %d", vault.logins)
	}
	if len(vault.mintedThrough) != 1 || vault.mintedThrough[0] == "kept-root" {
		t.Errorf("after the repair the pass runs on the session, got minted with %v", vault.mintedThrough)
	}
}

// The vault refuses the identity and there is no root token to repair with:
// a refusal naming the identity, the Secret, the binding, and the two ways
// back.
func TestEnsureInitialized_LoginRefused_NoRootTokenIsRefused(t *testing.T) {
	previous := operatorIdentity{Namespace: "old-system", ServiceAccount: "old-operator"}
	vault := newFakeVault(true, false, true).withAccessConfigured("planton", previous)
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: transitSeal(), InitSecretName: "my-vault-keys"})
	c := vaultFakeClient(t, transitCredentials())

	res := runEnsure(t, planton, c, srv)
	if res.Ready || res.Reason != v1.ComponentReasonConfigurationRefused || res.Object == nil || res.Object.Name != "my-vault-keys" {
		t.Fatalf("expected ConfigurationRefused naming the init Secret, got %+v", res)
	}
	mustContain(t, res.Message,
		"refused the operator's sign-in as planton-operator-system/planton-operator",
		"service account name not authorized",
		"Secret my-vault-keys in namespace planton does not exist",
		"install the operator under that identity",
		"recreate Secret my-vault-keys from the copy you kept",
		"ClusterRoleBinding planton-planton-openbao-auth-delegator",
		"ServiceAccount planton-openbao")
	if len(vault.mintedThrough) != 0 {
		t.Error("nothing is minted without a session")
	}
}

// The role is right and the login still fails: the vault cannot verify the
// token with the API server -- the binding is what to check.
func TestEnsureInitialized_LoginRefused_AfterRepairNamesTheBinding(t *testing.T) {
	vault := newFakeVault(true, false, false).withAccessConfigured("planton", testOperatorIdentity())
	vault.loginFails = true
	srv := vault.serve(t)
	planton := vaultTestPlatform(nil)
	c := vaultFakeClient(t, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao-init", Namespace: "planton"},
		Data:       map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("root-from-init")},
	})

	res := runEnsure(t, planton, c, srv)
	if res.Ready || res.Reason != v1.ComponentReasonConfigurationRefused || res.Object == nil || res.Object.Kind != "ClusterRoleBinding" {
		t.Fatalf("expected ConfigurationRefused naming the binding, got %+v", res)
	}
	mustContain(t, res.Message, "still refuses", "lookup failed", "system:auth-delegator", "planton-openbao", resources.OpenBAOKubernetesHost)
}

// The root token in the init Secret no longer works (revoked or replaced):
// the repair itself is refused and the sentence says how to mint a new one.
func TestEnsureInitialized_LoginRefused_StaleRootTokenIsRefused(t *testing.T) {
	previous := operatorIdentity{Namespace: "old-system", ServiceAccount: "old-operator"}
	vault := newFakeVault(true, false, false).withAccessConfigured("planton", previous)
	srv := vault.serve(t)
	planton := vaultTestPlatform(nil)
	c := vaultFakeClient(t, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao-init", Namespace: "planton"},
		Data:       map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("revoked-root")},
	})

	res := runEnsure(t, planton, c, srv)
	if res.Ready || res.Reason != v1.ComponentReasonConfigurationRefused {
		t.Fatalf("expected ConfigurationRefused, got %+v", res)
	}
	mustContain(t, res.Message, "could not repair its access", "permission denied", "revoked or replaced", "recovery quorum", resources.OpenBAOInitSecretRootTokenKey)
}

// The operator reads its identity from its mounted token's subject.
func TestIdentityFromServiceAccountToken(t *testing.T) {
	id, err := identityFromServiceAccountToken(fakeServiceAccountJWT("ops", "planton-operator"))
	if err != nil {
		t.Fatal(err)
	}
	if id.Namespace != "ops" || id.ServiceAccount != "planton-operator" || id.String() != "ops/planton-operator" {
		t.Errorf("identity = %+v", id)
	}
	for name, jwt := range map[string]string{
		"not a JWT":        "nope",
		"not an SA":        "h." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"alice"}`)) + ".s",
		"no account":       "h." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"system:serviceaccount:ops"}`)) + ".s",
		"undecodable body": "h.!!!.s",
	} {
		if _, err := identityFromServiceAccountToken(jwt); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
