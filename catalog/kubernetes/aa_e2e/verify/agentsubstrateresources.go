package verify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/kubernetes/aa_e2e/verify/internal/ateapipb"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The resource kinds that run on an Agent Substrate control plane: worker
// pools, sandbox configurations and CSI driver configurations. Each lane runs
// behind the control plane the runner deployed from its published prerequisite
// profile, and every proof that needs the API reaches it through that control
// plane's outputs (agentsubstrateapi.go), never through restated names.
//
//   - Every lane reads every field the scenario sets back from the live object
//     (the API server prunes what its definition does not serve), and the
//     cluster's admission of the object proves its upstream rules.
//   - A WorkerPool is proved serving: its workers roll out, its class
//     resolves to the expected SandboxConfig, and every worker registers with
//     the API in that class. The actor lane then proves the promise Substrate
//     exists for: an actor created from a template that selects the pool,
//     booted on the runtime of the SandboxConfig the pool resolves, reached
//     through the router over mutual TLS, suspended into the control plane's
//     snapshot location, and resumed with its files.
const (
	agentSubstrateWorkerPoolResource    = "workerpools.ate.dev"
	agentSubstrateSandboxConfigResource = "sandboxconfigs.ate.dev"
	agentSubstrateCSIDriverResource     = "csidriverconfigs.ate.dev"

	// The lanes' names in the API. Every run creates them in an atespace of
	// its own (laneAtespace), so a run never meets objects an interrupted run
	// left behind, and creating any of them is expected to succeed outright.
	agentSubstrateActorTemplate = "e2e-counter"
	agentSubstrateActorName     = "e2e-counter-1"

	// The control plane reads its snapshot-store key from the Secret the
	// KubernetesSecret in front of it publishes (its secret_name and
	// secret_namespace outputs), under the two keys its S3 credentials
	// contract names. The actor lane lists the snapshot with the very key the
	// control plane writes it with.
	agentSubstrateStoreAccessKeyID   = "access_key_id"
	agentSubstrateStoreSecretKey     = "secret_access_key"
	agentSubstrateStoreKind          = "KubernetesSeaweedFs"
	agentSubstrateStoreSecretKind    = "KubernetesSecret"
	agentSubstrateReleaseImagesIndex = "https://downloads.planton.ai/releases/agent-substrate/%s/images.json"
)

// The counter workload, upstream's own demo (demos/counter at the pinned
// release): an HTTP server on port 80 whose every request bumps a counter in
// memory and a counter in a file under its durable directory, and answers
// with both. Under VOLUMES fidelity a suspend keeps the durable directory and
// drops memory, so after a resume the file counter continues and the memory
// counter starts over: the two numbers together prove the actor really left
// memory and really came back from storage. Planton publishes the image with
// the release (images.json, key "counter").
const (
	counterImageKey        = "counter"
	counterCommand         = "/ko-app/counter"
	counterPort            = 80
	counterDataMount       = "/home/counter"
	counterMemoryPrefix    = "preserved memory count: "
	counterFilePrefix      = "preserved file counter: "
	counterActorMillicores = 1000
	counterActorMemory     = 512 << 20
	agentSubstrateHeader   = "ate-target-actor"
)

// ---------------------------------------------------------------------------
// SandboxConfig
// ---------------------------------------------------------------------------

// AgentSubstrateSandboxConfigVerifier proves a SandboxConfig is what the
// manifest declares. Nothing in the API names a config: a worker pool
// resolves it, which the actor lane proves end to end.
type AgentSubstrateSandboxConfigVerifier struct {
	ManifestPath string
	Name         string
	Class        string
}

func newAgentSubstrateSandboxConfigVerifier(manifestPath, name string) *AgentSubstrateSandboxConfigVerifier {
	spec := manifestSpecMap(manifestPath)
	return &AgentSubstrateSandboxConfigVerifier{
		ManifestPath: manifestPath,
		Name:         name,
		Class:        stringOr(specField(spec, "sandboxClass"), "gvisor"),
	}
}

func (v *AgentSubstrateSandboxConfigVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] Agent Substrate SandboxConfig %s (class %s)\n", v.Name, v.Class)
	if err := KubectlResourceExists(ctx, kubeconfig, agentSubstrateSandboxConfigResource, v.Name, ""); err != nil {
		return err
	}
	return proveProjectionReadback(ctx, kubeconfig, agentSubstrateSandboxConfigResource, v.Name, "", v.ManifestPath)
}

func (v *AgentSubstrateSandboxConfigVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, agentSubstrateSandboxConfigResource, v.Name, "")
}

// ---------------------------------------------------------------------------
// CSIDriverConfig
// ---------------------------------------------------------------------------

// AgentSubstrateCSIDriverConfigVerifier proves a CSIDriverConfig is what the
// manifest declares. Nothing in the API reports a driver config's health, and
// a lane with no CSI driver on the cluster cannot prove more than the object.
type AgentSubstrateCSIDriverConfigVerifier struct {
	ManifestPath string
	Name         string
}

func (v *AgentSubstrateCSIDriverConfigVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] Agent Substrate CSIDriverConfig %s\n", v.Name)
	if err := KubectlResourceExists(ctx, kubeconfig, agentSubstrateCSIDriverResource, v.Name, ""); err != nil {
		return err
	}
	return proveProjectionReadback(ctx, kubeconfig, agentSubstrateCSIDriverResource, v.Name, "", v.ManifestPath)
}

func (v *AgentSubstrateCSIDriverConfigVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	return KubectlResourceAbsent(ctx, kubeconfig, agentSubstrateCSIDriverResource, v.Name, "")
}

// ---------------------------------------------------------------------------
// WorkerPool
// ---------------------------------------------------------------------------

// AgentSubstrateWorkerPoolVerifier proves a WorkerPool serves: its workers
// roll out and register, its class resolves to the expected config, and, on
// the actor lane, an actor runs on it, sleeps into storage and comes back.
type AgentSubstrateWorkerPoolVerifier struct {
	ManifestPath string
	Name         string
	Namespace    string
	Replicas     int64
	Class        string
	// ConfigRef is the SandboxConfig the manifest names, empty when the pool
	// takes its class default.
	ConfigRef string
	// Labels are the pool's own labels, which an actor template selects.
	Labels map[string]string
	// ActorLifecycle marks the actor lane (behavioral-actor-lifecycle).
	ActorLifecycle bool
}

func newAgentSubstrateWorkerPoolVerifier(manifestPath, name, namespace string) *AgentSubstrateWorkerPoolVerifier {
	spec := manifestSpecMap(manifestPath)
	v := &AgentSubstrateWorkerPoolVerifier{
		ManifestPath:   manifestPath,
		Name:           name,
		Namespace:      namespace,
		Replicas:       intOr(specField(spec, "replicas"), 0),
		Class:          "gvisor",
		Labels:         map[string]string{},
		ActorLifecycle: strings.Contains(manifestPath, "behavioral-actor-lifecycle"),
	}
	if classes, ok := specField(spec, "sandboxClasses").([]interface{}); ok && len(classes) > 0 {
		if class, ok := classes[0].(map[string]interface{}); ok {
			v.Class = stringOr(specField(class, "name"), "gvisor")
			if ref, ok := specField(class, "configRef").(map[string]interface{}); ok {
				v.ConfigRef = stringOr(referenceValue(specField(ref, "name")), "")
			}
		}
	}
	if labels, ok := specField(spec, "labels").(map[string]interface{}); ok {
		for k, val := range labels {
			v.Labels[k] = fmt.Sprint(val)
		}
	}
	return v
}

// referenceValue reads a resolved reference field: the runner replaces a
// valueFrom with the literal before verification, so the manifest carries
// either a bare string or {value: ...}.
func referenceValue(field interface{}) interface{} {
	if m, ok := field.(map[string]interface{}); ok {
		return m["value"]
	}
	return field
}

func (v *AgentSubstrateWorkerPoolVerifier) VerifyExists(ctx context.Context, kubeconfig string) error {
	fmt.Printf("  [verify] Agent Substrate WorkerPool %s/%s (%d workers, class %s, actor lane=%t)\n", v.Namespace, v.Name, v.Replicas, v.Class, v.ActorLifecycle)
	if err := KubectlResourceExists(ctx, kubeconfig, agentSubstrateWorkerPoolResource, v.Name, v.Namespace); err != nil {
		return err
	}
	if err := proveProjectionReadback(ctx, kubeconfig, agentSubstrateWorkerPoolResource, v.Name, v.Namespace, v.ManifestPath); err != nil {
		return err
	}
	config, err := v.awaitResolved(ctx, kubeconfig)
	if err != nil {
		return err
	}
	if err := v.awaitWorkersReady(ctx, kubeconfig); err != nil {
		return err
	}

	outputs, err := agentSubstrateOutputsFromPrerequisite(ctx)
	if err != nil {
		return err
	}
	if err := applySubstrateProbe(ctx, kubeconfig, outputs, false); err != nil {
		return err
	}
	conn, err := dialSubstrateAPI(ctx, kubeconfig, outputs)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := v.awaitRegistered(ctx, conn); err != nil {
		return err
	}
	if !v.ActorLifecycle {
		return nil
	}
	return v.proveActorLifecycle(ctx, kubeconfig, conn, outputs, config)
}

