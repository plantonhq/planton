package aa_e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/plantonhq/planton/e2e/framework/provider"
)

// fakeStripe is a tiny in-memory Stripe: objects by path, each active or not, DELETE removes one.
type fakeStripe struct {
	mu      sync.Mutex
	objects map[string]string // path -> JSON body
	deleted []string
}

func (f *fakeStripe) put(path, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[path] = body
}

func (f *fakeStripe) serve(t *testing.T) (*Harness, func()) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		body, ok := f.objects[r.URL.Path]
		switch {
		case r.Method == http.MethodDelete && ok:
			delete(f.objects, r.URL.Path)
			f.deleted = append(f.deleted, r.URL.Path)
			_, _ = w.Write([]byte(`{"deleted":true}`))
		case ok:
			_, _ = w.Write([]byte(body))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"No such object"}}`))
		}
	}))
	h := NewHarness()
	h.client = NewClient("rk_test_key", "")
	h.client.baseURL = server.URL
	return h, server.Close
}

func newFake() *fakeStripe { return &fakeStripe{objects: map[string]string{}} }

// secondAct writes a second-act manifest declaring the upgrade's expectation and returns the
// contexts of both acts as the runner builds them.
func secondAct(t *testing.T, expectation string) (first, second context.Context) {
	dir := t.TempDir()
	firstPath := filepath.Join(dir, "minimal.yaml")
	secondPath := filepath.Join(dir, "minimal.to.yaml")
	annotation := ""
	if expectation != "" {
		annotation = "\n    " + provider.ExpectUpgradeAnnotation + ": " + expectation
	}
	body := "apiVersion: stripe.planton.dev/v1alpha1\nkind: StripePrice\nmetadata:\n  name: p\n  annotations:\n    planton.dev/e2e-second-act: minimal.yaml" + annotation + "\nspec: {}\n"
	for _, p := range []string{firstPath, secondPath} {
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	first = context.WithValue(context.Background(), provider.ManifestPathKey{}, firstPath)
	second = context.WithValue(context.Background(), provider.ManifestPathKey{}, secondPath)
	second = context.WithValue(second, provider.FirstActManifestPathKey{}, firstPath)
	return first, second
}

// An upgrade declared in-place must keep the object; declared replaced, it must yield a new
// object AND leave the old one the way the kind's destroy does (a price archived).
func TestHarness_JudgesTheUpgrade(t *testing.T) {
	f := newFake()
	h, stop := f.serve(t)
	defer stop()
	f.put("/v1/prices/price_old", `{"id":"price_old","active":true}`)
	f.put("/v1/prices/price_new", `{"id":"price_new","active":true}`)

	first, second := secondAct(t, provider.UpgradeInPlace)
	if err := h.VerifyDeployed(first, "stripeprice", map[string]interface{}{"id": "price_old"}); err != nil {
		t.Fatal(err)
	}
	if err := h.VerifyDeployed(second, "stripeprice", map[string]interface{}{"id": "price_old"}); err != nil {
		t.Errorf("an in-place upgrade that kept the object passes, got %v", err)
	}
	if err := h.VerifyDeployed(second, "stripeprice", map[string]interface{}{"id": "price_new"}); err == nil || !strings.Contains(err.Error(), "the change replaced it") {
		t.Errorf("an upgrade declared in-place that replaced the object must fail, got %v", err)
	}

	first, second = secondAct(t, provider.UpgradeReplaced)
	if err := h.VerifyDeployed(first, "stripeprice", map[string]interface{}{"id": "price_old"}); err != nil {
		t.Fatal(err)
	}
	if err := h.VerifyDeployed(second, "stripeprice", map[string]interface{}{"id": "price_new"}); err == nil || !strings.Contains(err.Error(), "the object the upgrade replaced") {
		t.Errorf("a replacement that left the old price active must fail on the old price, got %v", err)
	}
	f.put("/v1/prices/price_old", `{"id":"price_old","active":false}`)
	if err := h.VerifyDeployed(second, "stripeprice", map[string]interface{}{"id": "price_new"}); err != nil {
		t.Errorf("a replacement that archived the old price passes, got %v", err)
	}
	f.put("/v1/prices/price_old", `{"id":"price_old","active":true}`)
	if err := h.VerifyDeployed(second, "stripeprice", map[string]interface{}{"id": "price_old"}); err == nil || !strings.Contains(err.Error(), "kept its id") {
		t.Errorf("an upgrade declared replaced that kept the object must fail, got %v", err)
	}

	first, second = secondAct(t, "")
	if err := h.VerifyDeployed(first, "stripeprice", map[string]interface{}{"id": "price_new"}); err != nil {
		t.Fatal(err)
	}
	if err := h.VerifyDeployed(second, "stripeprice", map[string]interface{}{"id": "price_new"}); err == nil || !strings.Contains(err.Error(), provider.ExpectUpgradeAnnotation) {
		t.Errorf("a second act that declares nothing must be refused, got %v", err)
	}
}

// A signing secret is checked by shape and compared only by digest: it must come with every
// webhook endpoint, change when the endpoint is replaced, and never appear in an error.
func TestHarness_ChecksTheSigningSecret(t *testing.T) {
	f := newFake()
	h, stop := f.serve(t)
	defer stop()
	f.put("/v1/webhook_endpoints/we_old", `{"id":"we_old"}`)
	f.put("/v1/webhook_endpoints/we_new", `{"id":"we_new"}`)

	if err := h.VerifyDeployed(context.Background(), "stripewebhookendpoint", map[string]interface{}{"id": "we_old"}); err == nil || !strings.Contains(err.Error(), "empty after create") {
		t.Errorf("an endpoint without its secret must fail, got %v", err)
	}
	err := h.VerifyDeployed(context.Background(), "stripewebhookendpoint", map[string]interface{}{"id": "we_old", "secret": "not-a-secret-value"})
	if err == nil || !strings.Contains(err.Error(), "not a signing secret") || strings.Contains(err.Error(), "not-a-secret-value") {
		t.Errorf("a malformed secret must fail without being named, got %v", err)
	}

	first, second := secondAct(t, provider.UpgradeReplaced)
	if err := h.VerifyDeployed(first, "stripewebhookendpoint", map[string]interface{}{"id": "we_old", "secret": "whsec_one"}); err != nil {
		t.Fatal(err)
	}
	delete(f.objects, "/v1/webhook_endpoints/we_old")
	err = h.VerifyDeployed(second, "stripewebhookendpoint", map[string]interface{}{"id": "we_new", "secret": "whsec_one"})
	if err == nil || !strings.Contains(err.Error(), "old object's signing secret") || strings.Contains(err.Error(), "whsec_one") {
		t.Errorf("a replacement that kept the old secret must fail without naming it, got %v", err)
	}
	if err := h.VerifyDeployed(second, "stripewebhookendpoint", map[string]interface{}{"id": "we_new", "secret": "whsec_two"}); err != nil {
		t.Errorf("a replacement with a new secret over a deleted endpoint passes, got %v", err)
	}
}

// The out-of-band act deletes through Stripe's API and confirms it took, the recovery's new
// object must come with a new secret, and a kind Stripe never deletes is refused.
func TestHarness_DeletesOutOfBand(t *testing.T) {
	f := newFake()
	h, stop := f.serve(t)
	defer stop()
	f.put("/v1/webhook_endpoints/we_1", `{"id":"we_1"}`)
	ctx := context.WithValue(context.Background(), provider.ManifestPathKey{}, "/scenarios/deleted-outside-planton.yaml")
	tc := &provider.KindTestContext{Kind: "stripewebhookendpoint", Provider: "stripe"}

	if err := h.VerifyDeployed(ctx, "stripewebhookendpoint", map[string]interface{}{"id": "we_1", "secret": "whsec_one"}); err != nil {
		t.Fatal(err)
	}
	if err := h.DeleteOutOfBand(ctx, tc); err != nil {
		t.Fatalf("the out-of-band delete must succeed, got %v", err)
	}
	if len(f.deleted) != 1 || f.deleted[0] != "/v1/webhook_endpoints/we_1" {
		t.Fatalf("deleted %v, want exactly the deployed endpoint", f.deleted)
	}

	f.put("/v1/webhook_endpoints/we_2", `{"id":"we_2"}`)
	if err := h.VerifyDeployed(ctx, "stripewebhookendpoint", map[string]interface{}{"id": "we_2", "secret": "whsec_one"}); err == nil || !strings.Contains(err.Error(), "recreated") {
		t.Errorf("a recreated endpoint reporting the old secret must fail, got %v", err)
	}
	if err := h.VerifyDeployed(ctx, "stripewebhookendpoint", map[string]interface{}{"id": "we_2", "secret": "whsec_two"}); err != nil {
		t.Errorf("a recreated endpoint with a new secret passes, got %v", err)
	}

	f.put("/v1/prices/price_1", `{"id":"price_1","active":true}`)
	if err := h.VerifyDeployed(ctx, "stripeprice", map[string]interface{}{"id": "price_1"}); err != nil {
		t.Fatal(err)
	}
	if err := h.DeleteOutOfBand(ctx, &provider.KindTestContext{Kind: "stripeprice"}); err == nil || !strings.Contains(err.Error(), "does not apply") {
		t.Errorf("a kind Stripe only deactivates must be refused, got %v", err)
	}
}

// After the import round trip the state holds no signing secret, so a second act updated in place
// may report none; it must still never report a different one.
func TestHarness_InPlaceAfterImportHasNoSecret(t *testing.T) {
	f := newFake()
	h, stop := f.serve(t)
	defer stop()
	f.put("/v1/webhook_endpoints/we_1", `{"id":"we_1"}`)
	dir := t.TempDir()
	firstPath, secondPath := filepath.Join(dir, "minimal.yaml"), filepath.Join(dir, "minimal.to.yaml")
	body := "metadata:\n  annotations:\n    planton.dev/e2e-second-act: minimal.yaml\n    " + provider.ExpectUpgradeAnnotation + ": in-place\n"
	for _, p := range []string{firstPath, secondPath} {
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	first := context.WithValue(context.Background(), provider.ManifestPathKey{}, firstPath)
	second := context.WithValue(context.WithValue(context.Background(), provider.ManifestPathKey{}, secondPath), provider.FirstActManifestPathKey{}, firstPath)

	if err := h.VerifyDeployed(first, "stripewebhookendpoint", map[string]interface{}{"id": "we_1", "secret": "whsec_one"}); err != nil {
		t.Fatal(err)
	}
	if err := h.VerifyDeployed(second, "stripewebhookendpoint", map[string]interface{}{"id": "we_1"}); err != nil {
		t.Errorf("an in-place second act over an imported state reports no secret, which passes; got %v", err)
	}
	if err := h.VerifyDeployed(second, "stripewebhookendpoint", map[string]interface{}{"id": "we_1", "secret": "whsec_two"}); err == nil || !strings.Contains(err.Error(), "signing secret changed") {
		t.Errorf("an in-place second act reporting a different secret must fail, got %v", err)
	}
}
