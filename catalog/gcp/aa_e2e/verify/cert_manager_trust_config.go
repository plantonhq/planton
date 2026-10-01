package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// certManagerTrustConfigVerifier probes a Certificate Manager trust config
// via the certificatemanager API. The trust_config_id output is the full
// resource name, so the probe consumes it directly. Posture assertions: the
// platform attribution label (the live label-parity guard), and that the
// config holds trust material -- a trust store with at least one anchor, or
// at least one allowlisted certificate -- since a config with neither
// validates nothing.
type certManagerTrustConfigVerifier struct{}

func (v *certManagerTrustConfigVerifier) IDOutputKey() string { return "trust_config_id" }

func (v *certManagerTrustConfigVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	trustConfigID := outputs["trust_config_id"]
	if trustConfigID == "" {
		return errors.New("trust_config_id output missing after deploy")
	}

	trustConfig, err := svc.CertificateManager.Projects.Locations.TrustConfigs.Get(trustConfigID).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "trust config %s not found after deploy", trustConfigID)
	}

	// Live label-parity guard: both engines must have applied the platform
	// attribution labels identically.
	if trustConfig.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("trust config %s is missing the planton-ai_resource attribution label after deploy (labels: %v)",
			trustConfigID, trustConfig.Labels)
	}

	anchors := 0
	for _, store := range trustConfig.TrustStores {
		if store != nil {
			anchors += len(store.TrustAnchors)
		}
	}
	if anchors == 0 && len(trustConfig.AllowlistedCertificates) == 0 {
		return errors.Errorf("trust config %s holds no trust anchor and no allowlisted certificate after deploy", trustConfigID)
	}
	return nil
}

func (v *certManagerTrustConfigVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	trustConfigID := outputs["trust_config_id"]
	if trustConfigID == "" {
		return nil
	}

	_, err := svc.CertificateManager.Projects.Locations.TrustConfigs.Get(trustConfigID).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing trust config %s after destroy", trustConfigID)
	}
	return errors.Errorf("trust config %s still exists after destroy", trustConfigID)
}
