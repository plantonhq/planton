package module

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	v1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstigmer/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"sigs.k8s.io/yaml"
)

// TestEnginesApplyTheSameDocuments is the cross-engine parity proof: for each
// spec, what the Terraform module plans (a real `tofu plan` of iac/tf) must
// equal what this module renders: the chart's values, the Secrets' names and
// key names, the restore check's ConfigMap, the Stigmer, and the outputs. The
// two engines render with separate code, so this is the guard that they
// install the same thing. The cleanup marker is left out: it is the one
// object the engines do not share (gates.go, PARITY-EXCEPTION).
//
// It plans against a cluster (the plan-time preflight reads), so it runs when
// asked: STIGMER_TF_PARITY=1, with KUBE_CONFIG_PATH naming a cluster where
// namespace stigmer exists, no other Stigmer operator is installed, and
// `tofu` is on PATH.
func TestEnginesApplyTheSameDocuments(t *testing.T) {
	if os.Getenv("STIGMER_TF_PARITY") != "1" {
		t.Skip("set STIGMER_TF_PARITY=1 (with KUBE_CONFIG_PATH and tofu on PATH) to compare the engines' documents")
	}
	if os.Getenv("KUBE_CONFIG_PATH") == "" {
		t.Fatal("KUBE_CONFIG_PATH must name the cluster the plan reads")
	}
	pluginCache := t.TempDir()
	cases := map[string]*v1.KubernetesStigmerSpec{
		"minimal":                    paritySpec(),
		"bundled, every setting":     bundledEverything(),
		"bring your own, as values":  externalValues(),
		"bring your own, as secrets": externalSecrets(),
	}
	names := make([]string, 0, len(cases))
	for name := range cases {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		spec := cases[name]
		t.Run(name, func(t *testing.T) {
			plan := terraformPlan(t, spec, pluginCache)
			t.Logf("compared %d chart values, %d Secrets, %d documents, %d outputs", len(plan.chartValues), len(plan.secretKeys), len(plan.documents), len(plan.outputs))
			locals := &Locals{Spec: spec, Namespace: spec.GetNamespace().GetValue(), OperatorNamespace: spec.GetOperator().GetNamespace().GetValue(), Labels: parityLabels}

			if got, want := normalized(t, plan.chartValues), normalized(t, helmValues(spec)); got != want {
				t.Errorf("chart values differ:\nterraform: %s\npulumi:    %s", got, want)
			}

			wantSecrets := map[string][]string{}
			for _, s := range materializedSecrets(spec) {
				wantSecrets[s.Name] = sortedKeys(s.Data)
			}
			if got, want := normalized(t, plan.secretKeys), normalized(t, wantSecrets); got != want {
				t.Errorf("Secrets differ (name: key names):\nterraform: %s\npulumi:    %s", got, want)
			}

			stigmer := map[string]interface{}{
				"apiVersion": vars.StigmerAPIVersion,
				"kind":       vars.StigmerKind,
				"metadata":   stigmerMetadata(locals),
				"spec":       stigmerSpec(spec),
			}
			if got, want := plan.documents["Stigmer"], normalized(t, stigmer); got != want {
				t.Errorf("the Stigmer differs:\nterraform: %s\npulumi:    %s", got, want)
			}
			gate := map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "ConfigMap",
				"metadata":   map[string]interface{}{"name": vars.KeysGate, "namespace": locals.Namespace, "labels": parityLabels},
				"data":       keysGateData(locals.Namespace, spec.GetDatabase().GetBundled().GetRecoverFrom().GetServerName()),
			}
			if got, want := plan.documents["ConfigMap"], normalized(t, gate); got != want {
				t.Errorf("the restore check differs:\nterraform: %s\npulumi:    %s", got, want)
			}

			for name, value := range outputs(locals) {
				if fmt.Sprint(plan.outputs[name]) != value {
					t.Errorf("output %s: terraform %q, pulumi %q", name, plan.outputs[name], value)
				}
			}
		})
	}
}

// stigmerMetadata is the Stigmer's metadata as Resources declares it.
func stigmerMetadata(locals *Locals) map[string]interface{} {
	metadata := map[string]interface{}{"name": vars.StigmerName, "namespace": locals.Namespace, "labels": locals.Labels}
	if annotations := stigmerAnnotations(locals.Spec); annotations != nil {
		metadata["annotations"] = annotations
	}
	return metadata
}

var parityMetadata = &shared.CatalogObjectMetadata{Name: "parity"}

