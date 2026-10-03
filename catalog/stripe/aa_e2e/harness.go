// Package aa_e2e implements the E2E provider harness for Stripe, using Stripe's REST API to
// verify what the modules created, changed and left behind.
package aa_e2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/stripe/aa_e2e/verify"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	"github.com/plantonhq/planton/pkg/iac/provider/stripe/stripekey"
)

const (
	// envAPIKey is the dedicated test sandbox's key, the same variable the provider reads.
	envAPIKey = "STRIPE_API_KEY"
	// envAccount optionally names a Connect account to act on, as the provider reads it.
	envAccount = "STRIPE_ACCOUNT"
)

// Harness manages the Stripe E2E lifecycle. Stripe is a SaaS API: Setup proves the key belongs
// to a test sandbox and authenticates, and Teardown has nothing to remove.
//
// The key check is the harness's own. A lane passes the modules no provider config -- OpenTofu
// reads STRIPE_API_KEY from the environment -- so the credential loader's mode guard never runs
// here, and a live key would otherwise apply every scenario to a live account.
type Harness struct {
	client *Client

	// mu guards the deployed maps, written by VerifyDeployed and read by VerifyDestroyed, the
	// upgrade's judgment and the out-of-band act.
	mu               sync.Mutex
	deployedIDs      map[string]string            // manifest path + kind -> Stripe id
	deployedChildren map[string]map[string]string // manifest path + kind -> child key -> Stripe id
	// deployedSecrets holds a SHA-256 of each signing secret, never the secret, so a recreated or
	// replaced object can be shown to have come with a new one.
	deployedSecrets map[string]string // manifest path + kind -> hex digest
}

// NewHarness creates a Stripe harness; credentials are read in Setup.
func NewHarness() *Harness {
	return &Harness{
		deployedIDs:      make(map[string]string),
		deployedChildren: make(map[string]map[string]string),
		deployedSecrets:  make(map[string]string),
	}
}

// Setup refuses any key that is not a test-mode key, before the first API call, then proves the
// key authenticates. A refusal fails TestMain, so no lane runs.
func (h *Harness) Setup(ctx context.Context) error {
	apiKey := os.Getenv(envAPIKey)
	if apiKey == "" {
		return errors.Errorf("%s must be set to the dedicated test sandbox's key", envAPIKey)
	}
	if err := stripekey.RequireTestMode(apiKey); err != nil {
		return err
	}

	client := NewClient(apiKey, os.Getenv(envAccount))
	if err := client.VerifyConnectivity(); err != nil {
		return err
	}
	h.client = client
	fmt.Println("  [stripe] Authenticated with the test sandbox")
	return nil
}

// Teardown is a no-op: the harness creates nothing of its own.
func (h *Harness) Teardown(ctx context.Context) error {
	return nil
}

// VerifyDeployed checks the object the module created, reading its id from the outputs, and
// stores the id for VerifyDestroyed. A kind that folds children (a meter's alerts, a product's
// feature links, a Radar list's items) reports their ids as a map output, and each child is
// checked and stored too.
func (h *Harness) VerifyDeployed(ctx context.Context, kindDir string, outputs map[string]interface{}) error {
	v, err := verify.GetVerifier(kindDir)
	if err != nil {
		return err
	}
	id, _ := outputs["id"].(string)
	if id == "" {
		return errors.Errorf("no id found in the outputs of %s", kindDir)
	}
	var children map[string]string
	cv, folds := v.(verify.ChildVerifier)
	if folds {
		if children, err = childIDs(outputs[cv.ChildOutput()]); err != nil {
			return errors.Wrapf(err, "%s: output %s", kindDir, cv.ChildOutput())
		}
	}

	// A second act declared in place may follow the import round trip, whose state carries no
	// signing secret (Stripe returns it only at creation), so only an object this deploy created
	// must report one.
	firstAct, _ := ctx.Value(provider.FirstActManifestPathKey{}).(string)
	createdHere := firstAct == "" || upgradeExpectation(ctx) != provider.UpgradeInPlace
	secretDigest, err := secretDigestOf(kindDir, v, outputs, createdHere)
	if err != nil {
		return err
	}

	key := kindKey(ctx, kindDir)
	h.mu.Lock()
	previousID, previousSecret := h.deployedIDs[key], h.deployedSecrets[key]
	h.deployedIDs[key] = id
	h.deployedChildren[key] = children
	h.deployedSecrets[key] = secretDigest
	h.mu.Unlock()

	if err := v.VerifyExists(h.client, id); err != nil {
		return err
	}
	if folds {
		if err := cv.VerifyChildrenExist(h.client, id, children); err != nil {
			return err
		}
	}
	// The same manifest deployed again with a new object is a recreation (the out-of-band act's
	// recovery): a signing secret must be new with it.
	if previousID != "" && previousID != id && secretDigest != "" && secretDigest == previousSecret {
		return errors.Errorf("%s: %s was recreated as %s but reports the old object's signing secret", kindDir, previousID, id)
	}
	if firstAct != "" {
		return h.verifyUpgrade(ctx, v, kindDir, firstAct, id, secretDigest)
	}
	return nil
}

// upgradeExpectation reads the second act's declared expectation from the manifest on the
// context ("" when there is none, which verifyUpgrade refuses).
func upgradeExpectation(ctx context.Context) string {
	secondAct, _ := ctx.Value(provider.ManifestPathKey{}).(string)
	expectation, _ := runner.ManifestAnnotation(secondAct, provider.ExpectUpgradeAnnotation)
	return expectation
}

