package verify

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/logging/v2"
)

// firstLines truncates multi-line evidence to its first n lines -- enough to
// attribute a failure without flooding the phase output.
func firstLines(out string, n int) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

// plantonRunnerVerifier probes a Planton runner appliance via the Run
// Admin API v2. The appliance is a single-instance Cloud Run service, so
// the posture assertions confirm the service reconciled successfully (the
// Ready terminal condition — Cloud Run gates creation on first-revision
// readiness, so a runner that could not join its control plane never gets
// here) and that the module's scaling pin held (exactly one always-on
// instance; the singleton law — a second instance joining under the same
// runner name would revoke the first's key).
type plantonRunnerVerifier struct{}

func (v *plantonRunnerVerifier) IDOutputKey() string { return "service_short_name" }

// servicePath builds the projects/{p}/locations/{region}/services/{name}
// resource path the run API addresses services by.
func (v *plantonRunnerVerifier) servicePath(svc *Services, outputs map[string]string) (string, error) {
	name := outputs["service_short_name"]
	region := outputs["region"]
	if name == "" || region == "" {
		return "", errors.New("service_short_name or region output missing")
	}
	project := outputs["project_id"]
	if project == "" {
		project = svc.Project
	}
	return fmt.Sprintf("projects/%s/locations/%s/services/%s", project, region, name), nil
}

func (v *plantonRunnerVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	path, err := v.servicePath(svc, outputs)
	if err != nil {
		return errors.Wrap(err, "after deploy")
	}

	service, err := svc.Run.Projects.Locations.Services.Get(path).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "planton runner service %s not found after deploy", path)
	}

	// The Ready terminal condition is the API's own "reconciled and
	// serving" signal — for the runner it additionally proves the join
	// succeeded, because the container exits on a refused join and the
	// first revision would never turn ready.
	if service.TerminalCondition == nil || service.TerminalCondition.State != "CONDITION_SUCCEEDED" {
		state := "<none>"
		if service.TerminalCondition != nil {
			state = service.TerminalCondition.State
		}
		return errors.Errorf("planton runner service %s terminal condition is %q, want CONDITION_SUCCEEDED — a refused join (bad or unreachable token) keeps the first revision from turning ready", path, state)
	}

	// The singleton law, straight off the live template.
	if service.Template == nil || service.Template.Scaling == nil ||
		service.Template.Scaling.MinInstanceCount != 1 || service.Template.Scaling.MaxInstanceCount != 1 {
		return errors.Errorf("planton runner service %s must pin scaling to exactly one instance (a second instance joining under the same runner name would revoke the first's key)", path)
	}

	return nil
}

func (v *plantonRunnerVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	path, err := v.servicePath(svc, outputs)
	if err != nil {
		// Without a path there is nothing left to probe; treat as gone.
		return nil
	}

	_, err = svc.Run.Projects.Locations.Services.Get(path).Context(ctx).Do()
	if err == nil {
		return errors.Errorf("planton runner service %s still exists after destroy", path)
	}
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && apiErr.Code == 404 {
		return nil
	}
	return errors.Wrapf(err, "unexpected error probing planton runner service %s after destroy", path)
}

