package component

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	v1 "github.com/plantonhq/planton/operator/api/v1"
	"github.com/plantonhq/planton/operator/internal/resources"
)

// Every arm of the vault's initialization is pinned here against a fake
// OpenBao and a fake cluster: which init shape is sent for which seal, what
// the init Secret carries and who owns it, when the operator unseals and
// when it must not, and the sentence a person reads in every state the
// storage can be in. A real server proves the same arms in Docker
// (bootstrap's requires_docker suites); this is the offline bar.

// fakeVault is the OpenBao HTTP surface the component talks to: health,
// init, unseal, mounts. Its seal type decides which init shape it accepts,
// exactly as the real server does, and a cloud-sealed fake opens itself on
// the poll after init the way the real server's loop does.
type fakeVault struct {
	mu          sync.Mutex
	initialized bool
	sealed      bool
	cloudSeal   bool
	// opensItself makes a cloud-sealed fake flip to unsealed on the first
	// health poll after init (the server's own loop); false keeps it sealed
	// (a decrypt the identity lacks).
	opensItself bool
	initBodies  []map[string]int
	unsealCalls int
	mounts      map[string]bool
	rootToken   string
	mountTokens []string
}

func newFakeVault(initialized, sealed, cloudSeal bool) *fakeVault {
	return &fakeVault{initialized: initialized, sealed: sealed, cloudSeal: cloudSeal, opensItself: true, mounts: map[string]bool{}, rootToken: "root-from-init"}
}

func (f *fakeVault) serve(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/sys/health", func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		resp := map[string]any{"initialized": f.initialized, "sealed": f.sealed}
		if f.cloudSeal && f.initialized && f.sealed && f.opensItself {
			f.sealed = false // the server's own loop, one interval later
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/sys/init", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		var body map[string]int
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.initBodies = append(f.initBodies, body)
		if f.initialized {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"errors":["Vault is already initialized"]}`))
			return
		}
		_, hasSecret := body["secret_shares"]
		_, hasRecovery := body["recovery_shares"]
		if f.cloudSeal && hasSecret || !f.cloudSeal && hasRecovery {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"errors":["parameters not applicable to seal type"]}`))
			return
		}
		f.initialized, f.sealed = true, true
		resp := map[string]any{"root_token": f.rootToken, "keys": []string{}, "recovery_keys": []string{}}
		if f.cloudSeal {
			resp["recovery_keys"] = []string{"r1", "r2", "r3", "r4", "r5"}
		} else {
			resp["keys"] = []string{"k1", "k2", "k3", "k4", "k5"}
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/sys/unseal", func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.unsealCalls++
		if f.cloudSeal {
			t.Error("the operator must never call /sys/unseal under a cloud seal")
		}
		f.sealed = false
		_ = json.NewEncoder(w).Encode(map[string]any{"sealed": false})
	})
	mux.HandleFunc("/v1/sys/mounts", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.mountTokens = append(f.mountTokens, r.Header.Get("X-Vault-Token"))
		out := map[string]any{"sys/": map[string]any{}}
		for m := range f.mounts {
			out[m] = map[string]any{}
		}
		_ = json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("/v1/sys/mounts/", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.mounts[strings.TrimPrefix(r.URL.Path, "/v1/sys/mounts/")+"/"] = true
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func vaultTestPlatform(vault *v1.OpenBAOSpec) *v1.PlantonPlatform {
	planton := &v1.PlantonPlatform{
		TypeMeta:   metav1.TypeMeta{APIVersion: "planton.ai/v1", Kind: "PlantonPlatform"},
		ObjectMeta: metav1.ObjectMeta{Name: "planton", Namespace: "planton", UID: types.UID("1a2b3c4d-0000-4000-8000-000000000000")},
	}
	planton.Spec.Vault = vault
	return planton
}

func transitSeal() *v1.OpenBAOAutoUnsealSpec {
	return &v1.OpenBAOAutoUnsealSpec{Transit: &v1.OpenBAOTransitSealSpec{Address: "http://holder:8200", KeyName: "planton-unseal", CredentialsSecretName: "holder-token"}}
}

func transitCredentials() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "holder-token", Namespace: "planton"},
		Data:       map[string][]byte{resources.OpenBAOSealEnvTransitToken: []byte("s.token")},
	}
}

