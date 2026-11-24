package kubernetesstigmerv1alpha1

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesStigmerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesStigmerSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

// minimalStigmer is the default path: the operator in its own namespace,
// Stigmer in another, a model key, everything else bundled.
func minimalStigmer() *KubernetesStigmer {
	return &KubernetesStigmer{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesStigmer",
		Metadata:   &shared.CatalogObjectMetadata{Name: "stigmer"},
		Spec: &KubernetesStigmerSpec{
			Namespace: literalRef("stigmer"),
			Version:   "0.3.2",
			Operator: &KubernetesStigmerOperator{
				Namespace:       literalRef("stigmer-system"),
				CreateNamespace: true,
			},
			Owner: &KubernetesStigmerOwner{Email: "owner@acme.example"},
			Runner: &KubernetesStigmerRunner{
				Model: &KubernetesStigmerModel{Source: &KubernetesStigmerModel_Keys{
					Keys: &KubernetesStigmerModelKeys{AnthropicApiKey: "sk-ant-example"},
				}},
			},
		},
	}
}

func objectStore() *KubernetesStigmerObjectStore {
	return &KubernetesStigmerObjectStore{S3: &KubernetesStigmerS3{
		Bucket:   literalRef("acme-stigmer"),
		Endpoint: literalRef("http://seaweedfs-s3.storage.svc.cluster.local:8333"),
		Credentials: &KubernetesStigmerS3Credentials{Source: &KubernetesStigmerS3Credentials_AccessKey{
			AccessKey: &KubernetesStigmerS3AccessKey{AccessKeyId: "stigmer", SecretAccessKey: "s3-secret"},
		}},
	}}
}

func externalDatabase() *KubernetesStigmerDatabase {
	return &KubernetesStigmerDatabase{Postgresql: &KubernetesStigmerDatabase_External{
		External: &KubernetesStigmerExternalPostgres{
			Host: literalRef("db-rw.data.svc.cluster.local"),
			User: "stigmer",
			Password: &KubernetesStigmerPassword{Source: &KubernetesStigmerPassword_Value{
				Value: "db-password",
			}},
		},
	}}
}

func externalTemporal() *KubernetesStigmerTemporal {
	return &KubernetesStigmerTemporal{Mode: &KubernetesStigmerTemporal_External{
		External: &KubernetesStigmerExternalTemporal{HostPort: literalRef("temporal-frontend.temporal.svc.cluster.local:7233")},
	}}
}

func ingressWithIssuer() *KubernetesStigmerIngress {
	return &KubernetesStigmerIngress{
		Hostname: "stigmer.acme.example",
		Route: &KubernetesStigmerIngress_Tls{Tls: &KubernetesStigmerIngressTls{
			Certificate: &KubernetesStigmerIngressTls_Issuer{Issuer: &KubernetesStigmerIssuerRef{Name: literalRef("letsencrypt")}},
		}},
		ClassName: "nginx",
	}
}

func oidc() *KubernetesStigmerIdentity {
	return &KubernetesStigmerIdentity{Provider: &KubernetesStigmerIdentity_Oidc{Oidc: &KubernetesStigmerOidc{
		Issuer:   "https://auth.acme.example/realms/stigmer",
		Audience: "stigmer",
	}}}
}

func enterprise() *KubernetesStigmerEnterprise {
	return &KubernetesStigmerEnterprise{License: &KubernetesStigmerSecretValue{
		Source: &KubernetesStigmerSecretValue_Value{Value: "license-key"},
	}}
}

func expectValid(input *KubernetesStigmer) {
	gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
}

// expectRule asserts the declaration is refused by the named rule: a rule
// id the spec authors, or a field path a standard rule names.
func expectRule(input *KubernetesStigmer, rule string) {
	err := protovalidate.Validate(input)
	gomega.Expect(err).To(gomega.HaveOccurred())
	var validationErr *protovalidate.ValidationError
	gomega.Expect(errors.As(err, &validationErr)).To(gomega.BeTrue(), err.Error())
	var seen []string
	for _, violation := range validationErr.Violations {
		id := violation.Proto.GetRuleId()
		path := protovalidate.FieldPathString(violation.Proto.GetField())
		if id == rule || strings.Contains(path, rule) {
			return
		}
		seen = append(seen, id+" at "+path)
	}
	ginkgo.Fail(fmt.Sprintf("expected rule %q, got %v", rule, seen))
}