// awaitResolved waits for the pool's class to resolve and returns the
// SandboxConfig it runs: the one the manifest names, or the class default.
func (v *AgentSubstrateWorkerPoolVerifier) awaitResolved(ctx context.Context, kubeconfig string) (string, error) {
	deadline := time.Now().Add(3 * time.Minute)
	var resolved, config string
	for {
		raw, err := kubectlGetJSONPath(ctx, kubeconfig, agentSubstrateWorkerPoolResource, v.Name, v.Namespace,
			`{.status.sandboxClasses[0].conditions[?(@.type=="Resolved")].status}|{.status.sandboxClasses[0].configRef.name}`)
		if err == nil {
			resolved, config, _ = strings.Cut(strings.TrimSpace(raw), "|")
			if resolved == "True" && config != "" {
				break
			}
		}
		if time.Now().After(deadline) {
			return "", errors.Errorf("WorkerPool %s/%s never resolved its %s class (Resolved=%q, config %q)", v.Namespace, v.Name, v.Class, resolved, config)
		}
		time.Sleep(5 * time.Second)
	}
	if v.ConfigRef != "" && config != v.ConfigRef {
		return "", errors.Errorf("WorkerPool %s/%s runs SandboxConfig %s, but the manifest names %s", v.Namespace, v.Name, config, v.ConfigRef)
	}
	fmt.Printf("  [verify] the %s class resolved to SandboxConfig %s\n", v.Class, config)
	return config, nil
}

// awaitWorkersReady waits for the worker Deployment the controller renders
// under the pool's name to have every worker ready.
func (v *AgentSubstrateWorkerPoolVerifier) awaitWorkersReady(ctx context.Context, kubeconfig string) error {
	if v.Replicas == 0 {
		return nil
	}
	deadline := time.Now().Add(10 * time.Minute)
	var ready string
	for {
		out, err := kubectlGetJSONPath(ctx, kubeconfig, agentSubstrateWorkerPoolResource, v.Name, v.Namespace, "{.status.readyReplicas}")
		if err == nil {
			ready = strings.TrimSpace(out)
			if ready == fmt.Sprint(v.Replicas) {
				fmt.Printf("  [verify] %d workers ready\n", v.Replicas)
				return nil
			}
		}
		if time.Now().After(deadline) {
			return errors.Errorf("WorkerPool %s/%s has %q of %d workers ready", v.Namespace, v.Name, ready, v.Replicas)
		}
		time.Sleep(10 * time.Second)
	}
}

// awaitRegistered waits until the API lists every worker of the pool as
// active: each worker registered with its node agent and the API server.
func (v *AgentSubstrateWorkerPoolVerifier) awaitRegistered(ctx context.Context, conn *substrateAPI) error {
	deadline := time.Now().Add(5 * time.Minute)
	var active int64
	for {
		active = 0
		c, cancel, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
		if err != nil {
			return err
		}
		resp, err := conn.control.ListWorkers(c, &ateapipb.ListWorkersRequest{Options: &ateapipb.ListOptions{PageSize: 500}})
		cancel()
		if err == nil {
			for _, w := range resp.GetWorkers() {
				if w.GetWorkerPool() == v.Name && w.GetWorkerNamespace() == v.Namespace &&
					w.GetStatus().GetState() == ateapipb.WorkerState_WORKER_STATE_ACTIVE {
					classes := w.GetStatus().GetSandboxClasses()
					if len(classes) != 1 || classes[0].GetSandboxClass() != v.Class {
						return errors.Errorf("worker %s of pool %s/%s reports sandbox classes %v, want %s", w.GetWorkerPod(), v.Namespace, v.Name, classes, v.Class)
					}
					active++
				}
			}
			if active == v.Replicas {
				fmt.Printf("  [verify] the API lists %d active workers of the pool\n", active)
				return nil
			}
		}
		if time.Now().After(deadline) {
			return errors.Errorf("the API lists %d active workers of pool %s/%s, want %d (last error: %v)", active, v.Namespace, v.Name, v.Replicas, err)
		}
		time.Sleep(5 * time.Second)
	}
}

