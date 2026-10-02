package gcpcloudfunctionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestGcpCloudFunctionSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GcpCloudFunctionSpec Validation Suite")
}

var _ = Describe("GcpCloudFunctionSpec validations", func() {

	strVal := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	strRef := func(name string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
			},
		}
	}

	makeValidSpec := func() *GcpCloudFunctionSpec {
		return &GcpCloudFunctionSpec{
			ProjectId: strVal("my-gcp-project"),
			Region:    "us-central1",
			BuildConfig: &GcpCloudFunctionBuildConfig{
				Runtime:    "python312",
				EntryPoint: "hello_http",
				Source: &GcpCloudFunctionSource{
					StorageSource: &GcpCloudFunctionStorageSource{
						Bucket: strVal("my-source-bucket"),
						Object: "functions/hello-v1.zip",
					},
				},
			},
		}
	}

	Context("Required fields", func() {
		It("accepts a minimal valid spec", func() {
			Expect(protovalidate.Validate(makeValidSpec())).To(BeNil())
		})

		It("accepts a spec without project_id (ambient project)", func() {
			spec := makeValidSpec()
			spec.ProjectId = nil
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects spec with missing region", func() {
			spec := makeValidSpec()
			spec.Region = ""
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects spec with missing build_config", func() {
			spec := makeValidSpec()
			spec.BuildConfig = nil
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects build_config with missing runtime", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Runtime = ""
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects build_config with missing entry_point", func() {
			spec := makeValidSpec()
			spec.BuildConfig.EntryPoint = ""
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects build_config with missing source", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Source = nil
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Region validation", func() {
		It("accepts a multi-digit region", func() {
			spec := makeValidSpec()
			spec.Region = "europe-west12"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a zone as region", func() {
			spec := makeValidSpec()
			spec.Region = "us-central1-a"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Function name validation", func() {
		It("accepts a valid function name", func() {
			spec := makeValidSpec()
			spec.FunctionName = "my-function"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an uppercase function name", func() {
			spec := makeValidSpec()
			spec.FunctionName = "MyFunction"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a function name starting with a digit", func() {
			spec := makeValidSpec()
			spec.FunctionName = "1function"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Runtime validation", func() {
		It("accepts current runtimes", func() {
			for _, rt := range []string{"python313", "nodejs22", "go123", "java21", "dotnet8", "ruby33", "php83"} {
				spec := makeValidSpec()
				spec.BuildConfig.Runtime = rt
				Expect(protovalidate.Validate(spec)).To(BeNil())
			}
		})

		It("accepts runtimes newer than any hardcoded list (free-string contract)", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Runtime = "nodejs24"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a runtime with invalid characters", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Runtime = "Python 3.12"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Source arms (storage XOR repo)", func() {
		It("rejects a source with neither arm", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Source = &GcpCloudFunctionSource{}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a source with both arms", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Source.RepoSource = &GcpCloudFunctionRepoSource{
				RepoName:   "my-repo",
				BranchName: "main",
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a repo source with a branch", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Source = &GcpCloudFunctionSource{
				RepoSource: &GcpCloudFunctionRepoSource{
					RepoName:   "my-repo",
					BranchName: "main",
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a storage source without a bucket", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Source.StorageSource.Bucket = nil
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a storage source without an object", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Source.StorageSource.Object = ""
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a storage source bucket expressed as a reference", func() {
			spec := makeValidSpec()
			spec.BuildConfig.Source.StorageSource.Bucket = strRef("my-bucket-resource")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("Repo source revision exclusivity", func() {
		makeRepoSpec := func() *GcpCloudFunctionSpec {
			spec := makeValidSpec()
			spec.BuildConfig.Source = &GcpCloudFunctionSource{
				RepoSource: &GcpCloudFunctionRepoSource{RepoName: "my-repo"},
			}
			return spec
		}

		It("rejects a repo source with no revision pin", func() {
			Expect(protovalidate.Validate(makeRepoSpec())).NotTo(BeNil())
		})

		It("accepts exactly one revision pin", func() {
			spec := makeRepoSpec()
			spec.BuildConfig.Source.RepoSource.TagName = "v1.2.3"
			Expect(protovalidate.Validate(spec)).To(BeNil())

			spec = makeRepoSpec()
			spec.BuildConfig.Source.RepoSource.CommitSha = "abc123def456"
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects two revision pins", func() {
			spec := makeRepoSpec()
			spec.BuildConfig.Source.RepoSource.BranchName = "main"
			spec.BuildConfig.Source.RepoSource.TagName = "v1.2.3"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a repo source without a repo name", func() {
			spec := makeRepoSpec()
			spec.BuildConfig.Source.RepoSource.RepoName = ""
			spec.BuildConfig.Source.RepoSource.BranchName = "main"
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Service config", func() {
		withServiceConfig := func(sc *GcpCloudFunctionServiceConfig) *GcpCloudFunctionSpec {
			spec := makeValidSpec()
			spec.ServiceConfig = sc
			return spec
		}

		It("accepts memory quantity strings", func() {
			for _, m := range []string{"256M", "512M", "1Gi", "16Gi", "128Mi"} {
				spec := withServiceConfig(&GcpCloudFunctionServiceConfig{AvailableMemory: m})
				Expect(protovalidate.Validate(spec)).To(BeNil())
			}
		})

		It("rejects a bare-integer-with-unit-typo memory string", func() {
			spec := withServiceConfig(&GcpCloudFunctionServiceConfig{AvailableMemory: "256MB"})
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts fractional cpu", func() {
			spec := withServiceConfig(&GcpCloudFunctionServiceConfig{AvailableCpu: "0.5"})
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a non-numeric cpu", func() {
			spec := withServiceConfig(&GcpCloudFunctionServiceConfig{AvailableCpu: "one"})
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a timeout above 3600 seconds", func() {
			spec := withServiceConfig(&GcpCloudFunctionServiceConfig{TimeoutSeconds: 3601})
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects concurrency above 1000", func() {
			spec := withServiceConfig(&GcpCloudFunctionServiceConfig{MaxInstanceRequestConcurrency: 1001})
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a service account expressed as a reference", func() {
			spec := withServiceConfig(&GcpCloudFunctionServiceConfig{
				ServiceAccountEmail: strRef("my-sa-resource"),
			})
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a vpc connector expressed as a reference", func() {
			spec := withServiceConfig(&GcpCloudFunctionServiceConfig{
				VpcConnector: strRef("my-connector-resource"),
			})
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("Secret environment variables", func() {
		withSecretEnv := func(v *GcpCloudFunctionSecretEnvVar) *GcpCloudFunctionSpec {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				SecretEnvironmentVariables: []*GcpCloudFunctionSecretEnvVar{v},
			}
			return spec
		}

		It("accepts a valid secret env var", func() {
			spec := withSecretEnv(&GcpCloudFunctionSecretEnvVar{
				Key:     "DATABASE_PASSWORD",
				Secret:  "db-password",
				Version: "latest",
			})
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a secret env var without a key", func() {
			spec := withSecretEnv(&GcpCloudFunctionSecretEnvVar{Secret: "db-password"})
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a secret env var with neither a secret nor a value", func() {
			spec := withSecretEnv(&GcpCloudFunctionSecretEnvVar{Key: "DATABASE_PASSWORD"})
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a secret value the component stores", func() {
			spec := withSecretEnv(&GcpCloudFunctionSecretEnvVar{Key: "DATABASE_PASSWORD", Value: "hunter2"})
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a secret env var with both a secret and a value", func() {
			spec := withSecretEnv(&GcpCloudFunctionSecretEnvVar{Key: "DATABASE_PASSWORD", Secret: "db-password", Value: "hunter2"})
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("exactly one of secret"))
		})

		It("rejects a version or project_id beside a stored value", func() {
			spec := withSecretEnv(&GcpCloudFunctionSecretEnvVar{Key: "DATABASE_PASSWORD", Value: "hunter2", Version: "latest"})
			err := protovalidate.Validate(spec)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("apply only with secret"))
			spec = withSecretEnv(&GcpCloudFunctionSecretEnvVar{Key: "DATABASE_PASSWORD", Value: "hunter2", ProjectId: "shared-secrets"})
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
			spec = withSecretEnv(&GcpCloudFunctionSecretEnvVar{Key: "DATABASE_PASSWORD", Secret: "db-password", Version: "3", ProjectId: "shared-secrets"})
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an env var name starting with a digit", func() {
			spec := withSecretEnv(&GcpCloudFunctionSecretEnvVar{Key: "1BAD", Secret: "s"})
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Build worker pool", func() {
		It("takes the worker pool as a reference or a full literal name", func() {
			spec := makeValidSpec()
			spec.BuildConfig.WorkerPool = strRef("private-builds")
			Expect(protovalidate.Validate(spec)).To(BeNil())
			spec.BuildConfig.WorkerPool = strVal("projects/my-gcp-project/locations/us-central1/workerPools/private")
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a literal worker pool that is not a full name", func() {
			spec := makeValidSpec()
			spec.BuildConfig.WorkerPool = strVal("private")
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Secret volumes", func() {
		It("accepts a valid secret volume with versions", func() {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				SecretVolumes: []*GcpCloudFunctionSecretVolume{{
					MountPath: "/etc/secrets",
					Secret:    "tls-cert",
					Versions: []*GcpCloudFunctionSecretVolumeVersion{
						{Version: "latest", Path: "cert.pem"},
					},
				}},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a secret volume without a mount path", func() {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				SecretVolumes: []*GcpCloudFunctionSecretVolume{{Secret: "tls-cert"}},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects a volume version without a path", func() {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				SecretVolumes: []*GcpCloudFunctionSecretVolume{{
					MountPath: "/etc/secrets",
					Secret:    "tls-cert",
					Versions:  []*GcpCloudFunctionSecretVolumeVersion{{Version: "latest"}},
				}},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})
	})

	Context("Scaling bounds", func() {
		withScaling := func(min, max int32) *GcpCloudFunctionSpec {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				Scaling: &GcpCloudFunctionScalingConfig{
					MinInstanceCount: min,
					MaxInstanceCount: max,
				},
			}
			return spec
		}

		It("accepts valid scaling bounds", func() {
			Expect(protovalidate.Validate(withScaling(1, 10))).To(BeNil())
		})

		It("rejects min above max", func() {
			Expect(protovalidate.Validate(withScaling(10, 5))).NotTo(BeNil())
		})

		It("rejects max above 3000", func() {
			Expect(protovalidate.Validate(withScaling(0, 3001))).NotTo(BeNil())
		})
	})

	Context("Trigger coherence", func() {
		It("accepts an explicit HTTP trigger", func() {
			spec := makeValidSpec()
			spec.Trigger = &GcpCloudFunctionTrigger{TriggerType: GcpCloudFunctionTriggerType_HTTP}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects EVENT_TRIGGER without an event_trigger block", func() {
			spec := makeValidSpec()
			spec.Trigger = &GcpCloudFunctionTrigger{TriggerType: GcpCloudFunctionTriggerType_EVENT_TRIGGER}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a Pub/Sub event trigger with a topic reference", func() {
			spec := makeValidSpec()
			spec.Trigger = &GcpCloudFunctionTrigger{
				TriggerType: GcpCloudFunctionTriggerType_EVENT_TRIGGER,
				EventTrigger: &GcpCloudFunctionEventTrigger{
					EventType:   "google.cloud.pubsub.topic.v1.messagePublished",
					PubsubTopic: strRef("my-topic-resource"),
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects an event trigger without an event type", func() {
			spec := makeValidSpec()
			spec.Trigger = &GcpCloudFunctionTrigger{
				TriggerType:  GcpCloudFunctionTriggerType_EVENT_TRIGGER,
				EventTrigger: &GcpCloudFunctionEventTrigger{},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("rejects an event filter missing its value", func() {
			spec := makeValidSpec()
			spec.Trigger = &GcpCloudFunctionTrigger{
				TriggerType: GcpCloudFunctionTriggerType_EVENT_TRIGGER,
				EventTrigger: &GcpCloudFunctionEventTrigger{
					EventType:    "google.cloud.storage.object.v1.finalized",
					EventFilters: []*GcpCloudFunctionEventFilter{{Attribute: "bucket"}},
				},
			}
			Expect(protovalidate.Validate(spec)).NotTo(BeNil())
		})

		It("accepts a multi-region trigger region", func() {
			spec := makeValidSpec()
			spec.Trigger = &GcpCloudFunctionTrigger{
				TriggerType: GcpCloudFunctionTriggerType_EVENT_TRIGGER,
				EventTrigger: &GcpCloudFunctionEventTrigger{
					EventType:     "google.cloud.storage.object.v1.finalized",
					TriggerRegion: "us",
					EventFilters: []*GcpCloudFunctionEventFilter{
						{Attribute: "bucket", Value: "my-bucket"},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("Build update policy and traffic defaults", func() {
		It("accepts ON_DEPLOY update policy", func() {
			spec := makeValidSpec()
			spec.BuildConfig.UpdatePolicy = GcpCloudFunctionBuildUpdatePolicy_ON_DEPLOY
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts holding traffic off the latest revision", func() {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				AllTrafficOnLatestRevision: proto.Bool(false),
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("Direct VPC egress", func() {
		It("accepts a direct VPC interface with a network reference", func() {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				DirectVpcNetworkInterface: &GcpCloudFunctionDirectVpcNetworkInterface{
					Network: strRef("my-vpc"),
					Tags:    []string{"fn-egress"},
				},
				DirectVpcEgress: GcpCloudFunctionVpcEgressSetting_ALL_TRAFFIC,
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("accepts a direct VPC interface with a literal subnetwork only", func() {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				DirectVpcNetworkInterface: &GcpCloudFunctionDirectVpcNetworkInterface{
					Subnetwork: strVal("my-subnet"),
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("rejects a direct VPC interface with neither network nor subnetwork", func() {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				DirectVpcNetworkInterface: &GcpCloudFunctionDirectVpcNetworkInterface{
					Tags: []string{"fn-egress"},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("rejects combining a connector with a direct VPC interface", func() {
			spec := makeValidSpec()
			spec.ServiceConfig = &GcpCloudFunctionServiceConfig{
				VpcConnector: strVal("projects/p/locations/us-central1/connectors/c"),
				DirectVpcNetworkInterface: &GcpCloudFunctionDirectVpcNetworkInterface{
					Network: strVal("my-vpc"),
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})

	Context("Deletion policy", func() {
		It("accepts each deletion_policy value", func() {
			for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
				spec := makeValidSpec()
				spec.DeletionPolicy = v
				Expect(protovalidate.Validate(spec)).To(BeNil())
			}
		})

		It("rejects an invalid deletion_policy", func() {
			spec := makeValidSpec()
			spec.DeletionPolicy = "KEEP"
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})
})
