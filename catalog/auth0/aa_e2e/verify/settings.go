package verify

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

// The verifiers of kinds whose destroy deletes nothing. Each states its own
// destroy contract, source-verified against the provider at the pin, because
// the two differ in what "absent" means:
//
//   - Auth0CustomDomainVerification: verification is a one-time action and
//     the provider's delete is a no-op, so after destroy the custom domain is
//     STILL verified (it is removed only when its Auth0CustomDomain is).
//   - Auth0TenantSettings: a tenant always has settings and the provider's
//     delete is a no-op, so after destroy the settings object STILL answers
//     with the last-applied values.
//
// Asserting disappearance would fail every honest run; each verifier asserts
// the state its destroy leaves instead.

// customDomainVerificationVerifier verifies Auth0CustomDomainVerification by
// reading the custom domain it verified: verification succeeded only if Auth0
// reports the domain "ready".
type customDomainVerificationVerifier struct{}

func (*customDomainVerificationVerifier) IDOutput() string { return "custom_domain_id" }

func (*customDomainVerificationVerifier) VerifyExists(checker ResourceChecker, id string) error {
	return requireReadyDomain(checker, id, "after verification")
}

func (*customDomainVerificationVerifier) VerifyAbsent(checker ResourceChecker, id string) error {
	return requireReadyDomain(checker, id, "after the verification's destroy, which must leave the domain verified")
}

func requireReadyDomain(checker ResourceChecker, id, when string) error {
	domain, exists, err := checker.ReadResource(fmt.Sprintf("custom-domains/%s", url.PathEscape(id)))
	if err != nil {
		return errors.Wrap(err, "auth0customdomainverification: reading the custom domain failed")
	}
	if !exists {
		return errors.Errorf("auth0customdomainverification: custom domain %s does not exist %s", id, when)
	}
	if status, _ := domain["status"].(string); status != "ready" {
		return errors.Errorf("auth0customdomainverification: custom domain %s is %q %s, want \"ready\"", id, status, when)
	}
	return nil
}

// tenantSettingsVerifier verifies Auth0TenantSettings: the settings of the
// tenant the credential belongs to answer after deploy, and still answer after
// destroy.
type tenantSettingsVerifier struct{}

func (*tenantSettingsVerifier) IDOutput() string { return "" }

func (*tenantSettingsVerifier) VerifyExists(checker ResourceChecker, _ string) error {
	return requireTenantSettings(checker, "after deploy")
}

func (*tenantSettingsVerifier) VerifyAbsent(checker ResourceChecker, _ string) error {
	return requireTenantSettings(checker, "after destroy, which leaves the last-applied settings in place")
}

func requireTenantSettings(checker ResourceChecker, when string) error {
	exists, err := checker.ResourceExists("tenants/settings")
	if err != nil {
		return errors.Wrap(err, "auth0tenantsettings: reading the tenant settings failed")
	}
	if !exists {
		return errors.Errorf("auth0tenantsettings: the tenant's settings did not answer %s", when)
	}
	return nil
}