// verifyUpgrade judges a second act against the first: the second manifest declares whether the
// upgrade kept the object (in-place) or replaced it, and only Stripe's ids can tell. A replaced
// object must meet its kind's delete truth -- an old price archived, an old endpoint gone, an old
// tax registration still collecting -- and a signing secret changes with the object and only
// with it. A second act that declares nothing is refused, never passed unjudged.
func (h *Harness) verifyUpgrade(ctx context.Context, v verify.Verifier, kindDir, firstAct, id, secretDigest string) error {
	expectation := upgradeExpectation(ctx)
	firstKey := firstAct + "::" + kindDir
	h.mu.Lock()
	firstID, firstSecret := h.deployedIDs[firstKey], h.deployedSecrets[firstKey]
	h.mu.Unlock()
	if firstID == "" {
		return errors.Errorf("%s: no id stored for the first act (%s)", kindDir, firstAct)
	}

	switch expectation {
	case provider.UpgradeInPlace:
		if id != firstID {
			return errors.Errorf("%s: the upgrade is declared %s, but Stripe now holds %s where the first act created %s: the change replaced it",
				kindDir, provider.UpgradeInPlace, id, firstID)
		}
		// An empty secret here is the imported state's (the round trip ran before this act).
		if secretDigest != "" && secretDigest != firstSecret {
			return errors.Errorf("%s: the object was updated in place, but its signing secret changed", kindDir)
		}
		return nil
	case provider.UpgradeReplaced:
		if id == firstID {
			return errors.Errorf("%s: the upgrade is declared %s, but %s kept its id: the change was applied in place", kindDir, provider.UpgradeReplaced, id)
		}
		if secretDigest != "" && secretDigest == firstSecret {
			return errors.Errorf("%s: %s replaced %s but reports the old object's signing secret", kindDir, id, firstID)
		}
		return errors.Wrapf(v.VerifyDestroyed(h.client, firstID), "%s: the object the upgrade replaced (%s)", kindDir, firstID)
	default:
		return errors.Errorf("%s: the second act must declare %s: %s or %s (got %q)",
			kindDir, provider.ExpectUpgradeAnnotation, provider.UpgradeInPlace, provider.UpgradeReplaced, expectation)
	}
}

// DeleteOutOfBand deletes the deployed object through Stripe's API, the way a person in the
// Dashboard would, for the out-of-band act. Only a kind Stripe deletes outright qualifies.
func (h *Harness) DeleteOutOfBand(ctx context.Context, tc *provider.KindTestContext) error {
	v, err := verify.GetVerifier(tc.Kind)
	if err != nil {
		return err
	}
	d, ok := v.(verify.OutOfBandDeletable)
	if !ok {
		return errors.Errorf("%s: Stripe keeps this object after its destroy (deactivated or forgotten), so the out-of-band act does not apply", tc.Kind)
	}
	key := kindKey(ctx, tc.Kind)
	h.mu.Lock()
	id := h.deployedIDs[key]
	h.mu.Unlock()
	if id == "" {
		return errors.Errorf("no stored id for %s -- VerifyDeployed may not have run", tc.Kind)
	}
	if err := d.DeleteOutOfBand(h.client, id); err != nil {
		return err
	}
	fmt.Printf("  [stripe] deleted %s outside Planton\n", id)
	return nil
}

// secretDigestOf checks a kind's signing secret, when it reports one, and returns its SHA-256 in
// hex ("" when there is none). A secret is required only of an object this deploy created
// (createdHere); any reported secret must be one. The secret itself never leaves this function.
func secretDigestOf(kindDir string, v verify.Verifier, outputs map[string]interface{}, createdHere bool) (string, error) {
	sv, ok := v.(verify.SecretVerifier)
	if !ok || sv.SecretOutput() == "" {
		return "", nil
	}
	secret, _ := outputs[sv.SecretOutput()].(string)
	if secret != "" || createdHere {
		if err := verify.CheckSecretShape(kindDir, sv, secret); err != nil {
			return "", err
		}
	}
	if secret == "" {
		return "", nil
	}
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:]), nil
}

// VerifyDestroyed checks what destroy left, which differs by kind: a deleted endpoint is gone,
// while a deactivated configuration is still there, inactive.
func (h *Harness) VerifyDestroyed(ctx context.Context, kindDir string) error {
	v, err := verify.GetVerifier(kindDir)
	if err != nil {
		return err
	}
	key := kindKey(ctx, kindDir)
	h.mu.Lock()
	id, stored := h.deployedIDs[key]
	children := h.deployedChildren[key]
	h.mu.Unlock()
	if !stored {
		return errors.Errorf("no stored id for %s -- VerifyDeployed may not have run", kindDir)
	}
	if err := v.VerifyDestroyed(h.client, id); err != nil {
		return err
	}
	if cv, folds := v.(verify.ChildVerifier); folds {
		return cv.VerifyChildrenDestroyed(h.client, id, children)
	}
	return nil
}

// childIDs reads a map output of child ids. A kind that declared no children reports an empty
// map, or no output at all.
func childIDs(output interface{}) (map[string]string, error) {
	ids := map[string]string{}
	switch m := output.(type) {
	case nil:
	case map[string]string:
		for k, id := range m {
			ids[k] = id
		}
	case map[string]interface{}:
		for k, raw := range m {
			id, ok := raw.(string)
			if !ok || id == "" {
				return nil, errors.Errorf("child %q has no id (got %v)", k, raw)
			}
			ids[k] = id
		}
	default:
		return nil, errors.Errorf("want a map of child ids, got %T", output)
	}
	return ids, nil
}

// kindKey combines the manifest path from the context with the kind, so scenarios of
// one kind running side by side never share an id.
func kindKey(ctx context.Context, kindDir string) string {
	if mp, ok := ctx.Value(provider.ManifestPathKey{}).(string); ok && mp != "" {
		return mp + "::" + kindDir
	}
	return kindDir
}