func vaultFakeClient(t *testing.T, objs ...client.Object) client.Client {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(subOperatorScheme(t)).WithObjects(objs...).Build()
}

func initSecretOf(t *testing.T, c client.Client, name string) *corev1.Secret {
	t.Helper()
	var s corev1.Secret
	if err := c.Get(context.Background(), types.NamespacedName{Name: name, Namespace: "planton"}, &s); err != nil {
		t.Fatalf("init Secret %s: %v", name, err)
	}
	return &s
}

func runEnsure(t *testing.T, planton *v1.PlantonPlatform, c client.Client, srv *httptest.Server) Result {
	t.Helper()
	o := &OpenBAO{}
	seal := sealOptionsFrom(planton)
	initSecret, err := readInitSecret(context.Background(), c, vaultInitSecretName(planton), planton.Namespace)
	if err != nil {
		t.Fatal(err)
	}
	res, err := o.ensureInitialized(context.Background(), c, planton, seal, initSecret, srv.URL, srv.Client())
	if err != nil {
		t.Fatalf("ensureInitialized: %v", err)
	}
	return res
}

// A fresh vault under the built-in seal: init with secret_shares, the unseal
// keys and root token in the operator's own Secret with an owner reference,
// the operator unseals, the engines are mounted, Ready.
func TestEnsureInitialized_FreshShamir_OperatorOwnedSecret(t *testing.T) {
	vault := newFakeVault(false, true, false)
	srv := vault.serve(t)
	planton := vaultTestPlatform(nil)
	c := vaultFakeClient(t)

	res := runEnsure(t, planton, c, srv)
	if !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}
	if len(vault.initBodies) != 1 || vault.initBodies[0]["secret_shares"] != 5 || vault.initBodies[0]["secret_threshold"] != 3 {
		t.Errorf("init must ask for 5/3 key shares, got %v", vault.initBodies)
	}
	if vault.unsealCalls == 0 {
		t.Error("the operator unseals a built-in-seal vault after init")
	}
	if !vault.mounts["secret/"] || !vault.mounts["transit/"] {
		t.Errorf("the engines must be mounted, got %v", vault.mounts)
	}

	s := initSecretOf(t, c, "planton-openbao-init")
	if string(s.Data[resources.OpenBAOInitSecretRootTokenKey]) != "root-from-init" {
		t.Error("the root token must be written")
	}
	if len(s.Data[resources.OpenBAOInitSecretUnsealKeysKey]) == 0 || len(s.Data[resources.OpenBAOInitSecretRecoveryKeysKey]) != 0 {
		t.Errorf("a built-in-seal vault's Secret carries unseal-keys and no recovery-keys, got keys %v", dataKeys(s))
	}
	if len(s.OwnerReferences) != 1 || s.OwnerReferences[0].Name != "planton" {
		t.Errorf("the operator's own Secret is owner-referenced to the platform, got %v", s.OwnerReferences)
	}
	if s.Annotations[resources.OpenBAOSealAnnotation] != resources.OpenBAOSealShamir {
		t.Errorf("the fingerprint must record the built-in seal, got %q", s.Annotations[resources.OpenBAOSealAnnotation])
	}
	mustContain(t, s.Annotations[resources.OpenBAOInitSecretAnnotation], "Unseal keys", "deleted with the platform")
}

// The adopter names the Secret: written WITHOUT an owner reference, with the
// note that says they own it and what keeps it.
func TestEnsureInitialized_FreshShamir_AdopterNamedSecret(t *testing.T) {
	vault := newFakeVault(false, true, false)
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{InitSecretName: "my-vault-keys"})
	c := vaultFakeClient(t)

	if res := runEnsure(t, planton, c, srv); !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}
	s := initSecretOf(t, c, "my-vault-keys")
	if len(s.OwnerReferences) != 0 {
		t.Errorf("the adopter's Secret carries no owner reference, got %v", s.OwnerReferences)
	}
	if s.Labels["app.kubernetes.io/managed-by"] != resources.ManagedByLabel {
		t.Error("the Secret is labeled like every operator-written object")
	}
	mustContain(t, s.Annotations[resources.OpenBAOInitSecretAnnotation], "You own this Secret", "never deletes it", "keep a copy")
	if _, exists := readOrNil(t, c, "planton-openbao-init"); exists {
		t.Error("no operator-owned init Secret is created when the adopter named one")
	}
}

