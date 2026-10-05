package kuberneteslocustv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesLocust(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesLocust Suite")
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

func inlineScripts() *KubernetesLocustLoadTest_Inline {
	return &KubernetesLocustLoadTest_Inline{
		Inline: &KubernetesLocustInlineScripts{
			LocustfileContent: "from locust import HttpUser, task\n\nclass U(HttpUser):\n    @task\n    def index(self):\n        self.client.get(\"/\")\n",
		},
	}
}

var _ = ginkgo.Describe("KubernetesLocust Validation Tests", func() {
	var input *KubernetesLocust

	ginkgo.BeforeEach(func() {
		input = &KubernetesLocust{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesLocust",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "load-test",
			},
			Spec: &KubernetesLocustSpec{
				Namespace: literal("load-test"),
				LoadTest: &KubernetesLocustLoadTest{
					Scripts: inlineScripts(),
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (namespace + inline locustfile) should be valid", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full load-test surface should be valid", func() {
			input.Spec.LoadTest = &KubernetesLocustLoadTest{
				Name:    strPtr("checkout-soak"),
				Scripts: inlineScripts(),
				TargetHost: valueFrom(
					catalogkind.CatalogKind_KubernetesDeployment,
					"web-app", "status.outputs.service"),
				PipPackages:              []string{"faker==33.1.0", "pyjwt"},
				PipRequirementsConfigMap: "extra-requirements",
				Environment:              map[string]string{"CHECKOUT_PATH": "/api/checkout"},
				EnvFromSecrets:           []string{"shop-api-credentials"},
				EnvFromSecretKeys: []*KubernetesLocustSecretEnv{
					{SecretName: "payments-api", Keys: []string{"API_TOKEN"}},
				},
				Tags:        []string{"checkout", "browse"},
				ExcludeTags: []string{"admin"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("inline lib files with module names should be valid", func() {
			input.Spec.LoadTest.Scripts.(*KubernetesLocustLoadTest_Inline).Inline.LibFiles = map[string]string{
				"__init__.py": "",
				"helpers.py":  "def token():\n    return \"t\"\n",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("existing-ConfigMap scripts should be valid", func() {
			input.Spec.LoadTest.Scripts = &KubernetesLocustLoadTest_ExistingConfigMaps{
				ExistingConfigMaps: &KubernetesLocustExistingScriptConfigMaps{
					LocustfileConfigMap: "ci-locustfile",
					LocustfileName:      strPtr("perf_test.py"),
					LibConfigMap:        "ci-locust-lib",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("headless run should be valid", func() {
			input.Spec.LoadTest.Headless = true
			input.Spec.LoadTest.Environment = map[string]string{
				"LOCUST_USERS":      "100",
				"LOCUST_SPAWN_RATE": "10",
				"LOCUST_RUN_TIME":   "10m",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("worker HPA arm should be valid", func() {
			input.Spec.Workers = &KubernetesLocustWorkers{
				Autoscaling: &KubernetesLocustWorkers_Hpa{
					Hpa: &KubernetesLocustWorkerHpa{MaxReplicas: 20},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("worker KEDA arm with the login disabled should be valid", func() {
			input.Spec.WebUiAuth = &KubernetesLocustWebUiAuth{Enabled: boolPtr(false)}
			input.Spec.Workers = &KubernetesLocustWorkers{
				Autoscaling: &KubernetesLocustWorkers_Keda{
					Keda: &KubernetesLocustWorkerKeda{
						MaxReplicas:          50,
						TargetUsersPerWorker: int32Ptr(40),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("worker KEDA arm with custom triggers and the default login should be valid", func() {
			input.Spec.Workers = &KubernetesLocustWorkers{
				Autoscaling: &KubernetesLocustWorkers_Keda{
					Keda: &KubernetesLocustWorkerKeda{
						MaxReplicas:    50,
						CustomTriggers: "- type: prometheus\n  metadata:\n    serverAddress: http://prometheus:9090\n    query: locust_users\n    threshold: \"50\"\n",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("web-ui auth with custom username should be valid", func() {
			input.Spec.WebUiAuth = &KubernetesLocustWebUiAuth{
				Enabled:  boolPtr(true),
				Username: strPtr("perf-team"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("LoadBalancer service with annotations should be valid", func() {
			input.Spec.Service = &KubernetesLocustService{
				Type:        strPtr("LoadBalancer"),
				Annotations: map[string]string{"service.beta.kubernetes.io/aws-load-balancer-type": "nlb"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should be invalid", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing load_test should be invalid", func() {
			input.Spec.LoadTest = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("load_test without a scripts arm should be invalid", func() {
			input.Spec.LoadTest.Scripts = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("inline scripts without locustfile content should be invalid", func() {
			input.Spec.LoadTest.Scripts = &KubernetesLocustLoadTest_Inline{
				Inline: &KubernetesLocustInlineScripts{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("existing-ConfigMap scripts without the locustfile ConfigMap should be invalid", func() {
			input.Spec.LoadTest.Scripts = &KubernetesLocustLoadTest_ExistingConfigMaps{
				ExistingConfigMaps: &KubernetesLocustExistingScriptConfigMaps{},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("locustfile name without .py should be invalid", func() {
			input.Spec.LoadTest.Scripts = &KubernetesLocustLoadTest_ExistingConfigMaps{
				ExistingConfigMaps: &KubernetesLocustExistingScriptConfigMaps{
					LocustfileConfigMap: "ci-locustfile",
					LocustfileName:      strPtr("perf_test"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("lib file name with a space should be invalid", func() {
			input.Spec.LoadTest.Scripts.(*KubernetesLocustLoadTest_Inline).Inline.LibFiles = map[string]string{
				"my helpers.py": "",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("load-test name with uppercase should be invalid", func() {
			input.Spec.LoadTest.Name = strPtr("CheckoutSoak")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("pip package with a space should be invalid", func() {
			input.Spec.LoadTest.PipPackages = []string{"faker == 33.1.0"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("tag with a space should be invalid", func() {
			input.Spec.LoadTest.Tags = []string{"checkout flow"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("secret env without keys should be invalid", func() {
			input.Spec.LoadTest.EnvFromSecretKeys = []*KubernetesLocustSecretEnv{
				{SecretName: "payments-api"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("secret env without a secret name should be invalid", func() {
			input.Spec.LoadTest.EnvFromSecretKeys = []*KubernetesLocustSecretEnv{
				{Keys: []string{"API_TOKEN"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("negative worker replicas should be invalid", func() {
			input.Spec.Workers = &KubernetesLocustWorkers{Replicas: int32Ptr(-1)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("HPA without max replicas should be invalid", func() {
			input.Spec.Workers = &KubernetesLocustWorkers{
				Autoscaling: &KubernetesLocustWorkers_Hpa{
					Hpa: &KubernetesLocustWorkerHpa{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("HPA target CPU above 100 should be invalid", func() {
			input.Spec.Workers = &KubernetesLocustWorkers{
				Autoscaling: &KubernetesLocustWorkers_Hpa{
					Hpa: &KubernetesLocustWorkerHpa{
						MaxReplicas:                 10,
						TargetCpuUtilizationPercent: int32Ptr(150),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("KEDA without max replicas should be invalid", func() {
			input.Spec.Workers = &KubernetesLocustWorkers{
				Autoscaling: &KubernetesLocustWorkers_Keda{
					Keda: &KubernetesLocustWorkerKeda{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("KEDA default trigger with the default web-ui login should be invalid", func() {
			input.Spec.Workers = &KubernetesLocustWorkers{
				Autoscaling: &KubernetesLocustWorkers_Keda{
					Keda: &KubernetesLocustWorkerKeda{MaxReplicas: 50},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("KEDA default trigger with a headless run should be invalid", func() {
			input.Spec.LoadTest.Headless = true
			input.Spec.WebUiAuth = &KubernetesLocustWebUiAuth{Enabled: boolPtr(false)}
			input.Spec.Workers = &KubernetesLocustWorkers{
				Autoscaling: &KubernetesLocustWorkers_Keda{
					Keda: &KubernetesLocustWorkerKeda{MaxReplicas: 50},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("KEDA target users 0 should be invalid", func() {
			input.Spec.Workers = &KubernetesLocustWorkers{
				Autoscaling: &KubernetesLocustWorkers_Keda{
					Keda: &KubernetesLocustWorkerKeda{
						MaxReplicas:          10,
						TargetUsersPerWorker: int32Ptr(0),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("web-ui username with uppercase should be invalid", func() {
			input.Spec.WebUiAuth = &KubernetesLocustWebUiAuth{Username: strPtr("Admin")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("master log level outside the set should be invalid", func() {
			input.Spec.Master = &KubernetesLocustMaster{LogLevel: strPtr("TRACE")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("service type outside the set should be invalid", func() {
			input.Spec.Service = &KubernetesLocustService{Type: strPtr("ExternalName")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("wrong kind constant should be invalid", func() {
			input.Kind = "KubernetesLocustCluster"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
