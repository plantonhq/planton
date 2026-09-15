package module

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	kubernetesopenbaov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesopenbao/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"sigs.k8s.io/yaml"
)

// The rendering contract for the storage engine: which chart mode each
// engine drives, what the configuration string says, where the data volume
// is claimed, how the PostgreSQL connection reaches the server, and what
// must NEVER appear anywhere the module wrote. The Terraform module renders
// the SAME documents from the same spec (iac/tf/locals.tf); the expectations
// below are the cross-engine contract, and a run with OPENBAO_RENDER_DIR set
// writes each fixture's documents to disk so the Terraform rendering can be
// diffed against them byte for byte.

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(i int32) *int32 { return &i }
func strPtr(s string) *string { return &s }

// localsFor builds the module's locals the way Resources does, from a stack
// input carrying the given spec. References arrive at the module already
// resolved to values, so the fixtures use literals throughout. The manifest
// loader fills a present block's declared scalar defaults before either
// engine sees the spec (server.replicas 1, log_level info, log_format
// standard), so the fixtures are completed the same way — the rendered
// documents must match what the Terraform module renders from the loaded
// manifest, byte for byte.
func localsFor(spec *kubernetesopenbaov1alpha1.KubernetesOpenBaoSpec) *Locals {
	spec.Namespace = literal("openbao")
	spec.CreateNamespace = true
	if s := spec.Server; s != nil {
		if s.Replicas == nil {
			s.Replicas = int32Ptr(1)
		}
		if s.LogLevel == nil {
			s.LogLevel = strPtr("info")
		}
		if s.LogFormat == nil {
			s.LogFormat = strPtr("standard")
		}
	}
	return initializeLocals(nil, &kubernetesopenbaov1alpha1.KubernetesOpenBaoStackInput{
		Target: &kubernetesopenbaov1alpha1.KubernetesOpenBao{
			Metadata: &shared.CloudResourceMetadata{Name: "vault"},
			Spec:     spec,
		},
	})
}

func postgresqlStorage() *kubernetesopenbaov1alpha1.KubernetesOpenBaoServer_Postgresql {
	return &kubernetesopenbaov1alpha1.KubernetesOpenBaoServer_Postgresql{
		Postgresql: &kubernetesopenbaov1alpha1.KubernetesOpenBaoPostgresqlStorage{
			Host:     literal("identity-db-rw"),
			Database: "openbao",
			PasswordSecret: &kubernetesopenbaov1alpha1.KubernetesOpenBaoPostgresqlPasswordSecret{
				SecretName: literal("identity-db-app"),
			},
		},
	}
}

// fixtures are the seven shapes the cross-engine proof renders; the YAML
// manifests the Terraform side loads mirror these field for field.
func fixtures() map[string]*kubernetesopenbaov1alpha1.KubernetesOpenBaoSpec {
	pgFull := postgresqlStorage()
	pgFull.Postgresql.Port = int32Ptr(5433)
	pgFull.Postgresql.Username = strPtr("identity")
	pgFull.Postgresql.PasswordSecret.SecretKey = strPtr("pw")
	pgFull.Postgresql.SslMode = strPtr("disable")
	pgFull.Postgresql.MaxParallel = int32Ptr(32)
	return map[string]*kubernetesopenbaov1alpha1.KubernetesOpenBaoSpec{
		"raft-1": {},
		"raft-3-volume": {Server: &kubernetesopenbaov1alpha1.KubernetesOpenBaoServer{
			Storage: &kubernetesopenbaov1alpha1.KubernetesOpenBaoServer_Raft{Raft: &kubernetesopenbaov1alpha1.KubernetesOpenBaoRaftStorage{
				DataStorage: &kubernetesopenbaov1alpha1.KubernetesOpenBaoVolume{Size: strPtr("20Gi"), StorageClass: literal("fast")},
			}},
			Replicas: int32Ptr(3),
		}},
		"postgresql-1": {Server: &kubernetesopenbaov1alpha1.KubernetesOpenBaoServer{Storage: postgresqlStorage()}},
		"postgresql-3-full": {Server: &kubernetesopenbaov1alpha1.KubernetesOpenBaoServer{
			Storage:      pgFull,
			Replicas:     int32Ptr(3),
			AuditStorage: &kubernetesopenbaov1alpha1.KubernetesOpenBaoVolume{Size: strPtr("5Gi")},
		}},
		"dev": {Server: &kubernetesopenbaov1alpha1.KubernetesOpenBaoServer{Dev: &kubernetesopenbaov1alpha1.KubernetesOpenBaoDevMode{}}},
		"raft-1-backup": {
			Server: &kubernetesopenbaov1alpha1.KubernetesOpenBaoServer{
				Storage:  &kubernetesopenbaov1alpha1.KubernetesOpenBaoServer_Raft{Raft: &kubernetesopenbaov1alpha1.KubernetesOpenBaoRaftStorage{}},
				Replicas: int32Ptr(1),
			},
			Backup: &kubernetesopenbaov1alpha1.KubernetesOpenBaoBackup{
				ObjectStore: &kubernetesopenbaov1alpha1.KubernetesOpenBaoBackupObjectStore{
					Prefix: "vault",
					Backend: &kubernetesopenbaov1alpha1.KubernetesOpenBaoBackupObjectStore_S3{S3: &kubernetesopenbaov1alpha1.KubernetesOpenBaoS3ObjectStore{
						Bucket: "snapshots", Region: "us-east-1",
						AccessKeys: &kubernetesopenbaov1alpha1.KubernetesOpenBaoS3AccessKeys{AccessKeyId: "AKIA", SecretAccessKey: "s3-secret"},
					}},
				},
			},
		},
		"raft-3-aws-seal": {
			Server: &kubernetesopenbaov1alpha1.KubernetesOpenBaoServer{Replicas: int32Ptr(3)},
			AutoUnseal: &kubernetesopenbaov1alpha1.KubernetesOpenBaoAutoUnseal{
				Seal: &kubernetesopenbaov1alpha1.KubernetesOpenBaoAutoUnseal_AwsKms{AwsKms: &kubernetesopenbaov1alpha1.KubernetesOpenBaoAwsKmsSeal{
					Region: "us-east-1", KmsKeyId: "alias/vault", AccessKeyId: "AKIA", SecretAccessKey: "kms-secret",
				}},
			},
		},
	}
}