// A fresh vault under a cloud seal: init with recovery_shares and nothing
// else, recovery-keys written, the operator never unseals, the server opens
// itself and the pass finds it open.
func TestEnsureInitialized_FreshCloudSeal_ServerOpensItself(t *testing.T) {
	vault := newFakeVault(false, true, true)
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: transitSeal(), InitSecretName: "my-vault-keys"})
	c := vaultFakeClient(t, transitCredentials())

	res := runEnsure(t, planton, c, srv)
	if !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}
	body := vault.initBodies[0]
	if body["recovery_shares"] != 5 || body["recovery_threshold"] != 3 {
		t.Errorf("init must ask for a 5/3 recovery quorum, got %v", body)
	}
	if _, has := body["secret_shares"]; has {
		t.Errorf("init under a cloud seal must not carry secret_shares, got %v", body)
	}
	if vault.unsealCalls != 0 {
		t.Error("the operator never unseals a cloud-sealed vault")
	}
	s := initSecretOf(t, c, "my-vault-keys")
	if len(s.Data[resources.OpenBAOInitSecretRecoveryKeysKey]) == 0 || len(s.Data[resources.OpenBAOInitSecretUnsealKeysKey]) != 0 {
		t.Errorf("a cloud-sealed vault's Secret carries recovery-keys and no unseal-keys, got keys %v", dataKeys(s))
	}
	if got := s.Annotations[resources.OpenBAOSealAnnotation]; got != "transit http://holder:8200/transit/planton-unseal" {
		t.Errorf("fingerprint = %q", got)
	}
	mustContain(t, s.Annotations[resources.OpenBAOInitSecretAnnotation], "Recovery keys", "opens itself", "break-glass")
}

// The adopter pre-created the Secret empty (GitOps owns its existence): the
// operator fills it in place and adds no owner reference.
func TestEnsureInitialized_AdopterPreCreatedEmptySecretIsFilled(t *testing.T) {
	vault := newFakeVault(false, true, false)
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{InitSecretName: "my-vault-keys"})
	empty := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "my-vault-keys", Namespace: "planton", Labels: map[string]string{"team": "platform"}}}
	c := vaultFakeClient(t, empty)

	if res := runEnsure(t, planton, c, srv); !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}
	s := initSecretOf(t, c, "my-vault-keys")
	if len(s.Data[resources.OpenBAOInitSecretUnsealKeysKey]) == 0 || string(s.Data[resources.OpenBAOInitSecretRootTokenKey]) == "" {
		t.Errorf("the empty Secret must be filled with the keys, got %v", dataKeys(s))
	}
	if s.Labels["team"] != "platform" {
		t.Error("the adopter's own labels are kept")
	}
	if len(s.OwnerReferences) != 0 {
		t.Error("no owner reference on the adopter's Secret")
	}
}

// A Secret that already holds keys while the vault is not initialized is
// refused, never written over -- for the adopter's Secret and the
// operator's own alike -- and /sys/init is never called.
func TestEnsureInitialized_PopulatedSecretOnUninitializedVaultIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name   string
		vault  *v1.OpenBAOSpec
		secret string
		who    string
	}{
		{"adopter's", &v1.OpenBAOSpec{InitSecretName: "my-vault-keys"}, "my-vault-keys", "spec.vault.initSecretName"},
		{"operator's", nil, "planton-openbao-init", "The operator owns it"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vault := newFakeVault(false, true, false)
			srv := vault.serve(t)
			planton := vaultTestPlatform(tc.vault)
			foreign := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: tc.secret, Namespace: "planton"},
				Data:       map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["old"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("old-root")},
			}
			c := vaultFakeClient(t, foreign)

			res := runEnsure(t, planton, c, srv)
			if res.Ready || res.Reason != v1.ComponentReasonConfigurationRefused || res.Object == nil || res.Object.Name != tc.secret {
				t.Fatalf("expected ConfigurationRefused naming the Secret, got %+v", res)
			}
			mustContain(t, res.Message, "already holds a vault's keys", "not initialized", tc.who, "recoverFrom", "delete or rename")
			if len(vault.initBodies) != 0 {
				t.Error("/sys/init must not be called while another vault's keys sit in the Secret")
			}
			s := initSecretOf(t, c, tc.secret)
			if string(s.Data[resources.OpenBAOInitSecretRootTokenKey]) != "old-root" {
				t.Error("the foreign keys must be untouched")
			}
		})
	}
}

