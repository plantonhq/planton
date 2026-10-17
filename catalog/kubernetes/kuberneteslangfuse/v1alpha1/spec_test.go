package kuberneteslangfusev1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesLangfuse(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesLangfuse Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name, FieldPath: fieldPath},
		},
	}
}

// violations lists every violation of a validation error as both its rule id
// and its field path, so a test can name either.
func violations(err error) []string {
	gomega.Expect(err).NotTo(gomega.BeNil())
	var verr *protovalidate.ValidationError
	gomega.Expect(errors.As(err, &verr)).To(gomega.BeTrue())
	out := []string{}
	for _, v := range verr.Violations {
		out = append(out, v.Proto.GetRuleId(), protovalidate.FieldPathString(v.Proto.GetField()))
	}
	return out
}

func seaweedTarget() *KubernetesLangfuseBlobTarget {
	return &KubernetesLangfuseBlobTarget{Backend: &KubernetesLangfuseBlobTarget_S3Compatible{
		S3Compatible: &KubernetesLangfuseS3CompatibleTarget{
			Endpoint: valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "store", "status.outputs.s3_endpoint"),
			Bucket:   "langfuse",
			Credentials: &KubernetesLangfuseS3CompatibleTarget_CredentialsSecret{CredentialsSecret: &KubernetesLangfuseS3CredentialsSecret{
				SecretName: valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "store", "status.outputs.s3_credentials_secret_name"),
			}},
		},
	}}
}

func gcsTarget() *KubernetesLangfuseBlobTarget {
	return &KubernetesLangfuseBlobTarget{Backend: &KubernetesLangfuseBlobTarget_Gcs{
		Gcs: &KubernetesLangfuseGcsTarget{Bucket: literal("langfuse-events")},
	}}
}

func oauthClient() *KubernetesLangfuseOAuthClient {
	return &KubernetesLangfuseOAuthClient{ClientId: "client", ClientSecret: "$secret/sso"}
}