func serverValues(t *testing.T, locals *Locals) map[string]interface{} {
	t.Helper()
	values, err := buildHelmValues(locals)
	if err != nil {
		t.Fatalf("buildHelmValues: %v", err)
	}
	server, ok := values["server"].(map[string]interface{})
	if !ok {
		t.Fatalf("server values not rendered: %#v", values)
	}
	return server
}

const listenerBlock = "ui = true\n\nlistener \"tcp\" {\n  tls_disable = 1\n  address = \"[::]:8200\"\n  cluster_address = \"[::]:8201\"\n}\n\n"

func TestRender_UnsetServerIsOneRaftServer(t *testing.T) {
	locals := localsFor(fixtures()["raft-1"])
	if locals.Dev || locals.Storage != storageRaft || locals.Replicas != 1 {
		t.Fatalf("unset server must resolve to raft x1, got dev=%v storage=%q replicas=%d", locals.Dev, locals.Storage, locals.Replicas)
	}
	want := listenerBlock +
		"storage \"raft\" {\n  path = \"/openbao/data\"\n  retry_join {\n    leader_api_addr = \"http://vault-0.vault-internal:8200\"\n  }\n}\n\n" +
		"service_registration \"kubernetes\" {}\n"
	if locals.BaoConfigHcl != want {
		t.Fatalf("config:\n%s\nwant:\n%s", locals.BaoConfigHcl, want)
	}
	server := serverValues(t, locals)
	wantHa := map[string]interface{}{
		"enabled":          true,
		"replicas":         1,
		"raft":             map[string]interface{}{"enabled": true, "setNodeId": true, "config": want},
		"disruptionBudget": map[string]interface{}{"enabled": false},
	}
	if !reflect.DeepEqual(server["ha"], wantHa) {
		t.Fatalf("ha values: %#v\nwant: %#v", server["ha"], wantHa)
	}
	if !reflect.DeepEqual(server["dataStorage"], map[string]interface{}{"enabled": true}) {
		t.Fatalf("dataStorage: %#v", server["dataStorage"])
	}
	if _, ok := server["standalone"]; ok {
		t.Fatalf("the chart's standalone mode must never be driven: %#v", server)
	}
}

func TestRender_RaftAtThreeKeepsTheChartsQuorumBudgetAndClaimsTheVolume(t *testing.T) {
	server := serverValues(t, localsFor(fixtures()["raft-3-volume"]))
	ha := server["ha"].(map[string]interface{})
	if ha["replicas"] != 3 {
		t.Fatalf("replicas: %#v", ha["replicas"])
	}
	if !reflect.DeepEqual(ha["disruptionBudget"], map[string]interface{}{"enabled": true}) {
		t.Fatalf("Raft above one replica leaves the chart's quorum arithmetic in charge: %#v", ha["disruptionBudget"])
	}
	wantVolume := map[string]interface{}{"enabled": true, "size": "20Gi", "storageClass": "fast"}
	if !reflect.DeepEqual(server["dataStorage"], wantVolume) {
		t.Fatalf("dataStorage: %#v", server["dataStorage"])
	}
	config := ha["raft"].(map[string]interface{})["config"].(string)
	if strings.Count(config, "retry_join {") != 3 {
		t.Fatalf("three peers need three retry_join blocks:\n%s", config)
	}
}

