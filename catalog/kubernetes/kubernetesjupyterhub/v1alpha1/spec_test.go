package kubernetesjupyterhubv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesJupyterHub(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesJupyterHub Suite")
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
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

func testHubPostgres() *KubernetesJupyterHubPostgres {
	return &KubernetesJupyterHubPostgres{
		Host: literal("hub-pg-rw"),
		PasswordSecret: &KubernetesJupyterHubPasswordSecret{
			SecretName: literal("hub-pg-app"),
		},
	}
}

var _ = ginkgo.Describe("KubernetesJupyterHub Validation Tests", func() {
	var input *KubernetesJupyterHub

	ginkgo.BeforeEach(func() {
		input = &KubernetesJupyterHub{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesJupyterHub",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "notebooks",
			},
			Spec: &KubernetesJupyterHubSpec{
				Namespace: literal("jupyterhub"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (namespace only — every default applies) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "jupyterhub", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("sqlite-pvc hub database with sizing should be valid", func() {
			input.Spec.Hub = &KubernetesJupyterHubHub{
				Database: &KubernetesJupyterHubDatabase{
					Backend: &KubernetesJupyterHubDatabase_SqlitePvc{SqlitePvc: &KubernetesJupyterHubSqlitePvc{
						StorageSize:  strPtr("2Gi"),
						StorageClass: "fast-ssd",
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("postgres hub database composed from a KubernetesPostgres should be valid", func() {
			pg := testHubPostgres()
			pg.Host = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "hub-pg", "status.outputs.rw_service")
			pg.PasswordSecret.SecretName = valueFrom(catalogkind.CatalogKind_KubernetesPostgres, "hub-pg", "status.outputs.password_secret.name")
			input.Spec.Hub = &KubernetesJupyterHubHub{
				Database: &KubernetesJupyterHubDatabase{
					Backend: &KubernetesJupyterHubDatabase_Postgres{Postgres: pg},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("postgres tuning (port, database, user) should be valid", func() {
			pg := testHubPostgres()
			pg.Port = int32Ptr(5433)
			pg.DatabaseName = strPtr("hub_prod")
			pg.Username = strPtr("hub_svc")
			input.Spec.Hub = &KubernetesJupyterHubHub{
				Database: &KubernetesJupyterHubDatabase{
					Backend: &KubernetesJupyterHubDatabase_Postgres{Postgres: pg},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("mysql hub database arm should be valid", func() {
			input.Spec.Hub = &KubernetesJupyterHubHub{
				Database: &KubernetesJupyterHubDatabase{
					Backend: &KubernetesJupyterHubDatabase_Mysql{Mysql: &KubernetesJupyterHubMysql{
						Host: literal("hub-mysql"),
						PasswordSecret: &KubernetesJupyterHubMysqlPasswordSecret{
							SecretName: literal("hub-mysql-auth"),
						},
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("hub throttles and named servers should be valid", func() {
			input.Spec.Hub = &KubernetesJupyterHubHub{
				ConcurrentSpawnLimit:    int32Ptr(16),
				ActiveServerLimit:       int32Ptr(200),
				AllowNamedServers:       true,
				NamedServerLimitPerUser: int32Ptr(3),
				ShutdownOnLogout:        true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("shared-password auth with a module-generated password (empty message) should be valid", func() {
			input.Spec.Authentication = &KubernetesJupyterHubAuth{
				Method:     &KubernetesJupyterHubAuth_SharedPassword{SharedPassword: &KubernetesJupyterHubDummyAuth{}},
				AdminUsers: []string{"ada"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("shared-password auth with an existing Secret should be valid", func() {
			input.Spec.Authentication = &KubernetesJupyterHubAuth{
				Method: &KubernetesJupyterHubAuth_SharedPassword{SharedPassword: &KubernetesJupyterHubDummyAuth{
					PasswordSecret: &KubernetesJupyterHubExistingSecretRef{SecretName: "team-hub-password"},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("native auth with signup tuning should be valid", func() {
			input.Spec.Authentication = &KubernetesJupyterHubAuth{
				Method: &KubernetesJupyterHubAuth_Native{Native: &KubernetesJupyterHubNativeAuth{
					OpenSignup:            true,
					MinimumPasswordLength: int32Ptr(12),
				}},
				AllowedUsers: []string{"ada", "grace"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("github auth with org gating should be valid", func() {
			input.Spec.Authentication = &KubernetesJupyterHubAuth{
				Method: &KubernetesJupyterHubAuth_Github{Github: &KubernetesJupyterHubGithubAuth{
					ClientId:             "Iv1.abc123",
					ClientSecretSecret:   &KubernetesJupyterHubExistingSecretRef{SecretName: "gh-oauth", SecretKey: strPtr("client_secret")},
					OauthCallbackUrl:     "https://hub.example.com/hub/oauth_callback",
					AllowedOrganizations: []string{"example-org", "example-org:ml-team"},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("google auth with hosted domain should be valid", func() {
			input.Spec.Authentication = &KubernetesJupyterHubAuth{
				Method: &KubernetesJupyterHubAuth_Google{Google: &KubernetesJupyterHubGoogleAuth{
					ClientId:           "1234.apps.googleusercontent.com",
					ClientSecretSecret: &KubernetesJupyterHubExistingSecretRef{SecretName: "google-oauth"},
					OauthCallbackUrl:   "https://hub.example.com/hub/oauth_callback",
					HostedDomains:      []string{"example.com"},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("generic OIDC auth (Keycloak-shaped) should be valid", func() {
			input.Spec.Authentication = &KubernetesJupyterHubAuth{
				Method: &KubernetesJupyterHubAuth_Oidc{Oidc: &KubernetesJupyterHubOidcAuth{
					ClientId:           "jupyterhub",
					ClientSecretSecret: &KubernetesJupyterHubExistingSecretRef{SecretName: "kc-oauth", SecretKey: strPtr("client-secret")},
					OauthCallbackUrl:   "https://hub.example.com/hub/oauth_callback",
					AuthorizeUrl:       "https://kc.example.com/realms/eng/protocol/openid-connect/auth",
					TokenUrl:           "https://kc.example.com/realms/eng/protocol/openid-connect/token",
					UserdataUrl:        "https://kc.example.com/realms/eng/protocol/openid-connect/userinfo",
					Scopes:             []string{"openid", "email"},
					UsernameClaim:      strPtr("email"),
					LoginService:       strPtr("Keycloak"),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("proxy exposure and sizing should be valid", func() {
			input.Spec.Proxy = &KubernetesJupyterHubProxy{
				ServiceType: strPtr("LoadBalancer"),
				ServiceAnnotations: map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("single-user image, sizing and dynamic storage should be valid", func() {
			input.Spec.SingleUser = &KubernetesJupyterHubSingleUser{
				Image:               &KubernetesJupyterHubImage{Repository: "quay.io/jupyter/scipy-notebook", Tag: "2026-07-28"},
				MemoryGuarantee:     strPtr("2G"),
				MemoryLimit:         "4G",
				CpuGuarantee:        "0.5",
				CpuLimit:            "2",
				DefaultUrl:          "/lab",
				StartTimeoutSeconds: int32Ptr(600),
				ExtraEnv:            map[string]string{"MLFLOW_TRACKING_URI": "http://mlflow.mlflow:5000"},
				Storage: &KubernetesJupyterHubUserStorage{
					Mode: &KubernetesJupyterHubUserStorage_Dynamic{Dynamic: &KubernetesJupyterHubDynamicStorage{
						Capacity:     strPtr("20Gi"),
						StorageClass: "fast-ssd",
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("static shared storage should be valid", func() {
			input.Spec.SingleUser = &KubernetesJupyterHubSingleUser{
				Storage: &KubernetesJupyterHubUserStorage{
					Mode: &KubernetesJupyterHubUserStorage_Static{Static: &KubernetesJupyterHubStaticStorage{
						PvcName: "shared-homes",
						SubPath: strPtr("homes/{username}"),
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("ephemeral storage (none) should be valid", func() {
			input.Spec.SingleUser = &KubernetesJupyterHubSingleUser{
				Storage: &KubernetesJupyterHubUserStorage{
					Mode: &KubernetesJupyterHubUserStorage_None{None: &KubernetesJupyterHubNoStorage{}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("spawn-menu profiles should be valid", func() {
			input.Spec.SingleUser = &KubernetesJupyterHubSingleUser{
				Profiles: []*KubernetesJupyterHubProfile{
					{DisplayName: "Small", Description: "2G RAM", Default: true, MemoryGuarantee: "2G", MemoryLimit: "2G"},
					{DisplayName: "GPU workstation", Image: &KubernetesJupyterHubImage{Repository: "quay.io/jupyter/pytorch-notebook", Tag: "cuda12"}, CpuLimit: "8"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("scheduling machinery (placeholders + selectors) should be valid", func() {
			input.Spec.Scheduling = &KubernetesJupyterHubScheduling{
				UserSchedulerEnabled:    boolPtr(true),
				UserPlaceholderReplicas: int32Ptr(3),
				CoreNodeSelector:        map[string]string{"hub.jupyter.org/node-purpose": "core"},
				UserNodeSelector:        map[string]string{"hub.jupyter.org/node-purpose": "user"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("culling tuning should be valid", func() {
			input.Spec.Culling = &KubernetesJupyterHubCulling{
				Enabled:        boolPtr(true),
				TimeoutSeconds: int32Ptr(1800),
				EverySeconds:   int32Ptr(300),
				MaxAgeSeconds:  int32Ptr(28800),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("pre-puller toggles and network policy off should be valid", func() {
			input.Spec.PrePuller = &KubernetesJupyterHubPrePuller{
				HookEnabled:       boolPtr(false),
				ContinuousEnabled: boolPtr(false),
			}
			input.Spec.NetworkPolicyEnabled = boolPtr(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("helm values escape hatch should be valid", func() {
			input.Spec.HelmValues = "hub:\n  extraConfig:\n    myConfig: |\n      c.JupyterHub.log_level = 'DEBUG'\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("empty database oneof should fail", func() {
			input.Spec.Hub = &KubernetesJupyterHubHub{Database: &KubernetesJupyterHubDatabase{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("postgres without a password secret should fail", func() {
			input.Spec.Hub = &KubernetesJupyterHubHub{
				Database: &KubernetesJupyterHubDatabase{
					Backend: &KubernetesJupyterHubDatabase_Postgres{Postgres: &KubernetesJupyterHubPostgres{
						Host: literal("hub-pg-rw"),
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("postgres with an out-of-range port should fail", func() {
			pg := testHubPostgres()
			pg.Port = int32Ptr(70000)
			input.Spec.Hub = &KubernetesJupyterHubHub{
				Database: &KubernetesJupyterHubDatabase{
					Backend: &KubernetesJupyterHubDatabase_Postgres{Postgres: pg},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("postgres with an invalid database name should fail", func() {
			pg := testHubPostgres()
			pg.DatabaseName = strPtr("bad-name!")
			input.Spec.Hub = &KubernetesJupyterHubHub{
				Database: &KubernetesJupyterHubDatabase{
					Backend: &KubernetesJupyterHubDatabase_Postgres{Postgres: pg},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("named-server limit without allow_named_servers should fail", func() {
			input.Spec.Hub = &KubernetesJupyterHubHub{
				NamedServerLimitPerUser: int32Ptr(3),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("github auth without a client secret should fail", func() {
			input.Spec.Authentication = &KubernetesJupyterHubAuth{
				Method: &KubernetesJupyterHubAuth_Github{Github: &KubernetesJupyterHubGithubAuth{
					ClientId:         "Iv1.abc123",
					OauthCallbackUrl: "https://hub.example.com/hub/oauth_callback",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("github auth without a callback URL should fail", func() {
			input.Spec.Authentication = &KubernetesJupyterHubAuth{
				Method: &KubernetesJupyterHubAuth_Github{Github: &KubernetesJupyterHubGithubAuth{
					ClientId:           "Iv1.abc123",
					ClientSecretSecret: &KubernetesJupyterHubExistingSecretRef{SecretName: "gh-oauth"},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("OIDC auth missing its endpoints should fail", func() {
			input.Spec.Authentication = &KubernetesJupyterHubAuth{
				Method: &KubernetesJupyterHubAuth_Oidc{Oidc: &KubernetesJupyterHubOidcAuth{
					ClientId:           "jupyterhub",
					ClientSecretSecret: &KubernetesJupyterHubExistingSecretRef{SecretName: "kc-oauth"},
					OauthCallbackUrl:   "https://hub.example.com/hub/oauth_callback",
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("proxy service type outside the allowed set should fail", func() {
			input.Spec.Proxy = &KubernetesJupyterHubProxy{ServiceType: strPtr("ExternalName")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("profile without a display name should fail", func() {
			input.Spec.SingleUser = &KubernetesJupyterHubSingleUser{
				Profiles: []*KubernetesJupyterHubProfile{{Description: "no name"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("single-user image without a tag should fail", func() {
			input.Spec.SingleUser = &KubernetesJupyterHubSingleUser{
				Image: &KubernetesJupyterHubImage{Repository: "quay.io/jupyter/scipy-notebook"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("static storage without a PVC name should fail", func() {
			input.Spec.SingleUser = &KubernetesJupyterHubSingleUser{
				Storage: &KubernetesJupyterHubUserStorage{
					Mode: &KubernetesJupyterHubUserStorage_Static{Static: &KubernetesJupyterHubStaticStorage{}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("dynamic storage with a malformed capacity should fail", func() {
			input.Spec.SingleUser = &KubernetesJupyterHubSingleUser{
				Storage: &KubernetesJupyterHubUserStorage{
					Mode: &KubernetesJupyterHubUserStorage_Dynamic{Dynamic: &KubernetesJupyterHubDynamicStorage{
						Capacity: strPtr("ten gigs"),
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("culling timeout below the floor should fail", func() {
			input.Spec.Culling = &KubernetesJupyterHubCulling{TimeoutSeconds: int32Ptr(10)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("spawn start timeout below the floor should fail", func() {
			input.Spec.SingleUser = &KubernetesJupyterHubSingleUser{StartTimeoutSeconds: int32Ptr(5)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("negative placeholder replicas should fail", func() {
			input.Spec.Scheduling = &KubernetesJupyterHubScheduling{UserPlaceholderReplicas: int32Ptr(-1)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("wrong kind constant should fail", func() {
			input.Kind = "JupyterHub"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