// VerifyExpectedDeployFailure proves a fake-token deploy failed for EXACTLY
// the designed reason on the one substrate that gates creation on workload
// health. Three assertions, each with the provider's own evidence:
//
//  1. The engine's error is the revision-readiness class (Cloud Run refused
//     to finish creating a service whose first revision never turned ready)
//     -- a quota, permission, or image-pull failure fails here with the full
//     error text.
//  2. The Service OBJECT exists with the full authored template despite the
//     failed create: the token wired as a secret reference (never a plain
//     env value) and scaling pinned to exactly one instance -- proving the
//     module composed everything correctly and only readiness failed.
//  3. Cloud Logging carries the runner's own join-step error ("joining as
//     runner ...") for this service -- emitted only after the process read
//     the token from Secret Manager and REACHED the control plane. A
//     "dialing control-plane" line instead means network, not the token, and
//     the phase fails with that evidence. Log ingestion lags by seconds, so
//     the read polls bounded.
func (v *plantonRunnerVerifier) VerifyExpectedDeployFailure(ctx context.Context, svc *Services, serviceName, region, expectation string, deployErr error) error {
	if expectation != "revision-readiness" {
		return errors.Errorf("unsupported deploy-failure expectation %q for the runner (supported: revision-readiness)", expectation)
	}

	// 1. Classify the engine error: both engines surface Cloud Run's own
	// revision-readiness failure; anything else is NOT the designed failure.
	// The live phrasing (both engines, proven): "The user-provided container
	// failed the configured startup probe checks" -- the runner ran, its join
	// was refused, it exited, and the probe never answered. The Ready-
	// condition phrasings cover other provider versions of the same class.
	errText := deployErr.Error()
	readinessClass := strings.Contains(errText, "startup probe") ||
		strings.Contains(errText, "not ready") ||
		strings.Contains(errText, "Ready") ||
		strings.Contains(errText, "RevisionFailed")
	if !readinessClass {
		return errors.Errorf("the deploy failed, but not with the revision-readiness class this scenario designs for; engine error: %s", firstLines(errText, 8))
	}

	path := fmt.Sprintf("projects/%s/locations/%s/services/%s", svc.Project, region, serviceName)

	// 2. The service object survived the failed create with the authored
	// template intact.
	service, err := svc.Run.Projects.Locations.Services.Get(path).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "the failed create should leave the service object inspectable, but %s is not readable", path)
	}
	if service.Template == nil || service.Template.Scaling == nil ||
		service.Template.Scaling.MinInstanceCount != 1 || service.Template.Scaling.MaxInstanceCount != 1 {
		return errors.Errorf("service %s must pin scaling to exactly one instance even in the failed-create state", path)
	}
	tokenWiredAsSecret := false
	for _, c := range service.Template.Containers {
		for _, env := range c.Env {
			if env.Name == "PLANTON_RUNNER_TOKEN" && env.ValueSource != nil && env.ValueSource.SecretKeyRef != nil {
				tokenWiredAsSecret = true
			}
		}
	}
	if !tokenWiredAsSecret {
		return errors.Errorf("service %s must carry PLANTON_RUNNER_TOKEN as a secret reference (never a plain env value)", path)
	}

	// 3. Pin the cause from the workload's own log line.
	filter := fmt.Sprintf(`resource.type="cloud_run_revision" AND resource.labels.service_name=%q AND resource.labels.location=%q AND timestamp>=%q`,
		serviceName, region, time.Now().Add(-30*time.Minute).UTC().Format(time.RFC3339))
	deadline := time.Now().Add(3 * time.Minute)
	lastState := "no log entries read yet"
	for {
		resp, logErr := svc.Logging.Entries.List(&logging.ListLogEntriesRequest{
			ResourceNames: []string{"projects/" + svc.Project},
			Filter:        filter,
			OrderBy:       "timestamp desc",
			PageSize:      200,
		}).Context(ctx).Do()
		if logErr == nil {
			var dialing bool
			for _, entry := range resp.Entries {
				// The runner's own line for a refused join: "the control
				// plane at <endpoint> did not admit runner <name>: this
				// runner token was not accepted".
				if strings.Contains(entry.TextPayload, "did not admit runner") {
					fmt.Printf("  [verify] CAUSE: deploy failed at revision readiness, the service object carries the full authored template, and Cloud Logging holds the join-step error -- the ONLY failure is the refused join\n")
					return nil
				}
				if strings.Contains(entry.TextPayload, "dialing control-plane") {
					dialing = true
				}
			}
			if dialing {
				return errors.New("the runner failed DIALING the control plane (network), not joining with the token")
			}
			lastState = fmt.Sprintf("%d log entries, none carrying the join line yet", len(resp.Entries))
		} else {
			lastState = logErr.Error()
		}

		if time.Now().After(deadline) {
			return errors.Errorf("Cloud Logging never surfaced the runner's join-step error within the window; last state: %s", lastState)
		}
		time.Sleep(10 * time.Second)
	}
}