func TestRender_PostgresqlDrivesHaWithRaftOffAndNoVolume(t *testing.T) {
	locals := localsFor(fixtures()["postgresql-1"])
	if locals.Storage != storagePostgresql || locals.Replicas != 1 {
		t.Fatalf("storage=%q replicas=%d", locals.Storage, locals.Replicas)
	}
	want := listenerBlock + "storage \"postgresql\" {\n  ha_enabled = \"true\"\n}\n\nservice_registration \"kubernetes\" {}\n"
	if locals.BaoConfigHcl != want {
		t.Fatalf("config:\n%s\nwant:\n%s", locals.BaoConfigHcl, want)
	}
	server := serverValues(t, locals)
	wantHa := map[string]interface{}{
		"enabled":          true,
		"replicas":         1,
		"raft":             map[string]interface{}{"enabled": false},
		"config":           want,
		"disruptionBudget": map[string]interface{}{"enabled": false},
	}
	if !reflect.DeepEqual(server["ha"], wantHa) {
		t.Fatalf("ha values: %#v\nwant: %#v", server["ha"], wantHa)
	}
	if !reflect.DeepEqual(server["dataStorage"], map[string]interface{}{"enabled": false}) {
		t.Fatalf("PostgreSQL storage claims no data volume: %#v", server["dataStorage"])
	}
	wantEnv := map[string]interface{}{
		"PGHOST": "identity-db-rw", "PGPORT": "5432", "PGDATABASE": "openbao", "PGUSER": "app", "PGSSLMODE": "require",
	}
	if !reflect.DeepEqual(server["extraEnvironmentVars"], wantEnv) {
		t.Fatalf("plain env: %#v\nwant: %#v", server["extraEnvironmentVars"], wantEnv)
	}
	wantSecretEnv := []interface{}{
		map[string]interface{}{"envName": "PGPASSWORD", "secretName": "identity-db-app", "secretKey": "password"},
	}
	if !reflect.DeepEqual(server["extraSecretEnvironmentVars"], wantSecretEnv) {
		t.Fatalf("secret env: %#v\nwant: %#v", server["extraSecretEnvironmentVars"], wantSecretEnv)
	}
}

func TestRender_PostgresqlAtThreeAllowsAllButOneToBeDisrupted(t *testing.T) {
	locals := localsFor(fixtures()["postgresql-3-full"])
	want := listenerBlock + "storage \"postgresql\" {\n  ha_enabled = \"true\"\n  max_parallel = \"32\"\n}\n\nservice_registration \"kubernetes\" {}\n"
	if locals.BaoConfigHcl != want {
		t.Fatalf("config:\n%s\nwant:\n%s", locals.BaoConfigHcl, want)
	}
	server := serverValues(t, locals)
	ha := server["ha"].(map[string]interface{})
	if !reflect.DeepEqual(ha["disruptionBudget"], map[string]interface{}{"enabled": true, "maxUnavailable": 2}) {
		t.Fatalf("disruptionBudget: %#v", ha["disruptionBudget"])
	}
	wantEnv := map[string]interface{}{
		"PGHOST": "identity-db-rw", "PGPORT": "5433", "PGDATABASE": "openbao", "PGUSER": "identity", "PGSSLMODE": "disable",
	}
	if !reflect.DeepEqual(server["extraEnvironmentVars"], wantEnv) {
		t.Fatalf("plain env: %#v", server["extraEnvironmentVars"])
	}
	wantSecretEnv := []interface{}{
		map[string]interface{}{"envName": "PGPASSWORD", "secretName": "identity-db-app", "secretKey": "pw"},
	}
	if !reflect.DeepEqual(server["extraSecretEnvironmentVars"], wantSecretEnv) {
		t.Fatalf("secret env: %#v", server["extraSecretEnvironmentVars"])
	}
	// The audit volume is the server's on either engine.
	if !reflect.DeepEqual(server["auditStorage"], map[string]interface{}{"enabled": true, "size": "5Gi"}) {
		t.Fatalf("auditStorage: %#v", server["auditStorage"])
	}
}

