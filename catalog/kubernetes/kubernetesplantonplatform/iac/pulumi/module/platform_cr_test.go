package module

import (
	"reflect"
	"testing"

	kubernetesplantonplatformv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesplantonplatform/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

// The rendering contract for the database's backup and recovery
// declarations, locked here because no offline gate reads the rendered CR
// for what it must NOT carry. The Terraform module renders the SAME shape
// from the same spec (iac/tf/locals.tf); the maps below are the
// cross-engine contract.

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// localsFor builds the module's locals the way Resources does, from a stack
// input carrying the given spec. References arrive at the module already
// resolved to values, so the fixtures use literals throughout.
func localsFor(spec *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec) *Locals {
	spec.Namespace = literal("planton")
	spec.CreateNamespace = true
	spec.Version = "v0.0.62"
	return initializeLocals(nil, &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformStackInput{
		Target: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatform{
			Metadata: &shared.CloudResourceMetadata{Name: "acme"},
			Spec:     spec,
		},
	})
}

func r2Store(path string) *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore {
	return &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore{
		DestinationPath: path,
		Backend: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore_R2{
			R2: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformR2ObjectStore{
				AccountId:    literal("0123456789abcdef0123456789abcdef"),
				Jurisdiction: literal("default"),
				Credentials: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformR2Credentials{
					AccessKeyId:     literal("token-id"),
					SecretAccessKey: literal("token-sha256"),
				},
			},
		},
	}
}

func postgresqlBody(t *testing.T, locals *Locals) map[string]interface{} {
	t.Helper()
	database, ok := platformSpecBody(locals)["database"].(map[string]interface{})
	if !ok {
		t.Fatalf("spec.database not rendered: %#v", platformSpecBody(locals))
	}
	postgresql, ok := database["postgresql"].(map[string]interface{})
	if !ok {
		t.Fatalf("spec.database.postgresql not rendered: %#v", database)
	}
	return postgresql
}

func TestPlatformSpecBody_NoBackupRendersNothing(t *testing.T) {
	locals := localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		Database: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformDatabase{
			Postgresql: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformPostgresql{
				Replicas: proto.Int32(2),
			},
		},
		Prerequisites: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformPrerequisites{
			PostgresOperator: proto.String("auto"),
		},
	})
	spec := platformSpecBody(locals)

	postgresql := postgresqlBody(t, locals)
	for _, key := range []string{"backup", "recoverFrom"} {
		if _, present := postgresql[key]; present {
			t.Errorf("postgresql.%s rendered for a platform that declared none: %#v", key, postgresql)
		}
	}
	prerequisites := spec["prerequisites"].(map[string]interface{})
	if _, present := prerequisites["postgresBackupPlugin"]; present {
		t.Errorf("prerequisites.postgresBackupPlugin rendered when the manifest never set it: %#v", prerequisites)
	}
	if secrets := objectStoreSecrets(locals); len(secrets) != 0 {
		t.Errorf("no store declared, yet %d Secret(s) would be created: %#v", len(secrets), secrets)
	}
}

func TestPlatformSpecBody_R2BackupNamesTheMaterializedSecret(t *testing.T) {
	locals := localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		Database: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformDatabase{
			Postgresql: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformPostgresql{
				Backup: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformPostgresqlBackup{
					ObjectStore:     r2Store("s3://acme-platform-backups/platform"),
					RetentionPolicy: proto.String("30d"),
					Schedule:        proto.String("0 0 2 * * *"),
				},
			},
		},
		Prerequisites: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformPrerequisites{
			PostgresBackupPlugin: proto.String("auto"),
		},
	})

	want := map[string]interface{}{
		"objectStore": map[string]interface{}{
			"destinationPath": "s3://acme-platform-backups/platform",
			"r2": map[string]interface{}{
				"accountId":             "0123456789abcdef0123456789abcdef",
				"jurisdiction":          "default",
				"credentialsSecretName": "acme-postgres-backup-creds",
			},
		},
		"retentionPolicy": "30d",
		"schedule":        "0 0 2 * * *",
	}
	got := postgresqlBody(t, locals)["backup"]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("backup rendered\n got: %#v\nwant: %#v", got, want)
	}

	prerequisites := platformSpecBody(locals)["prerequisites"].(map[string]interface{})
	if prerequisites["postgresBackupPlugin"] != "auto" {
		t.Errorf("prerequisites.postgresBackupPlugin = %#v, want auto", prerequisites["postgresBackupPlugin"])
	}

	wantSecrets := []materializedSecret{{
		Name: "acme-postgres-backup-creds",
		Data: map[string]string{
			"ACCESS_KEY_ID":     "token-id",
			"SECRET_ACCESS_KEY": "token-sha256",
		},
	}}
	if got := objectStoreSecrets(locals); !reflect.DeepEqual(got, wantSecrets) {
		t.Errorf("Secrets to materialize\n got: %#v\nwant: %#v", got, wantSecrets)
	}
}