var _ = ginkgo.Describe("KubernetesLangfuse Validation Tests", func() {
	var input *KubernetesLangfuse

	ginkgo.BeforeEach(func() {
		input = &KubernetesLangfuse{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesLangfuse",
			Metadata:   &shared.CatalogObjectMetadata{Name: "langfuse"},
			Spec: &KubernetesLangfuseSpec{
				Namespace: literal("langfuse"),
				Postgres: &KubernetesLangfusePostgres{
					Host:     valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "db", "status.outputs.rw_service"),
					Username: "langfuse",
					PasswordSecret: &KubernetesLangfusePostgresPasswordSecret{
						SecretName: valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "db", "status.outputs.password_secret.name"),
					},
				},
				Clickhouse: &KubernetesLangfuseClickHouse{
					Host:     valueFrom(catalogkind.CatalogKind_KubernetesClickHouse, "ch", "status.outputs.service_name"),
					Username: "langfuse",
					PasswordSecret: &KubernetesLangfuseClickHousePasswordSecret{
						SecretName: valueFrom(catalogkind.CatalogKind_KubernetesClickHouse, "ch", "status.outputs.auth_secret_name"),
					},
				},
				Redis: &KubernetesLangfuseRedis{
					Topology: &KubernetesLangfuseRedis_Standalone{Standalone: &KubernetesLangfuseRedisStandalone{
						Host: valueFrom(catalogkind.CatalogKind_KubernetesValkey, "cache", "status.outputs.service"),
					}},
					Username: valueFrom(catalogkind.CatalogKind_KubernetesValkey, "cache", "status.outputs.username"),
					PasswordSecret: &KubernetesLangfuseRedisPasswordSecret{
						SecretName: valueFrom(catalogkind.CatalogKind_KubernetesValkey, "cache", "status.outputs.password_secret.name"),
						SecretKey:  valueFrom(catalogkind.CatalogKind_KubernetesValkey, "cache", "status.outputs.password_secret.key"),
					},
				},
				BlobStorage: &KubernetesLangfuseBlobStorage{Target: seaweedTarget()},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts the composed minimal spec", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a fully configured production spec", func() {
			spec := input.Spec
			spec.ExternalUrl = strPtr("https://langfuse.example.com")
			spec.Clickhouse.Cluster = &KubernetesLangfuseClickHouseCluster{
				Enabled: boolPtr(true),
				Name:    valueFrom(catalogkind.CatalogKind_KubernetesClickHouse, "ch", "status.outputs.cluster_name"),
			}
			spec.Clickhouse.DeleteMode = KubernetesLangfuseClickHouse_lightweight_update
			spec.Redis.Tls = &KubernetesLangfuseRedisTls{
				Enabled: true,
				Ca: &KubernetesLangfuseCertificateFile{
					Source: &KubernetesLangfuseCertificateFile_SecretName{SecretName: "redis-ca"},
				},
			}
			spec.BlobStorage.MediaUpload = &KubernetesLangfuseMediaUpload{
				PublicEndpoint: "https://s3.langfuse.example.com",
				Sse:            &KubernetesLangfuseServerSideEncryption{Algorithm: KubernetesLangfuseServerSideEncryption_aes256},
			}
			spec.BlobStorage.BatchExport = &KubernetesLangfuseBatchExport{PublicEndpoint: "https://s3.langfuse.example.com"}
			spec.Secrets = &KubernetesLangfuseInstanceSecrets{
				Salt:          "$secret/langfuse-salt",
				EncryptionKey: "$secret/langfuse-encryption-key",
			}
			spec.Init = &KubernetesLangfuseInit{
				Admin: &KubernetesLangfuseInitAdmin{Email: "admin@example.com"},
			}
			spec.Auth = &KubernetesLangfuseAuth{
				DisableSignUp: true,
				NewUserAccess: &KubernetesLangfuseNewUserAccess{
					OrganizationId:   "langfuse",
					OrganizationRole: KubernetesLangfuseNewUserAccess_MEMBER,
					ProjectRole:      KubernetesLangfuseNewUserAccess_VIEWER,
				},
				Sso: &KubernetesLangfuseSsoProviders{
					Okta: &KubernetesLangfuseIssuerSso{Client: oauthClient(), Issuer: "https://example.okta.com"},
				},
			}
			spec.AiFeatures = &KubernetesLangfuseAiFeatures{
				Provider:  KubernetesLangfuseAiFeatures_anthropic,
				Model:     "claude-opus-5",
				ApiKey:    "$secret/anthropic",
				Assistant: &KubernetesLangfuseAssistant{Enabled: true},
			}
			spec.Web = &KubernetesLangfuseWeb{
				Replicas: int32Ptr(2),
				Autoscaling: &KubernetesLangfuseAutoscaling{
					Hpa: &KubernetesLangfuseHpa{MinReplicas: int32Ptr(2), MaxReplicas: int32Ptr(6)},
				},
			}
			spec.ServiceAccount = &KubernetesLangfuseServiceAccount{
				WorkloadIdentity: &kubernetes.KubernetesWorkloadIdentity{Provider: &kubernetes.KubernetesWorkloadIdentity_Eks{
					Eks: &kubernetes.KubernetesWorkloadIdentityEksIrsa{RoleArn: literal("arn:aws:iam::123456789012:role/langfuse")},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts gcs targets for every upload type", func() {
			input.Spec.BlobStorage = &KubernetesLangfuseBlobStorage{
				Target:      gcsTarget(),
				MediaUpload: &KubernetesLangfuseMediaUpload{Target: gcsTarget()},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a Redis Sentinel topology with TLS to the Sentinels", func() {
			input.Spec.Redis.Topology = &KubernetesLangfuseRedis_Sentinel{Sentinel: &KubernetesLangfuseRedisSentinel{
				MasterName: "mymaster", Nodes: []string{"sentinel-0:26379"}, TlsEnabled: true,
			}}
			input.Spec.Redis.Tls = &KubernetesLangfuseRedisTls{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects an external_url with a path", func() {
			input.Spec.ExternalUrl = strPtr("https://example.com/langfuse")
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.external_url"))
		})

		ginkgo.It("rejects single sign-on on the localhost default", func() {
			input.Spec.Auth = &KubernetesLangfuseAuth{Sso: &KubernetesLangfuseSsoProviders{
				Google: &KubernetesLangfuseGoogleSso{Client: oauthClient()},
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.auth.sso_requires_external_url"))
		})

		ginkgo.It("rejects mixing gcs with an S3 target", func() {
			input.Spec.BlobStorage.MediaUpload = &KubernetesLangfuseMediaUpload{Target: gcsTarget()}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.blob_storage.one_family"))
		})

		ginkgo.It("rejects the same variable in extra_env and extra_env_from_secret", func() {
			input.Spec.ExtraEnv = map[string]string{"FOO": "bar"}
			input.Spec.ExtraEnvFromSecret = map[string]*kubernetes.KubernetesSecretKey{"FOO": {Name: "s", Key: "k"}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.extra_env.unique_names"))
		})

		ginkgo.It("rejects an Assistant without a model", func() {
			input.Spec.AiFeatures = &KubernetesLangfuseAiFeatures{Assistant: &KubernetesLangfuseAssistant{Enabled: true}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.ai_features.assistant_requires_model"))
		})

		ginkgo.It("rejects a provider without a model", func() {
			input.Spec.AiFeatures = &KubernetesLangfuseAiFeatures{Provider: KubernetesLangfuseAiFeatures_bedrock}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.ai_features.model_with_provider"))
		})

		ginkgo.It("rejects an api_key for bedrock", func() {
			input.Spec.AiFeatures = &KubernetesLangfuseAiFeatures{
				Provider: KubernetesLangfuseAiFeatures_bedrock, Model: "eu.anthropic.claude-opus-5", ApiKey: "$secret/key",
			}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.ai_features.api_key_by_provider"))
		})

		ginkgo.It("rejects openai without an api_key", func() {
			input.Spec.AiFeatures = &KubernetesLangfuseAiFeatures{Provider: KubernetesLangfuseAiFeatures_openai, Model: "gpt-5"}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.ai_features.api_key_by_provider"))
		})

		ginkgo.It("rejects a ClickHouse host with a scheme", func() {
			input.Spec.Clickhouse.Host = literal("https://clickhouse.example.com")
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.clickhouse.host.no_scheme"))
		})

		ginkgo.It("rejects a malformed read-only URL", func() {
			input.Spec.Clickhouse.ReadOnlyUrl = "reader.example.com"
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.clickhouse.read_only_url"))
		})

		ginkgo.It("rejects a Redis socket timeout under ten seconds", func() {
			input.Spec.Redis.SocketTimeoutMs = int32Ptr(5000)
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.redis.socket_timeout_ms"))
		})

		ginkgo.It("rejects Sentinel TLS without Redis TLS", func() {
			input.Spec.Redis.Topology = &KubernetesLangfuseRedis_Sentinel{Sentinel: &KubernetesLangfuseRedisSentinel{
				MasterName: "mymaster", Nodes: []string{"sentinel-0:26379"}, TlsEnabled: true,
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.redis.sentinel_tls_requires_tls"))
		})

		ginkgo.It("rejects a Redis Cluster node without a port", func() {
			input.Spec.Redis.Topology = &KubernetesLangfuseRedis_Cluster{Cluster: &KubernetesLangfuseRedisCluster{
				Nodes: []string{"redis-0"},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects half a mutual TLS pair", func() {
			input.Spec.Redis.Tls = &KubernetesLangfuseRedisTls{
				Enabled:           true,
				ClientCertificate: &KubernetesLangfuseSecretFile{SecretName: literal("redis-client")},
			}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.redis.tls.mutual_pair"))
		})

		ginkgo.It("rejects a prefix without a trailing slash", func() {
			input.Spec.BlobStorage.EventUpload = &KubernetesLangfuseEventUpload{Prefix: strPtr("events")}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.blob_storage.event_upload.prefix"))
		})

		ginkgo.It("rejects a KMS key on AES256 encryption", func() {
			input.Spec.BlobStorage.EventUpload = &KubernetesLangfuseEventUpload{Sse: &KubernetesLangfuseServerSideEncryption{
				Algorithm: KubernetesLangfuseServerSideEncryption_aes256,
				KmsKeyId:  literal("arn:aws:kms:us-east-1:123456789012:key/abc"),
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.blob_storage.sse.kms_key_requires_kms"))
		})

		ginkgo.It("accepts managed-secret references for the instance secrets and the project key pair", func() {
			input.Spec.Secrets = &KubernetesLangfuseInstanceSecrets{
				Salt:           "$secret/langfuse-salt",
				EncryptionKey:  "$secret/langfuse-encryption-key",
				NextauthSecret: "$secret/langfuse-nextauth-secret",
			}
			input.Spec.Init = &KubernetesLangfuseInit{Project: &KubernetesLangfuseInitProject{
				PublicKey: "pk-lf-langfuse", SecretKey: "$secret/langfuse-project-secret-key",
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("rejects half a project key pair", func() {
			input.Spec.Init = &KubernetesLangfuseInit{Project: &KubernetesLangfuseInitProject{PublicKey: "pk-lf-123"}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.init.project.key_pair"))
		})

		ginkgo.It("rejects a public key without its prefix", func() {
			input.Spec.Init = &KubernetesLangfuseInit{Project: &KubernetesLangfuseInitProject{
				PublicKey: "123", SecretKey: "sk-lf-123",
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.init.project.public_key"))
		})

		ginkgo.It("rejects a retention shorter than three days", func() {
			input.Spec.Init = &KubernetesLangfuseInit{Project: &KubernetesLangfuseInitProject{RetentionDays: int32Ptr(1)}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an admin without an email", func() {
			input.Spec.Init = &KubernetesLangfuseInit{Admin: &KubernetesLangfuseInitAdmin{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects NONE as a project role", func() {
			input.Spec.Auth = &KubernetesLangfuseAuth{NewUserAccess: &KubernetesLangfuseNewUserAccess{
				ProjectRole: KubernetesLangfuseNewUserAccess_NONE,
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an ID token algorithm for GitHub", func() {
			input.Spec.ExternalUrl = strPtr("https://langfuse.example.com")
			client := oauthClient()
			client.IdTokenSignedResponseAlg = KubernetesLangfuseOAuthClient_RS256
			input.Spec.Auth = &KubernetesLangfuseAuth{Sso: &KubernetesLangfuseSsoProviders{
				Github: &KubernetesLangfuseGitHubSso{Client: client},
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.auth.sso.github.no_id_token_alg"))
		})

		ginkgo.It("rejects GitHub Enterprise without a base URL", func() {
			input.Spec.ExternalUrl = strPtr("https://langfuse.example.com")
			input.Spec.Auth = &KubernetesLangfuseAuth{Sso: &KubernetesLangfuseSsoProviders{
				GithubEnterprise: &KubernetesLangfuseGitHubEnterpriseSso{Client: oauthClient()},
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.auth.sso.github_enterprise.base_url"))
		})

		ginkgo.It("rejects OIDC options on WorkOS", func() {
			input.Spec.ExternalUrl = strPtr("https://langfuse.example.com")
			client := oauthClient()
			client.Checks = []KubernetesLangfuseOAuthClient_Check{KubernetesLangfuseOAuthClient_pkce}
			input.Spec.Auth = &KubernetesLangfuseAuth{Sso: &KubernetesLangfuseSsoProviders{
				Workos: &KubernetesLangfuseWorkOsSso{Client: client},
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.auth.sso.workos.client_options"))
		})

		ginkgo.It("rejects an Authentik issuer outside its URL shape", func() {
			input.Spec.ExternalUrl = strPtr("https://langfuse.example.com")
			input.Spec.Auth = &KubernetesLangfuseAuth{Sso: &KubernetesLangfuseSsoProviders{
				Authentik: &KubernetesLangfuseAuthentikSso{Client: oauthClient(), Issuer: "https://auth.example.com"},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a custom provider without a name", func() {
			input.Spec.ExternalUrl = strPtr("https://langfuse.example.com")
			input.Spec.Auth = &KubernetesLangfuseAuth{Sso: &KubernetesLangfuseSsoProviders{
				Custom: &KubernetesLangfuseCustomSso{Client: oauthClient(), Issuer: "https://idp.example.com"},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects KEDA beside an HPA", func() {
			input.Spec.Worker = &KubernetesLangfuseWorker{Autoscaling: &KubernetesLangfuseAutoscaling{
				Hpa: &KubernetesLangfuseHpa{}, Keda: &KubernetesLangfuseKeda{},
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.autoscaling.keda_exclusive"))
		})

		ginkgo.It("rejects inverted HPA bounds", func() {
			input.Spec.Web = &KubernetesLangfuseWeb{Autoscaling: &KubernetesLangfuseAutoscaling{
				Hpa: &KubernetesLangfuseHpa{MinReplicas: int32Ptr(5), MaxReplicas: int32Ptr(2)},
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.autoscaling.hpa.replica_bounds"))
		})

		ginkgo.It("rejects a PodDisruptionBudget with both bounds", func() {
			input.Spec.Web = &KubernetesLangfuseWeb{Pdb: &KubernetesLangfusePdb{MinAvailable: "1", MaxUnavailable: "1"}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.pdb.one_bound"))
		})

		ginkgo.It("rejects surge settings on a Recreate strategy", func() {
			input.Spec.Worker = &KubernetesLangfuseWorker{Workload: &KubernetesLangfuseWorkloadShape{
				Strategy: &KubernetesLangfuseRolloutStrategy{Type: "Recreate", MaxSurge: "1"},
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.workload.strategy.recreate_bounds"))
		})

		ginkgo.It("rejects annotations on an existing ServiceAccount", func() {
			input.Spec.ServiceAccount = &KubernetesLangfuseServiceAccount{
				ExistingName: "langfuse", Annotations: map[string]string{"a": "b"},
			}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.service_account.existing_owns_annotations"))
		})

		ginkgo.It("rejects an unknown image pull policy", func() {
			input.Spec.Web = &KubernetesLangfuseWeb{Image: &KubernetesLangfuseImage{PullPolicy: "Sometimes"}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.image.pull_policy"))
		})

		ginkgo.It("rejects an allowlisted IP that is not an address", func() {
			input.Spec.Network = &KubernetesLangfuseNetwork{Allowlists: &KubernetesLangfuseOutboundAllowlists{
				Webhooks: &KubernetesLangfuseAllowlist{Ips: []string{"not-an-ip"}},
			}}
			gomega.Expect(violations(protovalidate.Validate(input))).To(gomega.ContainElement("spec.network.allowlist.ips"))
		})

		ginkgo.It("rejects a missing blob storage target", func() {
			input.Spec.BlobStorage = &KubernetesLangfuseBlobStorage{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an S3-compatible target without credentials", func() {
			input.Spec.BlobStorage.Target = &KubernetesLangfuseBlobTarget{Backend: &KubernetesLangfuseBlobTarget_S3Compatible{
				S3Compatible: &KubernetesLangfuseS3CompatibleTarget{Endpoint: literal("http://minio:9000"), Bucket: "langfuse"},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing Redis topology", func() {
			input.Spec.Redis.Topology = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
