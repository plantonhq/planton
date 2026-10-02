package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// bigQueryReservationGet reads a BigQuery Reservation API resource
// (reservation, assignment list, capacity commitment, reservation group) by
// its full name from the https://bigqueryreservation.googleapis.com/v1/
// endpoint (the pinned client library carries no Reservation client).
func bigQueryReservationGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "bigquery reservation resource",
		fmt.Sprintf("https://bigqueryreservation.googleapis.com/v1/%s", name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// bigQueryReservationAbsent is the destroy verdict the Reservation API
// verifiers share.
func bigQueryReservationAbsent(ctx context.Context, svc *Services, what, name string) error {
	if name == "" {
		return nil
	}
	_, status, err := bigQueryReservationGet(ctx, svc, name)
	return restAbsent(what, name, status, err)
}

// bigQueryReservationVerifier probes the reservation: it reads back under
// its name with the platform attribution label, and every exported
// assignment is in the reservation's assignment list (the API has no GET
// for a single assignment).
type bigQueryReservationVerifier struct{}

func (v *bigQueryReservationVerifier) IDOutputKey() string { return "name" }

func (v *bigQueryReservationVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	reservation, _, err := bigQueryReservationGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "bigquery reservation %s not found after deploy", name)
	}
	labels, _ := reservation["labels"].(map[string]interface{})
	if labels["planton-ai_resource"] != "true" {
		return errors.Errorf("bigquery reservation %s missing the planton-ai_resource attribution label after deploy", name)
	}

	expected := listOutput(outputs, "assignment_names")
	if len(expected) == 0 {
		return nil
	}
	list, _, err := bigQueryReservationGet(ctx, svc, name+"/assignments")
	if err != nil {
		return errors.Wrapf(err, "failed to list the assignments of bigquery reservation %s", name)
	}
	live := map[string]bool{}
	items, _ := list["assignments"].([]interface{})
	for _, item := range items {
		if assignment, ok := item.(map[string]interface{}); ok {
			live[stringField(assignment, "name")] = true
		}
	}
	for _, assignment := range expected {
		if !live[assignment] {
			return errors.Errorf("bigquery reservation %s is missing assignment %s after deploy", name, assignment)
		}
	}
	return nil
}

func (v *bigQueryReservationVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return bigQueryReservationAbsent(ctx, svc, "bigquery reservation", outputs["name"])
}

// bigQueryCapacityCommitmentVerifier probes the commitment under its name.
// It is registered for completeness; the kind's profile keeps it out of the
// live lane (a commitment is a multi-year purchase).
type bigQueryCapacityCommitmentVerifier struct{}

func (v *bigQueryCapacityCommitmentVerifier) IDOutputKey() string { return "name" }

func (v *bigQueryCapacityCommitmentVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	if _, _, err := bigQueryReservationGet(ctx, svc, name); err != nil {
		return errors.Wrapf(err, "bigquery capacity commitment %s not found after deploy", name)
	}
	return nil
}

func (v *bigQueryCapacityCommitmentVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return bigQueryReservationAbsent(ctx, svc, "bigquery capacity commitment", outputs["name"])
}

// bigQueryReservationGroupVerifier probes the group under its name.
type bigQueryReservationGroupVerifier struct{}

func (v *bigQueryReservationGroupVerifier) IDOutputKey() string { return "name" }

func (v *bigQueryReservationGroupVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	if _, _, err := bigQueryReservationGet(ctx, svc, name); err != nil {
		return errors.Wrapf(err, "bigquery reservation group %s not found after deploy", name)
	}
	if got := outputs["reservation_group_name"]; got != lastPathSegment(name) {
		return errors.Errorf("bigquery reservation group %s reservation_group_name output %q does not match its name", name, got)
	}
	return nil
}

func (v *bigQueryReservationGroupVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return bigQueryReservationAbsent(ctx, svc, "bigquery reservation group", outputs["name"])
}