func TestPlatformSpecBody_RecoverFromRendersBesideBackup(t *testing.T) {
	locals := localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		Database: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformDatabase{
			Postgresql: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformPostgresql{
				Backup: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformPostgresqlBackup{
					ObjectStore: r2Store("s3://acme-platform-backups/platform"),
				},
				RecoverFrom: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformPostgresqlRecoverFrom{
					ObjectStore: r2Store("s3://acme-platform-backups/platform"),
					ServerName:  "acme-postgres-1a2b3c4d",
					TargetTime:  "2026-09-13T20:30:00Z",
				},
			},
		},
	})

	want := map[string]interface{}{
		"objectStore": map[string]interface{}{
			"destinationPath": "s3://acme-platform-backups/platform",
			"r2": map[string]interface{}{
				"accountId":             "0123456789abcdef0123456789abcdef",
				"jurisdiction":          "default",
				"credentialsSecretName": "acme-postgres-recovery-creds",
			},
		},
		"serverName": "acme-postgres-1a2b3c4d",
		"targetTime": "2026-09-13T20:30:00Z",
	}
	got := postgresqlBody(t, locals)["recoverFrom"]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("recoverFrom rendered\n got: %#v\nwant: %#v", got, want)
	}

	secrets := objectStoreSecrets(locals)
	if len(secrets) != 2 || secrets[0].Name != "acme-postgres-backup-creds" || secrets[1].Name != "acme-postgres-recovery-creds" {
		t.Errorf("expected the backup Secret then the recovery Secret, got %#v", secrets)
	}
}

func TestObjectStoreBody_KeylessPosturesNameNoSecret(t *testing.T) {
	cases := map[string]struct {
		store *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore
		want  map[string]interface{}
	}{
		"s3 keyless": {
			store: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore{
				DestinationPath: "s3://acme-backups/platform",
				Backend: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore_S3{
					S3: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformS3ObjectStore{
						Region: "us-west-2", Keyless: true,
					},
				},
			},
			want: map[string]interface{}{
				"destinationPath": "s3://acme-backups/platform",
				"s3":              map[string]interface{}{"region": "us-west-2"},
			},
		},
		"gcs keyless": {
			store: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore{
				DestinationPath: "gs://acme-backups/platform",
				Backend: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore_Gcs{
					Gcs: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformGcsObjectStore{Keyless: true},
				},
			},
			want: map[string]interface{}{
				"destinationPath": "gs://acme-backups/platform",
				"gcs":             map[string]interface{}{},
			},
		},
		"azure keyless": {
			store: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore{
				DestinationPath: "https://acme.blob.core.windows.net/backups/platform",
				Backend: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore_AzureBlob{
					AzureBlob: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformAzureBlobObjectStore{
						StorageAccount: "acme", Keyless: true,
					},
				},
			},
			want: map[string]interface{}{
				"destinationPath": "https://acme.blob.core.windows.net/backups/platform",
				"azureBlob":       map[string]interface{}{"storageAccount": "acme"},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := objectStoreBody(tc.store, "creds", "ca")
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("objectStore rendered\n got: %#v\nwant: %#v", got, tc.want)
			}
			if data := objectStoreCredentialsData(tc.store); data != nil {
				t.Errorf("a keyless posture must materialize no Secret, got %#v", data)
			}
		})
	}
}

func TestObjectStoreBody_S3WithKeysAndPrivateCA(t *testing.T) {
	store := &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore{
		DestinationPath: "s3://backups/platform",
		Backend: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformObjectStore_S3{
			S3: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformS3ObjectStore{
				EndpointUrl:   "https://minio.minio-system.svc:9000",
				EndpointCaPem: "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n",
				AccessKeys: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformS3AccessKeys{
					AccessKeyId: "minio-access-key", SecretAccessKey: "minio-secret-key",
				},
			},
		},
	}
	want := map[string]interface{}{
		"destinationPath": "s3://backups/platform",
		"s3": map[string]interface{}{
			"endpointURL":           "https://minio.minio-system.svc:9000",
			"credentialsSecretName": "acme-postgres-backup-creds",
			"endpointCASecretRef": map[string]interface{}{
				"name": "acme-postgres-backup-endpoint-ca",
				"key":  "ca.crt",
			},
		},
	}
	got := objectStoreBody(store, "acme-postgres-backup-creds", "acme-postgres-backup-endpoint-ca")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("objectStore rendered\n got: %#v\nwant: %#v", got, want)
	}

	wantSecrets := []materializedSecret{
		{Name: "acme-postgres-backup-creds", Data: map[string]string{
			"ACCESS_KEY_ID": "minio-access-key", "SECRET_ACCESS_KEY": "minio-secret-key",
		}},
		{Name: "acme-postgres-backup-endpoint-ca", Data: map[string]string{
			"ca.crt": "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n",
		}},
	}
	if got := storeSecrets(store, "acme-postgres-backup-creds", "acme-postgres-backup-endpoint-ca"); !reflect.DeepEqual(got, wantSecrets) {
		t.Errorf("Secrets to materialize\n got: %#v\nwant: %#v", got, wantSecrets)
	}
}