var _ = ginkgo.Describe("KubernetesStigmerSpec", func() {

	ginkgo.Describe("valid declarations", func() {
		ginkgo.It("accepts the minimal install", func() {
			expectValid(minimalStigmer())
		})

		ginkgo.It("accepts a pre-created operator namespace", func() {
			in := minimalStigmer()
			in.Spec.Operator.CreateNamespace = false
			expectValid(in)
		})

		ginkgo.It("accepts the operator in the Stigmer's namespace when the kind does not create it", func() {
			in := minimalStigmer()
			in.Spec.Operator.Namespace = literalRef("stigmer")
			in.Spec.Operator.CreateNamespace = false
			expectValid(in)
		})

		ginkgo.It("accepts a model key kept in an existing Secret", func() {
			in := minimalStigmer()
			in.Spec.Runner.Model = &KubernetesStigmerModel{Source: &KubernetesStigmerModel_ExistingSecretName{
				ExistingSecretName: literalRef("model-keys"),
			}}
			expectValid(in)
		})

		ginkgo.It("accepts a gateway address beside a key", func() {
			in := minimalStigmer()
			in.Spec.Runner.Model.GetKeys().AnthropicBaseUrl = "http://models.ai.svc.cluster.local:8080"
			expectValid(in)
		})

		ginkgo.It("accepts your own Postgres with your own Temporal, no front door", func() {
			in := minimalStigmer()
			in.Spec.Database = externalDatabase()
			in.Spec.Temporal = externalTemporal()
			expectValid(in)
		})

		ginkgo.It("accepts a password read from a KubernetesPostgres Secret", func() {
			in := minimalStigmer()
			in.Spec.Database = externalDatabase()
			in.Spec.Database.GetExternal().Password = &KubernetesStigmerPassword{Source: &KubernetesStigmerPassword_ExistingSecret{
				ExistingSecret: &KubernetesStigmerPostgresPasswordSecret{Name: literalRef("stigmer-db-app")},
			}}
			in.Spec.Temporal = externalTemporal()
			expectValid(in)
		})

		ginkgo.It("accepts a bucket with backups and a restore", func() {
			in := minimalStigmer()
			in.Spec.ObjectStore = objectStore()
			in.Spec.Database = &KubernetesStigmerDatabase{Postgresql: &KubernetesStigmerDatabase_Bundled{
				Bundled: &KubernetesStigmerBundledPostgres{
					Instances:   int32Ptr(3),
					Backup:      &KubernetesStigmerPostgresBackup{Schedule: "0 0 3 * * *", Retention: "14d"},
					RecoverFrom: &KubernetesStigmerRecoverFrom{ServerName: "stigmer-20261120", TargetTime: "2026-11-20T08:00:00Z"},
				},
			}}
			expectValid(in)
		})

		ginkgo.It("accepts an exposed install with the bundled sign-in and sign-up", func() {
			in := minimalStigmer()
			in.Spec.Ingress = ingressWithIssuer()
			in.Spec.Identity = &KubernetesStigmerIdentity{Provider: &KubernetesStigmerIdentity_Bundled{
				Bundled: &KubernetesStigmerBundledIdentity{
					SignUpEmailDomains: []string{"acme.example"},
					Smtp:               &KubernetesStigmerSmtp{Host: "smtp.acme.example", From: "stigmer@acme.example"},
				},
			}}
			expectValid(in)
		})

		ginkgo.It("accepts a front door on a Gateway", func() {
			in := minimalStigmer()
			in.Spec.Ingress = &KubernetesStigmerIngress{
				Hostname: "stigmer.acme.example",
				Route:    &KubernetesStigmerIngress_Gateway{Gateway: &KubernetesStigmerGatewayRef{Name: literalRef("edge")}},
			}
			expectValid(in)
		})

		ginkgo.It("accepts Enterprise with OIDC and per-session sandboxes", func() {
			in := minimalStigmer()
			in.Spec.Identity = oidc()
			in.Spec.Enterprise = enterprise()
			in.Spec.ObjectStore = objectStore()
			in.Spec.Sandboxes = &KubernetesStigmerSandboxes{Driver: "agentSandbox",
				AgentSandbox: &KubernetesStigmerAgentSandbox{EgressAllowCidrs: []string{"10.20.0.0/16"}}}
			expectValid(in)
		})

		ginkgo.It("accepts substrate sandboxes with DNS egress lanes", func() {
			in := minimalStigmer()
			in.Spec.ObjectStore = objectStore()
			in.Spec.Sandboxes = &KubernetesStigmerSandboxes{Driver: "substrate", Substrate: &KubernetesStigmerSubstrate{
				Workers:    int32Ptr(2),
				EgressHttp: []string{"models.ai.svc.cluster.local:8080"},
			}}
			expectValid(in)
		})

		ginkgo.It("accepts extra settings that the operator does not own", func() {
			in := minimalStigmer()
			in.Spec.Server = &KubernetesStigmerServer{ExtraEnv: &kubernetes.ContainerEnv{
				Variables: []*kubernetes.EnvVar{{Name: "LOG_FORMAT", Source: &kubernetes.EnvVar_Value{Value: "json"}}},
				Secrets:   []*kubernetes.SecretEnvVar{{Name: "MCP_TOKEN", Source: &kubernetes.SecretEnvVar_Value{Value: "token"}}},
			}}
			expectValid(in)
		})

		ginkgo.It("accepts an upgrade released without a backup when it names the version", func() {
			in := minimalStigmer()
			in.Spec.UpgradeWithoutBackupTo = "0.3.2"
			expectValid(in)
		})
	})

	ginkgo.Describe("refusals", func() {
		ginkgo.It("refuses a version that is not exact semver", func() {
			in := minimalStigmer()
			in.Spec.Version = "v0.3.2"
			expectRule(in, "version.semver")
		})

		ginkgo.It("refuses a version before the first chart", func() {
			in := minimalStigmer()
			in.Spec.Version = "0.3.0"
			expectRule(in, "version.minimum")
		})

		ginkgo.It("refuses a missing version", func() {
			in := minimalStigmer()
			in.Spec.Version = ""
			expectRule(in, "required")
		})

		ginkgo.It("refuses creating the operator namespace in the Stigmer's own namespace", func() {
			in := minimalStigmer()
			in.Spec.Operator.Namespace = literalRef("stigmer")
			expectRule(in, "spec.operator_namespace.not_created_in_stigmer_namespace")
		})

		ginkgo.It("refuses a stale upgrade-without-backup version", func() {
			in := minimalStigmer()
			in.Spec.UpgradeWithoutBackupTo = "0.3.1"
			expectRule(in, "spec.upgrade_without_backup_to.matches_version")
		})

		ginkgo.It("refuses your own Postgres with the bundled Temporal", func() {
			in := minimalStigmer()
			in.Spec.Database = externalDatabase()
			expectRule(in, "spec.database.external_needs_temporal_external")
		})

		ginkgo.It("refuses Enterprise without OIDC", func() {
			in := minimalStigmer()
			in.Spec.Enterprise = enterprise()
			in.Spec.ObjectStore = objectStore()
			in.Spec.Sandboxes = &KubernetesStigmerSandboxes{Driver: "substrate"}
			expectRule(in, "spec.enterprise.needs_oidc")
		})

		ginkgo.It("refuses Enterprise on a plaintext external Temporal", func() {
			in := minimalStigmer()
			in.Spec.Enterprise = enterprise()
			in.Spec.Identity = oidc()
			in.Spec.Database = externalDatabase()
			in.Spec.Temporal = externalTemporal()
			in.Spec.ObjectStore = objectStore()
			in.Spec.Sandboxes = &KubernetesStigmerSandboxes{Driver: "substrate"}
			expectRule(in, "spec.enterprise.temporal_must_authenticate")
		})

		ginkgo.It("refuses Enterprise on the shared runner", func() {
			in := minimalStigmer()
			in.Spec.Enterprise = enterprise()
			in.Spec.Identity = oidc()
			expectRule(in, "spec.enterprise.needs_per_session_sandboxes")
		})

		ginkgo.It("refuses Enterprise without a license", func() {
			in := minimalStigmer()
			in.Spec.Enterprise = &KubernetesStigmerEnterprise{}
			expectRule(in, "license")
		})

		ginkgo.It("refuses backups without a bucket", func() {
			in := minimalStigmer()
			in.Spec.Database = &KubernetesStigmerDatabase{Postgresql: &KubernetesStigmerDatabase_Bundled{
				Bundled: &KubernetesStigmerBundledPostgres{Backup: &KubernetesStigmerPostgresBackup{}},
			}}
			expectRule(in, "spec.database.backup_needs_object_store")
		})

		ginkgo.It("refuses an artifacts volume beside a bucket", func() {
			in := minimalStigmer()
			in.Spec.ObjectStore = objectStore()
			in.Spec.Server = &KubernetesStigmerServer{Artifacts: &KubernetesStigmerVolume{Size: "20Gi"}}
			expectRule(in, "spec.server.artifacts_without_object_store")
		})

		ginkgo.It("refuses the bundled sign-in without a front door", func() {
			in := minimalStigmer()
			in.Spec.Identity = &KubernetesStigmerIdentity{Provider: &KubernetesStigmerIdentity_Bundled{Bundled: &KubernetesStigmerBundledIdentity{}}}
			expectRule(in, "spec.identity.bundled_needs_ingress")
		})

		ginkgo.It("refuses an exposed install without an owner", func() {
			in := minimalStigmer()
			in.Spec.Owner = nil
			in.Spec.Ingress = ingressWithIssuer()
			expectRule(in, "spec.owner.required_with_bundled_sign_in")
		})

		ginkgo.It("refuses your own Postgres behind the bundled sign-in", func() {
			in := minimalStigmer()
			in.Spec.Ingress = ingressWithIssuer()
			in.Spec.Database = externalDatabase()
			in.Spec.Temporal = externalTemporal()
			expectRule(in, "spec.database.external_without_bundled_sign_in")
		})

		ginkgo.It("refuses unauthenticated exposure beside a sign-in", func() {
			in := minimalStigmer()
			in.Spec.Ingress = ingressWithIssuer()
			in.Spec.Server = &KubernetesStigmerServer{AllowUnauthenticatedExposure: true}
			in.Spec.Identity = oidc()
			expectRule(in, "spec.server.unauthenticated_exposure_without_identity")
		})

		ginkgo.It("refuses per-session sandboxes without a bucket", func() {
			in := minimalStigmer()
			in.Spec.Sandboxes = &KubernetesStigmerSandboxes{Driver: "agentSandbox"}
			expectRule(in, "spec.sandboxes.per_session_needs_object_store")
		})

		ginkgo.It("refuses agent-sandbox tuning under another driver", func() {
			in := minimalStigmer()
			in.Spec.Sandboxes = &KubernetesStigmerSandboxes{AgentSandbox: &KubernetesStigmerAgentSandbox{}}
			expectRule(in, "sandboxes.agent_sandbox_needs_driver")
		})

		ginkgo.It("refuses substrate tuning under another driver", func() {
			in := minimalStigmer()
			in.Spec.ObjectStore = objectStore()
			in.Spec.Sandboxes = &KubernetesStigmerSandboxes{Driver: "agentSandbox", Substrate: &KubernetesStigmerSubstrate{}}
			expectRule(in, "sandboxes.substrate_needs_driver")
		})

		ginkgo.It("refuses an IP address as a substrate egress host", func() {
			in := minimalStigmer()
			in.Spec.ObjectStore = objectStore()
			in.Spec.Sandboxes = &KubernetesStigmerSandboxes{Driver: "substrate", Substrate: &KubernetesStigmerSubstrate{
				EgressHttp: []string{"10.0.0.7:8080"},
			}}
			expectRule(in, "substrate.egress_http.format")
		})

		ginkgo.It("refuses a gateway address with no key", func() {
			in := minimalStigmer()
			in.Spec.Runner.Model = &KubernetesStigmerModel{Source: &KubernetesStigmerModel_Keys{
				Keys: &KubernetesStigmerModelKeys{AnthropicBaseUrl: "http://models.ai.svc.cluster.local:8080"},
			}}
			expectRule(in, "model.keys.at_least_one")
		})

		ginkgo.It("refuses sign-up without mail", func() {
			in := minimalStigmer()
			in.Spec.Ingress = ingressWithIssuer()
			in.Spec.Identity = &KubernetesStigmerIdentity{Provider: &KubernetesStigmerIdentity_Bundled{
				Bundled: &KubernetesStigmerBundledIdentity{SignUpEmailDomains: []string{"acme.example"}},
			}}
			expectRule(in, "identity.bundled.sign_up_needs_smtp")
		})

		ginkgo.It("refuses a front door with neither a certificate nor a Gateway", func() {
			in := minimalStigmer()
			in.Spec.Ingress = &KubernetesStigmerIngress{Hostname: "stigmer.acme.example"}
			expectRule(in, "required")
		})

		ginkgo.It("refuses a single-label hostname", func() {
			in := minimalStigmer()
			in.Spec.Ingress = ingressWithIssuer()
			in.Spec.Ingress.Hostname = "stigmer"
			expectRule(in, "hostname")
		})

		ginkgo.It("refuses a plain-http OIDC issuer off loopback", func() {
			in := minimalStigmer()
			in.Spec.Identity = oidc()
			in.Spec.Identity.GetOidc().Issuer = "http://auth.acme.example"
			expectRule(in, "oidc.issuer.format")
		})

		ginkgo.It("refuses a client certificate without its key", func() {
			in := minimalStigmer()
			in.Spec.Database = externalDatabase()
			in.Spec.Temporal = externalTemporal()
			in.Spec.Temporal.GetExternal().Tls = &KubernetesStigmerTemporalTls{Certificates: &KubernetesStigmerTemporalTls_Pem{
				Pem: &KubernetesStigmerTemporalTlsPem{ClientCertificate: "-----BEGIN CERTIFICATE-----"},
			}}
			expectRule(in, "temporal.tls.pem.client_pair")
		})

		ginkgo.It("refuses a CIDR that is not one", func() {
			in := minimalStigmer()
			in.Spec.Server = &KubernetesStigmerServer{Egress: &KubernetesStigmerServerEgress{Fenced: true, AllowCidrs: []string{"10.0.0.0"}}}
			expectRule(in, "allow_cidrs")
		})

		ginkgo.It("refuses an unknown sandbox driver", func() {
			in := minimalStigmer()
			in.Spec.Sandboxes = &KubernetesStigmerSandboxes{Driver: "gvisor"}
			expectRule(in, "driver")
		})

		ginkgo.It("refuses a retention without a unit", func() {
			in := minimalStigmer()
			in.Spec.ObjectStore = objectStore()
			in.Spec.Database = &KubernetesStigmerDatabase{Postgresql: &KubernetesStigmerDatabase_Bundled{
				Bundled: &KubernetesStigmerBundledPostgres{Backup: &KubernetesStigmerPostgresBackup{Retention: "30"}},
			}}
			expectRule(in, "backup.retention.format")
		})

		ginkgo.It("refuses a server setting the operator owns", func() {
			in := minimalStigmer()
			in.Spec.Server = &KubernetesStigmerServer{ExtraEnv: &kubernetes.ContainerEnv{
				Variables: []*kubernetes.EnvVar{{Name: "DATABASE_URL", Source: &kubernetes.EnvVar_Value{Value: "postgres://x"}}},
			}}
			expectRule(in, "server.extra_env.not_operator_owned")
		})

		ginkgo.It("refuses a runner secret setting the operator owns", func() {
			in := minimalStigmer()
			in.Spec.Runner.ExtraEnv = &kubernetes.ContainerEnv{
				Secrets: []*kubernetes.SecretEnvVar{{Name: "STIGMER_TOKEN", Source: &kubernetes.SecretEnvVar_Value{Value: "x"}}},
			}
			expectRule(in, "runner.extra_env.not_operator_owned")
		})

		ginkgo.It("refuses env_from in extra settings", func() {
			in := minimalStigmer()
			in.Spec.Runner.ExtraEnv = &kubernetes.ContainerEnv{
				EnvFrom: []*kubernetes.EnvFromSource{{Source: &kubernetes.EnvFromSource_SecretRef{SecretRef: &kubernetes.SecretRef{Name: "all"}}}},
			}
			expectRule(in, "runner.extra_env.no_env_from")
		})

		ginkgo.It("refuses a setting declared both plain and secret", func() {
			in := minimalStigmer()
			in.Spec.Server = &KubernetesStigmerServer{ExtraEnv: &kubernetes.ContainerEnv{
				Variables: []*kubernetes.EnvVar{{Name: "MCP_TOKEN", Source: &kubernetes.EnvVar_Value{Value: "a"}}},
				Secrets:   []*kubernetes.SecretEnvVar{{Name: "MCP_TOKEN", Source: &kubernetes.SecretEnvVar_Value{Value: "b"}}},
			}}
			expectRule(in, "server.extra_env.unique_names")
		})

		ginkgo.It("accepts operator sizes by cpu, memory and ephemeral-storage", func() {
			in := minimalStigmer()
			in.Spec.Operator.Resources = &kubernetes.KubernetesCoreV1ApiResourceRequirements{
				Requests: map[string]string{"cpu": "100m", "memory": "256Mi", "ephemeral-storage": "1Gi"},
				Limits:   map[string]string{"memory": "1Gi"},
			}
			expectValid(in)
		})

		ginkgo.It("refuses an operator resource name the chart does not accept", func() {
			in := minimalStigmer()
			in.Spec.Operator.Resources = &kubernetes.KubernetesCoreV1ApiResourceRequirements{Requests: map[string]string{"nvidia.com/gpu": "1"}}
			expectRule(in, "operator.resources.chart_resources")
		})

		ginkgo.It("refuses operator resource claims, which the chart does not take", func() {
			in := minimalStigmer()
			in.Spec.Operator.Resources = &kubernetes.KubernetesCoreV1ApiResourceRequirements{
				Claims: []*kubernetes.KubernetesCoreV1ApiResourceClaim{{Name: "gpu"}},
			}
			expectRule(in, "operator.resources.chart_resources")
		})

		ginkgo.It("accepts runner sizes on the shared runner", func() {
			in := minimalStigmer()
			in.Spec.Runner.Resources = &kubernetes.KubernetesCoreV1ApiResourceRequirements{Requests: map[string]string{"cpu": "1", "memory": "2Gi"}}
			in.Spec.Runner.Storage = &KubernetesStigmerVolume{Size: "20Gi"}
			expectValid(in)
		})

		ginkgo.It("refuses runner sizes under a per-session sandbox driver, which runs no shared runner", func() {
			for _, sized := range []func(*KubernetesStigmerRunner){
				func(r *KubernetesStigmerRunner) {
					r.Resources = &kubernetes.KubernetesCoreV1ApiResourceRequirements{Requests: map[string]string{"cpu": "1"}}
				},
				func(r *KubernetesStigmerRunner) { r.Storage = &KubernetesStigmerVolume{Size: "20Gi"} },
			} {
				in := minimalStigmer()
				in.Spec.ObjectStore = objectStore()
				in.Spec.Sandboxes = &KubernetesStigmerSandboxes{Driver: "agentSandbox"}
				sized(in.Spec.Runner)
				expectRule(in, "spec.runner.sizes_need_the_shared_runner")
			}
		})

		ginkgo.It("refuses an owner email that is not one", func() {
			in := minimalStigmer()
			in.Spec.Owner.Email = "owner"
			expectRule(in, "owner.email.format")
		})
	})
})