func TestRender_DevRendersNoConfigNoVolumeNoBudget(t *testing.T) {
	locals := localsFor(fixtures()["dev"])
	if !locals.Dev || locals.Replicas != 1 || locals.BaoConfigHcl != "" {
		t.Fatalf("dev=%v replicas=%d config=%q", locals.Dev, locals.Replicas, locals.BaoConfigHcl)
	}
	server := serverValues(t, localsFor(fixtures()["dev"]))
	if !reflect.DeepEqual(server["dev"], map[string]interface{}{"enabled": true}) {
		t.Fatalf("dev: %#v", server["dev"])
	}
	for _, key := range []string{"ha", "standalone", "dataStorage"} {
		if _, ok := server[key]; ok {
			t.Fatalf("dev must not render %s: %#v", key, server[key])
		}
	}
}

func TestRender_SealAndDatabaseShareTheEnvSeamsSorted(t *testing.T) {
	spec := fixtures()["raft-3-aws-seal"]
	spec.Server.Storage = postgresqlStorage()
	server := serverValues(t, localsFor(spec))
	wantEnv := map[string]interface{}{
		"AWS_ACCESS_KEY_ID": "AKIA", "AWS_REGION": "us-east-1",
		"PGHOST": "identity-db-rw", "PGPORT": "5432", "PGDATABASE": "openbao", "PGUSER": "app", "PGSSLMODE": "require",
	}
	if !reflect.DeepEqual(server["extraEnvironmentVars"], wantEnv) {
		t.Fatalf("plain env: %#v", server["extraEnvironmentVars"])
	}
	wantSecretEnv := []interface{}{
		map[string]interface{}{"envName": "AWS_SECRET_ACCESS_KEY", "secretName": "vault-seal-credentials", "secretKey": "AWS_SECRET_ACCESS_KEY"},
		map[string]interface{}{"envName": "PGPASSWORD", "secretName": "identity-db-app", "secretKey": "password"},
	}
	if !reflect.DeepEqual(server["extraSecretEnvironmentVars"], wantSecretEnv) {
		t.Fatalf("secret env: %#v", server["extraSecretEnvironmentVars"])
	}
}

func TestRender_ActiveServiceExistsForEveryEngineButDev(t *testing.T) {
	for name, spec := range fixtures() {
		got := activeServiceName(localsFor(spec))
		want := "vault-active"
		if name == "dev" {
			want = ""
		}
		if got != want {
			t.Fatalf("%s: active_service %q, want %q", name, got, want)
		}
	}
}

// TestRender_NothingSensitiveLeavesTheReferencedSecret is the leak gate:
// across every fixture, the rendered configuration and the rendered chart
// values carry PGPASSWORD only as an environment-variable NAME, name the
// referenced Secret only as a secretName, and never carry the seal
// credential material as a value.
func TestRender_NothingSensitiveLeavesTheReferencedSecret(t *testing.T) {
	for name, spec := range fixtures() {
		locals := localsFor(spec)
		values, err := buildHelmValues(locals)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		doc, err := yaml.Marshal(values)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		rendered := string(doc) + locals.BaoConfigHcl
		for _, forbidden := range []string{"kms-secret", "s3-secret", "connection_url", "postgres://"} {
			if strings.Contains(rendered, forbidden) {
				t.Fatalf("%s: rendered documents carry %q:\n%s", name, forbidden, rendered)
			}
		}
		if strings.Contains(locals.BaoConfigHcl, "PGPASSWORD") || strings.Contains(locals.BaoConfigHcl, "identity-db-app") {
			t.Fatalf("%s: the configuration string must carry neither the password variable nor the Secret's name:\n%s", name, locals.BaoConfigHcl)
		}
		if strings.Count(rendered, "identity-db-app") != strings.Count(rendered, "secretName: identity-db-app") {
			t.Fatalf("%s: the referenced Secret's name appears somewhere other than as a secretName:\n%s", name, rendered)
		}
	}
}

// TestRender_WriteFixturesForCrossEngineDiff writes every fixture's
// configuration string and chart values (as the release renders them) to
// OPENBAO_RENDER_DIR so the Terraform module's rendering of the same
// manifests can be diffed byte for byte. Skips unless the directory is set.
func TestRender_WriteFixturesForCrossEngineDiff(t *testing.T) {
	dir := os.Getenv("OPENBAO_RENDER_DIR")
	if dir == "" {
		t.Skip("OPENBAO_RENDER_DIR not set")
	}
	for name, spec := range fixtures() {
		locals := localsFor(spec)
		values, err := buildHelmValues(locals)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		delete(values, "fullnameOverride") // re-pinned as a separate document on both engines
		doc, err := yaml.Marshal(values)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, name+".pulumi.values.yaml"), doc, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name+".pulumi.config.hcl"), []byte(locals.BaoConfigHcl), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