// A restored (or restarted) built-in-seal vault with its Secret present is
// unsealed from the Secret and its engines re-ensured with the root token.
func TestEnsureInitialized_RestoredShamir_UnsealsFromSecret(t *testing.T) {
	vault := newFakeVault(true, true, false)
	vault.mounts["secret/"], vault.mounts["transit/"] = true, true
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{InitSecretName: "my-vault-keys"})
	kept := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-vault-keys", Namespace: "planton"},
		Data:       map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1","k2","k3"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("kept-root")},
	}
	c := vaultFakeClient(t, kept)

	res := runEnsure(t, planton, c, srv)
	if !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}
	if vault.unsealCalls == 0 || vault.sealed {
		t.Error("the operator unseals with the kept shares")
	}
	if len(vault.mountTokens) == 0 || vault.mountTokens[0] != "kept-root" {
		t.Errorf("the engines are checked with the kept root token, got %v", vault.mountTokens)
	}
}

// A restored built-in-seal vault with no Secret is refused with the Secret
// named: the data came back from the archive, the keys did not.
func TestEnsureInitialized_RestoredShamir_NoSecretIsRefused(t *testing.T) {
	vault := newFakeVault(true, true, false)
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{InitSecretName: "my-vault-keys"})
	planton.Status.Backup = &v1.BackupStatus{RestoredFrom: "planton-postgres-deadbeef"}
	c := vaultFakeClient(t)

	res := runEnsure(t, planton, c, srv)
	if res.Ready || res.Reason != v1.ComponentReasonConfigurationRefused || res.Object == nil || res.Object.Name != "my-vault-keys" {
		t.Fatalf("expected ConfigurationRefused naming the Secret, got %+v", res)
	}
	mustContain(t, res.Message, "came back from archive planton-postgres-deadbeef", "my-vault-keys", "Recreate that Secret")
	if vault.unsealCalls != 0 {
		t.Error("nothing to unseal with")
	}
}

// A cloud-sealed vault that stays sealed: a young pod is opening; an old pod
// names the grant the identity lacks. The operator never unseals either.
func TestEnsureInitialized_CloudSealStaysSealed_SentenceByPodAge(t *testing.T) {
	for _, tc := range []struct {
		name  string
		age   time.Duration
		wants []string
	}{
		{"young pod is opening", 10 * time.Second, []string{"is opening it", "every five seconds"}},
		{"old pod names the grant", 5 * time.Minute, []string{"cannot decrypt with it", "decrypt on the transit key"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vault := newFakeVault(true, true, true)
			vault.opensItself = false
			srv := vault.serve(t)
			planton := vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: transitSeal(), InitSecretName: "my-vault-keys"})
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao-0", Namespace: "planton", Labels: map[string]string{"app.kubernetes.io/name": "planton-openbao"}},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning, StartTime: &metav1.Time{Time: time.Now().Add(-tc.age)}},
			}
			c := vaultFakeClient(t, transitCredentials(), pod)

			res := runEnsure(t, planton, c, srv)
			if res.Ready || res.Reason != v1.ComponentReasonDeploying {
				t.Fatalf("expected a Deploying wait, got %+v", res)
			}
			mustContain(t, res.Message, tc.wants...)
			if vault.unsealCalls != 0 {
				t.Error("the operator never unseals a cloud-sealed vault")
			}
		})
	}
}

