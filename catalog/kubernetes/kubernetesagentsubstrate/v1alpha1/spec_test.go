package kubernetesagentsubstratev1alpha1

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// The rules pinned here are the spec's whole validation surface: upstream's
// own startup and installer checks (a setting the installer or a component
// would refuse at start is refused before deploy), the choices that only
// work keyless on GKE, and the settings that would otherwise be silently
// ignored. Every rejection names the rule it expects, so a case cannot pass
// because some other rule happened to fire.

func TestKubernetesAgentSubstrate(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAgentSubstrate Suite")
}

func int32Ptr(i int32) *int32       { return &i }
func int64Ptr(i int64) *int64       { return &i }
func boolPtr(b bool) *bool          { return &b }
func stringPtr(s string) *string    { return &s }
func float64Ptr(f float64) *float64 { return &f }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: name}}}
}

func secretKey(name, key string) *KubernetesAgentSubstrateSecretKeyRef {
	return &KubernetesAgentSubstrateSecretKeyRef{Name: literal(name), Key: key}
}

func generic() *KubernetesAgentSubstratePlatform {
	return &KubernetesAgentSubstratePlatform{Kind: &KubernetesAgentSubstratePlatform_Generic{Generic: &KubernetesAgentSubstrateGeneric{
		ServiceAccountIssuer: "https://kubernetes.default.svc.cluster.local",
	}}}
}

func gke(gkePlatform *KubernetesAgentSubstrateGke) *KubernetesAgentSubstratePlatform {
	return &KubernetesAgentSubstratePlatform{Kind: &KubernetesAgentSubstratePlatform_Gke{Gke: gkePlatform}}
}

func bundled() *KubernetesAgentSubstrateDatabase {
	return &KubernetesAgentSubstrateDatabase{Kind: &KubernetesAgentSubstrateDatabase_Bundled{Bundled: &KubernetesAgentSubstrateBundledDatabase{}}}
}

func cloudSQL(c *KubernetesAgentSubstrateCloudSqlDatabase) *KubernetesAgentSubstrateDatabase {
	return &KubernetesAgentSubstrateDatabase{Kind: &KubernetesAgentSubstrateDatabase_CloudSql{CloudSql: c}}
}

func s3() *KubernetesAgentSubstrateStorage {
	return &KubernetesAgentSubstrateStorage{Kind: &KubernetesAgentSubstrateStorage_S3{S3: &KubernetesAgentSubstrateS3Storage{
		Endpoint:    literal("http://main-s3.seaweedfs.svc.cluster.local:8333"),
		Bucket:      "snapshots",
		Credentials: &KubernetesAgentSubstrateS3Credentials{SecretName: literal("substrate-s3")},
	}}}
}

func gcs() *KubernetesAgentSubstrateStorage {
	return &KubernetesAgentSubstrateStorage{Kind: &KubernetesAgentSubstrateStorage_Gcs{Gcs: &KubernetesAgentSubstrateGcsStorage{Bucket: literal("acme-snapshots")}}}
}

// envoyRouter is the default router with the one client it requires.
func envoyRouter() *KubernetesAgentSubstrateRouter {
	return &KubernetesAgentSubstrateRouter{AllowedClientSpiffeIds: []string{"spiffe://cluster.local/ns/app/sa/web"}}
}

func disabledProvider() *KubernetesAgentSubstrateEgress {
	return &KubernetesAgentSubstrateEgress{CredentialProvider: &KubernetesAgentSubstrateCredentialProvider{
		Kind: &KubernetesAgentSubstrateCredentialProvider_Disabled{Disabled: &KubernetesAgentSubstrateCredentialProviderDisabled{}},
	}}
}

