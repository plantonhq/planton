package resources

import (
	"encoding/json"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// The derived resource names, asserted as literals: these are the names a
// human will see in kubectl output and docs, so a rename must fail a test.
const (
	wantRunnerName         = "planton-runner"
	wantRunnerStatePVCName = "planton-runner-state"
	wantRunnerCloudOpsName = "planton-runner-cloudops"
)

// testRunnerConfig mirrors the shipped default: builds ON (the component
// threads the effective spec.build toggle here, and its default is true --
// builds power Service Hub). The opt-out shape is tested explicitly.
func testRunnerConfig() RunnerConfig {
	return RunnerConfig{
		CRName:       "planton",
		Namespace:    "default",
		Version:      "v1.0.0",
		OrgSlug:      "default",
		BuildEnabled: true,
	}
}

func testRunnerConfigBuildsOff() RunnerConfig {
	cfg := testRunnerConfig()
	cfg.BuildEnabled = false
	return cfg
}

func TestRunnerNames(t *testing.T) {
	if got := RunnerDeploymentName("planton"); got != wantRunnerName {
		t.Errorf("deployment = %s, want planton-runner", got)
	}
	if got := RunnerServiceAccountName("planton"); got != wantRunnerName {
		t.Errorf("serviceaccount = %s, want planton-runner", got)
	}
	if got := RunnerServiceName("planton"); got != wantRunnerName {
		t.Errorf("service = %s, want planton-runner", got)
	}
	if got := RunnerStatePVCName("planton"); got != wantRunnerStatePVCName {
		t.Errorf("pvc = %s, want planton-runner-state", got)
	}
	if got := RunnerCloudOpsSecretName("planton"); got != wantRunnerCloudOpsName {
		t.Errorf("secret = %s, want planton-runner-cloudops", got)
	}
	if got := RunnerServiceFQDN("planton", "default"); got != "planton-runner.default.svc.cluster.local" {
		t.Errorf("service FQDN = %s, want planton-runner.default.svc.cluster.local", got)
	}
	// The slug IS the ServiceAccount name -- the badge identity-binding
	// grammar (the control plane refuses a registration whose declared
	// ServiceAccount differs from its slug). One derivation, asserted equal.
	if got := RunnerSlug("planton"); got != RunnerServiceAccountName("planton") {
		t.Errorf("slug = %s, must equal the ServiceAccount name %s", got, RunnerServiceAccountName("planton"))
	}
}

// The task queue is derived identically on both sides of the dispatch: the
// control plane env and the runner's identity document must agree or deploys
// sit pending forever. The derivation mirrors the runner binary's
// TaskQueuePrefix + channel-identifier convention.
func TestRunnerTaskQueueDerivation(t *testing.T) {
	if got := RunnerChannelIdentifier("planton", "acme"); got != "org.acme.runner.planton-runner" {
		t.Errorf("channel = %s, want org.acme.runner.planton-runner", got)
	}
	if got := RunnerTaskQueue("planton", "acme"); got != "iac-operation.org.acme.runner.planton-runner" {
		t.Errorf("queue = %s, want iac-operation.org.acme.runner.planton-runner", got)
	}
}

// The CloudOps token must never be mistakable for a Planton API key (pak_):
// the two are different trust relationships (control-plane->runner perimeter
// vs a control-plane credential) and a shared prefix would invite pasting one
// where the other belongs.
func TestGenerateRunnerCloudOpsToken(t *testing.T) {
	token, err := GenerateRunnerCloudOpsToken()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(token, "pcot_") {
		t.Errorf("token %q must carry the pcot_ prefix", token)
	}
	if strings.HasPrefix(token, "pak_") {
		t.Errorf("token %q must never look like a control-plane API key", token)
	}
	// 256-bit entropy: 32 bytes base64url unpadded = 43 chars.
	if len(token) != len("pcot_")+43 {
		t.Errorf("token length = %d, want %d (pcot_ + 43 chars of base64url)", len(token), len("pcot_")+43)
	}
	second, err := GenerateRunnerCloudOpsToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == second {
		t.Error("two minted tokens must never collide")
	}
}

// The identity document is SECRET-FREE: identity coordinates and the
// in-cluster endpoint, never a credential of any kind -- the runner's proof
// is its projected badge, so nothing in this document could admit or
// impersonate anything.
func TestRunnerIdentityDocumentJSON(t *testing.T) {
	out := RunnerIdentityDocumentJSON("planton", "default", "acme")
	var doc map[string]string
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("the identity document must be valid JSON: %v", err)
	}
	if doc["type"] != "planton_runner" {
		t.Errorf("type = %q, want planton_runner", doc["type"])
	}
	if doc["org"] != "acme" || doc["runner"] != "planton-runner" {
		t.Errorf("identity = %s/%s, want acme/planton-runner", doc["org"], doc["runner"])
	}
	if doc["channel_identifier"] != "org.acme.runner.planton-runner" {
		t.Errorf("channel_identifier = %q, want org.acme.runner.planton-runner", doc["channel_identifier"])
	}
	if doc["planton_api_endpoint"] != "planton-control-plane.default.svc.cluster.local:80" {
		t.Errorf("planton_api_endpoint = %q, want the in-cluster control-plane Service",
			doc["planton_api_endpoint"])
	}
	// NO credential and NO tunnel material, structurally: the badge is the
	// proof, and the tunnel is the remote runner's transport.
	for _, absent := range []string{"api_key", "tunnel_endpoint", "ca_certificate",
		"agent_certificate", "agent_private_key"} {
		if _, ok := doc[absent]; ok {
			t.Errorf("%s must be absent from the badge runner's identity document", absent)
		}
	}
}

