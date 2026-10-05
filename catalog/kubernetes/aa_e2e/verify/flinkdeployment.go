package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// FlinkDeploymentVerifier checks an operator-managed Flink cluster to
// the point a customer's stream is actually flowing: the
// FlinkDeployment CR reports jobManagerDeploymentStatus READY (the
// operator's "running and ready to receive REST API calls" state —
// JobManagerDeploymentStatus.java:24), the naming contract holds
// (`<name>-rest`, the REST Service the outputs publish), and THE REST
// CONTRACT on every lane: GET /config through a port-forward must
// answer with the DECLARED Flink version. A session cluster holds zero
// TaskManagers by design in native mode, so its own REST surface
// answering correctly IS its definition of ready — there is nothing
// else to observe.
//
// Application lanes add THE STREAM PROOF: a job reaching state RUNNING
// — a streaming job's steady state (a batch job would finish; a
// streaming job that is "working" is one that is RUNNING and stays
// there) — plus TaskManagers materialized on demand by the native
// integration. Pods being up proves nothing here: the job graph must
// have been submitted, scheduled, and its tasks deployed onto
// registered TaskManager slots.
//
// The behavioral-recovery lane adds THE RECOVERY PROOF: completed
// checkpoints observed through the REST API (checkpoints landing in
// the composed S3 store IS the S3-seam proof — the counter only moves
// when Flink wrote state to s3:// and acknowledged it), then DELETE
// the JobManager pod, wait for a UID-verified Ready replacement, and
// assert a job returns to RUNNING with checkpoint continuity —
// recovery through HA metadata + checkpoint restore, the exact
// mechanism the state.highAvailability block exists to buy.
type FlinkDeploymentVerifier struct {
	Namespace string
	Name      string
	// SessionMode marks a manifest with no job block — an empty Flink
	// runtime accepting external submissions.
	SessionMode bool
	// RecoveryProof marks the behavioral-recovery lane (recognized by
	// the dispatcher from the manifest path).
	RecoveryProof bool
	// FlinkVersion is the spec enum (e.g. "v2_1") — the REST /config
	// must report a version under the matching prefix ("2.1").
	FlinkVersion string
}

// flinkRestLocalPort is the workstation side of the port-forward to the
// JobManager REST Service (8081).
const flinkRestLocalPort = "18881"

