package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// certManagerIssuanceConfigVerifier probes a Certificate Manager certificate
// issuance config via the certificatemanager API. The issuance_config_id
// output is the full resource name, so the probe consumes it directly.
// Posture assertions: the platform attribution label (the live label-parity
// guard), and that the config names a Certificate Authority Service pool --
// the one thing it exists to carry. Issuing a certificate through the config
// is the certificate's concern, not this probe's.
type certManagerIssuanceConfigVerifier struct{}

func (v *certManagerIssuanceConfigVerifier) IDOutputKey() string { return "issuance_config_id" }

func (v *certManagerIssuanceConfigVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	issuanceConfigID := outputs["issuance_config_id"]
	if issuanceConfigID == "" {
		return errors.New("issuance_config_id output missing after deploy")
	}

	issuanceConfig, err := svc.CertificateManager.Projects.Locations.CertificateIssuanceConfigs.Get(issuanceConfigID).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "certificate issuance config %s not found after deploy", issuanceConfigID)
	}

	// Live label-parity guard: both engines must have applied the platform
	// attribution labels identically.
	if issuanceConfig.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("certificate issuance config %s is missing the planton-ai_resource attribution label after deploy (labels: %v)",
			issuanceConfigID, issuanceConfig.Labels)
	}

	if issuanceConfig.CertificateAuthorityConfig == nil ||
		issuanceConfig.CertificateAuthorityConfig.CertificateAuthorityServiceConfig == nil ||
		issuanceConfig.CertificateAuthorityConfig.CertificateAuthorityServiceConfig.CaPool == "" {
		return errors.Errorf("certificate issuance config %s names no CA pool after deploy", issuanceConfigID)
	}
	return nil
}

func (v *certManagerIssuanceConfigVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	issuanceConfigID := outputs["issuance_config_id"]
	if issuanceConfigID == "" {
		return nil
	}

	_, err := svc.CertificateManager.Projects.Locations.CertificateIssuanceConfigs.Get(issuanceConfigID).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing certificate issuance config %s after destroy", issuanceConfigID)
	}
	return errors.Errorf("certificate issuance config %s still exists after destroy", issuanceConfigID)
}