var parityLabels = map[string]string{
	"planton.ai/resource": "true",
	"planton.ai/name":     "parity",
	"planton.ai/kind":     kindName,
}

type plannedModule struct {
	chartValues map[string]interface{}
	secretKeys  map[string][]string
	documents   map[string]string
	outputs     map[string]interface{}
}

// terraformPlan plans iac/tf for the spec and returns what it would apply.
func terraformPlan(t *testing.T, spec *v1.KubernetesStigmerSpec, pluginCache string) plannedModule {
	t.Helper()
	dir := t.TempDir()
	sources, err := filepath.Glob(filepath.Join("..", "..", "tf", "*.tf"))
	if err != nil || len(sources) == 0 {
		t.Fatalf("the Terraform module's files: %v", err)
	}
	for _, source := range append(sources, filepath.Join("..", "..", "tf", ".terraform.lock.hcl")) {
		raw, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, filepath.Base(source)), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	manifest := &v1.KubernetesStigmer{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesStigmer",
		Metadata:   parityMetadata,
		Spec:       spec,
	}
	if err := generators.WriteVarFile(manifest, filepath.Join(dir, "terraform.tfvars")); err != nil {
		t.Fatal(err)
	}
	tofu(t, dir, pluginCache, "init", "-input=false", "-backend=false", "-lockfile=readonly")
	tofu(t, dir, pluginCache, "plan", "-input=false", "-out=plan.bin")
	var plan struct {
		PlannedValues struct {
			Outputs map[string]struct {
				Value interface{} `json:"value"`
			} `json:"outputs"`
			RootModule struct {
				Resources []struct {
					Type   string                 `json:"type"`
					Values map[string]interface{} `json:"values"`
				} `json:"resources"`
			} `json:"root_module"`
		} `json:"planned_values"`
	}
	if err := json.Unmarshal(tofu(t, dir, pluginCache, "show", "-json", "plan.bin"), &plan); err != nil {
		t.Fatal(err)
	}
	out := plannedModule{chartValues: map[string]interface{}{}, secretKeys: map[string][]string{}, documents: map[string]string{}, outputs: map[string]interface{}{}}
	for _, r := range plan.PlannedValues.RootModule.Resources {
		switch r.Type {
		case "helm_release":
			values := r.Values["values"].([]interface{})
			if err := yaml.Unmarshal([]byte(values[0].(string)), &out.chartValues); err != nil {
				t.Fatal(err)
			}
		case "kubernetes_secret_v1":
			metadata := r.Values["metadata"].([]interface{})[0].(map[string]interface{})
			data, _ := r.Values["data"].(map[string]interface{})
			keys := make([]string, 0, len(data))
			for k := range data {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			out.secretKeys[metadata["name"].(string)] = keys
		case "kubectl_manifest":
			var object map[string]interface{}
			if err := yaml.Unmarshal([]byte(r.Values["yaml_body"].(string)), &object); err != nil {
				t.Fatal(err)
			}
			out.documents[object["kind"].(string)] = normalized(t, object)
		}
	}
	for name, o := range plan.PlannedValues.Outputs {
		out.outputs[name] = o.Value
	}
	return out
}

func tofu(t *testing.T, dir, pluginCache string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("tofu", append(args, "-no-color")...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "TF_PLUGIN_CACHE_DIR="+pluginCache)
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		t.Fatalf("tofu %v: %v\n%s\n%s", args, err, out, stderr)
	}
	return out
}

// normalized is a value's canonical JSON, so integers and floats, and key
// order, never read as differences.
func normalized(t *testing.T, value interface{}) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var generic interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	canonical, _ := json.Marshal(generic)
	return string(canonical)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ---- the specs both tests render --------------------------------------------

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func i32(v int32) *int32 { return &v }

func paritySpec() *v1.KubernetesStigmerSpec {
	return &v1.KubernetesStigmerSpec{
		Namespace: literal("stigmer"),
		Version:   "0.3.2",
		Operator:  &v1.KubernetesStigmerOperator{Namespace: literal("stigmer-system"), CreateNamespace: true},
		Owner:     &v1.KubernetesStigmerOwner{Email: "owner@acme.example"},
		Runner: &v1.KubernetesStigmerRunner{Model: &v1.KubernetesStigmerModel{Source: &v1.KubernetesStigmerModel_Keys{
			Keys: &v1.KubernetesStigmerModelKeys{AnthropicApiKey: "sk-ant-example"},
		}}},
	}
}

