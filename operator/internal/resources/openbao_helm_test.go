package resources

import (
	neturl "net/url"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// testOpenBAOOptions is the built-in-seal default every values test starts
// from; seal tests set Seal on top of it.
func testOpenBAOOptions() OpenBAOHelmOptions {
	return OpenBAOHelmOptions{
		CRName:                    "my-planton",
		Namespace:                 "planton",
		StoragePasswordSecretName: PostgreSQLVaultRoleSecretName("my-planton"),
	}
}

func TestOpenBAOHelmValues_FullnameOverride(t *testing.T) {
	vals := OpenBAOHelmValues(testOpenBAOOptions())
	if vals["fullnameOverride"] != "my-planton-openbao" {
		t.Errorf("expected fullnameOverride my-planton-openbao, got %v", vals["fullnameOverride"])
	}
}

func TestOpenBAOHelmValues_Global(t *testing.T) {
	vals := OpenBAOHelmValues(testOpenBAOOptions())
	global, ok := vals["global"].(map[string]any)
	if !ok {
		t.Fatal("expected global to be a map")
	}
	if global["enabled"] != true {
		t.Error("expected global enabled")
	}
	if global["tlsDisable"] != true {
		t.Error("expected global tlsDisable true")
	}
}

func TestOpenBAOHelmValues_Standalone(t *testing.T) {
	vals := OpenBAOHelmValues(testOpenBAOOptions())
	server, ok := vals["server"].(map[string]any)
	if !ok {
		t.Fatal("expected server to be a map")
	}
	standalone, ok := server["standalone"].(map[string]any)
	if !ok {
		t.Fatal("expected standalone to be a map")
	}
	if standalone["enabled"] != true {
		t.Error("expected standalone enabled")
	}
	config, ok := standalone["config"].(string)
	if !ok || config == "" {
		t.Fatal("expected non-empty standalone config string")
	}
	if strings.Contains(config, `seal "`) {
		t.Errorf("the built-in seal is simply not configured; found a seal stanza in:\n%s", config)
	}
}

func TestOpenBAOHelmValues_HADisabled(t *testing.T) {
	vals := OpenBAOHelmValues(testOpenBAOOptions())
	server := vals["server"].(map[string]any)
	ha, ok := server["ha"].(map[string]any)
	if !ok {
		t.Fatal("expected ha to be a map")
	}
	if ha["enabled"] != false {
		t.Error("expected ha disabled")
	}
}

// The vault's storage is the platform's PostgreSQL: the config names the
// backend and its pool cap and NOTHING credential-bearing (the chart renders
// it into a ConfigMap); the connection URL is a plain variable with no
// password in it; the password is projected from the vault role's own
// credential Secret -- the projection every consumer of the platform's
// database uses -- and the chart's volume is off.
func TestOpenBAOHelmValues_Storage(t *testing.T) {
	vals := OpenBAOHelmValues(testOpenBAOOptions())
	server := vals["server"].(map[string]any)

	config := server["standalone"].(map[string]any)["config"].(string)
	if !strings.Contains(config, `storage "postgresql"`) {
		t.Errorf("config must name the postgresql backend, got:\n%s", config)
	}
	if strings.Contains(config, `storage "file"`) || strings.Contains(config, "/openbao/data") {
		t.Errorf("config must not name a file backend or a data path, got:\n%s", config)
	}
	if !strings.Contains(config, `max_parallel = "32"`) {
		t.Errorf("config must cap the backend's pool against the cluster's shared connection budget, got:\n%s", config)
	}
	for _, forbidden := range []string{"connection_url", "postgres://", "password"} {
		if strings.Contains(config, forbidden) {
			t.Errorf("config must carry nothing credential-bearing; found %q in:\n%s", forbidden, config)
		}
	}

	dataStorage, ok := server["dataStorage"].(map[string]any)
	if !ok || dataStorage["enabled"] != false {
		t.Errorf("dataStorage must be disabled -- the vault has no volume; got %v", server["dataStorage"])
	}

	plain, ok := server["extraEnvironmentVars"].(map[string]any)
	if !ok {
		t.Fatalf("expected plain environment variables, got %v", server["extraEnvironmentVars"])
	}
	url, _ := plain[OpenBAOStorageURLEnv].(string)
	if url != "postgres://openbao@my-planton-postgres-rw.planton.svc.cluster.local:5432/openbao?sslmode=disable" {
		t.Errorf("the storage URL must be a plain variable naming role, host, and database with no password, got %q", url)
	}
	parsed, err := neturl.Parse(url)
	if err != nil {
		t.Fatalf("the storage URL must parse: %v", err)
	}
	if _, has := parsed.User.Password(); has || parsed.User.Username() != PostgreSQLVaultRole {
		t.Errorf("the storage URL must name the role and carry no password, got %q", url)
	}

	env, ok := server["extraSecretEnvironmentVars"].([]any)
	if !ok || len(env) != 1 {
		t.Fatalf("expected exactly one secret environment variable under the built-in seal, got %v", server["extraSecretEnvironmentVars"])
	}
	entry := env[0].(map[string]any)
	if entry["envName"] != OpenBAOStoragePasswordEnv || entry["secretName"] != "my-planton-postgres-openbao" || entry["secretKey"] != BasicAuthPasswordKey {
		t.Errorf("the storage password must reach the process as %s from the role Secret's %s key, got %v", OpenBAOStoragePasswordEnv, BasicAuthPasswordKey, entry)
	}
}

// The connection URL is composed in one place: the vault's own role and
// database on the platform's primary, with the platform's in-cluster TLS
// posture, and no password -- that is a separate variable.
func TestOpenBAOStorageURL(t *testing.T) {
	got := OpenBAOStorageURL("my-planton", "planton")
	want := "postgres://openbao@my-planton-postgres-rw.planton.svc.cluster.local:5432/openbao?sslmode=disable"
	if got != want {
		t.Errorf("URL\n got: %s\nwant: %s", got, want)
	}
}

func TestOpenBAOHelmValues_UI(t *testing.T) {
	vals := OpenBAOHelmValues(testOpenBAOOptions())
	ui, ok := vals["ui"].(map[string]any)
	if !ok {
		t.Fatal("expected ui to be a map")
	}
	if ui["enabled"] != true {
		t.Error("expected ui enabled")
	}
}

func TestOpenBAOHelmValues_InjectorDisabled(t *testing.T) {
	vals := OpenBAOHelmValues(testOpenBAOOptions())
	injector, ok := vals["injector"].(map[string]any)
	if !ok {
		t.Fatal("expected injector to be a map")
	}
	if injector["enabled"] != false {
		t.Error("expected injector disabled")
	}
}

func TestOpenBAOReleaseName(t *testing.T) {
	if got := openbaoReleaseName("my-planton"); got != "my-planton-openbao" {
		t.Errorf("expected my-planton-openbao, got %s", got)
	}
}

func TestOpenBAOInitSecretName(t *testing.T) {
	if got := OpenBAOInitSecretName("my-planton"); got != "my-planton-openbao-init" {
		t.Errorf("expected my-planton-openbao-init, got %s", got)
	}
}

// The init Secret's note must explain the key material in plain language
// for the seal the vault runs under and for who owns the Secret: what each
// key does, the cost of deleting it, and what keeps it. Deployed by default,
// this Secret exists on every install.
func TestOpenBAOInitSecretNote(t *testing.T) {
	gcp := &OpenBAOSealOptions{GcpKms: &OpenBAOGcpKmsSealOptions{Project: "p", Region: "global", KeyRing: "ring", CryptoKey: "key"}}
	cases := []struct {
		name        string
		seal        *OpenBAOSealOptions
		adoptersOwn bool
		want        []string
		forbid      []string
	}{
		{"shamir, operator-owned", nil, false,
			[]string{"Unseal keys", "my-planton-openbao", "unseals it with them", "stays locked", "deleted with the platform", "spec.vault.initSecretName"},
			[]string{"Recovery keys", "never deletes"}},
		{"shamir, adopter-owned", nil, true,
			[]string{"Unseal keys", "You own this Secret", "never deletes it", "deleting the PlantonPlatform leaves it standing", "namespace the declaration owns", "keep a copy"},
			[]string{"Recovery keys", "deleted with the platform"}},
		{"cloud seal, adopter-owned", gcp, true,
			[]string{"Recovery keys", "Google Cloud KMS key", "opens itself", "never unseal anything", "break-glass", "You own this Secret", "keep a copy"},
			[]string{"Unseal keys", "stays locked"}},
		{"cloud seal, operator-owned", gcp, false,
			[]string{"Recovery keys", "break-glass is gone", "deleted with the platform"},
			[]string{"Unseal keys"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			note := OpenBAOInitSecretNote("my-planton", tc.seal, tc.adoptersOwn)
			for _, want := range tc.want {
				if !strings.Contains(note, want) {
					t.Errorf("note must say %q, got: %s", want, note)
				}
			}
			for _, forbid := range tc.forbid {
				if strings.Contains(note, forbid) {
					t.Errorf("note must not say %q here, got: %s", forbid, note)
				}
			}
		})
	}
}

// Deployed on every default install, the vault must schedule honestly:
// explicit requests, a memory limit, and (house pattern) no CPU limit.
func TestOpenBAOHelmValues_Resources(t *testing.T) {
	vals := OpenBAOHelmValues(testOpenBAOOptions())
	server := vals["server"].(map[string]any)
	res, ok := server["resources"].(map[string]any)
	if !ok {
		t.Fatal("expected server.resources to be set -- the chart ships none")
	}
	requests := res["requests"].(map[string]any)
	if requests["cpu"] == "" || requests["memory"] == "" {
		t.Errorf("expected cpu+memory requests, got %v", requests)
	}
	limits := res["limits"].(map[string]any)
	if limits["memory"] == "" {
		t.Errorf("expected a memory limit, got %v", limits)
	}
	if _, hasCPULimit := limits["cpu"]; hasCPULimit {
		t.Error("no CPU limit by design (requests-only, the house pattern)")
	}
}

func TestOpenBAOServiceHost(t *testing.T) {
	expected := "my-planton-openbao.planton-system.svc.cluster.local"
	if got := OpenBAOServiceHost("my-planton", "planton-system"); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestOpenBAOAPIAddr(t *testing.T) {
	expected := "http://my-planton-openbao.default.svc.cluster.local:8200"
	if got := OpenBAOAPIAddr("my-planton", "default"); got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

// renderOpenBAO renders the chart for the options and returns the server
// container, the ServiceAccount, and the config ConfigMap's HCL.
func renderOpenBAO(t *testing.T, opts OpenBAOHelmOptions) (container map[string]any, serviceAccount *unstructured.Unstructured, config string, all []*unstructured.Unstructured) {
	t.Helper()
	objs, err := RenderHelmChart(LoadOpenBAOChart(), opts.CRName+"-openbao", opts.Namespace, OpenBAOHelmValues(opts))
	if err != nil {
		t.Fatalf("failed to render OpenBAO chart: %v", err)
	}
	for _, obj := range objs {
		switch obj.GetKind() {
		case "StatefulSet":
			containers, _, _ := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
			if len(containers) == 0 {
				t.Fatal("rendered StatefulSet has no containers")
			}
			container = containers[0].(map[string]any)
		case "ServiceAccount":
			serviceAccount = obj
		case "ConfigMap":
			if data, _, _ := unstructured.NestedStringMap(obj.Object, "data"); data["extraconfig-from-values.hcl"] != "" {
				config = data["extraconfig-from-values.hcl"]
			}
		}
	}
	if container == nil {
		t.Fatal("expected a StatefulSet in the rendered objects")
	}
	return container, serviceAccount, config, objs
}

func envByName(envs []any) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, raw := range envs {
		if e, ok := raw.(map[string]any); ok {
			out[e["name"].(string)] = e
		}
	}
	return out
}

func TestOpenBAOHelmValues_ChartRendering(t *testing.T) {
	if len(LoadOpenBAOChart()) == 0 {
		t.Fatal("OpenBAO chart data is empty")
	}
	opts := OpenBAOHelmOptions{CRName: "test", Namespace: "default", StoragePasswordSecretName: PostgreSQLVaultRoleSecretName("test")}
	c, _, config, objs := renderOpenBAO(t, opts)

	kinds := make(map[string]bool)
	for _, obj := range objs {
		kinds[obj.GetKind()] = true
		if obj.GetKind() == "ClusterRoleBinding" && obj.GetName() == "test-openbao-server-binding" {
			t.Error("auth-delegator ClusterRoleBinding must be disabled -- the operator RBAC cannot grant tokenreview permissions")
		}
	}
	if !kinds["StatefulSet"] || !kinds["Service"] {
		t.Errorf("expected a StatefulSet and a Service in the rendered objects, got %v", kinds)
	}
	if config == "" || !strings.Contains(config, `storage "postgresql"`) {
		t.Errorf("the chart must render the operator's config into its ConfigMap, got %q", config)
	}

	// Render-level assertion (the inert-values-key lesson): the chart must
	// actually thread server.resources onto the container.
	cpu, _, _ := unstructured.NestedString(c, "resources", "requests", "cpu")
	memLimit, _, _ := unstructured.NestedString(c, "resources", "limits", "memory")
	if cpu == "" || memLimit == "" {
		t.Errorf("rendered container must carry the requests + memory limit, got resources=%v", c["resources"])
	}

	// The storage seams, at the render: the URL is a plain variable with no
	// password, the password is projected from the role Secret's key, and
	// the data mount the file backend needed is gone with the volume.
	envs, _, _ := unstructured.NestedSlice(c, "env")
	byName := envByName(envs)
	urlEnv := byName[OpenBAOStorageURLEnv]
	if urlEnv == nil {
		t.Fatalf("rendered container carries no %s variable; env=%v", OpenBAOStorageURLEnv, envs)
	}
	if v, _ := urlEnv["value"].(string); v != OpenBAOStorageURL("test", "default") || strings.Contains(v, "password") {
		t.Errorf("%s must be the plain, password-less URL, got %v", OpenBAOStorageURLEnv, urlEnv)
	}
	passEnv := byName[OpenBAOStoragePasswordEnv]
	if passEnv == nil {
		t.Fatalf("rendered container carries no %s variable; env=%v", OpenBAOStoragePasswordEnv, envs)
	}
	secretName, _, _ := unstructured.NestedString(passEnv, "valueFrom", "secretKeyRef", "name")
	secretKey, _, _ := unstructured.NestedString(passEnv, "valueFrom", "secretKeyRef", "key")
	if secretName != "test-postgres-openbao" || secretKey != BasicAuthPasswordKey {
		t.Errorf("%s must be projected from the role Secret's %s key, got %v", OpenBAOStoragePasswordEnv, BasicAuthPasswordKey, passEnv)
	}
	mounts, _, _ := unstructured.NestedSlice(c, "volumeMounts")
	for _, raw := range mounts {
		if m, ok := raw.(map[string]any); ok && m["mountPath"] == "/openbao/data" {
			t.Errorf("the vault has no volume, yet the container mounts a data path: %v", m)
		}
	}
	for _, obj := range objs {
		if obj.GetKind() != "StatefulSet" {
			continue
		}
		if vcts, found, _ := unstructured.NestedSlice(obj.Object, "spec", "volumeClaimTemplates"); found && len(vcts) > 0 {
			t.Errorf("the vault has no volume, yet the StatefulSet carries claim templates: %v", vcts)
		}
	}
}

// Every seal arm at the render: the stanza lands in the config ConfigMap
// with the key it names and nothing else; identifiers ride plain variables;
// the credential rides a variable projected from the named Secret; the
// keyless identity rides the ServiceAccount's annotations; and no credential
// value appears anywhere in any rendered document.
func TestOpenBAOHelmValues_SealRendering(t *testing.T) {
	const credsSecret = "vault-seal-creds"
	cases := []struct {
		name        string
		seal        *OpenBAOSealOptions
		annotations map[string]string
		wantStanza  []string
		wantPlain   map[string]string
		wantSecret  map[string]string // env name -> key
	}{
		{
			name: "aws kms, keyless through IRSA",
			seal: &OpenBAOSealOptions{AwsKms: &OpenBAOAwsKmsSealOptions{Region: "us-east-1", KmsKeyID: "alias/vault-unseal"}},
			annotations: map[string]string{
				"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/vault-unseal",
			},
			wantStanza: []string{`seal "awskms" {`, `region = "us-east-1"`, `kms_key_id = "alias/vault-unseal"`},
			wantPlain:  map[string]string{OpenBAOSealEnvAwsRegion: "us-east-1"},
		},
		{
			name:       "aws kms, static key pair",
			seal:       &OpenBAOSealOptions{AwsKms: &OpenBAOAwsKmsSealOptions{Region: "eu-west-1", KmsKeyID: "1234abcd", AccessKeyID: "AKIAEXAMPLE", CredentialsSecretName: credsSecret}},
			wantStanza: []string{`seal "awskms" {`, `kms_key_id = "1234abcd"`},
			wantPlain:  map[string]string{OpenBAOSealEnvAwsRegion: "eu-west-1", OpenBAOSealEnvAwsAccessKeyID: "AKIAEXAMPLE"},
			wantSecret: map[string]string{OpenBAOSealEnvAwsSecretAccessKey: OpenBAOSealEnvAwsSecretAccessKey},
		},
		{
			name: "gcp kms, keyless through workload identity",
			seal: &OpenBAOSealOptions{GcpKms: &OpenBAOGcpKmsSealOptions{Project: "planton-foundation", Region: "global", KeyRing: "vault-unseal", CryptoKey: "management"}},
			annotations: map[string]string{
				"iam.gke.io/gcp-service-account": "vault-unseal@planton-foundation.iam.gserviceaccount.com",
			},
			wantStanza: []string{`seal "gcpckms" {`, `project = "planton-foundation"`, `region = "global"`, `key_ring = "vault-unseal"`, `crypto_key = "management"`},
		},
		{
			name:       "azure key vault, service principal",
			seal:       &OpenBAOSealOptions{AzureKeyVault: &OpenBAOAzureKeyVaultSealOptions{VaultName: "planton-kv", KeyName: "vault-unseal", TenantID: "tenant-1", ClientID: "client-1", CredentialsSecretName: credsSecret}},
			wantStanza: []string{`seal "azurekeyvault" {`, `vault_name = "planton-kv"`, `key_name = "vault-unseal"`, `tenant_id = "tenant-1"`, `client_id = "client-1"`},
			wantSecret: map[string]string{OpenBAOSealEnvAzureClientSecret: OpenBAOSealEnvAzureClientSecret},
		},
		{
			name:       "transit",
			seal:       &OpenBAOSealOptions{Transit: &OpenBAOTransitSealOptions{Address: "https://bao.example.com:8200", KeyName: "planton-unseal", CredentialsSecretName: credsSecret}},
			wantStanza: []string{`seal "transit" {`, `address = "https://bao.example.com:8200"`, `key_name = "planton-unseal"`, `mount_path = "transit/"`},
			wantSecret: map[string]string{OpenBAOSealEnvTransitToken: OpenBAOSealEnvTransitToken},
		},
	}
	// Values a leak would surface as. None is ever handed to the renderer
	// as a value -- the CR names a Secret -- so the grep is a guard on the
	// rendering's shape, not on these strings finding a way in.
	forbidden := []string{"client_secret", "secret_key", "token =", "credentials ="}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := testOpenBAOOptions()
			opts.Seal = tc.seal
			opts.ServiceAccountAnnotations = tc.annotations
			c, sa, config, objs := renderOpenBAO(t, opts)

			for _, want := range tc.wantStanza {
				if !strings.Contains(config, want) {
					t.Errorf("config must carry %q, got:\n%s", want, config)
				}
			}
			for _, f := range forbidden {
				if strings.Contains(config, f) {
					t.Errorf("config must carry no credential parameter; found %q in:\n%s", f, config)
				}
			}

			envs, _, _ := unstructured.NestedSlice(c, "env")
			byName := envByName(envs)
			for name, value := range tc.wantPlain {
				e := byName[name]
				if e == nil || e["value"] != value {
					t.Errorf("plain variable %s must be %q, got %v", name, value, e)
				}
			}
			for name, key := range tc.wantSecret {
				e := byName[name]
				if e == nil {
					t.Fatalf("credential variable %s missing; env=%v", name, envs)
				}
				secretName, _, _ := unstructured.NestedString(e, "valueFrom", "secretKeyRef", "name")
				secretKey, _, _ := unstructured.NestedString(e, "valueFrom", "secretKeyRef", "key")
				if secretName != credsSecret || secretKey != key {
					t.Errorf("%s must be projected from %s's %s key, got %v", name, credsSecret, key, e)
				}
				if _, plain := e["value"]; plain {
					t.Errorf("%s must never be a plain value, got %v", name, e)
				}
			}
			if tc.wantSecret == nil {
				for name := range byName {
					if name == OpenBAOSealEnvAwsSecretAccessKey || name == OpenBAOSealEnvAzureClientSecret || name == OpenBAOSealEnvTransitToken {
						t.Errorf("a keyless arm must project no credential variable, found %s", name)
					}
				}
			}

			if len(tc.annotations) > 0 {
				if sa == nil {
					t.Fatal("expected a ServiceAccount in the rendered objects")
				}
				for k, v := range tc.annotations {
					if sa.GetAnnotations()[k] != v {
						t.Errorf("ServiceAccount must carry %s=%s, got %v", k, v, sa.GetAnnotations())
					}
				}
			}

			// The whole render, as YAML, must not carry a credential
			// value: the only Secret data the chart could render is none,
			// because every credential is a reference.
			for _, obj := range objs {
				raw, err := yaml.Marshal(obj.Object)
				if err != nil {
					t.Fatal(err)
				}
				if obj.GetKind() == "Secret" {
					t.Errorf("the chart must render no Secret; every credential is a reference, found %s", obj.GetName())
				}
				for _, f := range forbidden {
					if strings.Contains(string(raw), f) {
						t.Errorf("%s/%s carries %q", obj.GetKind(), obj.GetName(), f)
					}
				}
			}
		})
	}
}