func TestRunnerCloudOpsSecret(t *testing.T) {
	secret := RunnerCloudOpsSecret("planton", "default", "pcot_token", nil)
	if secret.Name != wantRunnerCloudOpsName {
		t.Errorf("name = %s, want planton-runner-cloudops", secret.Name)
	}
	if string(secret.Data[RunnerCloudOpsSecretKeyToken]) != "pcot_token" {
		t.Error("cloudops-auth-token entry must hold the direct-dial bearer (both sides read it)")
	}
	// Exactly ONE key: the Secret carries the perimeter token and nothing
	// else -- no api-key, no credentials.json, no identity material of any
	// kind survives on this install.
	if len(secret.Data) != 1 {
		t.Errorf("data = %v, want exactly the cloudops-auth-token entry", secret.Data)
	}
}

// The Service is the control plane's direct-dial door for live cloud
// operations (gRPC), and -- with builds on, the default -- Tekton's
// CloudEvents door on port 80, so the sink URL needs no explicit port.
func TestRunnerService(t *testing.T) {
	svc := RunnerService(testRunnerConfig())
	if svc.Name != wantRunnerName {
		t.Errorf("name = %s, want planton-runner", svc.Name)
	}
	if len(svc.Spec.Ports) != 2 {
		t.Fatalf("ports = %+v, want grpc + webhook in the default (builds-on) shape", svc.Spec.Ports)
	}
	if svc.Spec.Ports[0].Name != "grpc" || svc.Spec.Ports[0].Port != 50051 ||
		svc.Spec.Ports[0].TargetPort.IntValue() != 50051 {
		t.Errorf("ports[0] = %+v, want grpc 50051->50051", svc.Spec.Ports[0])
	}
	if svc.Spec.Ports[1].Name != "webhook" || svc.Spec.Ports[1].Port != 80 ||
		svc.Spec.Ports[1].TargetPort.String() != "webhook" {
		t.Errorf("ports[1] = %+v, want webhook 80->webhook (portless sink URL)", svc.Spec.Ports[1])
	}
	if svc.Spec.Selector["app.kubernetes.io/name"] != "runner" {
		t.Errorf("selector = %v, must select the runner pod", svc.Spec.Selector)
	}
}

