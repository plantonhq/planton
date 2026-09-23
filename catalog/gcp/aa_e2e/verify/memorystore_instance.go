package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// memorystoreInstanceVerifier probes a Memorystore (Valkey) instance via the
// Memorystore REST API. The pinned google.golang.org/api line predates the
// typed memorystore client, so the probe is a plain authenticated GET on the
// instance's documented resource path — existence, ACTIVE state, and the
// discovery-endpoint posture are asserted from the JSON body.
type memorystoreInstanceVerifier struct{}

func (v *memorystoreInstanceVerifier) IDOutputKey() string { return "name" }

// memorystoreInstance is the subset of the API's Instance object the
// verifier asserts on.
type memorystoreInstance struct {
	Name      string `json:"name"`
	State     string `json:"state"`
	Uid       string `json:"uid"`
	Endpoints []struct {
		Connections []struct {
			PscAutoConnection struct {
				IpAddress string `json:"ipAddress"`
			} `json:"pscAutoConnection"`
		} `json:"connections"`
	} `json:"endpoints"`
}

func (v *memorystoreInstanceVerifier) get(ctx context.Context, svc *Services, name string) (*memorystoreInstance, int, error) {
	instance := &memorystoreInstance{}
	status, err := googleRestGet(ctx, svc, "memorystore instance", fmt.Sprintf("https://memorystore.googleapis.com/v1/%s", name), instance)
	if err != nil {
		return nil, status, err
	}
	return instance, status, nil
}

func (v *memorystoreInstanceVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}

	instance, _, err := v.get(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "memorystore instance %s not found after deploy", name)
	}
	if instance.State != "ACTIVE" {
		return errors.Errorf("memorystore instance %s state is %q, want ACTIVE", name, instance.State)
	}
	if uid := outputs["instance_uid"]; uid != "" && instance.Uid != uid {
		return errors.Errorf("memorystore instance %s uid mismatch: output %q, live %q", name, uid, instance.Uid)
	}
	// The discovery address in the outputs must be one of the live PSC
	// endpoint IPs — the address applications will actually connect to.
	if addr := outputs["discovery_address"]; addr != "" {
		found := false
		for _, ep := range instance.Endpoints {
			for _, conn := range ep.Connections {
				if conn.PscAutoConnection.IpAddress == addr {
					found = true
				}
			}
		}
		if !found {
			return errors.Errorf("memorystore instance %s: discovery_address output %q matches no live PSC endpoint", name, addr)
		}
	}
	return nil
}

func (v *memorystoreInstanceVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}

	_, status, err := v.get(ctx, svc, name)
	return restAbsent("memorystore instance", name, status, err)
}