// An open vault whose Secret does not exist is refused with the Secret named
// and the ways back: the operator and the control plane sign in with the
// token it holds.
func TestEnsureInitialized_OpenVaultWithoutSecretIsRefused(t *testing.T) {
	vault := newFakeVault(true, false, true)
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: transitSeal(), InitSecretName: "my-vault-keys"})
	planton.Status.Backup = &v1.BackupStatus{RestoredFrom: "planton-postgres-deadbeef"}
	c := vaultFakeClient(t, transitCredentials())

	res := runEnsure(t, planton, c, srv)
	if res.Ready || res.Reason != v1.ComponentReasonConfigurationRefused || res.Object == nil || res.Object.Name != "my-vault-keys" {
		t.Fatalf("expected ConfigurationRefused naming the Secret, got %+v", res)
	}
	mustContain(t, res.Message, "came back from archive planton-postgres-deadbeef", "transit key opened it", "does not exist", "Recreate it from the copy you kept", "renamed spec.vault.initSecretName")
}

// The steady state re-asserts the note and the fingerprint: a Secret
// recreated from its data alone regains its pin; a recorded fingerprint is
// re-asserted with its own value, never the declaration's.
func TestEnsureInitialized_OpenVault_AnnotationsReasserted(t *testing.T) {
	vault := newFakeVault(true, false, false)
	vault.mounts["secret/"], vault.mounts["transit/"] = true, true
	srv := vault.serve(t)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{InitSecretName: "my-vault-keys"})

	bare := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-vault-keys", Namespace: "planton"},
		Data:       map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1"]`), resources.OpenBAOInitSecretRootTokenKey: []byte("r")},
	}
	c := vaultFakeClient(t, bare)
	if res := runEnsure(t, planton, c, srv); !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}
	s := initSecretOf(t, c, "my-vault-keys")
	if s.Annotations[resources.OpenBAOSealAnnotation] != resources.OpenBAOSealShamir {
		t.Errorf("a Secret without a fingerprint regains one from the open vault's declaration, got %q", s.Annotations[resources.OpenBAOSealAnnotation])
	}
	mustContain(t, s.Annotations[resources.OpenBAOInitSecretAnnotation], "Unseal keys", "You own this Secret")
	if string(s.Data[resources.OpenBAOInitSecretRootTokenKey]) != "r" {
		t.Error("the annotation pass never touches the data")
	}

	recorded := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-vault-keys", Namespace: "planton", Annotations: map[string]string{resources.OpenBAOSealAnnotation: "gcpKms p/global/r/k"}},
		Data:       bare.Data,
	}
	c2 := vaultFakeClient(t, recorded)
	vault2 := newFakeVault(true, false, false)
	vault2.mounts["secret/"], vault2.mounts["transit/"] = true, true
	if res := runEnsure(t, planton, c2, vault2.serve(t)); !res.Ready {
		t.Fatalf("expected Ready, got %+v", res)
	}
	if got := initSecretOf(t, c2, "my-vault-keys").Annotations[resources.OpenBAOSealAnnotation]; got != "gcpKms p/global/r/k" {
		t.Errorf("a recorded fingerprint is never rewritten by the steady-state pass, got %q", got)
	}
}