// proveActorLifecycle is upstream's durable-directory lifecycle
// (internal/e2e/suites/demo TestDurableDirLifecycle) on this pool: create,
// resume, call through the router, suspend into the control plane's snapshot
// location, find the snapshot in the bucket, resume, call again. The actor
// boots the default version of the SandboxConfig the pool resolved (config),
// the only runtime the pool's workers run.
func (v *AgentSubstrateWorkerPoolVerifier) proveActorLifecycle(ctx context.Context, kubeconfig string, conn *substrateAPI, outputs map[string]interface{}, config string) (err error) {
	if len(v.Labels) == 0 {
		return errors.New("the actor lane's pool must carry labels: an actor template selects its pool by them")
	}
	images, err := releaseImages(ctx, fmt.Sprint(outputs["version"]))
	if err != nil {
		return err
	}
	atespace := laneAtespace("actors")
	if err := ensureAtespace(ctx, conn, atespace); err != nil {
		return err
	}
	// Whatever happens below, the actor, its template and the atespace go
	// before the verifier returns: a running actor left on a worker would
	// hold the pool's teardown for the worker's hour-long drain.
	actor := &ateapipb.ObjectRef{Atespace: atespace, Name: agentSubstrateActorName}
	defer func() { err = withLaneCleanup(err, cleanUpLaneAtespace(ctx, conn, atespace, actor)) }()

	template := counterTemplate(atespace, agentSubstrateActorTemplate, images[counterImageKey], v.Labels)
	if err := createTemplate(ctx, conn, template); err != nil {
		return errors.Wrapf(err, "creating the counter template on pool %s/%s (SandboxConfig %s)", v.Namespace, v.Name, config)
	}
	if err := awaitGoldenSnapshot(ctx, conn, &ateapipb.ObjectRef{Atespace: atespace, Name: agentSubstrateActorTemplate}); err != nil {
		return err
	}

	c, cancel, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
	if err != nil {
		return err
	}
	_, err = conn.control.CreateActor(c, &ateapipb.CreateActorRequest{Actor: &ateapipb.Actor{
		Metadata:      &ateapipb.ResourceMetadata{Atespace: atespace, Name: agentSubstrateActorName},
		ActorTemplate: &ateapipb.ObjectRef{Atespace: atespace, Name: agentSubstrateActorTemplate},
	}})
	cancel()
	if err != nil {
		return errors.Wrap(err, "creating the counter actor")
	}

	router := fmt.Sprint(outputs["router_tls_endpoint"])
	if err := resumeAndCall(ctx, kubeconfig, conn, actor, router, 1, 1, "after creation"); err != nil {
		return err
	}

	c, cancel, err = conn.as(ctx, agentSubstrateProbeClient, 2*time.Minute)
	if err != nil {
		return err
	}
	_, err = conn.control.SuspendActor(c, &ateapipb.SuspendActorRequest{Actor: actor})
	cancel()
	if err != nil {
		return errors.Wrap(err, "suspending the counter actor")
	}
	suspended, err := awaitActorState(ctx, conn, actor, ateapipb.ActorState_ACTOR_STATE_SUSPENDED, 5*time.Minute)
	if err != nil {
		return err
	}
	snapshot := durableSnapshot(suspended.GetStatus())
	if snapshot == nil {
		return errors.Errorf("the suspended actor reports no completed snapshot in storage; snapshots: %v", suspended.GetStatus().GetSnapshots())
	}
	if snapshot.GetFidelity() != ateapipb.SnapshotFidelity_SNAPSHOT_FIDELITY_VOLUMES {
		return errors.Errorf("the actor's snapshot has fidelity %v, want VOLUMES (the template's)", snapshot.GetFidelity())
	}
	uri := snapshot.GetDurableSnapshot().GetObject().GetSnapshotUri()
	if location := strings.TrimSuffix(fmt.Sprint(outputs["storage_location"]), "/") + "/"; !strings.HasPrefix(uri, location) {
		return errors.Errorf("the actor's snapshot is at %s, outside the control plane's snapshot location %s (the storage_location output)", uri, location)
	}
	if err := proveSnapshotInBucket(ctx, kubeconfig, uri); err != nil {
		return err
	}

	// After the resume the file counter continues (the durable directory came
	// back from the snapshot) and the memory counter starts over (VOLUMES
	// fidelity: the actor really left memory).
	if err := resumeAndCall(ctx, kubeconfig, conn, actor, router, 1, 2, "after suspend and resume"); err != nil {
		return err
	}
	fmt.Printf("  [verify] the actor slept into %s and woke with its files\n", uri)
	return nil
}

