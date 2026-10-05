package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// cloudSqlInstanceVerifier probes a Cloud SQL instance via the sqladmin API.
// Posture assertions distinguish public-only vs private-IP instances from the
// outputs the deploy produced.
type cloudSqlInstanceVerifier struct{}

func (v *cloudSqlInstanceVerifier) IDOutputKey() string { return "instance_name" }

func (v *cloudSqlInstanceVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["instance_name"]
	if name == "" {
		return errors.New("instance_name output missing after deploy")
	}

	inst, err := svc.SqlAdmin.Instances.Get(svc.Project, name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "cloud sql instance %s not found after deploy", name)
	}
	if inst.State != "RUNNABLE" {
		return errors.Errorf("cloud sql instance %s state is %q, want RUNNABLE", name, inst.State)
	}

	wantPublic := outputs["public_ip"] != ""
	wantPrivate := outputs["private_ip"] != ""

	var hasPublic, hasPrivate bool
	for _, ip := range inst.IpAddresses {
		switch ip.Type {
		case "PRIMARY":
			if ip.IpAddress != "" {
				hasPublic = true
			}
		case "PRIVATE":
			hasPrivate = true
		}
	}

	// Assert connectivity posture only when outputs signal intent.
	// Proxy-only public instances (ipv4 on, no authorized networks) still
	// carry a PRIMARY address; an empty public_ip output must not be treated
	// as "private-only".
	if wantPrivate {
		if !hasPrivate {
			return errors.Errorf("cloud sql instance %s: private_ip output set but instance has no PRIVATE address", name)
		}
		if !wantPublic && hasPublic {
			return errors.Errorf("cloud sql instance %s: private-only instance has a PRIMARY public address", name)
		}
	}
	if wantPublic && !hasPublic {
		return errors.Errorf("cloud sql instance %s: public_ip output set but instance has no PRIMARY address", name)
	}

	if conn := outputs["connection_name"]; conn != "" && inst.ConnectionName != conn {
		return errors.Errorf("cloud sql instance %s connection_name mismatch: output %q, live %q", name, conn, inst.ConnectionName)
	}
	return nil
}

func (v *cloudSqlInstanceVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["instance_name"]
	if name == "" {
		return nil
	}

	_, err := svc.SqlAdmin.Instances.Get(svc.Project, name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing cloud sql instance %s after destroy", name)
	}
	return errors.Errorf("cloud sql instance %s still exists after destroy", name)
}
