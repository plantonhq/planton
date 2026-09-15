package kubernetesplantonplatformv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesPlantonPlatformSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesPlantonPlatformSpec Validation Tests")
}

func strPtr(value string) *string { return &value }

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// gatewayRef is a valueFrom against a KubernetesGateway resource; an empty
// fieldPath leans on the field's annotated default.
func gatewayRef(name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return refTo(cloudresourcekind.CloudResourceKind_KubernetesGateway, name, fieldPath)
}

// refTo is a valueFrom against any catalog resource — the shape a manifest
// uses to follow another resource's output instead of typing its value.
func refTo(kind cloudresourcekind.CloudResourceKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

// minimalValidPlatform is the zero-config case: a namespace and a version
// — the whole platform boots from exactly this, reachable over
// port-forward with the first console visitor becoming the admin.
func minimalValidPlatform() *KubernetesPlantonPlatform {
	return &KubernetesPlantonPlatform{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesPlantonPlatform",
		Metadata: &shared.CloudResourceMetadata{
			Name: "planton",
		},
		Spec: &KubernetesPlantonPlatformSpec{
			Namespace:       literalRef("planton"),
			CreateNamespace: true,
			Version:         "v0.0.45",
		},
	}
}

var _ = ginkgo.Describe("KubernetesPlantonPlatformSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for the zero-config platform", func() {
			err := protovalidate.Validate(minimalValidPlatform())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept platform-wide storage settings", func() {
			input := minimalValidPlatform()
			input.Spec.Storage = &KubernetesPlantonPlatformStorage{
				StorageClassName: "gp3",
				Size:             "800Gi",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an ingress with hostname and cert-manager TLS", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:  true,
				Hostname: "planton.example.com",
				Tls: &KubernetesPlantonPlatformIngressTls{
					Issuer: &KubernetesPlantonPlatformCertManagerIssuer{
						Name: "letsencrypt",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an ingress with a brought certificate Secret", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:  true,
				Hostname: "planton.example.com",
				Tls: &KubernetesPlantonPlatformIngressTls{
					SecretName: "planton-tls",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a magic-DNS ingress (enabled, no hostname, no tls)", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{Enabled: true}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		// The Gateway API front door: the Gateway's listener already serves
		// the hostname over HTTPS, so no tls block is needed.
		ginkgo.It("should accept a Gateway API front door pinned to one listener, no tls", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:  true,
				Hostname: "planton.example.com",
				GatewayRef: &KubernetesPlantonPlatformGatewayRef{
					Name:        literalRef("main"),
					Namespace:   literalRef("istio-ingress"),
					SectionName: "https",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a Gateway API front door with a cert-manager issuer", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:    true,
				Hostname:   "planton.example.com",
				GatewayRef: &KubernetesPlantonPlatformGatewayRef{Name: literalRef("main")},
				Tls: &KubernetesPlantonPlatformIngressTls{
					Issuer: &KubernetesPlantonPlatformCertManagerIssuer{
						Name: "letsencrypt",
						Kind: strPtr("ClusterIssuer"),
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		// The reachability declaration: every word is accepted on an enabled
		// door, and "private" is accepted even on a disabled one because it is
		// true there (a port-forward door is reached only from the machine
		// running it). Only the contradiction is refused, in the invalid block.
		ginkgo.It("should accept every reachability word on an enabled door", func() {
			for _, word := range []string{"auto", "public", "private"} {
				input := minimalValidPlatform()
				input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
					Enabled:      true,
					Hostname:     "planton.example.com",
					Reachability: strPtr(word),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil(), "reachability %q", word)
			}
		})

		ginkgo.It("should accept reachability private on a disabled ingress", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{Reachability: strPtr("private")}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept runner workload identity and database growth", func() {
			input := minimalValidPlatform()
			replicas := int32(2)
			input.Spec.Runner = &KubernetesPlantonPlatformRunner{
				ServiceAccountAnnotations: map[string]string{
					"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/planton-runner",
				},
			}
			input.Spec.Database = &KubernetesPlantonPlatformDatabase{
				Postgresql: &KubernetesPlantonPlatformPostgresql{
					Replicas:    &replicas,
					StorageSize: "20Gi",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a module-release override that names a published release", func() {
			// The one legitimate use: routing around a retracted artifact set. The
			// platform resolves modules at its own catalog release when this is unset.
			input := minimalValidPlatform()
			input.Spec.ControlPlane = &KubernetesPlantonPlatformControlPlane{IacModulesVersion: "v0.5.60"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())

			// Unset is the shape every install should have, and a controlPlane block
			// declared for other reasons (replicas) must not trip the pattern.
			unset := minimalValidPlatform()
			replicas := int32(1)
			unset.Spec.ControlPlane = &KubernetesPlantonPlatformControlPlane{Replicas: &replicas}
			gomega.Expect(protovalidate.Validate(unset)).To(gomega.BeNil())
		})

		ginkgo.It("should accept the AWS secret backend with its config", func() {
			input := minimalValidPlatform()
			input.Spec.Bootstrap = &KubernetesPlantonPlatformBootstrap{
				SecretBackend: &KubernetesPlantonPlatformSecretBackend{
					Type: "awsSecretsManager",
					AwsSecretsManager: &KubernetesPlantonPlatformAwsSecretsManager{
						Region:    "us-east-1",
						KmsKeyArn: "arn:aws:kms:us-east-1:123456789012:key/abc",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a license from a Secret reference", func() {
			input := minimalValidPlatform()
			input.Spec.License = &KubernetesPlantonPlatformLicense{
				SecretKeyRef: &KubernetesPlantonPlatformSecretKeyRef{
					Name: "planton-license",
					Key:  "key",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a pre-seeded admin email", func() {
			input := minimalValidPlatform()
			input.Spec.Identity = &KubernetesPlantonPlatformIdentity{
				AdminEmail: "admin@example.com",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		// Email: one declaration for both senders. The three ways into a relay
		// and the Resend arm are each accepted on their own; the contradictions
		// are refused in the invalid block, in the platform's own words.
		ginkgo.It("should accept an SMTP relay with a username and password Secret", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From:    &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				ReplyTo: "it-help@acme.com",
				Smtp: &KubernetesPlantonPlatformEmailSmtp{
					Host:                  "smtp.office365.com",
					CredentialsSecretName: "planton-email",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an SMTP relay signed in through OAuth2, with a private CA bundle", func() {
			input := minimalValidPlatform()
			port := int32(587)
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com", Name: strPtr("Acme Planton")},
				Smtp: &KubernetesPlantonPlatformEmailSmtp{
					Host:     "smtp.office365.com",
					Port:     &port,
					Security: strPtr("starttls"),
					Oauth2: &KubernetesPlantonPlatformEmailSmtpOauth2{
						User:            "no-reply@planton.acme.com",
						TokenUrl:        "https://login.microsoftonline.com/tenant/oauth2/v2.0/token",
						Scope:           "https://outlook.office365.com/.default",
						ClientId:        "app-registration-client-id",
						ClientSecretRef: &KubernetesPlantonPlatformSecretKeyRef{Name: "planton-email-oauth2", Key: "client-secret"},
					},
					CaBundleSecretRef: &KubernetesPlantonPlatformSecretKeyRef{Name: "corp-ca", Key: "ca.crt"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a credential-free plaintext relay (an internal smart host that admits the cluster's address)", func() {
			input := minimalValidPlatform()
			port := int32(25)
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				Smtp: &KubernetesPlantonPlatformEmailSmtp{
					Host:     "smtp-relay.corp.acme.com",
					Port:     &port,
					Security: strPtr("none"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept every security word on a credential-free relay", func() {
			for _, word := range []string{"starttls", "tls", "none"} {
				input := minimalValidPlatform()
				input.Spec.Email = &KubernetesPlantonPlatformEmail{
					From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
					Smtp: &KubernetesPlantonPlatformEmailSmtp{Host: "smtp-relay.corp.acme.com", Security: strPtr(word)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil(), "security %q", word)
			}
		})

		ginkgo.It("should accept Resend with the API key by reference", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				Resend: &KubernetesPlantonPlatformEmailResend{
					ApiKeySecretRef: &KubernetesPlantonPlatformSecretKeyRef{Name: "planton-email", Key: "api-key"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		// ---- the database's backup and recovery ------------------------------------

		ginkgo.It("should accept a backup to R2 composed from the Cloudflare kinds by reference", func() {
			input := backupCapablePlatform()
			input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore:     r2StoreByReference("s3://acme-platform-backups/platform"),
				RetentionPolicy: strPtr("30d"),
				Schedule:        strPtr("0 0 2 * * *"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a backup to R2 with every value literal and no jurisdiction", func() {
			input := backupCapablePlatform()
			store := r2StoreByReference("s3://acme-platform-backups/platform")
			store.GetR2().AccountId = literalRef("0123456789abcdef0123456789abcdef")
			store.GetR2().Jurisdiction = nil
			store.GetR2().Credentials.AccessKeyId = literalRef("token-id")
			store.GetR2().Credentials.SecretAccessKey = literalRef("token-sha256")
			input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{ObjectStore: store})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every retention unit the store enforces", func() {
			for _, retention := range []string{"7d", "8w", "6m"} {
				input := backupCapablePlatform()
				input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
					ObjectStore:     r2StoreByReference("s3://acme-platform-backups/platform"),
					RetentionPolicy: strPtr(retention),
				})
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "retention %q", retention)
			}
		})

		ginkgo.It("should accept S3 keyless and S3 with access keys", func() {
			keyless := backupCapablePlatform()
			keyless.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: &KubernetesPlantonPlatformObjectStore{
					DestinationPath: "s3://acme-backups/platform",
					Backend: &KubernetesPlantonPlatformObjectStore_S3{S3: &KubernetesPlantonPlatformS3ObjectStore{
						Region: "us-west-2", Keyless: true,
					}},
				},
				ServiceAccountAnnotations: map[string]string{"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/platform-backups"},
			})
			gomega.Expect(protovalidate.Validate(keyless)).To(gomega.BeNil())

			keyed := backupCapablePlatform()
			keyed.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: &KubernetesPlantonPlatformObjectStore{
					DestinationPath: "s3://backups/platform",
					Backend: &KubernetesPlantonPlatformObjectStore_S3{S3: &KubernetesPlantonPlatformS3ObjectStore{
						EndpointUrl: "http://minio.minio-system.svc:9000",
						AccessKeys:  &KubernetesPlantonPlatformS3AccessKeys{AccessKeyId: "minio-access-key", SecretAccessKey: "minio-secret-key"},
					}},
				},
			})
			gomega.Expect(protovalidate.Validate(keyed)).To(gomega.BeNil())
		})

		ginkgo.It("should accept GCS keyless and GCS with a service-account key", func() {
			keyless := backupCapablePlatform()
			keyless.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: &KubernetesPlantonPlatformObjectStore{
					DestinationPath: "gs://acme-backups/platform",
					Backend:         &KubernetesPlantonPlatformObjectStore_Gcs{Gcs: &KubernetesPlantonPlatformGcsObjectStore{Keyless: true}},
				},
			})
			gomega.Expect(protovalidate.Validate(keyless)).To(gomega.BeNil())

			keyed := backupCapablePlatform()
			keyed.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: &KubernetesPlantonPlatformObjectStore{
					DestinationPath: "gs://acme-backups/platform",
					Backend:         &KubernetesPlantonPlatformObjectStore_Gcs{Gcs: &KubernetesPlantonPlatformGcsObjectStore{ServiceAccountKeyJson: "{}"}},
				},
			})
			gomega.Expect(protovalidate.Validate(keyed)).To(gomega.BeNil())
		})

		ginkgo.It("should accept Azure Blob keyless and Azure Blob with a connection string", func() {
			keyless := backupCapablePlatform()
			keyless.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: &KubernetesPlantonPlatformObjectStore{
					DestinationPath: "https://acme.blob.core.windows.net/backups/platform",
					Backend: &KubernetesPlantonPlatformObjectStore_AzureBlob{AzureBlob: &KubernetesPlantonPlatformAzureBlobObjectStore{
						StorageAccount: "acme", Keyless: true,
					}},
				},
			})
			gomega.Expect(protovalidate.Validate(keyless)).To(gomega.BeNil())

			keyed := backupCapablePlatform()
			keyed.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: &KubernetesPlantonPlatformObjectStore{
					DestinationPath: "https://acme.blob.core.windows.net/backups/platform",
					Backend: &KubernetesPlantonPlatformObjectStore_AzureBlob{AzureBlob: &KubernetesPlantonPlatformAzureBlobObjectStore{
						StorageAccount: "acme", ConnectionString: "DefaultEndpointsProtocol=https;AccountName=acme;AccountKey=...",
					}},
				},
			})
			gomega.Expect(protovalidate.Validate(keyed)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a recovery from R2 beside the platform's own backup, with and without a target time", func() {
			input := backupCapablePlatform()
			input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: r2StoreByReference("s3://acme-platform-backups/platform"),
			})
			input.Spec.Database.Postgresql.RecoverFrom = &KubernetesPlantonPlatformPostgresqlRecoverFrom{
				ObjectStore: r2StoreByReference("s3://acme-platform-backups/platform"),
				ServerName:  "acme-postgres-1a2b3c4d",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			for _, at := range []string{"2026-09-13T20:30:00Z", "2026-09-13T20:30:00.123456+05:30"} {
				input.Spec.Database.Postgresql.RecoverFrom.TargetTime = at
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "target_time %q", at)
			}
		})

		ginkgo.It("should accept every word of the backup plugin prerequisite", func() {
			for _, word := range []string{"", "auto", "skip"} {
				input := minimalValidPlatform()
				input.Spec.Prerequisites = &KubernetesPlantonPlatformPrerequisites{PostgresBackupPlugin: strPtr(word)}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "postgres_backup_plugin %q", word)
			}
		})

		// ---- the vault's seal and keys ------------------------------------------------

		ginkgo.It("should accept a vault with only a named keys Secret (the built-in seal, keys you own)", func() {
			input := minimalValidPlatform()
			input.Spec.Vault = &KubernetesPlantonPlatformVault{InitSecretName: "planton-vault-keys"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an AWS KMS seal, keyless and with static keys", func() {
			keyless := minimalValidPlatform()
			keyless.Spec.Vault = &KubernetesPlantonPlatformVault{
				AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{Seal: &KubernetesPlantonPlatformVaultAutoUnseal_AwsKms{
					AwsKms: &KubernetesPlantonPlatformVaultAwsKmsSeal{Region: "us-west-2", KmsKeyId: "alias/planton-vault-unseal"},
				}},
				ServiceAccountAnnotations: map[string]string{"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/planton-vault-unseal"},
			}
			gomega.Expect(protovalidate.Validate(keyless)).To(gomega.BeNil())

			keyed := minimalValidPlatform()
			keyed.Spec.Vault = &KubernetesPlantonPlatformVault{
				AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{Seal: &KubernetesPlantonPlatformVaultAutoUnseal_AwsKms{
					AwsKms: &KubernetesPlantonPlatformVaultAwsKmsSeal{
						Region: "us-west-2", KmsKeyId: "alias/planton-vault-unseal",
						AccessKeyId: "AKIA-example", SecretAccessKey: "secret",
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(keyed)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a GCP KMS seal by reference and by literal, with and without a workload identity", func() {
			byReference := minimalValidPlatform()
			byReference.Spec.Vault = &KubernetesPlantonPlatformVault{
				AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{Seal: &KubernetesPlantonPlatformVaultAutoUnseal_GcpKms{
					GcpKms: &KubernetesPlantonPlatformVaultGcpKmsSeal{
						Project:                        refTo(cloudresourcekind.CloudResourceKind_GcpProject, "acme-platform", "status.outputs.project_id"),
						Region:                         "global",
						KeyRing:                        refTo(cloudresourcekind.CloudResourceKind_GcpKmsKeyRing, "planton-vault-unseal", "status.outputs.key_ring_name"),
						CryptoKey:                      refTo(cloudresourcekind.CloudResourceKind_GcpKmsKey, "planton-vault-unseal", "status.outputs.key_name"),
						WorkloadIdentityServiceAccount: refTo(cloudresourcekind.CloudResourceKind_GcpServiceAccount, "planton-vault-unseal", "status.outputs.email"),
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(byReference)).To(gomega.BeNil())

			literal := minimalValidPlatform()
			literal.Spec.Vault = &KubernetesPlantonPlatformVault{
				AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{Seal: &KubernetesPlantonPlatformVaultAutoUnseal_GcpKms{
					GcpKms: &KubernetesPlantonPlatformVaultGcpKmsSeal{
						Project: literalRef("acme-platform"), Region: "us-central1",
						KeyRing: literalRef("planton-vault-unseal"), CryptoKey: literalRef("planton-vault-unseal"),
					},
				}},
				ServiceAccountAnnotations: map[string]string{"iam.gke.io/gcp-service-account": "planton-vault-unseal@acme-platform.iam.gserviceaccount.com"},
			}
			gomega.Expect(protovalidate.Validate(literal)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an Azure Key Vault seal, keyless and with a service principal", func() {
			keyless := minimalValidPlatform()
			keyless.Spec.Vault = &KubernetesPlantonPlatformVault{
				AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{Seal: &KubernetesPlantonPlatformVaultAutoUnseal_AzureKeyVault{
					AzureKeyVault: &KubernetesPlantonPlatformVaultAzureKeyVaultSeal{VaultName: "acme-kv", KeyName: "planton-vault-unseal", TenantId: "tenant"},
				}},
				ServiceAccountAnnotations: map[string]string{"azure.workload.identity/client-id": "00000000-0000-0000-0000-000000000000"},
			}
			gomega.Expect(protovalidate.Validate(keyless)).To(gomega.BeNil())

			keyed := minimalValidPlatform()
			keyed.Spec.Vault = &KubernetesPlantonPlatformVault{
				AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{Seal: &KubernetesPlantonPlatformVaultAutoUnseal_AzureKeyVault{
					AzureKeyVault: &KubernetesPlantonPlatformVaultAzureKeyVaultSeal{
						VaultName: "acme-kv", KeyName: "planton-vault-unseal", TenantId: "tenant",
						ClientId: "client", ClientSecret: "secret",
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(keyed)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a transit seal with and without the default mount path", func() {
			input := minimalValidPlatform()
			input.Spec.Vault = &KubernetesPlantonPlatformVault{
				AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{Seal: &KubernetesPlantonPlatformVaultAutoUnseal_Transit{
					Transit: &KubernetesPlantonPlatformVaultTransitSeal{Address: "http://key-holder.openbao.svc:8200", KeyName: "autounseal", Token: "s.token"},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.Vault.AutoUnseal.GetTransit().MountPath = strPtr("keys/")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a backup beside a seal, beside a named keys Secret, and beside an opted-out vault with a cloud backend", func() {
			sealed := minimalValidPlatform()
			sealed.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{ObjectStore: r2StoreByReference("s3://acme-platform-backups/platform")})
			sealed.Spec.Vault = &KubernetesPlantonPlatformVault{
				AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{Seal: &KubernetesPlantonPlatformVaultAutoUnseal_Transit{
					Transit: &KubernetesPlantonPlatformVaultTransitSeal{Address: "http://key-holder.openbao.svc:8200", KeyName: "autounseal"},
				}},
			}
			gomega.Expect(protovalidate.Validate(sealed)).To(gomega.BeNil())

			named := backupCapablePlatform()
			named.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{ObjectStore: r2StoreByReference("s3://acme-platform-backups/platform")})
			gomega.Expect(protovalidate.Validate(named)).To(gomega.BeNil())

			off := false
			optedOut := minimalValidPlatform()
			optedOut.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{ObjectStore: r2StoreByReference("s3://acme-platform-backups/platform")})
			optedOut.Spec.Vault = &KubernetesPlantonPlatformVault{Enabled: &off}
			optedOut.Spec.Bootstrap = &KubernetesPlantonPlatformBootstrap{SecretBackend: &KubernetesPlantonPlatformSecretBackend{
				Type:              "awsSecretsManager",
				AwsSecretsManager: &KubernetesPlantonPlatformAwsSecretsManager{Region: "us-east-1", KmsKeyArn: "arn:aws:kms:us-east-1:123456789012:key/abc"},
			}}
			gomega.Expect(protovalidate.Validate(optedOut)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should fail when version is missing", func() {
			input := minimalValidPlatform()
			input.Spec.Version = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should refuse a module-release override that is not an exact release tag", func() {
			// Only exact releases publish module artifacts; a pre-release, a bare
			// number, or a branch would 404 at every module download.
			for _, notARelease := range []string{"v0.5.60-rc.1", "0.5.60", "main", "latest"} {
				input := minimalValidPlatform()
				input.Spec.ControlPlane = &KubernetesPlantonPlatformControlPlane{IacModulesVersion: notARelease}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil(), notARelease)
			}
		})

		ginkgo.It("should fail when namespace is missing", func() {
			input := minimalValidPlatform()
			input.Spec.Namespace = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on TLS without a hostname", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled: true,
				Tls: &KubernetesPlantonPlatformIngressTls{
					SecretName: "planton-tls",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on TLS with BOTH a Secret and an issuer", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:  true,
				Hostname: "planton.example.com",
				Tls: &KubernetesPlantonPlatformIngressTls{
					SecretName: "planton-tls",
					Issuer: &KubernetesPlantonPlatformCertManagerIssuer{
						Name: "letsencrypt",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		// The two Gateway rules mirror the operator's CRD word for word, so a
		// declaration the catalog accepts is one the API server accepts.
		ginkgo.It("should fail when gateway_ref and ingress_class_name name two front doors", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:          true,
				Hostname:         "planton.example.com",
				IngressClassName: "nginx",
				GatewayRef:       &KubernetesPlantonPlatformGatewayRef{Name: literalRef("main")},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on gateway_ref with a brought certificate Secret (the listener owns it)", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:    true,
				Hostname:   "planton.example.com",
				GatewayRef: &KubernetesPlantonPlatformGatewayRef{Name: literalRef("main")},
				Tls:        &KubernetesPlantonPlatformIngressTls{SecretName: "planton-tls"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on reachability public with the ingress disabled (a port-forward door is never public)", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{Reachability: strPtr("public")}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("reached only through kubectl port-forward"))
		})

		ginkgo.It("should fail on a reachability word outside auto, public, private", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:      true,
				Hostname:     "planton.example.com",
				Reachability: strPtr("internet"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on a gateway_ref without a name", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:    true,
				Hostname:   "planton.example.com",
				GatewayRef: &KubernetesPlantonPlatformGatewayRef{Namespace: literalRef("istio-ingress")},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		// The Gateway may be Planton's own: both fields are KubernetesGateway
		// foreign keys, so an infra chart wires them with valueFrom and the
		// platform deploys after the Gateway it attaches to.
		ginkgo.It("should accept gateway_ref name and namespace as references to a KubernetesGateway", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:  true,
				Hostname: "planton.example.com",
				GatewayRef: &KubernetesPlantonPlatformGatewayRef{
					Name:        gatewayRef("management-gateway", ""),
					Namespace:   gatewayRef("management-gateway", ""),
					SectionName: "https",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a referenced Gateway name beside a literal namespace", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:  true,
				Hostname: "planton.example.com",
				GatewayRef: &KubernetesPlantonPlatformGatewayRef{
					Name:      gatewayRef("management-gateway", "status.outputs.gateway_name"),
					Namespace: literalRef("istio-ingress"),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should fail on a gateway_ref name that is present but empty", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:    true,
				Hostname:   "planton.example.com",
				GatewayRef: &KubernetesPlantonPlatformGatewayRef{Name: literalRef("")},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on a gateway_ref namespace that is present but empty", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:    true,
				Hostname:   "planton.example.com",
				GatewayRef: &KubernetesPlantonPlatformGatewayRef{Name: literalRef("main"), Namespace: literalRef("")},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on TLS with NEITHER a Secret nor an issuer", func() {
			input := minimalValidPlatform()
			input.Spec.Ingress = &KubernetesPlantonPlatformIngress{
				Enabled:  true,
				Hostname: "planton.example.com",
				Tls:      &KubernetesPlantonPlatformIngressTls{},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on the AWS secret backend without its config", func() {
			input := minimalValidPlatform()
			input.Spec.Bootstrap = &KubernetesPlantonPlatformBootstrap{
				SecretBackend: &KubernetesPlantonPlatformSecretBackend{
					Type: "awsSecretsManager",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on a license with BOTH key and secret reference", func() {
			input := minimalValidPlatform()
			input.Spec.License = &KubernetesPlantonPlatformLicense{
				Key: "plk_FAKE_PLACEHOLDER_VALUE",
				SecretKeyRef: &KubernetesPlantonPlatformSecretKeyRef{
					Name: "planton-license",
					Key:  "key",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on a malformed storage size", func() {
			input := minimalValidPlatform()
			input.Spec.Storage = &KubernetesPlantonPlatformStorage{
				Size: "ten gigs",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on a malformed admin email", func() {
			input := minimalValidPlatform()
			input.Spec.Identity = &KubernetesPlantonPlatformIdentity{
				AdminEmail: "not-an-email",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on an unknown IaC provisioner", func() {
			input := minimalValidPlatform()
			provisioner := "pulumi"
			input.Spec.Bootstrap = &KubernetesPlantonPlatformBootstrap{
				IacProvisioner: &provisioner,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on an out-of-range gateway port", func() {
			input := minimalValidPlatform()
			port := int32(70000)
			input.Spec.Gateway = &KubernetesPlantonPlatformGateway{
				LocalPort: &port,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		// Email: the three rules the platform's own definition enforces,
		// mirrored so a manifest is refused here in the same words it would be
		// refused by the cluster.
		ginkgo.It("should fail on email with BOTH an SMTP relay and a Resend account", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From:   &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				Smtp:   &KubernetesPlantonPlatformEmailSmtp{Host: "smtp.office365.com"},
				Resend: &KubernetesPlantonPlatformEmailResend{ApiKeySecretRef: &KubernetesPlantonPlatformSecretKeyRef{Name: "planton-email", Key: "api-key"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("never both, never neither"))
		})

		ginkgo.It("should fail on email with neither provider arm", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on email without a from address", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{},
				Smtp: &KubernetesPlantonPlatformEmailSmtp{Host: "smtp.office365.com"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on an SMTP relay with BOTH a credentials Secret and OAuth2", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				Smtp: &KubernetesPlantonPlatformEmailSmtp{
					Host:                  "smtp.office365.com",
					CredentialsSecretName: "planton-email",
					Oauth2: &KubernetesPlantonPlatformEmailSmtpOauth2{
						User: "u", TokenUrl: "https://login.microsoftonline.com/t/oauth2/v2.0/token", Scope: "s", ClientId: "c",
						ClientSecretRef: &KubernetesPlantonPlatformSecretKeyRef{Name: "n", Key: "k"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("smtp authenticates one way"))
		})

		ginkgo.It("should fail on credentials over a plaintext connection (security: none would send them in the clear)", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				Smtp: &KubernetesPlantonPlatformEmailSmtp{
					Host:                  "smtp-relay.corp.acme.com",
					Security:              strPtr("none"),
					CredentialsSecretName: "planton-email",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("would send credentials in the clear"))
		})

		ginkgo.It("should fail on a security word outside starttls, tls, none", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				Smtp: &KubernetesPlantonPlatformEmailSmtp{Host: "smtp.office365.com", Security: strPtr("ssl")},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on an OAuth2 token URL that is not https", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				Smtp: &KubernetesPlantonPlatformEmailSmtp{
					Host: "smtp.office365.com",
					Oauth2: &KubernetesPlantonPlatformEmailSmtpOauth2{
						User: "u", TokenUrl: "http://login.microsoftonline.com/t/oauth2/v2.0/token", Scope: "s", ClientId: "c",
						ClientSecretRef: &KubernetesPlantonPlatformSecretKeyRef{Name: "n", Key: "k"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("https://"))
		})

		ginkgo.It("should fail on an out-of-range SMTP port", func() {
			input := minimalValidPlatform()
			port := int32(0)
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From: &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				Smtp: &KubernetesPlantonPlatformEmailSmtp{Host: "smtp.office365.com", Port: &port},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on Resend without the API key reference", func() {
			input := minimalValidPlatform()
			input.Spec.Email = &KubernetesPlantonPlatformEmail{
				From:   &KubernetesPlantonPlatformEmailFrom{Address: "no-reply@planton.acme.com"},
				Resend: &KubernetesPlantonPlatformEmailResend{},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		// ---- the database's backup and recovery ------------------------------------

		ginkgo.It("should fail on a backup without an object store", func() {
			input := backupCapablePlatform()
			input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{RetentionPolicy: strPtr("30d")})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on an object store without a backend arm, or without a destination path", func() {
			noArm := backupCapablePlatform()
			noArm.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: &KubernetesPlantonPlatformObjectStore{DestinationPath: "s3://acme-backups/platform"},
			})
			gomega.Expect(protovalidate.Validate(noArm)).NotTo(gomega.BeNil())

			noPath := backupCapablePlatform()
			noPath.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: r2StoreByReference(""),
			})
			gomega.Expect(protovalidate.Validate(noPath)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail when the destination path's scheme does not match the backend", func() {
			cases := map[string]*KubernetesPlantonPlatformObjectStore{
				"the s3 backend stores at an s3:// destination path": {
					DestinationPath: "gs://acme-backups/platform",
					Backend:         &KubernetesPlantonPlatformObjectStore_S3{S3: &KubernetesPlantonPlatformS3ObjectStore{Keyless: true}},
				},
				"the gcs backend stores at a gs:// destination path": {
					DestinationPath: "s3://acme-backups/platform",
					Backend:         &KubernetesPlantonPlatformObjectStore_Gcs{Gcs: &KubernetesPlantonPlatformGcsObjectStore{Keyless: true}},
				},
				"the azure_blob backend stores at an https:// destination path": {
					DestinationPath: "s3://acme-backups/platform",
					Backend: &KubernetesPlantonPlatformObjectStore_AzureBlob{AzureBlob: &KubernetesPlantonPlatformAzureBlobObjectStore{
						StorageAccount: "acme", Keyless: true,
					}},
				},
				"the r2 backend stores at an s3:// destination path": r2StoreByReference("gs://acme-backups/platform"),
			}
			for message, store := range cases {
				input := backupCapablePlatform()
				input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{ObjectStore: store})
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil(), message)
				gomega.Expect(err.Error()).To(gomega.ContainSubstring(message))
			}
		})

		ginkgo.It("should fail on a retention that is not a positive number of days, weeks, or months", func() {
			for _, retention := range []string{"0d", "30x", "d30", "30"} {
				input := backupCapablePlatform()
				input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
					ObjectStore:     r2StoreByReference("s3://acme-platform-backups/platform"),
					RetentionPolicy: strPtr(retention),
				})
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil(), "retention %q", retention)
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("retention_policy must be a positive number of days, weeks, or months"))
			}
		})

		ginkgo.It("should fail on a five-field cron schedule and name the missing seconds field", func() {
			input := backupCapablePlatform()
			input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: r2StoreByReference("s3://acme-platform-backups/platform"),
				Schedule:    strPtr("0 2 * * *"),
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("SIX-field cron expression"))
		})

		ginkgo.It("should fail on an R2 store missing its account, its credentials, or its secret key", func() {
			noAccount := r2StoreByReference("s3://acme-platform-backups/platform")
			noAccount.GetR2().AccountId = nil
			noCredentials := r2StoreByReference("s3://acme-platform-backups/platform")
			noCredentials.GetR2().Credentials = nil
			noSecret := r2StoreByReference("s3://acme-platform-backups/platform")
			noSecret.GetR2().Credentials.SecretAccessKey = nil
			for name, store := range map[string]*KubernetesPlantonPlatformObjectStore{
				"account": noAccount, "credentials": noCredentials, "secret key": noSecret,
			} {
				input := backupCapablePlatform()
				input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{ObjectStore: store})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "missing %s", name)
			}
		})

		ginkgo.It("should fail on an R2 account id that is not 32 hex characters, or an unknown jurisdiction", func() {
			badAccount := backupCapablePlatform()
			store := r2StoreByReference("s3://acme-platform-backups/platform")
			store.GetR2().AccountId = literalRef("acme")
			badAccount.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{ObjectStore: store})
			err := protovalidate.Validate(badAccount)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("32-hex-character Cloudflare account id"))

			badJurisdiction := backupCapablePlatform()
			store = r2StoreByReference("s3://acme-platform-backups/platform")
			store.GetR2().Jurisdiction = literalRef("europe")
			badJurisdiction.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{ObjectStore: store})
			err = protovalidate.Validate(badJurisdiction)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("jurisdiction must be one of"))
		})

		ginkgo.It("should fail on S3 with both postures, neither posture, or keyless against a compatible endpoint", func() {
			both := &KubernetesPlantonPlatformS3ObjectStore{
				Keyless:    true,
				AccessKeys: &KubernetesPlantonPlatformS3AccessKeys{AccessKeyId: "a", SecretAccessKey: "b"},
			}
			neither := &KubernetesPlantonPlatformS3ObjectStore{Region: "us-west-2"}
			compatibleKeyless := &KubernetesPlantonPlatformS3ObjectStore{EndpointUrl: "http://minio.minio-system.svc:9000", Keyless: true}
			for name, s3 := range map[string]*KubernetesPlantonPlatformS3ObjectStore{
				"both": both, "neither": neither, "compatible keyless": compatibleKeyless,
			} {
				input := backupCapablePlatform()
				input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
					ObjectStore: &KubernetesPlantonPlatformObjectStore{
						DestinationPath: "s3://acme-backups/platform",
						Backend:         &KubernetesPlantonPlatformObjectStore_S3{S3: s3},
					},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), name)
			}
		})

		ginkgo.It("should fail on an S3 endpoint that is not an http(s) URL, and on access keys missing a half", func() {
			badEndpoint := backupCapablePlatform()
			badEndpoint.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: &KubernetesPlantonPlatformObjectStore{
					DestinationPath: "s3://backups/platform",
					Backend: &KubernetesPlantonPlatformObjectStore_S3{S3: &KubernetesPlantonPlatformS3ObjectStore{
						EndpointUrl: "minio.minio-system.svc:9000",
						AccessKeys:  &KubernetesPlantonPlatformS3AccessKeys{AccessKeyId: "a", SecretAccessKey: "b"},
					}},
				},
			})
			gomega.Expect(protovalidate.Validate(badEndpoint)).NotTo(gomega.BeNil())

			halfKeys := backupCapablePlatform()
			halfKeys.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
				ObjectStore: &KubernetesPlantonPlatformObjectStore{
					DestinationPath: "s3://backups/platform",
					Backend: &KubernetesPlantonPlatformObjectStore_S3{S3: &KubernetesPlantonPlatformS3ObjectStore{
						AccessKeys: &KubernetesPlantonPlatformS3AccessKeys{AccessKeyId: "a"},
					}},
				},
			})
			gomega.Expect(protovalidate.Validate(halfKeys)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail on GCS with both postures or neither", func() {
			for name, gcs := range map[string]*KubernetesPlantonPlatformGcsObjectStore{
				"both":    {Keyless: true, ServiceAccountKeyJson: "{}"},
				"neither": {},
			} {
				input := backupCapablePlatform()
				input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
					ObjectStore: &KubernetesPlantonPlatformObjectStore{
						DestinationPath: "gs://acme-backups/platform",
						Backend:         &KubernetesPlantonPlatformObjectStore_Gcs{Gcs: gcs},
					},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), name)
			}
		})

		ginkgo.It("should fail on Azure Blob with both postures, neither, or no storage account", func() {
			for name, azure := range map[string]*KubernetesPlantonPlatformAzureBlobObjectStore{
				"both":       {StorageAccount: "acme", Keyless: true, ConnectionString: "x"},
				"neither":    {StorageAccount: "acme"},
				"no account": {Keyless: true},
			} {
				input := backupCapablePlatform()
				input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{
					ObjectStore: &KubernetesPlantonPlatformObjectStore{
						DestinationPath: "https://acme.blob.core.windows.net/backups/platform",
						Backend:         &KubernetesPlantonPlatformObjectStore_AzureBlob{AzureBlob: azure},
					},
				})
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), name)
			}
		})

		ginkgo.It("should fail on a recovery without a server name, or with a target time that is not RFC 3339", func() {
			noServer := minimalValidPlatform()
			noServer.Spec.Database = &KubernetesPlantonPlatformDatabase{Postgresql: &KubernetesPlantonPlatformPostgresql{
				RecoverFrom: &KubernetesPlantonPlatformPostgresqlRecoverFrom{
					ObjectStore: r2StoreByReference("s3://acme-platform-backups/platform"),
				},
			}}
			gomega.Expect(protovalidate.Validate(noServer)).NotTo(gomega.BeNil())

			badTime := minimalValidPlatform()
			badTime.Spec.Database = &KubernetesPlantonPlatformDatabase{Postgresql: &KubernetesPlantonPlatformPostgresql{
				RecoverFrom: &KubernetesPlantonPlatformPostgresqlRecoverFrom{
					ObjectStore: r2StoreByReference("s3://acme-platform-backups/platform"),
					ServerName:  "acme-postgres-1a2b3c4d",
					TargetTime:  "2026-09-13 20:30:00",
				},
			}}
			err := protovalidate.Validate(badTime)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("target_time is an RFC 3339 timestamp"))
		})

		ginkgo.It("should fail on an unknown backup plugin word", func() {
			input := minimalValidPlatform()
			input.Spec.Prerequisites = &KubernetesPlantonPlatformPrerequisites{PostgresBackupPlugin: strPtr("install")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		// ---- the vault's seal and keys ------------------------------------------------

		ginkgo.It("should refuse a backup whose vault keys would die with the platform", func() {
			input := minimalValidPlatform()
			input.Spec.Database = withBackup(&KubernetesPlantonPlatformPostgresqlBackup{ObjectStore: r2StoreByReference("s3://acme-platform-backups/platform")})
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("set vault.init_secret_name to a Secret you own"))

			// An explicit vault block that names nothing is the same posture.
			input.Spec.Vault = &KubernetesPlantonPlatformVault{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should refuse a seal, a keys Secret, or an identity on an opted-out vault", func() {
			off := false
			for name, vault := range map[string]*KubernetesPlantonPlatformVault{
				"seal": {Enabled: &off, AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{Seal: &KubernetesPlantonPlatformVaultAutoUnseal_Transit{
					Transit: &KubernetesPlantonPlatformVaultTransitSeal{Address: "http://key-holder.openbao.svc:8200", KeyName: "autounseal"},
				}}},
				"keys Secret": {Enabled: &off, InitSecretName: "planton-vault-keys"},
				"identity":    {Enabled: &off, ServiceAccountAnnotations: map[string]string{"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/x"}},
			} {
				input := minimalValidPlatform()
				input.Spec.Vault = vault
				input.Spec.Bootstrap = &KubernetesPlantonPlatformBootstrap{SecretBackend: &KubernetesPlantonPlatformSecretBackend{
					Type:              "awsSecretsManager",
					AwsSecretsManager: &KubernetesPlantonPlatformAwsSecretsManager{Region: "us-east-1", KmsKeyArn: "arn:aws:kms:us-east-1:123456789012:key/abc"},
				}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil(), name)
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("vault.enabled: false opts out"), name)
			}
		})

		ginkgo.It("should refuse the platform secret backend on an opted-out vault", func() {
			off := false
			input := minimalValidPlatform()
			input.Spec.Vault = &KubernetesPlantonPlatformVault{Enabled: &off}
			input.Spec.Bootstrap = &KubernetesPlantonPlatformBootstrap{SecretBackend: &KubernetesPlantonPlatformSecretBackend{Type: "platform"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("re-enable the vault or use type awsSecretsManager"))
		})

		ginkgo.It("should refuse an auto_unseal block that names no seal", func() {
			input := minimalValidPlatform()
			input.Spec.Vault = &KubernetesPlantonPlatformVault{AutoUnseal: &KubernetesPlantonPlatformVaultAutoUnseal{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should refuse each seal arm missing what its key needs", func() {
			cases := map[string]*KubernetesPlantonPlatformVaultAutoUnseal{
				"aws without a region": {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_AwsKms{AwsKms: &KubernetesPlantonPlatformVaultAwsKmsSeal{KmsKeyId: "alias/x"}}},
				"aws without a key":    {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_AwsKms{AwsKms: &KubernetesPlantonPlatformVaultAwsKmsSeal{Region: "us-west-2"}}},
				"gcp without a project": {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_GcpKms{GcpKms: &KubernetesPlantonPlatformVaultGcpKmsSeal{
					Region: "global", KeyRing: literalRef("ring"), CryptoKey: literalRef("key"),
				}}},
				"gcp without a region": {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_GcpKms{GcpKms: &KubernetesPlantonPlatformVaultGcpKmsSeal{
					Project: literalRef("p"), KeyRing: literalRef("ring"), CryptoKey: literalRef("key"),
				}}},
				"gcp without a key ring": {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_GcpKms{GcpKms: &KubernetesPlantonPlatformVaultGcpKmsSeal{
					Project: literalRef("p"), Region: "global", CryptoKey: literalRef("key"),
				}}},
				"gcp without a crypto key": {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_GcpKms{GcpKms: &KubernetesPlantonPlatformVaultGcpKmsSeal{
					Project: literalRef("p"), Region: "global", KeyRing: literalRef("ring"),
				}}},
				"azure without a vault name": {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_AzureKeyVault{AzureKeyVault: &KubernetesPlantonPlatformVaultAzureKeyVaultSeal{KeyName: "k", TenantId: "t"}}},
				"azure without a key name":   {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_AzureKeyVault{AzureKeyVault: &KubernetesPlantonPlatformVaultAzureKeyVaultSeal{VaultName: "v", TenantId: "t"}}},
				"azure without a tenant":     {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_AzureKeyVault{AzureKeyVault: &KubernetesPlantonPlatformVaultAzureKeyVaultSeal{VaultName: "v", KeyName: "k"}}},
				"transit without an address": {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_Transit{Transit: &KubernetesPlantonPlatformVaultTransitSeal{KeyName: "k"}}},
				"transit without a key name": {Seal: &KubernetesPlantonPlatformVaultAutoUnseal_Transit{Transit: &KubernetesPlantonPlatformVaultTransitSeal{Address: "http://key-holder:8200"}}},
			}
			for name, seal := range cases {
				input := minimalValidPlatform()
				input.Spec.Vault = &KubernetesPlantonPlatformVault{AutoUnseal: seal}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), name)
			}
		})
	})
})

// backupCapablePlatform is the zero-config platform with the one thing a
// backup declaration requires of the vault: keys that outlive the platform.
// The archive carries the vault's data, and the spec refuses a backup whose
// vault keys would die with the platform; naming the keys Secret is the
// bare-metal answer (a cloud seal is the other), so every backup fixture
// starts here and each refusal below fires for its own reason, never for
// the missing keys.
func backupCapablePlatform() *KubernetesPlantonPlatform {
	input := minimalValidPlatform()
	input.Spec.Vault = &KubernetesPlantonPlatformVault{InitSecretName: "planton-vault-keys"}
	return input
}

// withBackup wraps a backup declaration in the database block, the way a
// manifest declares it under spec.database.postgresql.backup.
func withBackup(backup *KubernetesPlantonPlatformPostgresqlBackup) *KubernetesPlantonPlatformDatabase {
	return &KubernetesPlantonPlatformDatabase{
		Postgresql: &KubernetesPlantonPlatformPostgresql{Backup: backup},
	}
}

// r2StoreByReference is the composed-reference shape the estate declares:
// account and jurisdiction from the bucket resource, the key pair from the
// token resource. Nothing typed.
func r2StoreByReference(path string) *KubernetesPlantonPlatformObjectStore {
	bucket := func(fieldPath string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind: cloudresourcekind.CloudResourceKind_CloudflareR2Bucket, Name: "acme-platform-backups", FieldPath: fieldPath,
			},
		}}
	}
	token := func(fieldPath string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind: cloudresourcekind.CloudResourceKind_CloudflareAccountApiToken, Name: "acme-platform-backups-writer", FieldPath: fieldPath,
			},
		}}
	}
	return &KubernetesPlantonPlatformObjectStore{
		DestinationPath: path,
		Backend: &KubernetesPlantonPlatformObjectStore_R2{R2: &KubernetesPlantonPlatformR2ObjectStore{
			AccountId:    bucket("status.outputs.account_id"),
			Jurisdiction: bucket("status.outputs.jurisdiction"),
			Credentials: &KubernetesPlantonPlatformR2Credentials{
				AccessKeyId:     token("status.outputs.r2_access_key_id"),
				SecretAccessKey: token("status.outputs.r2_secret_access_key"),
			},
		}},
	}
}