func (v *AgentSubstrateWorkerPoolVerifier) VerifyAbsent(ctx context.Context, kubeconfig string) error {
	if err := KubectlResourceAbsent(ctx, kubeconfig, agentSubstrateWorkerPoolResource, v.Name, v.Namespace); err != nil {
		return err
	}
	if err := KubectlResourceAbsent(ctx, kubeconfig, "deployment", v.Name, v.Namespace); err != nil {
		return errors.Wrap(err, "the pool's worker Deployment must go with the pool")
	}
	return removeSubstrateProbe(ctx, kubeconfig)
}

// ---------------------------------------------------------------------------
// Shared API steps
// ---------------------------------------------------------------------------

// counterTemplate is upstream's counter template (demos/counter) with the
// lane's image and pool selector: one container serving port 80, a durable
// directory at /home/counter, the limits every actor must declare, and
// VOLUMES fidelity (see the counter constants). It names no storage
// location, so its snapshots go under the control plane's own.
func counterTemplate(atespace, name, image string, selector map[string]string) *ateapipb.ActorTemplate {
	t := &ateapipb.ActorTemplate{
		Metadata: &ateapipb.ResourceMetadata{Atespace: atespace, Name: name},
		Containers: []*ateapipb.Container{{
			Name:         "counter",
			Image:        image,
			Command:      []string{counterCommand},
			WakeupProbe:  &ateapipb.WakeupProbe{HttpGet: &ateapipb.HttpGetAction{Path: "/readyz", Port: counterPort}},
			VolumeMounts: []*ateapipb.VolumeMount{{Name: "data", MountPath: counterDataMount}},
		}},
		Volumes: []*ateapipb.Volume{{Name: "data", DurableDir: &ateapipb.DurableDirVolumeSource{}}},
		Resources: &ateapipb.ResourceRequirements{Limits: &ateapipb.Resources{
			CpuMillicores: counterActorMillicores,
			MemoryBytes:   counterActorMemory,
		}},
		SnapshotConfig: &ateapipb.SnapshotConfig{
			PreferredFidelity: ateapipb.SnapshotFidelity_SNAPSHOT_FIDELITY_VOLUMES,
		},
	}
	if len(selector) > 0 {
		t.WorkerSelector = &ateapipb.Selector{MatchLabels: selector}
	}
	return t
}