func resources() *kubernetes.KubernetesCoreV1ApiResourceRequirements {
	return &kubernetes.KubernetesCoreV1ApiResourceRequirements{
		Requests: map[string]string{"cpu": "250m", "memory": "512Mi"},
		Limits:   map[string]string{"memory": "1Gi"},
	}
}

func sampleExtraEnv() *kubernetes.ContainerEnv {
	return &kubernetes.ContainerEnv{
		Variables: []*kubernetes.EnvVar{
			{Name: "LOG_FORMAT", Source: &kubernetes.EnvVar_Value{Value: "json"}},
			{Name: "TEAM", Source: &kubernetes.EnvVar_ConfigMapKeyRef{ConfigMapKeyRef: &kubernetes.ConfigMapKeyRef{Name: "team", Key: "name", Optional: true}}},
			{Name: "POD_IP", Source: &kubernetes.EnvVar_FieldRef{FieldRef: &kubernetes.ObjectFieldRef{FieldPath: "status.podIP"}}},
			{Name: "CPU_LIMIT", Source: &kubernetes.EnvVar_ResourceFieldRef{ResourceFieldRef: &kubernetes.ResourceFieldRef{Resource: "limits.cpu", Divisor: "1m"}}},
		},
		Secrets: []*kubernetes.SecretEnvVar{
			{Name: "MCP_TOKEN", Source: &kubernetes.SecretEnvVar_Value{Value: "token"}},
			{Name: "SEARCH_KEY", Source: &kubernetes.SecretEnvVar_SecretRef{SecretRef: &kubernetes.KubernetesSecretKeyRef{Name: "search", Key: "key"}}},
		},
	}
}

func s3(credentials *v1.KubernetesStigmerS3Credentials) *v1.KubernetesStigmerObjectStore {
	return &v1.KubernetesStigmerObjectStore{S3: &v1.KubernetesStigmerS3{
		Bucket:         literal("acme-stigmer"),
		Region:         strPtr("eu-west-1"),
		Endpoint:       literal("http://seaweedfs-s3.storage.svc.cluster.local:8333"),
		Credentials:    credentials,
		RunnerEndpoint: "http://seaweedfs.example:8333",
	}}
}

func strPtr(s string) *string { return &s }