var _ = ginkgo.Describe("KubernetesAgentSubstrate Validation Tests", func() {
	var input *KubernetesAgentSubstrate

	ginkgo.BeforeEach(func() {
		input = &KubernetesAgentSubstrate{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAgentSubstrate",
			Metadata:   &shared.CatalogObjectMetadata{Name: "agent-substrate"},
			Spec: &KubernetesAgentSubstrateSpec{
				Namespace: literal("ate-system"),
				Platform:  generic(),
				Database:  bundled(),
				Storage:   s3(),
				Router:    envoyRouter(),
				Egress:    disabledProvider(),
			},
		}
	})

	valid := func() {
		ginkgo.GinkgoHelper()
		gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
	}
	refusedBy := func(ruleID string) {
		ginkgo.GinkgoHelper()
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.HaveOccurred())
		var validationErr *protovalidate.ValidationError
		gomega.Expect(errors.As(err, &validationErr)).To(gomega.BeTrue(), "expected a validation error, got %v", err)
		ruleIDs := make([]string, 0, len(validationErr.Violations))
		for _, violation := range validationErr.Violations {
			ruleIDs = append(ruleIDs, violation.Proto.GetRuleId())
		}
		gomega.Expect(ruleIDs).To(gomega.ContainElement(ruleID), "refused, but not by %s: %v", ruleID, err)
	}
	refusedOnField := func(field string) {
		ginkgo.GinkgoHelper()
		err := protovalidate.Validate(input)
		var validationErr *protovalidate.ValidationError
		gomega.Expect(errors.As(err, &validationErr)).To(gomega.BeTrue(), "expected a validation error, got %v", err)
		fields := make([]string, 0, len(validationErr.Violations))
		for _, violation := range validationErr.Violations {
			elements := violation.Proto.GetField().GetElements()
			fields = append(fields, elements[len(elements)-1].GetFieldName())
		}
		gomega.Expect(fields).To(gomega.ContainElement(field), "refused, but not on %s: %v", field, err)
	}

	ginkgo.Describe("the namespace", func() {
		ginkgo.It("refuses any literal but upstream's", func() {
			input.Spec.Namespace = literal("substrate")
			refusedBy("namespace_fixed")
		})

		ginkgo.It("accepts a reference to the KubernetesNamespace that owns it", func() {
			input.Spec.Namespace = ref("ate-system")
			valid()
		})
	})

	ginkgo.Describe("the required choices", func() {
		ginkgo.It("accepts the minimal spec: platform, database, storage, the router's client and credential provider", func() {
			valid()
		})

		for _, field := range []string{"platform", "database", "storage", "egress"} {
			ginkgo.It("requires "+field+", which upstream's installer has no safe default for", func() {
				switch field {
				case "platform":
					input.Spec.Platform = nil
				case "database":
					input.Spec.Database = nil
				case "storage":
					input.Spec.Storage = nil
				case "egress":
					input.Spec.Egress = nil
				}
				refusedOnField(field)
			})
		}

		ginkgo.It("requires a credential provider choice, as upstream's installer does", func() {
			input.Spec.Egress = &KubernetesAgentSubstrateEgress{}
			refusedOnField("credential_provider")
		})

		ginkgo.It("requires the issuer on a generic cluster, because a wrong one refuses every caller", func() {
			input.Spec.Platform = &KubernetesAgentSubstratePlatform{Kind: &KubernetesAgentSubstratePlatform_Generic{Generic: &KubernetesAgentSubstrateGeneric{}}}
			refusedOnField("service_account_issuer")
		})

		ginkgo.It("refuses an issuer upstream's OIDC rules refuse", func() {
			for _, issuer := range []string{"http://kubernetes.default.svc", "https://user@issuer.example", "https://issuer.example?x=1"} {
				input.Spec.Platform = &KubernetesAgentSubstratePlatform{Kind: &KubernetesAgentSubstratePlatform_Generic{Generic: &KubernetesAgentSubstrateGeneric{ServiceAccountIssuer: issuer}}}
				refusedBy("service_account_issuer_https")
			}
		})
	})

	ginkgo.Describe("the version", func() {
		ginkgo.It("accepts a release tag", func() {
			for _, version := range []string{"v0.6.0", "v0.6.12", "v1.0.0"} {
				input.Spec.Version = stringPtr(version)
				valid()
			}
		})

		ginkgo.It("refuses anything but vMAJOR.MINOR.PATCH, the form Planton publishes releases under", func() {
			for _, version := range []string{"0.6.0", "v0.6", "latest", "v0.6.0-rc.1"} {
				input.Spec.Version = stringPtr(version)
				refusedOnField("version")
			}
		})
	})

	ginkgo.Describe("preview gates", func() {
		ginkgo.It("accepts upstream's gates and ALL", func() {
			input.Spec.PreviewGates = []string{"ExternalVolumes", "GoldenEgressPolicy", "PerTemplateSnapshotLocation"}
			valid()
			input.Spec.PreviewGates = []string{"ALL"}
			valid()
		})

		ginkgo.It("refuses a gate upstream does not know, which every component refuses to start with", func() {
			input.Spec.PreviewGates = []string{"Teleportation"}
			refusedOnField("preview_gates")
		})
	})

	ginkgo.Describe("GKE", func() {
		ginkgo.It("accepts the keyless shape: Cloud SQL with IAM login and GCS moved as the transfer service account", func() {
			input.Spec.Platform = gke(&KubernetesAgentSubstrateGke{
				ClusterId:                      literal("projects/acme/locations/us-central1/clusters/agents"),
				ApiServerServiceAccount:        literal("ate-api-server@acme.iam.gserviceaccount.com"),
				NodeAgentServiceAccount:        literal("atelet@acme.iam.gserviceaccount.com"),
				SnapshotTransferServiceAccount: literal("ate-snapshot-transfer@acme.iam.gserviceaccount.com"),
			})
			input.Spec.Database = cloudSQL(&KubernetesAgentSubstrateCloudSqlDatabase{InstanceConnectionName: literal("acme:us-central1:substrate")})
			input.Spec.Storage = gcs()
			valid()
		})

		ginkgo.It("refuses GCS without the transfer service account every snapshot moves as", func() {
			input.Spec.Platform = gke(&KubernetesAgentSubstrateGke{ClusterId: literal("projects/acme/locations/us-central1/clusters/agents")})
			input.Spec.Storage = gcs()
			refusedBy("gcs_needs_snapshot_transfer_service_account")
			input.Spec.Storage = s3()
			valid()
		})

		ginkgo.It("refuses a cluster ID that is not a full resource ID", func() {
			input.Spec.Platform = gke(&KubernetesAgentSubstrateGke{ClusterId: literal("agents")})
			refusedBy("gke_cluster_id_format")
		})

		ginkgo.It("refuses a bundled database's own resources with size10, whose postgresql.conf assumes its own sizing", func() {
			input.Spec.ClusterSize = KubernetesAgentSubstrateSpec_size10
			valid()
			input.Spec.Database.GetBundled().Resources = &kubernetes.ContainerResources{Limits: &kubernetes.CpuMemory{Cpu: "2", Memory: "2Gi"}}
			refusedBy("size10_sizes_the_bundled_database")
			input.Spec.ClusterSize = KubernetesAgentSubstrateSpec_size0
			valid()
		})

		ginkgo.It("refuses GCS off GKE, where it has no keyless identity", func() {
			input.Spec.Storage = gcs()
			refusedBy("gcs_needs_gke")
		})

		ginkgo.It("refuses Cloud SQL off GKE", func() {
			input.Spec.Database = cloudSQL(&KubernetesAgentSubstrateCloudSqlDatabase{InstanceConnectionName: literal("acme:us-central1:substrate")})
			refusedBy("cloud_sql_needs_gke")
		})

		ginkgo.It("refuses Cloud SQL IAM login without the service account it logs in as", func() {
			input.Spec.Platform = gke(&KubernetesAgentSubstrateGke{ClusterId: literal("projects/acme/locations/us-central1/clusters/agents")})
			input.Spec.Database = cloudSQL(&KubernetesAgentSubstrateCloudSqlDatabase{InstanceConnectionName: literal("acme:us-central1:substrate")})
			refusedBy("cloud_sql_iam_needs_api_server_identity")
		})

		ginkgo.It("takes a password login through the proxy only with a connection string, and never both", func() {
			input.Spec.Platform = gke(&KubernetesAgentSubstrateGke{ClusterId: literal("projects/acme/locations/us-central1/clusters/agents")})
			passwordLogin := &KubernetesAgentSubstrateCloudSqlDatabase{InstanceConnectionName: literal("acme:us-central1:substrate"), IamAuthentication: boolPtr(false)}
			input.Spec.Database = cloudSQL(passwordLogin)
			refusedBy("cloud_sql_password_needs_connection_string")
			passwordLogin.ReadWriteConnectionString = secretKey("substrate-db", "read_write")
			valid()
			iamLogin := &KubernetesAgentSubstrateCloudSqlDatabase{
				InstanceConnectionName:    literal("acme:us-central1:substrate"),
				ReadWriteConnectionString: secretKey("substrate-db", "read_write"),
			}
			input.Spec.Platform = gke(&KubernetesAgentSubstrateGke{
				ClusterId:               literal("projects/acme/locations/us-central1/clusters/agents"),
				ApiServerServiceAccount: literal("ate-api-server@acme.iam.gserviceaccount.com"),
			})
			input.Spec.Database = cloudSQL(iamLogin)
			refusedBy("cloud_sql_iam_has_no_connection_string")
		})

		ginkgo.It("refuses a connection name that is not project:region:instance", func() {
			input.Spec.Platform = gke(&KubernetesAgentSubstrateGke{
				ClusterId:               literal("projects/acme/locations/us-central1/clusters/agents"),
				ApiServerServiceAccount: literal("ate-api-server@acme.iam.gserviceaccount.com"),
			})
			input.Spec.Database = cloudSQL(&KubernetesAgentSubstrateCloudSqlDatabase{InstanceConnectionName: literal("substrate")})
			refusedBy("cloud_sql_connection_name_format")
		})
	})

	ginkgo.Describe("the external database", func() {
		ginkgo.It("requires the read/write connection string", func() {
			input.Spec.Database = &KubernetesAgentSubstrateDatabase{Kind: &KubernetesAgentSubstrateDatabase_External{External: &KubernetesAgentSubstrateExternalDatabase{}}}
			refusedOnField("read_write_connection_string")
		})

		ginkgo.It("accepts both pools, custom roles and a server CA", func() {
			input.Spec.Database = &KubernetesAgentSubstrateDatabase{
				Kind: &KubernetesAgentSubstrateDatabase_External{External: &KubernetesAgentSubstrateExternalDatabase{
					ReadWriteConnectionString: secretKey("substrate-db", "read_write"),
					OwnerConnectionString:     secretKey("substrate-db", "owner"),
					Roles:                     &KubernetesAgentSubstrateDatabaseRoles{ReadWrite: stringPtr("rw"), Owner: stringPtr("own"), Schema: stringPtr("ate")},
					ServerCaCertificate:       "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n",
				}},
				PoolMaxConnections: int32Ptr(32),
			}
			valid()
		})

		ginkgo.It("refuses a server CA that is not a PEM certificate", func() {
			input.Spec.Database = &KubernetesAgentSubstrateDatabase{Kind: &KubernetesAgentSubstrateDatabase_External{External: &KubernetesAgentSubstrateExternalDatabase{
				ReadWriteConnectionString: secretKey("substrate-db", "read_write"),
				ServerCaCertificate:       "not a certificate",
			}}}
			refusedBy("server_ca_certificate_pem")
		})
	})

	ginkgo.Describe("storage", func() {
		ginkgo.It("refuses an S3 endpoint that is not a URL", func() {
			input.Spec.Storage.GetS3().Endpoint = literal("main-s3:8333")
			refusedBy("s3_endpoint_url")
		})

		ginkgo.It("requires the S3 bucket and credentials", func() {
			input.Spec.Storage.GetS3().Bucket = ""
			refusedOnField("bucket")
			input.Spec.Storage = s3()
			input.Spec.Storage.GetS3().Credentials = nil
			refusedOnField("credentials")
		})

		ginkgo.It("refuses a bundled volume size that is not a quantity", func() {
			input.Spec.Database.GetBundled().VolumeSize = stringPtr("big")
			refusedBy("bundled_volume_size_quantity")
		})

		ginkgo.It("takes a snapshot prefix that stays inside the bucket", func() {
			input.Spec.Storage.SnapshotPrefix = "substrate/prod"
			valid()
			for _, prefix := range []string{"/substrate", "substrate/", "a//b", "a/../b", "a/./b", `a\b`, "a?b", "a#b", "user@x"} {
				input.Spec.Storage.SnapshotPrefix = prefix
				refusedBy("snapshot_prefix_path")
			}
		})
	})

	ginkgo.Describe("the egress gateway", func() {
		ginkgo.It("accepts the Kubernetes Secrets provider with grants, and an external provider", func() {
			input.Spec.Egress.CredentialProvider = &KubernetesAgentSubstrateCredentialProvider{Kind: &KubernetesAgentSubstrateCredentialProvider_KubernetesSecrets{
				KubernetesSecrets: &KubernetesAgentSubstrateKubernetesSecretsProvider{Grants: []*KubernetesAgentSubstrateNamespaceGrant{{Atespace: "team-a", AllowedNamespaces: []string{"team-a-secrets"}}}},
			}}
			valid()
			input.Spec.Egress.CredentialProvider = &KubernetesAgentSubstrateCredentialProvider{Kind: &KubernetesAgentSubstrateCredentialProvider_External{
				External: &KubernetesAgentSubstrateExternalCredentialProvider{Name: "secretmanager.googleapis.com", Address: "gsm-credential-provider.ate-system.svc:50051"},
			}}
			valid()
		})

		ginkgo.It("refuses an external provider whose address is not host:port", func() {
			input.Spec.Egress.CredentialProvider = &KubernetesAgentSubstrateCredentialProvider{Kind: &KubernetesAgentSubstrateCredentialProvider_External{
				External: &KubernetesAgentSubstrateExternalCredentialProvider{Name: "secretmanager.googleapis.com", Address: "https://gsm"},
			}}
			refusedBy("external_provider_address_host_port")
		})

		ginkgo.It("refuses a grant with no namespaces, which would grant nothing", func() {
			input.Spec.Egress.CredentialProvider = &KubernetesAgentSubstrateCredentialProvider{Kind: &KubernetesAgentSubstrateCredentialProvider_KubernetesSecrets{
				KubernetesSecrets: &KubernetesAgentSubstrateKubernetesSecretsProvider{Grants: []*KubernetesAgentSubstrateNamespaceGrant{{Atespace: "team-a"}}},
			}}
			refusedOnField("allowed_namespaces")
		})

		ginkgo.It("refuses an additional ext_proc filter with agentgateway, which has no Envoy configuration to splice it into", func() {
			input.Spec.Egress.AdditionalExtProcService = &KubernetesAgentSubstrateExtProcService{Namespace: "policy", Service: "filter", Port: 9443}
			valid()
			input.Spec.Router = &KubernetesAgentSubstrateRouter{Dataplane: KubernetesAgentSubstrateRouter_agentgateway}
			refusedBy("extproc_needs_envoy")
		})
	})

	ginkgo.Describe("the router", func() {
		ginkgo.It("requires a client for the Envoy router, which lets only the clients it names reach actors", func() {
			input.Spec.Router = nil
			refusedBy("envoy_router_needs_allowed_clients")
			input.Spec.Router = &KubernetesAgentSubstrateRouter{}
			refusedBy("envoy_router_needs_allowed_clients")
			input.Spec.Router = &KubernetesAgentSubstrateRouter{Dataplane: KubernetesAgentSubstrateRouter_agentgateway}
			valid()
		})

		ginkgo.It("refuses a client identity that is not a SPIFFE ID", func() {
			input.Spec.Router = &KubernetesAgentSubstrateRouter{AllowedClientSpiffeIds: []string{"router-client"}}
			refusedBy("allowed_client_spiffe_id_format")
		})

		ginkgo.It("refuses a zero route timeout, which would cut every request", func() {
			input.Spec.Router.RouteTimeout = stringPtr("0s")
			refusedBy("route_timeout_format")
		})

		ginkgo.It("refuses an ext_proc limit below the parking lot, which the router refuses to start with", func() {
			input.Spec.Router.ExtProcMaxRequests = int32Ptr(128)
			input.Spec.Router.ParkedRequests = &KubernetesAgentSubstrateParkedRequests{Max: int32Ptr(256)}
			refusedBy("router_ext_proc_covers_parked_requests")
			input.Spec.Router.ExtProcMaxRequests = int32Ptr(512)
			valid()
			input.Spec.Router = envoyRouter()
			input.Spec.Router.ExtProcMaxRequests = int32Ptr(512)
			refusedBy("router_ext_proc_covers_parked_requests") // below the unset parking lot's 1024
			input.Spec.Router.ExtProcMaxRequests = int32Ptr(0)
			valid()
		})

		ginkgo.It("refuses a drain shorter than the parking budget, which the router refuses to start with", func() {
			input.Spec.Router.DrainTimeout = "4s"
			refusedBy("router_drain_covers_parked_budget") // below the unset budget's 5s
			input.Spec.Router.DrainTimeout = "20s"
			input.Spec.Router.ParkedRequests = &KubernetesAgentSubstrateParkedRequests{Budget: stringPtr("30s")}
			refusedBy("router_drain_covers_parked_budget")
			input.Spec.Router.ParkedRequests.Budget = stringPtr("8s")
			valid()
		})
	})

	ginkgo.Describe("the node agent's image cache", func() {
		ginkgo.It("refuses a low mark at or above the high mark, which would collect forever", func() {
			input.Spec.NodeAgent = &KubernetesAgentSubstrateNodeAgent{ImageCache: &KubernetesAgentSubstrateImageCache{HighPercent: int32Ptr(70), LowPercent: int32Ptr(75)}}
			refusedBy("image_cache_low_below_high")
		})
	})

	ginkgo.Describe("the node agent's snapshots and caches", func() {
		ginkgo.It("refuses an upload delay past the detach delay, which the node agent refuses to start with", func() {
			input.Spec.NodeAgent = &KubernetesAgentSubstrateNodeAgent{Snapshots: &KubernetesAgentSubstrateNodeSnapshots{UploadDelay: stringPtr("40m")}}
			refusedBy("snapshots_upload_within_detach") // past the unset detach delay's 30m
			input.Spec.NodeAgent.Snapshots.DetachDelay = stringPtr("1h")
			valid()
		})

		ginkgo.It("refuses the zero delays and workers the node agent refuses to start with", func() {
			input.Spec.NodeAgent = &KubernetesAgentSubstrateNodeAgent{Snapshots: &KubernetesAgentSubstrateNodeSnapshots{UploadDelay: stringPtr("0s")}}
			refusedBy("snapshots_upload_delay_format")
			input.Spec.NodeAgent.Snapshots = &KubernetesAgentSubstrateNodeSnapshots{UploadWorkers: int32Ptr(0)}
			refusedOnField("upload_workers")
			input.Spec.NodeAgent.Snapshots = &KubernetesAgentSubstrateNodeSnapshots{UploadTimeout: stringPtr("0s")}
			valid() // a zero upload timeout turns it off
		})

		ginkgo.It("refuses an upload memory budget that is not a quantity", func() {
			input.Spec.NodeAgent = &KubernetesAgentSubstrateNodeAgent{Snapshots: &KubernetesAgentSubstrateNodeSnapshots{UploadMemoryBudget: stringPtr("lots")}}
			refusedBy("snapshots_upload_memory_budget_quantity")
		})

		ginkgo.It("refuses a zero fetch timeout, which the node agent refuses to start with", func() {
			input.Spec.NodeAgent = &KubernetesAgentSubstrateNodeAgent{SandboxAssetCache: &KubernetesAgentSubstrateSandboxAssetCache{FetchTimeout: stringPtr("0s")}}
			refusedBy("sandbox_asset_cache_fetch_timeout_format")
			input.Spec.NodeAgent = &KubernetesAgentSubstrateNodeAgent{SnapshotCache: &KubernetesAgentSubstrateSnapshotCache{FetchTimeout: stringPtr("0s")}}
			refusedBy("snapshot_cache_fetch_timeout_format")
		})
	})

	ginkgo.Describe("telemetry", func() {
		ginkgo.It("refuses a sampler ratio for a sampler that takes none", func() {
			input.Spec.Telemetry = &KubernetesAgentSubstrateTelemetry{TracesSampler: "always_on", TracesSamplerRatio: float64Ptr(0.5)}
			refusedBy("traces_sampler_ratio_needs_ratio_sampler")
		})

		ginkgo.It("refuses an exporter upstream does not know", func() {
			input.Spec.Telemetry = &KubernetesAgentSubstrateTelemetry{MetricsExporters: []string{"zipkin"}}
			refusedOnField("metrics_exporters")
			input.Spec.Telemetry = &KubernetesAgentSubstrateTelemetry{WorkerTracesExporters: []string{"zipkin"}}
			refusedOnField("worker_traces_exporters")
		})
	})

	ginkgo.Describe("the API server", func() {
		ginkgo.It("refuses an actor JWT issuer the identity server refuses: https://<host>, no path", func() {
			for _, issuer := range []string{"http://idp.example", "https://idp.example/actors", "https://idp.example/", "https://user@idp.example"} {
				input.Spec.ApiServer = &KubernetesAgentSubstrateApiServer{ActorJwt: &KubernetesAgentSubstrateActorJwt{Issuer: issuer}}
				refusedBy("actor_jwt_issuer_https")
			}
			input.Spec.ApiServer = &KubernetesAgentSubstrateApiServer{ActorJwt: &KubernetesAgentSubstrateActorJwt{Issuer: "https://idp.example.com:8443"}}
			valid()
		})

		ginkgo.It("refuses a template resync interval under upstream's 1s floor", func() {
			for _, interval := range []string{"500ms", "0.5s", "0s"} {
				input.Spec.ApiServer = &KubernetesAgentSubstrateApiServer{TemplateResyncInterval: stringPtr(interval)}
				refusedBy("template_resync_interval_format")
			}
			input.Spec.ApiServer = &KubernetesAgentSubstrateApiServer{TemplateResyncInterval: stringPtr("1s")}
			valid()
		})
	})

	ginkgo.Describe("the complete shape", func() {
		ginkgo.It("accepts every setting at once", func() {
			resources := &kubernetes.ContainerResources{Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "128Mi"}}
			input.Spec = &KubernetesAgentSubstrateSpec{
				Namespace:       ref("ate-system"),
				CreateNamespace: false,
				Crds:            &KubernetesAgentSubstrateCrds{Install: boolPtr(true), KeepOnUninstall: boolPtr(true)},
				Platform:        generic(),
				Database: &KubernetesAgentSubstrateDatabase{
					Kind: &KubernetesAgentSubstrateDatabase_Bundled{Bundled: &KubernetesAgentSubstrateBundledDatabase{
						StorageClass: "standard", VolumeSize: stringPtr("20Gi"), Resources: resources,
					}},
					PoolMaxConnections: int32Ptr(16),
				},
				Storage: s3(),
				ApiServer: &KubernetesAgentSubstrateApiServer{
					Replicas: int32Ptr(3), Resources: resources, LogLevel: LogLevel_debug,
					Authorization:          &KubernetesAgentSubstrateAuthorization{BootstrapOwners: []string{"spiffe://cluster.local/ns/ops/sa/admin"}},
					ActorJwt:               &KubernetesAgentSubstrateActorJwt{Algorithm: KubernetesAgentSubstrateActorJwt_rs256, Issuer: "https://idp.example.com"},
					DrainDelay:             stringPtr("10s"),
					DrainTimeout:           stringPtr("20s"),
					TemplateResyncInterval: stringPtr("30s"),
				},
				Controller: &KubernetesAgentSubstrateController{Resources: resources, LogLevel: LogLevel_warn, WorkerSyncConcurrency: int32Ptr(4)},
				NodeAgent: &KubernetesAgentSubstrateNodeAgent{
					Resources: resources, LogLevel: LogLevel_error,
					NodeSelector:           map[string]string{"pool": "sandboxes"},
					Tolerations:            []*kubernetes.WorkloadToleration{{Key: "sandbox", Operator: "Exists", Effect: "NoSchedule"}},
					ActorStatsPollInterval: stringPtr("30s"), DrainDelay: stringPtr("1s"), DrainTimeout: stringPtr("4m"),
					ImageCache: &KubernetesAgentSubstrateImageCache{
						GcPeriod: stringPtr("10m"), HighPercent: int32Ptr(90), LowPercent: int32Ptr(70),
						MaxBytes: int64Ptr(100 << 30), MinAge: stringPtr("5m"), DryRun: true,
					},
					AssetStagingCapacityGib: int32Ptr(16),
					Snapshots: &KubernetesAgentSubstrateNodeSnapshots{
						UploadDelay: stringPtr("2m"), DetachDelay: stringPtr("1h"), UploadTimeout: stringPtr("3m"), UploadThroughputMib: int32Ptr(50),
						UploadWorkers: int32Ptr(4), DetachWorkers: int32Ptr(3), ReportWorkers: int32Ptr(2), ReportGiveUpAfter: stringPtr("20m"),
						UploadMemoryBudget: stringPtr("1Gi"),
					},
					SandboxAssetCache: &KubernetesAgentSubstrateSandboxAssetCache{
						MinAge: stringPtr("2h"), FetchTimeout: stringPtr("10m"),
						Eviction: &KubernetesAgentSubstrateFileCacheEviction{GcPeriod: stringPtr("10m"), MaxBytes: int64Ptr(20 << 30), DryRun: true},
					},
					SnapshotCache: &KubernetesAgentSubstrateSnapshotCache{
						MinAge: stringPtr("20m"), FetchTimeout: stringPtr("15m"),
						Eviction: &KubernetesAgentSubstrateFileCacheEviction{GcPeriod: stringPtr("1m"), MaxBytes: int64Ptr(10 << 30)},
					},
				},
				Router: &KubernetesAgentSubstrateRouter{
					Replicas: int32Ptr(2), Resources: resources, LogLevel: LogLevel_debug,
					AllowedClientSpiffeIds: []string{"spiffe://cluster.local/ns/app/sa/frontend"},
					RouteTimeout:           stringPtr("2m"),
					ParkedRequests: &KubernetesAgentSubstrateParkedRequests{
						Budget: stringPtr("8s"), Max: int32Ptr(2048), RetryInterval: stringPtr("50ms"),
						RetryFactor: float64Ptr(1.5), RetryJitter: float64Ptr(0.2),
					},
					ExtProcMaxRequests: int32Ptr(4096),
					DrainDelay:         stringPtr("10s"),
					DrainTimeout:       "25s",
				},
				Egress: &KubernetesAgentSubstrateEgress{
					CredentialProvider: &KubernetesAgentSubstrateCredentialProvider{Kind: &KubernetesAgentSubstrateCredentialProvider_KubernetesSecrets{
						KubernetesSecrets: &KubernetesAgentSubstrateKubernetesSecretsProvider{
							Grants:    []*KubernetesAgentSubstrateNamespaceGrant{{Atespace: "team-a", AllowedNamespaces: []string{"team-a"}}},
							Resources: resources, LogLevel: LogLevel_info,
							KubeApiQps: float64Ptr(20), KubeApiBurst: int32Ptr(40), DrainGrace: stringPtr("3s"),
						},
					}},
					Replicas: int32Ptr(2), Resources: resources, LogLevel: LogLevel_warn,
					PolicyCacheTtl:           stringPtr("5s"),
					InterceptCertificateTtl:  stringPtr("30m"),
					EnvoyConcurrency:         int32Ptr(4),
					AdditionalExtProcService: &KubernetesAgentSubstrateExtProcService{Namespace: "policy", Service: "filter", Port: 9443},
					DrainDelay:               stringPtr("5s"),
				},
				PodCertificates: &KubernetesAgentSubstratePodCertificates{Resources: resources, WorkersPerSigner: int32Ptr(2), KubeApiQps: float64Ptr(50), KubeApiBurst: int32Ptr(100)},
				Telemetry: &KubernetesAgentSubstrateTelemetry{
					OtlpEndpoint:     "http://otel-collector.observability.svc:4317",
					TracesExporters:  []string{"otlp"},
					MetricsExporters: []string{"otlp", "prometheus"},
					LogsExporters:    []string{"console"},
					TracesSampler:    "parentbased_traceidratio", TracesSamplerRatio: float64Ptr(0.25),
					MetricExportIntervalMs: int32Ptr(10000), MetricExportTimeoutMs: int32Ptr(5000),
					WorkerTracesExporters:  []string{"none"},
					WorkerMetricsExporters: []string{"otlp"},
					WorkerLogsExporters:    []string{"otlp"},
				},
				ClusterSize:        KubernetesAgentSubstrateSpec_size0, // size10 excludes the bundled database's own resources set above
				CordonControlPlane: true,
				ImageRegistry:      "us-central1-docker.pkg.dev",
				ImagePullSecrets:   []string{"mirror-pull"},
				Version:            stringPtr("v0.6.0"),
				PreviewGates:       []string{"ExternalVolumes"},
				IdentityServer:     &KubernetesAgentSubstrateIdentityServer{Replicas: int32Ptr(3), Resources: resources, LogLevel: LogLevel_debug, DrainGrace: stringPtr("10s")},
			}
			input.Spec.Storage.SnapshotPrefix = "substrate/prod"
			valid()
		})
	})
})
