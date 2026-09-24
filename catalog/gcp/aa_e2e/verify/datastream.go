package verify

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// datastreamGet reads one Datastream resource by its full name (private
// connections, connection profiles, streams) from the global
// https://datastream.googleapis.com/v1/ endpoint. The decoded body is a
// generic map; the status code tells a 404 from any other error.
func datastreamGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "datastream resource",
		fmt.Sprintf("https://datastream.googleapis.com/v1/%s", name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// datastreamAbsent is the destroy verdict every Datastream verifier shares.
func datastreamAbsent(ctx context.Context, svc *Services, what, name string) error {
	if name == "" {
		return nil
	}
	_, status, err := datastreamGet(ctx, svc, name)
	return restAbsent(what, name, status, err)
}

// datastreamReadBack reads a Datastream resource after deploy and checks
// what every one of them carries: the platform attribution label (the
// cross-engine label canary) and an exported id equal to its name's last
// segment.
func datastreamReadBack(ctx context.Context, svc *Services, what string, outputs map[string]string, idKey string) (map[string]interface{}, error) {
	name := outputs["name"]
	if name == "" {
		return nil, errors.New("name output missing after deploy")
	}
	resource, _, err := datastreamGet(ctx, svc, name)
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

// datastreamPrivateConnectionVerifier probes the private connection: it
// reads back CREATED -- the peering or PSC interface is in place -- under
// its name.
type datastreamPrivateConnectionVerifier struct{}

func (v *datastreamPrivateConnectionVerifier) IDOutputKey() string { return "name" }

func (v *datastreamPrivateConnectionVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	connection, err := datastreamReadBack(ctx, svc, "datastream private connection", outputs, "private_connection_id")
	if err != nil {
		return err
	}
	if state := stringField(connection, "state"); state != "CREATED" {
		return errors.Errorf("datastream private connection %s is %q after deploy, want CREATED", outputs["name"], state)
	}
	return nil
}

func (v *datastreamPrivateConnectionVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return datastreamAbsent(ctx, svc, "datastream private connection", outputs["name"])
}

// datastreamConnectionProfileVerifier probes the connection profile under
// its name.
type datastreamConnectionProfileVerifier struct{}

func (v *datastreamConnectionProfileVerifier) IDOutputKey() string { return "name" }

func (v *datastreamConnectionProfileVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	_, err := datastreamReadBack(ctx, svc, "datastream connection profile", outputs, "connection_profile_id")
	return err
}

func (v *datastreamConnectionProfileVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return datastreamAbsent(ctx, svc, "datastream connection profile", outputs["name"])
}

// datastreamStreamVerifier probes the stream: it reads back under its name
// in a state that is not a failure (the scenarios create it NOT_STARTED, so
// no data moves).
type datastreamStreamVerifier struct{}

func (v *datastreamStreamVerifier) IDOutputKey() string { return "name" }

func (v *datastreamStreamVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	stream, err := datastreamReadBack(ctx, svc, "datastream stream", outputs, "stream_id")
	if err != nil {
		return err
	}
	if state := stringField(stream, "state"); strings.HasPrefix(state, "FAILED") {
		return errors.Errorf("datastream stream %s is %q after deploy", outputs["name"], state)
	}
	return nil
}

func (v *datastreamStreamVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return datastreamAbsent(ctx, svc, "datastream stream", outputs["name"])
}