func bundledEverything() *v1.KubernetesStigmerSpec {
	spec := paritySpec()
	tolerationSeconds := int64(60)
	exists, noExecute := "Exists", "NoExecute"
	spec.Operator = &v1.KubernetesStigmerOperator{
		Namespace:        literal("stigmer-system"),
		CreateNamespace:  true,
		ImageRepository:  "mirror.example.com/stigmer/stigmer-operator",
		ImagePullPolicy:  "Always",
		ImagePullSecrets: []string{"mirror-pull"},
		Replicas:         i32(2),
		Resources:        &v1.KubernetesStigmerOperatorResources{Requests: map[string]string{"cpu": "100m"}, Limits: map[string]string{"memory": "1Gi"}},
		LogLevel:         "debug",
		NodeSelector:     map[string]string{"kubernetes.io/os": "linux"},
		Tolerations:      []*kubernetes.KubernetesCoreV1ApiToleration{{Key: "platform", Operator: &exists, Effect: &noExecute, TolerationSeconds: &tolerationSeconds}},
	}
	spec.Image = &v1.KubernetesStigmerImage{Registry: "mirror.example.com/stigmer", PullPolicy: "IfNotPresent", PullSecrets: []string{"mirror-pull"}}
	spec.Owner = &v1.KubernetesStigmerOwner{Email: "owner@acme.example", Name: "Priya"}
	spec.Server = &v1.KubernetesStigmerServer{
		Resources:        resources(),
		Storage:          &v1.KubernetesStigmerVolume{Size: "8Gi", StorageClassName: "fast"},
		Help:             &v1.KubernetesStigmerHelp{Url: "https://help.acme.example", Label: "Ask platform"},
		InternalAccounts: []string{"*@acme.example"},
		Egress:           &v1.KubernetesStigmerServerEgress{Fenced: true, AllowCidrs: []string{"10.20.0.0/16"}},
		ExtraEnv:         sampleExtraEnv(),
	}
	spec.Runner.Resources = resources()
	spec.Runner.Storage = &v1.KubernetesStigmerVolume{Size: "20Gi"}
	spec.Runner.ExtraEnv = &kubernetes.ContainerEnv{Secrets: []*kubernetes.SecretEnvVar{{Name: "RUNNER_TOKEN", Source: &kubernetes.SecretEnvVar_Value{Value: "x"}}}}
	spec.Runner.Model.GetKeys().AnthropicBaseUrl = "http://models.ai.svc.cluster.local:8080"
	spec.Database = &v1.KubernetesStigmerDatabase{Postgresql: &v1.KubernetesStigmerDatabase_Bundled{Bundled: &v1.KubernetesStigmerBundledPostgres{
		Instances:   i32(3),
		Storage:     &v1.KubernetesStigmerVolume{Size: "50Gi"},
		Resources:   resources(),
		Backup:      &v1.KubernetesStigmerPostgresBackup{Schedule: "0 0 3 * * *", Retention: "14d"},
		RecoverFrom: &v1.KubernetesStigmerRecoverFrom{ServerName: "stigmer-20261120", TargetTime: "2026-11-20T08:00:00Z"},
	}}}
	spec.Temporal = &v1.KubernetesStigmerTemporal{Mode: &v1.KubernetesStigmerTemporal_Bundled{Bundled: &v1.KubernetesStigmerBundledTemporal{Resources: resources()}}}
	spec.ObjectStore = s3(&v1.KubernetesStigmerS3Credentials{Source: &v1.KubernetesStigmerS3Credentials_AccessKey{
		AccessKey: &v1.KubernetesStigmerS3AccessKey{AccessKeyId: "stigmer", SecretAccessKey: "s3-secret"},
	}})
	spec.Ingress = &v1.KubernetesStigmerIngress{
		Hostname:           "stigmer.acme.example",
		Route:              &v1.KubernetesStigmerIngress_Tls{Tls: &v1.KubernetesStigmerIngressTls{Certificate: &v1.KubernetesStigmerIngressTls_Issuer{Issuer: &v1.KubernetesStigmerIssuerRef{Name: literal("letsencrypt"), Kind: "ClusterIssuer"}}}},
		ClassName:          "nginx",
		Annotations:        map[string]string{"nginx.ingress.kubernetes.io/backend-protocol": "GRPC"},
		ServiceAnnotations: map[string]string{"traefik.ingress.kubernetes.io/service.serversscheme": "h2c"},
	}
	spec.Identity = &v1.KubernetesStigmerIdentity{Provider: &v1.KubernetesStigmerIdentity_Bundled{Bundled: &v1.KubernetesStigmerBundledIdentity{
		SignUpEmailDomains: []string{"acme.example"},
		Smtp: &v1.KubernetesStigmerSmtp{Host: "smtp.acme.example", Port: i32(587), From: "stigmer@acme.example", Starttls: true,
			Credentials: &v1.KubernetesStigmerSmtpCredentials{Source: &v1.KubernetesStigmerSmtpCredentials_Login{Login: &v1.KubernetesStigmerSmtpLogin{Username: "stigmer", Password: "mail"}}}},
	}}}
	spec.Sandboxes = &v1.KubernetesStigmerSandboxes{Driver: "substrate", IdleSeconds: i32(300), Substrate: &v1.KubernetesStigmerSubstrate{
		Workers:              i32(2),
		WorkerResources:      resources(),
		WorkerNodeSelector:   map[string]string{"pool": "agents"},
		WorkerTolerations:    []*kubernetes.KubernetesCoreV1ApiToleration{{Key: "agents", Operator: &exists}},
		SandboxCpuMillicores: i32(1000),
		SandboxMemoryMib:     i32(1024),
		EgressHttp:           []string{"models.ai.svc.cluster.local:8080"},
	}}
	spec.Prerequisites = &v1.KubernetesStigmerPrerequisites{PostgresOperator: "auto", PostgresBackupPlugin: "skip", AgentSandbox: "auto", Substrate: "auto"}
	spec.UpgradeWithoutBackupTo = "0.3.2"
	return spec
}