// ---- the vault's seal and keys ------------------------------------------------
//
// The vault renders like the object stores: identifiers pass through, a
// declared credential VALUE becomes the seal-credentials Secret the CR names
// (keyed by the environment variable the seal wrapper reads), a keyless arm
// names none, and the volume keys the vault no longer has never render.

func vaultBody(t *testing.T, locals *Locals) map[string]interface{} {
	t.Helper()
	vault, ok := platformSpecBody(locals)["vault"].(map[string]interface{})
	if !ok {
		t.Fatalf("spec.vault not rendered: %#v", platformSpecBody(locals))
	}
	return vault
}

func TestPlatformSpecBody_VaultWithoutASealRendersOnlyWhatWasDeclared(t *testing.T) {
	locals := localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		Vault: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVault{
			InitSecretName: "planton-vault-keys",
		},
	})
	want := map[string]interface{}{"initSecretName": "planton-vault-keys"}
	if got := vaultBody(t, locals); !reflect.DeepEqual(got, want) {
		t.Errorf("vault rendered\n got: %#v\nwant: %#v", got, want)
	}
	if secrets := materializedSecrets(locals); len(secrets) != 0 {
		t.Errorf("no seal credential declared, yet %d Secret(s) would be created: %#v", len(secrets), secrets)
	}
}

func TestPlatformSpecBody_AwsKmsSealWithKeysNamesTheMaterializedSecret(t *testing.T) {
	locals := localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		Vault: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVault{
			AutoUnseal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal{
				Seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal_AwsKms{
					AwsKms: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAwsKmsSeal{
						Region: "us-west-2", KmsKeyId: "alias/planton-vault-unseal",
						AccessKeyId: "AKIA-example", SecretAccessKey: "secret",
					},
				},
			},
		},
	})
	want := map[string]interface{}{
		"autoUnseal": map[string]interface{}{
			"awsKms": map[string]interface{}{
				"region":                "us-west-2",
				"kmsKeyId":              "alias/planton-vault-unseal",
				"accessKeyId":           "AKIA-example",
				"credentialsSecretName": "acme-openbao-seal-creds",
			},
		},
	}
	if got := vaultBody(t, locals); !reflect.DeepEqual(got, want) {
		t.Errorf("vault rendered\n got: %#v\nwant: %#v", got, want)
	}
	wantSecrets := []materializedSecret{{
		Name: "acme-openbao-seal-creds",
		Data: map[string]string{"AWS_SECRET_ACCESS_KEY": "secret"},
	}}
	if got := materializedSecrets(locals); !reflect.DeepEqual(got, wantSecrets) {
		t.Errorf("Secrets to materialize\n got: %#v\nwant: %#v", got, wantSecrets)
	}
}

func TestPlatformSpecBody_KeylessSealsNameNoSecret(t *testing.T) {
	cases := map[string]struct {
		seal *kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal
		want map[string]interface{}
	}{
		"aws keyless": {
			seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal{
				Seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal_AwsKms{
					AwsKms: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAwsKmsSeal{Region: "us-west-2", KmsKeyId: "alias/planton-vault-unseal"},
				},
			},
			want: map[string]interface{}{"awsKms": map[string]interface{}{"region": "us-west-2", "kmsKeyId": "alias/planton-vault-unseal"}},
		},
		"gcp": {
			seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal{
				Seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal_GcpKms{
					GcpKms: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultGcpKmsSeal{
						Project: literal("acme-platform"), Region: "global",
						KeyRing: literal("planton-vault-unseal"), CryptoKey: literal("planton-vault-unseal"),
					},
				},
			},
			want: map[string]interface{}{"gcpKms": map[string]interface{}{
				"project": "acme-platform", "region": "global", "keyRing": "planton-vault-unseal", "cryptoKey": "planton-vault-unseal",
			}},
		},
		"azure keyless": {
			seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal{
				Seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal_AzureKeyVault{
					AzureKeyVault: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAzureKeyVaultSeal{VaultName: "acme-kv", KeyName: "unseal", TenantId: "tenant"},
				},
			},
			want: map[string]interface{}{"azureKeyVault": map[string]interface{}{"vaultName": "acme-kv", "keyName": "unseal", "tenantId": "tenant"}},
		},
		"transit without a token": {
			seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal{
				Seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal_Transit{
					Transit: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultTransitSeal{Address: "http://key-holder.openbao.svc:8200", KeyName: "autounseal"},
				},
			},
			want: map[string]interface{}{"transit": map[string]interface{}{"address": "http://key-holder.openbao.svc:8200", "keyName": "autounseal"}},
		},
	}
	for name, tc := range cases {
		locals := localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
			Vault: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVault{AutoUnseal: tc.seal},
		})
		if got := vaultBody(t, locals)["autoUnseal"]; !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: autoUnseal rendered\n got: %#v\nwant: %#v", name, got, tc.want)
		}
		if secrets := materializedSecrets(locals); len(secrets) != 0 {
			t.Errorf("%s: keyless, yet %d Secret(s) would be created: %#v", name, len(secrets), secrets)
		}
	}
}

