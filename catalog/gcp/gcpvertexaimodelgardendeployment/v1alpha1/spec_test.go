package gcpvertexaimodelgardendeploymentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiModelGardenDeploymentSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpVertexAiModelGardenDeploymentSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// A small Hugging Face model on Model Garden's recommended shape.
	minimal := func() *GcpVertexAiModelGardenDeployment {
		return &GcpVertexAiModelGardenDeployment{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiModelGardenDeployment",
			Metadata:   &shared.CloudResourceMetadata{Name: "qwen-small"},
			Spec: &GcpVertexAiModelGardenDeploymentSpec{
				Location:           "us-central1",
				HuggingFaceModelId: "Qwen/Qwen3-0.6B",
				ModelConfig:        &GcpVertexAiModelGardenDeploymentModelConfig{AcceptEula: true},
			},
		}
	}

	probe := func() *GcpVertexAiModelGardenDeploymentProbe {
		return &GcpVertexAiModelGardenDeploymentProbe{
			PeriodSeconds:    proto.Int32(10),
			TimeoutSeconds:   proto.Int32(5),
			FailureThreshold: proto.Int32(3),
			Handler: &GcpVertexAiModelGardenDeploymentProbe_HttpGet{HttpGet: &GcpVertexAiModelGardenDeploymentHttpGetAction{
				Path: "/health",
				Port: proto.Int32(8080),
			}},
		}
	}

	ginkgo.It("should accept a Hugging Face model on the recommended shape", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a publisher model on a GPU with every lever", func() {
		msg := minimal()
		msg.Spec.ProjectId = nameRef("ai-project")
		msg.Spec.HuggingFaceModelId = ""
		msg.Spec.PublisherModelName = "publishers/google/models/gemma@gemma-1.1-2b-it"
		msg.Spec.ModelConfig = &GcpVertexAiModelGardenDeploymentModelConfig{
			AcceptEula:              true,
			HuggingFaceAccessToken:  "hf_secret",
			HuggingFaceCacheEnabled: true,
			ModelDisplayName:        "gemma-2b",
			ContainerSpec: &GcpVertexAiModelGardenDeploymentContainerSpec{
				ImageUri:           "us-docker.pkg.dev/vertex-ai/vertex-vision-model-garden-dockers/pytorch-vllm-serve:latest",
				Command:            []string{"python"},
				Args:               []string{"-m", "vllm.entrypoints.api_server"},
				Env:                []*GcpVertexAiModelGardenDeploymentEnvVar{{Name: "MODEL_ID", Value: "gemma-2b"}},
				Ports:              []*GcpVertexAiModelGardenDeploymentPort{{ContainerPort: proto.Int32(8080)}},
				GrpcPorts:          []*GcpVertexAiModelGardenDeploymentPort{{ContainerPort: proto.Int32(9000)}},
				PredictRoute:       "/generate",
				HealthRoute:        "/ping",
				DeploymentTimeout:  "3600s",
				SharedMemorySizeMb: proto.Int64(16384),
				StartupProbe:       probe(),
				LivenessProbe:      probe(),
				HealthProbe:        probe(),
			},
		}
		msg.Spec.DeployConfig = &GcpVertexAiModelGardenDeploymentDeployConfig{
			DedicatedResources: &GcpVertexAiModelGardenDeploymentDedicatedResources{
				MachineSpec: &GcpVertexAiModelGardenDeploymentMachineSpec{
					MachineType:      "g2-standard-12",
					AcceleratorType:  "NVIDIA_L4",
					AcceleratorCount: proto.Int32(1),
					ReservationAffinity: &GcpVertexAiModelGardenDeploymentReservationAffinity{
						ReservationAffinityType: "SPECIFIC_RESERVATION",
						Key:                     "compute.googleapis.com/reservation-name",
						Values:                  []string{"projects/ai-project/zones/us-central1-a/reservations/l4-pool"},
					},
				},
				MinReplicaCount:      1,
				MaxReplicaCount:      proto.Int32(3),
				RequiredReplicaCount: proto.Int32(1),
				Spot:                 true,
				AutoscalingMetricSpecs: []*GcpVertexAiModelGardenDeploymentAutoscalingMetricSpec{{
					MetricName: "aiplatform.googleapis.com/prediction/online/accelerator/duty_cycle",
					Target:     proto.Int32(70),
				}},
			},
			FastTryoutEnabled: true,
			SystemLabels:      map[string]string{"source": "model-garden"},
		}
		msg.Spec.EndpointConfig = &GcpVertexAiModelGardenDeploymentEndpointConfig{
			EndpointDisplayName:      "gemma-2b-endpoint",
			DedicatedEndpointEnabled: true,
			PrivateServiceConnectConfig: &GcpVertexAiModelGardenDeploymentPrivateServiceConnectConfig{
				EnablePrivateServiceConnect: true,
				ProjectAllowlist:            []*foreignkeyv1.StringValueOrRef{litRef("consumer-project")},
				PscAutomationConfig: &GcpVertexAiModelGardenDeploymentPscAutomationConfig{
					ProjectId: litRef("consumer-project"),
					Network:   nameRef("consumer-vpc"),
				},
			},
		}
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require exactly one of publisher_model_name and hugging_face_model_id", func() {
		msg := minimal()
		msg.Spec.PublisherModelName = "publishers/google/models/gemma@gemma-1.1-2b-it"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.HuggingFaceModelId = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject malformed model names", func() {
		msg := minimal()
		msg.Spec.HuggingFaceModelId = "Qwen3-0.6B"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.HuggingFaceModelId = ""
		msg.Spec.PublisherModelName = "google/gemma"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should pair accelerator type and count and bound replicas", func() {
		msg := minimal()
		msg.Spec.DeployConfig = &GcpVertexAiModelGardenDeploymentDeployConfig{
			DedicatedResources: &GcpVertexAiModelGardenDeploymentDedicatedResources{
				MachineSpec:     &GcpVertexAiModelGardenDeploymentMachineSpec{MachineType: "g2-standard-12", AcceleratorType: "NVIDIA_L4"},
				MinReplicaCount: 1,
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeployConfig = &GcpVertexAiModelGardenDeploymentDeployConfig{
			DedicatedResources: &GcpVertexAiModelGardenDeploymentDedicatedResources{
				MinReplicaCount: 3,
				MaxReplicaCount: proto.Int32(2),
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeployConfig = &GcpVertexAiModelGardenDeploymentDeployConfig{
			DedicatedResources: &GcpVertexAiModelGardenDeploymentDedicatedResources{MinReplicaCount: 0},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should tie reservation key and values to SPECIFIC_RESERVATION", func() {
		msg := minimal()
		msg.Spec.DeployConfig = &GcpVertexAiModelGardenDeploymentDeployConfig{
			DedicatedResources: &GcpVertexAiModelGardenDeploymentDedicatedResources{
				MachineSpec: &GcpVertexAiModelGardenDeploymentMachineSpec{
					MachineType:         "n1-standard-4",
					ReservationAffinity: &GcpVertexAiModelGardenDeploymentReservationAffinity{ReservationAffinityType: "SPECIFIC_RESERVATION"},
				},
				MinReplicaCount: 1,
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.DeployConfig.DedicatedResources.MachineSpec.ReservationAffinity = &GcpVertexAiModelGardenDeploymentReservationAffinity{
			ReservationAffinityType: "ANY_RESERVATION",
			Key:                     "compute.googleapis.com/reservation-name",
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one probe handler and a timeout within the period", func() {
		msg := minimal()
		msg.Spec.ModelConfig.ContainerSpec = &GcpVertexAiModelGardenDeploymentContainerSpec{
			ImageUri:    "us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-13:latest",
			HealthProbe: &GcpVertexAiModelGardenDeploymentProbe{PeriodSeconds: proto.Int32(10)},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.ModelConfig.ContainerSpec.HealthProbe = probe()
		msg.Spec.ModelConfig.ContainerSpec.HealthProbe.TimeoutSeconds = proto.Int32(30)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a container spec without an image and routes that do not start with a slash", func() {
		msg := minimal()
		msg.Spec.ModelConfig.ContainerSpec = &GcpVertexAiModelGardenDeploymentContainerSpec{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.ModelConfig.ContainerSpec = &GcpVertexAiModelGardenDeploymentContainerSpec{
			ImageUri:     "us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-13:latest",
			PredictRoute: "predict",
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