// releaseImages reads the image index Planton publishes with the release the
// control plane reports in its version output.
func releaseImages(ctx context.Context, version string) (map[string]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(agentSubstrateReleaseImagesIndex, version), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "fetching the image index of Agent Substrate %s", version)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("the image index of Agent Substrate %s answered HTTP %d", version, resp.StatusCode)
	}
	var index struct {
		Images map[string]string `json:"images"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, errors.Wrapf(err, "reading the image index of Agent Substrate %s", version)
	}
	if index.Images[counterImageKey] == "" {
		return nil, errors.Errorf("the image index of Agent Substrate %s names no %q image", version, counterImageKey)
	}
	return index.Images, nil
}

func ensureAtespace(ctx context.Context, conn *substrateAPI, name string) error {
	deadline := time.Now().Add(3 * time.Minute)
	for {
		c, cancel, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
		if err != nil {
			return err
		}
		_, err = conn.control.CreateAtespace(c, &ateapipb.CreateAtespaceRequest{Atespace: &ateapipb.Atespace{Metadata: &ateapipb.ResourceMetadata{Name: name}}})
		cancel()
		if err == nil || status.Code(err) == codes.AlreadyExists {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.Wrapf(err, "creating atespace %s", name)
		}
		time.Sleep(5 * time.Second)
	}
}

func deleteAtespace(ctx context.Context, conn *substrateAPI, name string) error {
	c, cancel, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
	if err != nil {
		return err
	}
	_, err = conn.control.DeleteAtespace(c, &ateapipb.DeleteAtespaceRequest{Atespace: &ateapipb.ObjectRef{Name: name}})
	cancel()
	if err != nil && status.Code(err) != codes.NotFound {
		return errors.Wrapf(err, "deleting atespace %s", name)
	}
	return nil
}

func createTemplate(ctx context.Context, conn *substrateAPI, t *ateapipb.ActorTemplate) error {
	c, cancel, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
	if err != nil {
		return err
	}
	defer cancel()
	_, err = conn.control.CreateActorTemplate(c, &ateapipb.CreateActorTemplateRequest{ActorTemplate: t})
	return err
}

// laneAtespace names a fresh atespace for one run of a lane.
func laneAtespace(purpose string) string {
	return fmt.Sprintf("e2e-ate-%s-%06x", purpose, rand.Uint32()&0xffffff)
}

// cleanUpLaneAtespace deletes everything a lane created in its atespace: the
// actor (in any state, waiting until it is gone, so its worker is free before
// the pool is destroyed), every template, then the atespace itself. It runs
// on a context the lane's own cancellation cannot end, since it also runs
// after a lane failed for running out of time.
func cleanUpLaneAtespace(ctx context.Context, conn *substrateAPI, atespace string, actor *ateapipb.ObjectRef) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Minute)
	defer cancel()
	if actor != nil {
		c, cancelCall, err := conn.as(ctx, agentSubstrateProbeClient, 2*time.Minute)
		if err != nil {
			return err
		}
		_, err = conn.control.DeleteActor(c, &ateapipb.DeleteActorRequest{Actor: actor})
		cancelCall()
		if err != nil && status.Code(err) != codes.NotFound {
			return errors.Wrapf(err, "deleting actor %s/%s", actor.GetAtespace(), actor.GetName())
		}
		for {
			c, cancelCall, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
			if err != nil {
				return err
			}
			_, err = conn.control.GetActor(c, &ateapipb.GetActorRequest{Actor: actor})
			cancelCall()
			if status.Code(err) == codes.NotFound {
				break
			}
			if ctx.Err() != nil {
				return errors.Errorf("actor %s/%s was never deleted (last: %v)", actor.GetAtespace(), actor.GetName(), err)
			}
			time.Sleep(5 * time.Second)
		}
	}
	c, cancelCall, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
	if err != nil {
		return err
	}
	templates, err := conn.control.ListActorTemplates(c, &ateapipb.ListActorTemplatesRequest{Atespace: atespace})
	cancelCall()
	if err != nil {
		return errors.Wrapf(err, "listing the actor templates in atespace %s", atespace)
	}
	for _, t := range templates.GetActorTemplates() {
		if err := deleteTemplate(ctx, conn, atespace, t.GetMetadata().GetName()); err != nil {
			return err
		}
	}
	return deleteAtespace(ctx, conn, atespace)
}

// withLaneCleanup folds a lane's cleanup into its result: the lane's own
// failure wins and the cleanup's is printed beside it.
func withLaneCleanup(laneErr, cleanupErr error) error {
	if cleanupErr == nil {
		return laneErr
	}
	if laneErr == nil {
		return errors.Wrap(cleanupErr, "cleaning up the lane's API objects")
	}
	fmt.Printf("  [verify] cleaning up the lane's API objects also failed: %v\n", cleanupErr)
	return laneErr
}

func deleteTemplate(ctx context.Context, conn *substrateAPI, atespace, name string) error {
	c, cancel, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
	if err != nil {
		return err
	}
	_, err = conn.control.DeleteActorTemplate(c, &ateapipb.DeleteActorTemplateRequest{ActorTemplate: &ateapipb.ObjectRef{Atespace: atespace, Name: name}})
	cancel()
	if err != nil && status.Code(err) != codes.NotFound {
		return errors.Wrapf(err, "deleting actor template %s", name)
	}
	return nil
}

// awaitGoldenSnapshot waits for the counter template's golden snapshot, the
// boot image every actor of the template resumes from, the way upstream's
// WaitForSubstrateTemplateReady does.
func awaitGoldenSnapshot(ctx context.Context, conn *substrateAPI, ref *ateapipb.ObjectRef) error {
	deadline := time.Now().Add(10 * time.Minute)
	var last *ateapipb.GoldenSnapshotStatus
	for {
		c, cancel, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
		if err != nil {
			return err
		}
		t, err := conn.control.GetActorTemplate(c, &ateapipb.GetActorTemplateRequest{ActorTemplate: ref})
		cancel()
		if err == nil {
			last = t.GetStatus().GetGoldenSnapshotStatus()
			if last.GetGoldenTag().GetName() != "" {
				fmt.Printf("  [verify] the counter template's golden snapshot is ready\n")
				return nil
			}
			if msg := last.GetErrorMessage(); msg != "" {
				return errors.Errorf("the counter template failed to build its golden snapshot: %s", msg)
			}
		}
		if time.Now().After(deadline) {
			return errors.Errorf("the counter template's golden snapshot was never built (last status %v, err %v)", last, err)
		}
		time.Sleep(3 * time.Second)
	}
}

// resumeAndCall resumes the actor (waiting out a saturated pool, as
// upstream's ResumeActorAwaitCapacity does), waits for it to run, and calls
// it through the router, expecting the two counters.
func resumeAndCall(ctx context.Context, kubeconfig string, conn *substrateAPI, actor *ateapipb.ObjectRef, router string, wantMemory, wantFile int, stage string) error {
	deadline := time.Now().Add(5 * time.Minute)
	for {
		c, cancel, err := conn.as(ctx, agentSubstrateProbeClient, 2*time.Minute)
		if err != nil {
			return err
		}
		_, err = conn.control.ResumeActor(c, &ateapipb.ResumeActorRequest{Actor: actor})
		cancel()
		if err == nil {
			break
		}
		if status.Code(err) != codes.ResourceExhausted || time.Now().After(deadline) {
			return errors.Wrapf(err, "resuming the counter actor (%s)", stage)
		}
		time.Sleep(3 * time.Second)
	}
	if _, err := awaitActorState(ctx, conn, actor, ateapipb.ActorState_ACTOR_STATE_RUNNING, 5*time.Minute); err != nil {
		return err
	}
	body, err := callThroughRouter(ctx, kubeconfig, router, actor)
	if err != nil {
		return errors.Wrapf(err, "calling the counter actor through the router (%s)", stage)
	}
	if !strings.Contains(body, counterMemoryPrefix+fmt.Sprint(wantMemory)) || !strings.Contains(body, counterFilePrefix+fmt.Sprint(wantFile)) {
		return errors.Errorf("%s the actor answered %q, want memory count %d and file counter %d", stage, strings.TrimSpace(body), wantMemory, wantFile)
	}
	fmt.Printf("  [verify] %s the actor answered through the router: memory %d, file %d\n", stage, wantMemory, wantFile)
	return nil
}

func awaitActorState(ctx context.Context, conn *substrateAPI, actor *ateapipb.ObjectRef, want ateapipb.ActorState, budget time.Duration) (*ateapipb.Actor, error) {
	deadline := time.Now().Add(budget)
	var last *ateapipb.Actor
	for {
		c, cancel, err := conn.as(ctx, agentSubstrateProbeClient, 30*time.Second)
		if err != nil {
			return nil, err
		}
		got, err := conn.control.GetActor(c, &ateapipb.GetActorRequest{Actor: actor})
		cancel()
		if err == nil {
			last = got
			if got.GetStatus().GetState() == want {
				return got, nil
			}
			if got.GetStatus().GetState() == ateapipb.ActorState_ACTOR_STATE_CRASHED {
				return nil, errors.Errorf("the counter actor crashed: %s", got.GetStatus().GetCrash().GetMessage())
			}
		}
		if time.Now().After(deadline) {
			return nil, errors.Errorf("the counter actor never reached %v (last state %v, err %v)", want, last.GetStatus().GetState(), err)
		}
		time.Sleep(2 * time.Second)
	}
}

// durableSnapshot is the newest completed snapshot in storage, chosen the way
// upstream's demo suite chooses it.
func durableSnapshot(s *ateapipb.ActorStatus) *ateapipb.Snapshot {
	var best *ateapipb.Snapshot
	for _, snap := range s.GetSnapshots() {
		st := snap.GetDurableSnapshot()
		if st.GetStatus() == ateapipb.SnapshotStorageStatus_SNAPSHOT_STORAGE_STATUS_COMPLETED && st.GetObject().GetSnapshotUri() != "" {
			if best == nil || snap.GetGeneration() > best.GetGeneration() {
				best = snap
			}
		}
	}
	return best
}

// callThroughRouter makes the request a product makes, from inside the
// cluster: the probe's router caller presents its pod certificate to the
// router's TLS endpoint (the router_tls_endpoint output) with
// `ate-target-actor: <atespace>/<actor>`. A freshly resumed actor can take a
// moment to accept connections, so the call is retried briefly.
func callThroughRouter(ctx context.Context, kubeconfig, router string, actor *ateapipb.ObjectRef) (string, error) {
	deadline := time.Now().Add(1 * time.Minute)
	var last error
	for time.Now().Before(deadline) {
		code, body, err := callRouter(ctx, kubeconfig, agentSubstrateRouterCaller, router, actor.GetAtespace()+"/"+actor.GetName())
		if err == nil && code == http.StatusOK {
			return body, nil
		}
		last = errors.Errorf("HTTP %d %v: %s", code, err, firstLines(body, 3))
		time.Sleep(3 * time.Second)
	}
	return "", last
}

// proveSnapshotInBucket lists the snapshot's objects in the snapshot store
// itself, independently of Substrate's own report: over a port-forward to the
// store's S3 Service (the KubernetesSeaweedFs prerequisite's s3_endpoint
// output) with the very key the control plane writes snapshots with.
func proveSnapshotInBucket(ctx context.Context, kubeconfig, snapshotURI string) error {
	u, err := url.Parse(snapshotURI)
	if err != nil || u.Host == "" {
		return errors.Errorf("the snapshot URI %q names no bucket", snapshotURI)
	}
	bucket, prefix := u.Host, strings.TrimPrefix(u.Path, "/")
	stores := provider.DependencyOutputsFrom(ctx)[agentSubstrateStoreKind]
	if len(stores) != 1 {
		return errors.Errorf("the actor lane lists its snapshot in the %s deployed in front of it, and %d were deployed", agentSubstrateStoreKind, len(stores))
	}
	var endpoint string
	for _, outputs := range stores {
		endpoint = fmt.Sprint(outputs["s3_endpoint"])
	}
	e, err := url.Parse(endpoint)
	if err != nil || e.Hostname() == "" {
		return errors.Errorf("the store's s3_endpoint output %q is not a URL", endpoint)
	}
	labels := strings.Split(e.Hostname(), ".")
	if len(labels) < 2 {
		return errors.Errorf("the store's s3_endpoint output %q does not name a Service", endpoint)
	}
	secrets := provider.DependencyOutputsFrom(ctx)[agentSubstrateStoreSecretKind]
	if len(secrets) != 1 {
		return errors.Errorf("the actor lane reads the snapshot-store key from the %s deployed in front of it, and %d were deployed", agentSubstrateStoreSecretKind, len(secrets))
	}
	var secretName, secretNamespace string
	for _, outputs := range secrets {
		secretName, secretNamespace = fmt.Sprint(outputs["secret_name"]), fmt.Sprint(outputs["secret_namespace"])
	}
	accessKey, err := kubectlGetJSONPath(ctx, kubeconfig, "secret", secretName, secretNamespace, "{.data."+agentSubstrateStoreAccessKeyID+"}")
	if err != nil {
		return errors.Wrap(err, "reading the control plane's snapshot-store key")
	}
	secretKey, err := kubectlGetJSONPath(ctx, kubeconfig, "secret", secretName, secretNamespace, "{.data."+agentSubstrateStoreSecretKey+"}")
	if err != nil {
		return errors.Wrap(err, "reading the control plane's snapshot-store key")
	}
	id, err := decodeBase64(accessKey)
	if err != nil {
		return err
	}
	key, err := decodeBase64(secretKey)
	if err != nil {
		return err
	}

	localPort, err := freeLocalPort()
	if err != nil {
		return err
	}
	stop, err := startPortForward(ctx, kubeconfig, "svc/"+labels[0], labels[1], fmt.Sprintf("%d:%s", localPort, e.Port()))
	if err != nil {
		return errors.Wrap(err, "forwarding a local port to the snapshot store")
	}
	defer stop()
	client := awss3.New(awss3.Options{
		Region:       "us-east-1",
		BaseEndpoint: awssdk.String(fmt.Sprintf("http://127.0.0.1:%d", localPort)),
		Credentials:  credentials.NewStaticCredentialsProvider(id, key, ""),
		UsePathStyle: true,
	})
	out, err := client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{Bucket: awssdk.String(bucket), Prefix: awssdk.String(prefix)})
	if err != nil {
		return errors.Wrapf(err, "listing %s in the snapshot store", snapshotURI)
	}
	if len(out.Contents) == 0 {
		return errors.Errorf("the actor reports its snapshot at %s, but the bucket holds no object under it", snapshotURI)
	}
	fmt.Printf("  [verify] the snapshot store holds %d object(s) under %s\n", len(out.Contents), snapshotURI)
	return nil
}

func decodeBase64(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return "", errors.Wrap(err, "decoding a Secret value")
	}
	return string(raw), nil
}
