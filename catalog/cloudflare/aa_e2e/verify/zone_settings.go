// The zone-settings singleton verifiers. These kinds manage zone-scoped
// SETTINGS surfaces that always exist while the zone exists: the settings
// object predates the deploy and survives the destroy (the provider's
// destroy is a NO-OP for the class -- measured at v5.23.0: zone_setting,
// waiting_room_settings, the cache toggles, universal_ssl_setting,
// total_tls, zone_auto_origin_tls_kex, certificate_authorities_hostname_
// associations all have literally empty Delete implementations).
//
// The verification contract that follows:
//
//   - VerifyExists asserts the kind's settings surface ANSWERS for the
//     deployed zone -- the deploy targeted a real, readable settings object.
//     Value-level assertion (did always_use_https actually become "on"?)
//     needs per-scenario expectations the harness does not carry; the proof
//     lane owns that design call and records it on the queue entry.
//   - VerifyAbsent asserts the surface STILL ANSWERS after destroy -- the
//     honest no-op-destroy posture. Asserting disappearance would fail every
//     honest run. Zero-orphan cleanup comes from the run-scoped zone fixture:
//     the abandoned settings die with the throwaway zone, so nothing outlives
//     the lane.
//
// One shared struct, not three bespoke ones: unlike the per-kind destroy
// contrasts seen elsewhere in the catalog, these three kinds carry exactly
// the same contract and differ only in which settings endpoint they probe.
package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// settingsSingletonVerifier verifies a scoped settings surface whose object
// always exists while its scope does (no-op destroy class). The scope is a
// zone in the common case; account-scoped singletons (the Zero Trust
// organization, the Gateway configuration) set idKey to "account_id".
type settingsSingletonVerifier struct {
	component string
	// pathFormat is the settings surface's GET path with one %s for the
	// scope id.
	pathFormat string
	// idKey names the output carrying the scope id. Empty means
	// "zone_id" (the original zone-singleton class -- existing
	// registrations stay untouched).
	idKey string
}

// identityKey resolves the scope-id output key, defaulting to zone_id.
func (v *settingsSingletonVerifier) identityKey() string {
	if v.idKey != "" {
		return v.idKey
	}
	return "zone_id"
}

// IDOutputKey: settings singletons have no resource id -- the scope IS the
// identity (the fallback-origin precedent).
func (v *settingsSingletonVerifier) IDOutputKey() string { return v.identityKey() }

func (v *settingsSingletonVerifier) VerifyExists(ctx context.Context, api API, outputs map[string]string) error {
	return v.assertSurfaceAnswers(ctx, api, outputs, "after deploy")
}

// VerifyAbsent asserts the settings surface still answers -- destroy is a
// no-op for this class, so "gone" would be a false expectation. Do not "fix"
// this into an absence check: it would fail every honest run (live-validated
// 2026-08-27 across the zone-settings, cache-settings, and zone-TLS lanes:
// every surface answered after destroy exactly as this contract asserts).
func (v *settingsSingletonVerifier) VerifyAbsent(ctx context.Context, api API, outputs map[string]string) error {
	return v.assertSurfaceAnswers(ctx, api, outputs, "after destroy (settings surfaces persist -- destroy is a no-op for this class)")
}

func (v *settingsSingletonVerifier) assertSurfaceAnswers(ctx context.Context, api API, outputs map[string]string, when string) error {
	scopeID := outputs[v.identityKey()]
	if scopeID == "" {
		return errors.Errorf("%s outputs carry no %s -- cannot verify", v.component, v.identityKey())
	}
	path := fmt.Sprintf(v.pathFormat, scopeID)
	exists, err := api.ResourceExists(ctx, path)
	if err != nil {
		return errors.Wrapf(err, "%s settings-surface probe failed %s", v.component, when)
	}
	if !exists {
		return errors.Errorf("%s settings surface for scope %s does not answer %s (GET %s)",
			v.component, scopeID, when, path)
	}
	return nil
}