// Opting out of builds retires the webhook door: Tekton events have nowhere
// to land, and nothing should pretend otherwise.
func TestRunnerService_BuildOptOut(t *testing.T) {
	svc := RunnerService(testRunnerConfigBuildsOff())
	if len(svc.Spec.Ports) != 1 || svc.Spec.Ports[0].Name != "grpc" {
		t.Errorf("ports = %+v, want only grpc with builds off", svc.Spec.Ports)
	}
}

// The build Role's verbs mirror the runner's own readiness probe exactly
// (rbac-build + rbac-logs assert these needs via SelfSubjectAccessReview) --
// a drift here surfaces as a named readiness failure, and this test names
// the contract at its source.
func TestRunnerBuildRole_ExactVerbs(t *testing.T) {
	role := RunnerBuildRole(testRunnerConfig())
	if role.Name != "planton-runner-build" {
		t.Errorf("name = %s, want planton-runner-build", role.Name)
	}
	if role.Namespace != "default" {
		t.Errorf("namespace = %s, want the CR namespace (where TEKTON_NAMESPACE points)", role.Namespace)
	}

	type rule struct {
		group, resource string
		verbs           string
	}
	got := make([]rule, 0, len(role.Rules))
	for _, r := range role.Rules {
		for _, res := range r.Resources {
			got = append(got, rule{r.APIGroups[0], res, strings.Join(r.Verbs, ",")})
		}
	}
	want := []rule{
		{"tekton.dev", "pipelineruns", "create,list,watch,deletecollection"},
		{"tekton.dev", "taskruns", "list,watch"},
		{"", "secrets", "create,get,deletecollection"},
		{"", "serviceaccounts", "create,get,deletecollection"},
		{"", "configmaps", "get,create,update,patch,deletecollection"},
		{"rbac.authorization.k8s.io", "roles", "create,deletecollection"},
		{"rbac.authorization.k8s.io", "rolebindings", "create,deletecollection"},
		{"", "pods", "get,list,watch"},
		{"", "pods/log", "get"},
	}
	if len(got) != len(want) {
		t.Fatalf("rules = %+v, want exactly %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("rule[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestRunnerBuildRoleBinding_BindsRunnerServiceAccount(t *testing.T) {
	binding := RunnerBuildRoleBinding(testRunnerConfig())
	if binding.RoleRef.Name != "planton-runner-build" || binding.RoleRef.Kind != "Role" {
		t.Errorf("roleRef = %+v, want the build Role", binding.RoleRef)
	}
	if len(binding.Subjects) != 1 || binding.Subjects[0].Kind != "ServiceAccount" ||
		binding.Subjects[0].Name != wantRunnerName || binding.Subjects[0].Namespace != "default" {
		t.Errorf("subjects = %+v, want the runner's dedicated ServiceAccount", binding.Subjects)
	}
}

// The ServiceAccount exists even without annotations (granting cloud access
// later is a pure annotation edit); with annotations it carries the
// workload-identity binding verbatim.
func TestRunnerServiceAccount(t *testing.T) {
	cfg := testRunnerConfig()
	sa := RunnerServiceAccount(cfg)
	if sa.Name != wantRunnerName {
		t.Errorf("name = %s, want planton-runner", sa.Name)
	}
	if len(sa.Annotations) != 0 {
		t.Errorf("annotations = %v, want none by default", sa.Annotations)
	}

	cfg.ServiceAccountAnnotations = map[string]string{
		"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/planton-runner",
	}
	sa = RunnerServiceAccount(cfg)
	if sa.Annotations["eks.amazonaws.com/role-arn"] != "arn:aws:iam::123456789012:role/planton-runner" {
		t.Error("the IRSA annotation must land on the ServiceAccount verbatim")
	}
}

func TestRunnerStatePVC(t *testing.T) {
	pvc := RunnerStatePVC(testRunnerConfig())
	if pvc.Name != wantRunnerStatePVCName {
		t.Errorf("name = %s, want planton-runner-state", pvc.Name)
	}
	size := pvc.Spec.Resources.Requests["storage"]
	if size.String() != RunnerDefaultStorageSize {
		t.Errorf("default size = %s, want %s", size.String(), RunnerDefaultStorageSize)
	}

	cfg := testRunnerConfig()
	cfg.StorageSize = resource.MustParse("10Gi")
	pvc = RunnerStatePVC(cfg)
	size = pvc.Spec.Resources.Requests["storage"]
	if size.String() != "10Gi" {
		t.Errorf("size = %s, want the declared 10Gi", size.String())
	}
}

func TestRunnerStatePVC_StorageClass(t *testing.T) {
	// Unpinned MUST be nil, not "": an empty string means "only bind
	// pre-provisioned volumes" to Kubernetes, silently hanging the claim.
	pvc := RunnerStatePVC(testRunnerConfig())
	if pvc.Spec.StorageClassName != nil {
		t.Errorf("storageClassName = %q, want nil when unpinned", *pvc.Spec.StorageClassName)
	}

	cfg := testRunnerConfig()
	cfg.StorageClassName = "trident"
	pvc = RunnerStatePVC(cfg)
	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName != "trident" {
		t.Errorf("storageClassName = %v, want trident", pvc.Spec.StorageClassName)
	}
}

func TestRunnerDeployment_Image(t *testing.T) {
	deploy := RunnerDeployment(testRunnerConfig())
	if got := deploy.Spec.Template.Spec.Containers[0].Image; got != RunnerDefaultImageRepo+":v1.0.0" {
		t.Errorf("image = %s, want %s:v1.0.0 (platform version by default)", got, RunnerDefaultImageRepo)
	}

	// The image's entrypoint is the bare binary by contract; a Deployment
	// that passes no subcommand prints help and CrashLoops with no error
	// (caught live on the first runner-enabled boot).
	if args := deploy.Spec.Template.Spec.Containers[0].Args; len(args) != 1 || args[0] != "start" {
		t.Errorf("args = %v, want exactly [start]", args)
	}

	cfg := testRunnerConfig()
	cfg.ImageRepository = "example.com/runner"
	cfg.ImageTag = "custom"
	deploy = RunnerDeployment(cfg)
	if got := deploy.Spec.Template.Spec.Containers[0].Image; got != "example.com/runner:custom" {
		t.Errorf("image = %s, want example.com/runner:custom", got)
	}
}

// The worker's boot contract: dual mode (IaC worker + the CloudOps surface,
// whose perimeter is the auth token -- the runner refuses to bind beyond
// loopback without one), polling the control plane's DISPATCH namespace,
// state on the PVC path, the SECRET-FREE identity document inline, and the
// projected badge as the one and only proof. Builds are ON in the default
// shape: worker and webhook together, placement in the CR namespace, the
// webhook port stated explicitly.
func TestRunnerDeployment_EnvContract(t *testing.T) {
	deploy := RunnerDeployment(testRunnerConfig())
	envMap := envVarMap(deploy.Spec.Template.Spec.Containers[0].Env)

	if envMap["EXECUTION_MODE"] != "dual" {
		t.Errorf("EXECUTION_MODE = %q, want dual", envMap["EXECUTION_MODE"])
	}
	if envMap["TUNNEL_ENABLED"] != "false" {
		t.Error("the tunnel must be off for the in-cluster runner (it shares the control plane's network)")
	}
	if envMap["WEBHOOK_ENABLED"] != "true" || envMap["BUILD_WORKER_ENABLED"] != "true" {
		t.Error("builds are the default: the build worker and the Tekton webhook must both render ON")
	}
	if envMap["TEKTON_NAMESPACE"] != "default" {
		t.Errorf("TEKTON_NAMESPACE = %q, want the CR namespace (build placement beside the runner)",
			envMap["TEKTON_NAMESPACE"])
	}
	if envMap["WEBHOOK_PORT"] != "8086" {
		t.Errorf("WEBHOOK_PORT = %q, want 8086 stated explicitly", envMap["WEBHOOK_PORT"])
	}
	assertSecretEnv(t, deploy.Spec.Template.Spec.Containers[0].Env, "CLOUDOPS_AUTH_TOKEN",
		wantRunnerCloudOpsName, RunnerCloudOpsSecretKeyToken)
	if envMap["TEMPORAL_NAMESPACE"] != "platform.pipelines" {
		t.Errorf("TEMPORAL_NAMESPACE = %q, want platform.pipelines (the dispatch namespace, "+
			"not the runner's default fallback)", envMap["TEMPORAL_NAMESPACE"])
	}
	if envMap["TEMPORAL_SERVICE_ADDRESS"] != TemporalFrontendEndpoint("planton", "default") {
		t.Errorf("TEMPORAL_SERVICE_ADDRESS = %q, want the in-cluster frontend", envMap["TEMPORAL_SERVICE_ADDRESS"])
	}
	if envMap["PLANTON_RUNNER_CREDENTIALS"] != RunnerIdentityDocumentJSON("planton", "default", "default") {
		t.Errorf("PLANTON_RUNNER_CREDENTIALS = %q, want the inline secret-free identity document",
			envMap["PLANTON_RUNNER_CREDENTIALS"])
	}
	if envMap["PLANTON_RUNNER_WORKLOAD_IDENTITY_TOKEN_FILE"] != "/var/run/secrets/planton.ai/token" {
		t.Errorf("PLANTON_RUNNER_WORKLOAD_IDENTITY_TOKEN_FILE = %q, want the projected badge path",
			envMap["PLANTON_RUNNER_WORKLOAD_IDENTITY_TOKEN_FILE"])
	}
	if _, ok := envMap["PLANTON_RUNNER_CREDENTIALS_FILE"]; ok {
		t.Error("PLANTON_RUNNER_CREDENTIALS_FILE must be absent: no credential file exists to mount")
	}
	if envMap["PLANTON_LOCAL_IAC_STATE_DIR"] != "/var/lib/planton/iac-state" {
		t.Errorf("PLANTON_LOCAL_IAC_STATE_DIR = %q, want the PVC mount", envMap["PLANTON_LOCAL_IAC_STATE_DIR"])
	}
}

// The opt-out shape states its contract too: both flags render an explicit
// "false" (never rely on the binary's default-on), and the build-only env
// and ports are ABSENT, not empty -- a disabled capability leaves no
// half-wired surface behind.
func TestRunnerDeployment_BuildOptOut(t *testing.T) {
	deploy := RunnerDeployment(testRunnerConfigBuildsOff())
	container := deploy.Spec.Template.Spec.Containers[0]
	envMap := envVarMap(container.Env)

	if envMap["WEBHOOK_ENABLED"] != "false" || envMap["BUILD_WORKER_ENABLED"] != "false" {
		t.Error("opt-out must render both build flags explicitly false")
	}
	for _, absent := range []string{"TEKTON_NAMESPACE", "WEBHOOK_PORT"} {
		if _, ok := envMap[absent]; ok {
			t.Errorf("%s must be absent when builds are off, not empty", absent)
		}
	}
	if len(container.Ports) != 1 || container.Ports[0].Name != "grpc" {
		t.Errorf("ports = %+v, want only grpc with builds off", container.Ports)
	}
}

// With builds on, the pod exposes the webhook container port Tekton's sink
// posts to.
func TestRunnerDeployment_WebhookPort(t *testing.T) {
	container := RunnerDeployment(testRunnerConfig()).Spec.Template.Spec.Containers[0]
	found := false
	for _, port := range container.Ports {
		if port.Name == "webhook" && port.ContainerPort == 8086 {
			found = true
		}
	}
	if !found {
		t.Errorf("ports = %+v, want a webhook port 8086 with builds on", container.Ports)
	}
}

// One replica on a Recreate strategy: the state volume is ReadWriteOnce and
// two workers mutating the same state directory would corrupt it.
func TestRunnerDeployment_SingleWriterDiscipline(t *testing.T) {
	deploy := RunnerDeployment(testRunnerConfig())
	if *deploy.Spec.Replicas != 1 {
		t.Errorf("replicas = %d, want exactly 1", *deploy.Spec.Replicas)
	}
	if deploy.Spec.Strategy.Type != appsv1.RecreateDeploymentStrategyType {
		t.Errorf("strategy = %s, want Recreate (RWO volume handover)", deploy.Spec.Strategy.Type)
	}
}

// Readiness probes the worker-poll key (SERVING means deploys can land);
// liveness deliberately probes only the overall server so a Temporal outage
// reads as not-ready, never as a kill loop.
func TestRunnerDeployment_Probes(t *testing.T) {
	container := RunnerDeployment(testRunnerConfig()).Spec.Template.Spec.Containers[0]

	if container.ReadinessProbe == nil || container.ReadinessProbe.GRPC == nil ||
		container.ReadinessProbe.GRPC.Service == nil ||
		*container.ReadinessProbe.GRPC.Service != RunnerWorkerHealthService {
		t.Error("readiness must probe the worker-poll health key")
	}
	if container.LivenessProbe == nil || container.LivenessProbe.GRPC == nil {
		t.Fatal("expected a gRPC liveness probe")
	}
	if container.LivenessProbe.GRPC.Service != nil && *container.LivenessProbe.GRPC.Service != "" {
		t.Error("liveness must probe the overall server, not the worker key")
	}
}

func TestRunnerDeployment_Volumes(t *testing.T) {
	spec := RunnerDeployment(testRunnerConfig()).Spec.Template.Spec

	if spec.ServiceAccountName != wantRunnerName {
		t.Errorf("serviceAccountName = %s, want the dedicated account", spec.ServiceAccountName)
	}

	byName := map[string]bool{}
	for _, vol := range spec.Volumes {
		byName[vol.Name] = true
		switch vol.Name {
		case "badge-token":
			if vol.Projected == nil || len(vol.Projected.Sources) != 1 ||
				vol.Projected.Sources[0].ServiceAccountToken == nil {
				t.Fatal("badge-token must be a projected ServiceAccount token volume")
			}
			proj := vol.Projected.Sources[0].ServiceAccountToken
			if proj.Audience != RunnerBadgeAudience {
				t.Errorf("badge audience = %q, want %s (what the verifier expects)", proj.Audience, RunnerBadgeAudience)
			}
			if proj.Path != "token" {
				t.Errorf("badge path = %q, want token", proj.Path)
			}
			if proj.ExpirationSeconds == nil || *proj.ExpirationSeconds != 3600 {
				t.Errorf("badge expiration = %v, want 3600 (short-lived, kubelet-rotated)", proj.ExpirationSeconds)
			}
		case "iac-state":
			if vol.PersistentVolumeClaim == nil || vol.PersistentVolumeClaim.ClaimName != wantRunnerStatePVCName {
				t.Error("iac-state volume must mount the state PVC")
			}
		case "iac-cache":
			if vol.EmptyDir == nil {
				t.Error("the cache is disposable and must ride an emptyDir, not the state volume")
			}
		}
		// The old credentials Secret volume must never come back: the badge
		// volume is the ONLY identity material on the pod.
		if vol.Name == "credentials" {
			t.Error("a credentials volume exists -- the badge runner mounts no credential Secret")
		}
	}
	for _, want := range []string{"badge-token", "iac-state", "iac-cache"} {
		if !byName[want] {
			t.Errorf("missing volume %s", want)
		}
	}
}

// Static cloud credentials are the fallback path for clusters without
// workload identity: a customer-owned Secret envFrom-injected into the pod.
func TestRunnerDeployment_CloudCredentialsSecret(t *testing.T) {
	deploy := RunnerDeployment(testRunnerConfig())
	if len(deploy.Spec.Template.Spec.Containers[0].EnvFrom) != 0 {
		t.Error("no envFrom without a declared cloud credentials Secret")
	}

	cfg := testRunnerConfig()
	cfg.CloudCredentialsSecretName = "my-aws-keys"
	deploy = RunnerDeployment(cfg)
	envFrom := deploy.Spec.Template.Spec.Containers[0].EnvFrom
	if len(envFrom) != 1 || envFrom[0].SecretRef == nil || envFrom[0].SecretRef.Name != "my-aws-keys" {
		t.Error("the declared cloud credentials Secret must be envFrom-injected")
	}
}
