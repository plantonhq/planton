package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// privatecaGet reads one Certificate Authority Service resource by its full
// name (pools, authorities, templates, certificates) from the global
// https://privateca.googleapis.com/v1/ endpoint. The decoded body is a
// generic map; the status code tells a 404 from any other error.
func privatecaGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "certificate authority service resource",
		fmt.Sprintf("https://privateca.googleapis.com/v1/%s", name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// privatecaReadBack reads a CA Service resource after deploy and checks what
// every one of them carries: the platform attribution label (the
// cross-engine label canary) and an exported id equal to its name's last
// segment.
func privatecaReadBack(ctx context.Context, svc *Services, what string, outputs map[string]string, idKey string) (map[string]interface{}, error) {
	name := outputs["name"]
	if name == "" {
		return nil, errors.New("name output missing after deploy")
	}
	resource, _, err := privatecaGet(ctx, svc, name)
	if err != nil {
		return nil, errors.Wrapf(err, "%s %s not found after deploy", what, name)
	}
	labels, _ := resource["labels"].(map[string]interface{})
	if labels["planton-ai_resource"] != "true" {
		return nil, errors.Errorf("%s %s missing the planton-ai_resource attribution label after deploy", what, name)
	}
	if got := outputs[idKey]; got != lastPathSegment(name) {
		return nil, errors.Errorf("%s %s %s output %q does not match its name", what, name, idKey, got)
	}
	return resource, nil
}

// privatecaAbsent is the destroy verdict for pools and templates: a 404.
func privatecaAbsent(ctx context.Context, svc *Services, what, name string) error {
	if name == "" {
		return nil
	}
	_, status, err := privatecaGet(ctx, svc, name)
	return restAbsent(what, name, status, err)
}

// privatecaPoolVerifier probes the CA pool under its name and checks the
// tier the manifest chose is the one Google holds.
type privatecaPoolVerifier struct{}

func (v *privatecaPoolVerifier) IDOutputKey() string { return "name" }

func (v *privatecaPoolVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	pool, err := privatecaReadBack(ctx, svc, "ca pool", outputs, "ca_pool_id")
	if err != nil {
		return err
	}
	if tier := stringField(pool, "tier"); tier != "ENTERPRISE" && tier != "DEVOPS" {
		return errors.Errorf("ca pool %s has tier %q after deploy", outputs["name"], tier)
	}
	return nil
}

func (v *privatecaPoolVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return privatecaAbsent(ctx, svc, "ca pool", outputs["name"])
}

// privatecaAuthorityVerifier probes the certificate authority: it reads back
// ENABLED (or STAGED, when the manifest asked for it) with a CA certificate
// chain. After destroy it is gone, or DELETED -- skip_grace_period removes
// it at once, but Google may report the DELETED state until the purge runs.
type privatecaAuthorityVerifier struct{}

func (v *privatecaAuthorityVerifier) IDOutputKey() string { return "name" }

func (v *privatecaAuthorityVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	authority, err := privatecaReadBack(ctx, svc, "certificate authority", outputs, "certificate_authority_id")
	if err != nil {
		return err
	}
	if state := stringField(authority, "state"); state != "ENABLED" && state != "STAGED" {
		return errors.Errorf("certificate authority %s is %q after deploy, want ENABLED or STAGED", outputs["name"], state)
	}
	if outputs["pem_ca_certificate"] == "" {
		return errors.Errorf("certificate authority %s exported no pem_ca_certificate", outputs["name"])
	}
	return nil
}

func (v *privatecaAuthorityVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	authority, status, err := privatecaGet(ctx, svc, name)
	if err == nil && stringField(authority, "state") == "DELETED" {
		return nil
	}
	return restAbsent("certificate authority", name, status, err)
}

// privatecaTemplateVerifier probes the certificate template under its name.
type privatecaTemplateVerifier struct{}

func (v *privatecaTemplateVerifier) IDOutputKey() string { return "name" }

func (v *privatecaTemplateVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	_, err := privatecaReadBack(ctx, svc, "certificate template", outputs, "template_id")
	return err
}

func (v *privatecaTemplateVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return privatecaAbsent(ctx, svc, "certificate template", outputs["name"])
}

// privatecaCertificateVerifier probes the issued certificate: it reads back
// unrevoked with a PEM certificate. Destroy revokes rather than deletes, so
// the destroy verdict is revocation details present (or a 404).
type privatecaCertificateVerifier struct{}

func (v *privatecaCertificateVerifier) IDOutputKey() string { return "name" }

func (v *privatecaCertificateVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	certificate, err := privatecaReadBack(ctx, svc, "certificate", outputs, "certificate_id")
	if err != nil {
		return err
	}
	if _, revoked := certificate["revocationDetails"]; revoked {
		return errors.Errorf("certificate %s is revoked right after deploy", outputs["name"])
	}
	if stringField(certificate, "pemCertificate") == "" || outputs["pem_certificate"] == "" {
		return errors.Errorf("certificate %s has no PEM certificate after deploy", outputs["name"])
	}
	return nil
}

func (v *privatecaCertificateVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	certificate, status, err := privatecaGet(ctx, svc, name)
	if err == nil {
		if _, revoked := certificate["revocationDetails"]; revoked {
			return nil
		}
		return errors.Errorf("certificate %s is still unrevoked after destroy", name)
	}
	return restAbsent("certificate", name, status, err)
}