func TestPlatformSpecBody_AzureAndTransitCredentialsRideTheSecretUnderTheirEnvNames(t *testing.T) {
	azure := localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		Vault: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVault{
			AutoUnseal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal{
				Seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal_AzureKeyVault{
					AzureKeyVault: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAzureKeyVaultSeal{
						VaultName: "acme-kv", KeyName: "unseal", TenantId: "tenant", ClientId: "client", ClientSecret: "sp-secret",
					},
				},
			},
		},
	})
	wantAzure := map[string]interface{}{
		"vaultName": "acme-kv", "keyName": "unseal", "tenantId": "tenant", "clientId": "client",
		"credentialsSecretName": "acme-openbao-seal-creds",
	}
	if got := vaultBody(t, azure)["autoUnseal"].(map[string]interface{})["azureKeyVault"]; !reflect.DeepEqual(got, wantAzure) {
		t.Errorf("azureKeyVault rendered\n got: %#v\nwant: %#v", got, wantAzure)
	}
	if got := materializedSecrets(azure); len(got) != 1 || !reflect.DeepEqual(got[0].Data, map[string]string{"AZURE_CLIENT_SECRET": "sp-secret"}) {
		t.Errorf("azure seal Secret\n got: %#v", got)
	}

	transit := localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		Vault: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVault{
			AutoUnseal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal{
				Seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal_Transit{
					Transit: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultTransitSeal{
						Address: "http://key-holder.openbao.svc:8200", KeyName: "autounseal", MountPath: proto.String("keys/"), Token: "s.token",
					},
				},
			},
		},
	})
	wantTransit := map[string]interface{}{
		"address": "http://key-holder.openbao.svc:8200", "keyName": "autounseal", "mountPath": "keys/",
		"credentialsSecretName": "acme-openbao-seal-creds",
	}
	if got := vaultBody(t, transit)["autoUnseal"].(map[string]interface{})["transit"]; !reflect.DeepEqual(got, wantTransit) {
		t.Errorf("transit rendered\n got: %#v\nwant: %#v", got, wantTransit)
	}
	if got := materializedSecrets(transit); len(got) != 1 || !reflect.DeepEqual(got[0].Data, map[string]string{"VAULT_TOKEN": "s.token"}) {
		t.Errorf("transit seal Secret\n got: %#v", got)
	}
}

func TestPlatformSpecBody_GcpWorkloadIdentityBecomesTheAnnotationAndExplicitEntriesWin(t *testing.T) {
	locals := localsFor(&kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformSpec{
		Vault: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVault{
			AutoUnseal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal{
				Seal: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultAutoUnseal_GcpKms{
					GcpKms: &kubernetesplantonplatformv1alpha1.KubernetesPlantonPlatformVaultGcpKmsSeal{
						Project: literal("acme-platform"), Region: "global",
						KeyRing: literal("planton-vault-unseal"), CryptoKey: literal("planton-vault-unseal"),
						WorkloadIdentityServiceAccount: literal("planton-vault-unseal@acme-platform.iam.gserviceaccount.com"),
					},
				},
			},
			ServiceAccountAnnotations: map[string]string{"example.com/team": "platform"},
		},
	})
	want := map[string]interface{}{
		"iam.gke.io/gcp-service-account": "planton-vault-unseal@acme-platform.iam.gserviceaccount.com",
		"example.com/team":               "platform",
	}
	if got := vaultBody(t, locals)["serviceAccountAnnotations"]; !reflect.DeepEqual(got, want) {
		t.Errorf("serviceAccountAnnotations rendered\n got: %#v\nwant: %#v", got, want)
	}

	// The same annotation declared explicitly overrides the arm's identity.
	locals.Spec.Vault.ServiceAccountAnnotations["iam.gke.io/gcp-service-account"] = "override@acme-platform.iam.gserviceaccount.com"
	got := vaultBody(t, locals)["serviceAccountAnnotations"].(map[string]interface{})
	if got["iam.gke.io/gcp-service-account"] != "override@acme-platform.iam.gserviceaccount.com" {
		t.Errorf("an explicit annotation must win over the arm's identity, got %#v", got)
	}
}