func (v *FlinkDeploymentVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] flinkdeployment %q in namespace %q (session mode %v, flink version %s)\n",
		v.Name, v.Namespace, v.SessionMode, v.FlinkVersion)

	// READY is the operator's own "REST API reachable" gate
	// (isRestApiAvailable() == this == READY in the enum) — everything
	// after this rides the REST surface it vouches for.
	if err := v.waitForJobManagerReady(ctx, kubeconfig, 10*time.Minute); err != nil {
		return err
	}

	// The naming contract the outputs publish.
	restSvc := v.Name + "-rest"
	if err := KubectlResourceExists(ctx, kubeconfig, "service", restSvc, v.Namespace); err != nil {
		return errors.Wrap(err, "the REST service not found")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	base := "http://127.0.0.1:" + flinkRestLocalPort

	cancel, err := openServiceTunnel(ctx, kubeconfig, v.Namespace, restSvc, flinkRestLocalPort, "8081")
	if err != nil {
		return err
	}

	// THE REST CONTRACT: the version the cluster reports is the version
	// the spec declared — a wrong image or a mis-wired service would
	// answer here first.
	if err := v.proveRestConfig(ctx, client, base, 4*time.Minute); err != nil {
		cancel()
		return err
	}

	if v.SessionMode {
		cancel()
		return nil
	}

	// THE STREAM PROOF: exactly one job (the spec's single job block)
	// in RUNNING state.
	jobId, err := v.waitForRunningJob(ctx, client, base, 8*time.Minute, true)
	if err != nil {
		cancel()
		return err
	}
	fmt.Printf("  [verify] THE STREAM PROOF: job %s is RUNNING — the stream reached its steady state\n", jobId)

	// TaskManagers are requested on demand by the native integration —
	// a RUNNING job with zero registered TaskManagers would mean the
	// REST answer lied about task deployment.
	if err := v.assertTaskManagersRegistered(ctx, client, base, 4*time.Minute); err != nil {
		cancel()
		return err
	}

	if !v.RecoveryProof {
		cancel()
		return nil
	}

	// THE RECOVERY PROOF, stage 1: checkpoints completing into the
	// composed S3 store (the scenario sets a 10s interval, so the first
	// completion is quick once the job runs).
	completed, err := v.waitForCompletedCheckpoints(ctx, client, base, jobId, 5*time.Minute)
	if err != nil {
		cancel()
		return err
	}
	fmt.Printf("  [verify] THE RECOVERY PROOF: %d checkpoint(s) completed into the composed S3 store before the kill\n", completed)

	// Stage 2: kill the JobManager. Close the tunnel across the
	// replacement window and reopen fresh after — the old tunnel dies
	// silently with its pod. The JobManager pod is selected by the
	// native-integration labels the operator's own e2e suite uses:
	// `app=<name>,component=jobmanager` (e2e-tests/utils.sh:167; the
	// operator watches JobManager Deployments by the same component
	// label, EventSourceUtils.java:95). The replacement pod carries a
	// new Deployment-generated name, so the selector-keyed
	// delete-and-await is the honest recovery wait.
	cancel()
	jmSelector := "app=" + v.Name + ",component=jobmanager"
	if err := deletePodAwaitReplacement(ctx, kubeconfig, v.Namespace, jmSelector, 8*time.Minute); err != nil {
		return errors.Wrap(err, "the JobManager pod did not recover after deletion")
	}
	fmt.Printf("  [verify] THE RECOVERY PROOF: JobManager replacement is Ready\n")

	cancel, err = openServiceTunnel(ctx, kubeconfig, v.Namespace, restSvc, flinkRestLocalPort, "8081")
	if err != nil {
		return errors.Wrap(err, "re-establishing the port-forward after the JobManager replacement")
	}
	defer cancel()

	// Stage 3: the job must return to RUNNING (the recovered JobManager
	// re-submits it from the HA metadata in S3 — the job id may differ,
	// so re-discover rather than assume).
	newJobId, err := v.waitForRunningJob(ctx, client, base, 8*time.Minute, false)
	if err != nil {
		return errors.Wrap(err, "no job returned to RUNNING after the JobManager replacement")
	}
	fmt.Printf("  [verify] THE RECOVERY PROOF: job %s RUNNING again after the JobManager replacement\n", newJobId)

	// Stage 4: checkpoint continuity, both halves. First the RESTORE
	// truth: the recovered execution must report latest.restored — the
	// coordinator's own record that it initialized FROM a checkpoint in
	// the store (CheckpointingStatistics.LatestCheckpoints, field
	// "restored", nullable — null would mean a fresh boot that merely
	// resembles recovery). Then the WRITE path: fresh checkpoints must
	// complete after the restore.
	restoredId, err := v.waitForRestoredCheckpoint(ctx, client, base, newJobId, 4*time.Minute)
	if err != nil {
		return errors.Wrap(err, "the recovered job never reported a RESTORED checkpoint — it may have rebooted empty instead of restoring")
	}
	fmt.Printf("  [verify] THE RECOVERY PROOF: the recovered job RESTORED checkpoint %d from the composed S3 store\n", restoredId)
	completed, err = v.waitForCompletedCheckpoints(ctx, client, base, newJobId, 5*time.Minute)
	if err != nil {
		return errors.Wrap(err, "no checkpoints completed after the recovery")
	}
	fmt.Printf("  [verify] THE RECOVERY PROOF: %d checkpoint(s) completed after recovery — HA metadata + checkpoint restore held\n", completed)
	return nil
}

// waitForRestoredCheckpoint polls /jobs/<id>/checkpoints until
// latest.restored is non-null, returning the restored checkpoint id.
// The field is the checkpoint coordinator's own restore record
// (CheckpointingStatistics.java: FIELD_NAME_LATEST_CHECKPOINTS
// "latest" → FIELD_NAME_RESTORED "restored", @Nullable
// RestoredCheckpointStatistics with id/restore_timestamp) — the one
// signal that distinguishes a genuine state restore from a job that
// rebooted empty and started checkpointing from scratch.
func (v *FlinkDeploymentVerifier) waitForRestoredCheckpoint(ctx context.Context, client *http.Client, base, jobId string, budget time.Duration) (int64, error) {
	deadline := time.Now().Add(budget)
	for time.Now().Before(deadline) {
		body, err := flinkGET(ctx, client, base+"/jobs/"+jobId+"/checkpoints", 60*time.Second)
		if err == nil {
			var stats struct {
				Latest struct {
					Restored *struct {
						Id int64 `json:"id"`
					} `json:"restored"`
				} `json:"latest"`
			}
			if jsonErr := json.Unmarshal([]byte(body), &stats); jsonErr == nil && stats.Latest.Restored != nil {
				return stats.Latest.Restored.Id, nil
			}
		}
		time.Sleep(10 * time.Second)
	}
	return 0, errors.Errorf("latest.restored never appeared for job %s within %s", jobId, budget)
}

func (v *FlinkDeploymentVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, "flinkdeployments.flink.apache.org", v.Name, v.Namespace); err != nil {
		return err
	}
	if err := KubectlResourceAbsent(ctx, kubeconfig, "service", v.Name+"-rest", v.Namespace); err != nil {
		return errors.Wrap(err, "the REST service never deleted after the CR was removed")
	}
	// Native-integration pods (JobManager AND TaskManagers) all carry
	// app=<name>; the operator tears them down asynchronously.
	return waitForNoPodsBySelector(ctx, kubeconfig, v.Namespace, "app="+v.Name, 3*time.Minute)
}