// Before the chart renders: a seal arm's credentials Secret must exist with
// its key, and the declared seal must be the recorded one.
func TestPreflightSeal(t *testing.T) {
	o := &OpenBAO{}
	planton := vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: transitSeal(), InitSecretName: "my-vault-keys"})
	seal := sealOptionsFrom(planton)

	t.Run("credentials Secret missing", func(t *testing.T) {
		res, _, err := o.preflightSeal(context.Background(), vaultFakeClient(t), planton, seal)
		if err != nil || res == nil || res.Reason != v1.ComponentReasonConfigurationRefused || res.Object.Name != "holder-token" {
			t.Fatalf("expected a refusal naming holder-token, got %+v %v", res, err)
		}
		mustContain(t, res.Message, `"holder-token"`, "spec.vault.autoUnseal.transit.credentialsSecretName", "VAULT_TOKEN", "the vault is not started")
	})
	t.Run("credentials Secret without its key", func(t *testing.T) {
		wrong := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "holder-token", Namespace: "planton"}, Data: map[string][]byte{"token": []byte("x")}}
		res, _, err := o.preflightSeal(context.Background(), vaultFakeClient(t, wrong), planton, seal)
		if err != nil || res == nil || res.Reason != v1.ComponentReasonConfigurationRefused {
			t.Fatalf("expected a refusal, got %+v %v", res, err)
		}
		mustContain(t, res.Message, `has no "VAULT_TOKEN" key`)
	})
	t.Run("keyless arm needs nothing", func(t *testing.T) {
		gcp := vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: &v1.OpenBAOAutoUnsealSpec{GcpKms: &v1.OpenBAOGcpKmsSealSpec{Project: "p", Region: "global", KeyRing: "r", CryptoKey: "k"}}})
		res, _, err := o.preflightSeal(context.Background(), vaultFakeClient(t), gcp, sealOptionsFrom(gcp))
		if err != nil || res != nil {
			t.Fatalf("a keyless arm preflights nothing, got %+v %v", res, err)
		}
	})
	t.Run("declared seal differs from the recorded one", func(t *testing.T) {
		recorded := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "my-vault-keys", Namespace: "planton", Annotations: map[string]string{resources.OpenBAOSealAnnotation: resources.OpenBAOSealShamir}},
			Data:       map[string][]byte{resources.OpenBAOInitSecretUnsealKeysKey: []byte(`["k1"]`)},
		}
		res, _, err := o.preflightSeal(context.Background(), vaultFakeClient(t, transitCredentials(), recorded), planton, seal)
		if err != nil || res == nil || res.Reason != v1.ComponentReasonConfigurationRefused || res.Object.Name != "my-vault-keys" {
			t.Fatalf("expected a refusal naming the init Secret, got %+v %v", res, err)
		}
		mustContain(t, res.Message, `initialized with the seal "shamir"`, `now says "transit http://holder:8200/transit/planton-unseal"`, "decided when the vault is created", "restore the declaration")
	})
	t.Run("same seal passes and hands the Secret on", func(t *testing.T) {
		recorded := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "my-vault-keys", Namespace: "planton", Annotations: map[string]string{resources.OpenBAOSealAnnotation: seal.Fingerprint()}},
		}
		res, got, err := o.preflightSeal(context.Background(), vaultFakeClient(t, transitCredentials(), recorded), planton, seal)
		if err != nil || res != nil || got == nil {
			t.Fatalf("expected no refusal and the Secret read, got %+v %v %v", res, got, err)
		}
	})
}

// A pod that runs but never answers is a crash-loop above all, and a
// crash-looping pod still reports phase Running: the health-failure arm goes
// through the classifier, which names the pod and its exit, and under a
// cloud seal the seal-check clause is appended.
func TestEnsureInitialized_HealthFailureIsClassified_WithSealHint(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("not json"))
	}))
	t.Cleanup(dead.Close)
	planton := vaultTestPlatform(&v1.OpenBAOSpec{AutoUnseal: &v1.OpenBAOAutoUnsealSpec{GcpKms: &v1.OpenBAOGcpKmsSealSpec{Project: "p", Region: "global", KeyRing: "r", CryptoKey: "k"}}})
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "planton-openbao", Namespace: "planton"},
		Spec:       appsv1.StatefulSetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app.kubernetes.io/name": "planton-openbao"}}},
	}
	pod := runningPod("planton-openbao-0", 2*time.Minute, 4)
	pod.Labels = map[string]string{"app.kubernetes.io/name": "planton-openbao"}
	pod.Status.ContainerStatuses[0].State = corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}}
	pod.Status.ContainerStatuses[0].LastTerminationState = corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 1}}
	c := vaultFakeClient(t, sts, &pod)

	o := &OpenBAO{}
	res, err := o.ensureInitialized(context.Background(), c, planton, sealOptionsFrom(planton), nil, dead.URL, dead.Client())
	if err != nil {
		t.Fatal(err)
	}
	if res.Reason != v1.ComponentReasonCrashLooping {
		t.Fatalf("expected CrashLooping from the classifier, got %+v", res)
	}
	mustContain(t, res.Message, "keeps exiting", "kubectl logs", "Under a Google Cloud KMS key", "roles/cloudkms.viewer", "Error configuring seal")
}

func dataKeys(s *corev1.Secret) []string {
	keys := make([]string, 0, len(s.Data))
	for k := range s.Data {
		keys = append(keys, k)
	}
	return keys
}

func readOrNil(t *testing.T, c client.Client, name string) (*corev1.Secret, bool) {
	t.Helper()
	s, err := readInitSecret(context.Background(), c, name, "planton")
	if err != nil {
		t.Fatal(err)
	}
	return s, s != nil
}