func externalValues() *v1.KubernetesStigmerSpec {
	spec := paritySpec()
	spec.Database = &v1.KubernetesStigmerDatabase{Postgresql: &v1.KubernetesStigmerDatabase_External{External: &v1.KubernetesStigmerExternalPostgres{
		Host: literal("db-rw.data.svc.cluster.local"), Port: i32(5432), User: "stigmer", Database: strPtr("stigmer"),
		Password: &v1.KubernetesStigmerPassword{Source: &v1.KubernetesStigmerPassword_Value{Value: "db-password"}},
		SslMode:  "require",
	}}}
	spec.Temporal = &v1.KubernetesStigmerTemporal{Mode: &v1.KubernetesStigmerTemporal_External{External: &v1.KubernetesStigmerExternalTemporal{
		HostPort:  literal("acme.tmprl.cloud:7233"),
		Namespace: strPtr("acme.stigmer"),
		Tls: &v1.KubernetesStigmerTemporalTls{ServerName: "acme.tmprl.cloud", Certificates: &v1.KubernetesStigmerTemporalTls_Pem{Pem: &v1.KubernetesStigmerTemporalTlsPem{
			CaCertificate: "-----BEGIN CERTIFICATE-----\nCA\n-----END CERTIFICATE-----\n", ClientCertificate: "cert", ClientPrivateKey: "key",
		}}},
		ApiKey: &v1.KubernetesStigmerSecretValue{Source: &v1.KubernetesStigmerSecretValue_Value{Value: "temporal-key"}},
	}}}
	spec.ObjectStore = s3(&v1.KubernetesStigmerS3Credentials{Source: &v1.KubernetesStigmerS3Credentials_ExistingSecretName{ExistingSecretName: literal("bucket-keys")}})
	spec.Ingress = &v1.KubernetesStigmerIngress{Hostname: "stigmer.acme.example", Route: &v1.KubernetesStigmerIngress_Gateway{
		Gateway: &v1.KubernetesStigmerGatewayRef{Name: literal("edge"), Namespace: literal("gateways"), SectionName: "https"},
	}}
	spec.Identity = &v1.KubernetesStigmerIdentity{Provider: &v1.KubernetesStigmerIdentity_Oidc{Oidc: &v1.KubernetesStigmerOidc{
		Issuer: "https://auth.acme.example/realms/stigmer", Audience: "stigmer", ConsoleClientId: "console", NativeClientId: "native", DiscoveryUrl: "http://keycloak.auth.svc:8080/realms/stigmer/.well-known/openid-configuration",
	}}}
	spec.Enterprise = &v1.KubernetesStigmerEnterprise{
		License:                    &v1.KubernetesStigmerSecretValue{Source: &v1.KubernetesStigmerSecretValue_Value{Value: "license"}},
		OrganizationCreators:       []string{"ida_01jexampleexampleexampleex"},
		AccessHistoryRetentionDays: i32(90),
		OutboundEgress:             "relaxed",
	}
	spec.Sandboxes = &v1.KubernetesStigmerSandboxes{Driver: "agentSandbox", AgentSandbox: &v1.KubernetesStigmerAgentSandbox{EgressAllowCidrs: []string{"10.20.0.0/16"}}}
	spec.Server = &v1.KubernetesStigmerServer{Artifacts: &v1.KubernetesStigmerVolume{Size: "30Gi"}}
	return spec
}

func externalSecrets() *v1.KubernetesStigmerSpec {
	spec := externalValues()
	spec.Runner.Model = &v1.KubernetesStigmerModel{Source: &v1.KubernetesStigmerModel_ExistingSecretName{ExistingSecretName: literal("model-keys")}}
	spec.Database.GetExternal().Password = &v1.KubernetesStigmerPassword{Source: &v1.KubernetesStigmerPassword_ExistingSecret{
		ExistingSecret: &v1.KubernetesStigmerPostgresPasswordSecret{Name: literal("stigmer-db-app")},
	}}
	spec.Temporal.GetExternal().Tls = &v1.KubernetesStigmerTemporalTls{Certificates: &v1.KubernetesStigmerTemporalTls_ExistingSecret{
		ExistingSecret: &v1.KubernetesStigmerTemporalTlsSecret{Name: literal("temporal-tls"), CaKey: "ca.crt", CertificateKey: "tls.crt", PrivateKeyKey: "tls.key"},
	}}
	spec.Temporal.GetExternal().ApiKey = &v1.KubernetesStigmerSecretValue{Source: &v1.KubernetesStigmerSecretValue_ExistingSecret{
		ExistingSecret: &v1.KubernetesStigmerSecretKeyRef{Name: literal("temporal-key"), Key: "key"},
	}}
	spec.Ingress = &v1.KubernetesStigmerIngress{Hostname: "stigmer.acme.example", Route: &v1.KubernetesStigmerIngress_Tls{
		Tls: &v1.KubernetesStigmerIngressTls{Certificate: &v1.KubernetesStigmerIngressTls_SecretName{SecretName: "stigmer-tls"}},
	}}
	spec.Enterprise.License = &v1.KubernetesStigmerSecretValue{Source: &v1.KubernetesStigmerSecretValue_ExistingSecret{
		ExistingSecret: &v1.KubernetesStigmerSecretKeyRef{Name: literal("license"), Key: "STIGMER_LICENSE_KEY"},
	}}
	return spec
}