// waitForJobManagerReady polls the CR until the operator reports
// jobManagerDeploymentStatus READY (JobManagerDeploymentStatus.java:24
// — "JobManager is running and ready to receive REST API calls"),
// printing progress: first boot pays the Flink image pull and, on the
// recovery lane, the S3 filesystem plugin activation.
func (v *FlinkDeploymentVerifier) waitForJobManagerReady(ctx context.Context, kubeconfig string, budget time.Duration) error {
	start := time.Now()
	deadline := start.Add(budget)
	var lastStatus string
	var lastProgress time.Time
	for time.Now().Before(deadline) {
		status, _ := kubectlGetJSONPath(ctx, kubeconfig, "flinkdeployments.flink.apache.org", v.Name, v.Namespace,
			"{.status.jobManagerDeploymentStatus}")
		lastStatus = strings.TrimSpace(status)
		if lastStatus == "READY" {
			fmt.Printf("  [verify] FlinkDeployment CR reports the JobManager READY\n")
			return nil
		}
		if time.Since(lastProgress) >= 30*time.Second {
			// Surface the operator's own diagnosis while waiting — it
			// records reconciliation errors on the status.
			errMsg, _ := kubectlGetJSONPath(ctx, kubeconfig, "flinkdeployments.flink.apache.org", v.Name, v.Namespace,
				"{.status.error}")
			progress := fmt.Sprintf("  [verify] waiting on the FlinkDeployment CR: jobManagerDeploymentStatus %q (elapsed %s)",
				lastStatus, time.Since(start).Round(time.Second))
			if strings.TrimSpace(errMsg) != "" {
				progress += fmt.Sprintf(" — operator error: %s", firstLines(errMsg, 1))
			}
			fmt.Println(progress)
			lastProgress = time.Now()
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("the FlinkDeployment CR never reported the JobManager READY within %s (last status %q)", budget, lastStatus)
}

// proveRestConfig asserts THE REST CONTRACT: /config answers 200 with a
// flink-version under the declared prefix (v2_1 -> "2.1").
func (v *FlinkDeploymentVerifier) proveRestConfig(ctx context.Context, client *http.Client, base string, budget time.Duration) error {
	body, err := flinkGET(ctx, client, base+"/config", budget)
	if err != nil {
		return errors.Wrap(err, "the REST /config never answered")
	}
	var config struct {
		FlinkVersion string `json:"flink-version"`
	}
	if err := json.Unmarshal([]byte(body), &config); err != nil {
		return errors.Wrapf(err, "parsing /config: %s", firstLines(body, 2))
	}
	want := flinkVersionPrefix(v.FlinkVersion)
	if want != "" && !strings.HasPrefix(config.FlinkVersion, want) {
		return errors.Errorf("/config reports flink-version %q, the spec declared %s (prefix %q)",
			config.FlinkVersion, v.FlinkVersion, want)
	}
	fmt.Printf("  [verify] REST CONTRACT: /config answered flink-version %q (declared %s)\n", config.FlinkVersion, v.FlinkVersion)
	return nil
}

// flinkVersionPrefix maps the spec enum to the version prefix the REST
// API must report: v2_1 -> "2.1".
func flinkVersionPrefix(specVersion string) string {
	return strings.ReplaceAll(strings.TrimPrefix(specVersion, "v"), "_", ".")
}

// waitForRunningJob polls /jobs/overview until a job reaches RUNNING,
// returning its id. exactlyOne additionally asserts the overview holds
// a single job — the application spec declares exactly one job block,
// so a second entry would mean a double submission. After a JobManager
// replacement the recovered execution may carry a NEW id, so callers
// re-discover instead of pinning the old one.
func (v *FlinkDeploymentVerifier) waitForRunningJob(ctx context.Context, client *http.Client, base string, budget time.Duration, exactlyOne bool) (string, error) {
	start := time.Now()
	deadline := start.Add(budget)
	var lastSeen string
	var lastProgress time.Time
	for time.Now().Before(deadline) {
		body, err := flinkGET(ctx, client, base+"/jobs/overview", 60*time.Second)
		if err == nil {
			var overview struct {
				Jobs []struct {
					Jid   string `json:"jid"`
					State string `json:"state"`
				} `json:"jobs"`
			}
			if jsonErr := json.Unmarshal([]byte(body), &overview); jsonErr == nil {
				states := make([]string, 0, len(overview.Jobs))
				running := ""
				for _, job := range overview.Jobs {
					states = append(states, job.Jid+"="+job.State)
					if job.State == "RUNNING" {
						running = job.Jid
					}
				}
				lastSeen = strings.Join(states, ", ")
				if running != "" {
					if exactlyOne && len(overview.Jobs) != 1 {
						return "", errors.Errorf("expected exactly one job, the overview holds %d (%s)", len(overview.Jobs), lastSeen)
					}
					return running, nil
				}
			}
		}
		if time.Since(lastProgress) >= 30*time.Second {
			fmt.Printf("  [verify] waiting on the job overview: [%s] (elapsed %s)\n",
				lastSeen, time.Since(start).Round(time.Second))
			lastProgress = time.Now()
		}
		time.Sleep(10 * time.Second)
	}
	return "", errors.Errorf("no job reached RUNNING within %s (last overview: [%s])", budget, lastSeen)
}

// assertTaskManagersRegistered proves the native integration
// materialized workers: /taskmanagers must list at least one registered
// TaskManager.
func (v *FlinkDeploymentVerifier) assertTaskManagersRegistered(ctx context.Context, client *http.Client, base string, budget time.Duration) error {
	deadline := time.Now().Add(budget)
	var lastCount int
	for time.Now().Before(deadline) {
		body, err := flinkGET(ctx, client, base+"/taskmanagers", 60*time.Second)
		if err == nil {
			var listing struct {
				TaskManagers []struct {
					Id string `json:"id"`
				} `json:"taskmanagers"`
			}
			if jsonErr := json.Unmarshal([]byte(body), &listing); jsonErr == nil {
				lastCount = len(listing.TaskManagers)
				if lastCount >= 1 {
					fmt.Printf("  [verify] THE STREAM PROOF: %d TaskManager(s) registered — native mode materialized workers on demand\n", lastCount)
					return nil
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
	return errors.Errorf("no TaskManagers registered within %s (last count %d)", budget, lastCount)
}

// waitForCompletedCheckpoints polls /jobs/<id>/checkpoints until
// counts.completed >= 1, returning the observed count. A completed
// checkpoint is Flink's own acknowledgment that state was WRITTEN to
// the configured store and confirmed — the counter cannot move unless
// the s3:// path composed onto the store actually accepted the bytes.
func (v *FlinkDeploymentVerifier) waitForCompletedCheckpoints(ctx context.Context, client *http.Client, base, jobId string, budget time.Duration) (int, error) {
	start := time.Now()
	deadline := start.Add(budget)
	var lastCompleted int
	var lastProgress time.Time
	for time.Now().Before(deadline) {
		body, err := flinkGET(ctx, client, base+"/jobs/"+jobId+"/checkpoints", 60*time.Second)
		if err == nil {
			var checkpoints struct {
				Counts struct {
					Completed int `json:"completed"`
				} `json:"counts"`
			}
			if jsonErr := json.Unmarshal([]byte(body), &checkpoints); jsonErr == nil {
				lastCompleted = checkpoints.Counts.Completed
				if lastCompleted >= 1 {
					return lastCompleted, nil
				}
			}
		}
		if time.Since(lastProgress) >= 30*time.Second {
			fmt.Printf("  [verify] waiting on checkpoints for job %s: completed %d (elapsed %s)\n",
				jobId, lastCompleted, time.Since(start).Round(time.Second))
			lastProgress = time.Now()
		}
		time.Sleep(10 * time.Second)
	}
	return 0, errors.Errorf("no checkpoint completed for job %s within %s", jobId, budget)
}

// flinkGET performs one GET against the REST API with retries across
// tunnel warm-up, succeeding on 200 only.
func flinkGET(ctx context.Context, client *http.Client, endpoint string, budget time.Duration) (string, error) {
	deadline := time.Now().Add(budget)
	var lastBody string
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return "", err
		}
		resp, err := client.Do(req)
		if err == nil {
			out := drainBody(resp)
			lastBody = out
			if resp.StatusCode == http.StatusOK {
				return out, nil
			}
			lastErr = errors.Errorf("HTTP %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		time.Sleep(5 * time.Second)
	}
	return lastBody, errors.Wrapf(lastErr, "last body: %s", firstLines(lastBody, 2))
}

// flinkDeploymentSessionMode marks a manifest with no job block — a
// SESSION cluster (an empty runtime accepting external submissions).
func flinkDeploymentSessionMode(spec map[string]interface{}) bool {
	_, hasJob := spec["job"].(map[string]interface{})
	return !hasJob
}

// flinkDeploymentVersion reads the declared Flink version enum (e.g.
// "v2_1" — both manifest key forms tolerated) for the REST /config
// contract assertion.
func flinkDeploymentVersion(spec map[string]interface{}) string {
	for _, key := range []string{"flink_version", "flinkVersion"} {
		if version, ok := spec[key].(string); ok && version != "" {
			return version
		}
	}
	return ""
}
