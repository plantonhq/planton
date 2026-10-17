package kubernetes

import (
	"errors"
	"testing"

	"buf.build/go/protovalidate"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

// The core/v1 library's rules are the checks the Pod API runs when it creates
// a pod, moved to authoring time: closed sets, quantities, ports, one source
// per volume and per handler, a selector operator's value count, and the keys
// the served schema requires. Each rule has one refusing case here, and a
// pod a person would really write passes.

func corePtr[T any](v T) *T { return &v }

func coreLiteral(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func corePod() *KubernetesCoreV1ApiPodSpec {
	return &KubernetesCoreV1ApiPodSpec{
		RestartPolicy: corePtr("Always"),
		Containers: []*KubernetesCoreV1ApiContainer{{
			Name:  "app",
			Image: "registry.k8s.io/e2e-test-images/agnhost:2.53",
			Ports: []*KubernetesCoreV1ApiContainerPort{{ContainerPort: corePtr(int32(8888)), Protocol: corePtr("TCP")}},
			Resources: &KubernetesCoreV1ApiResourceRequirements{
				Requests: map[string]string{"cpu": "500m", "memory": "1Gi"},
				Limits:   map[string]string{"memory": "2Gi"},
			},
			ReadinessProbe: &KubernetesCoreV1ApiProbe{
				HttpGet: &KubernetesCoreV1ApiHTTPGetAction{Path: "/healthz", Port: "8888"},
			},
			Env: []*KubernetesCoreV1ApiEnvVar{{
				Name: "API_KEY",
				ValueFrom: &KubernetesCoreV1ApiEnvVarSource{
					SecretKeyRef: &KubernetesCoreV1ApiSecretKeySelector{Name: coreLiteral("agent-credentials"), Key: "api-key"},
				},
			}},
		}},
		Volumes: []*KubernetesCoreV1ApiVolume{{
			Name:     "scratch",
			EmptyDir: &KubernetesCoreV1ApiEmptyDirVolumeSource{SizeLimit: corePtr("1Gi")},
		}},
		Tolerations: []*KubernetesCoreV1ApiToleration{{Key: "sandbox.gke.io/runtime", Operator: corePtr("Equal"), Value: "gvisor", Effect: corePtr("NoSchedule")}},
	}
}

func expectRule(t *testing.T, msg proto.Message, ruleID string) {
	t.Helper()
	err := protovalidate.Validate(msg)
	if ruleID == "" {
		if err != nil {
			t.Fatalf("want valid, got %v", err)
		}
		return
	}
	var validationErr *protovalidate.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("want a violation of %s, got %v", ruleID, err)
	}
	for _, v := range validationErr.Violations {
		if v.Proto.GetRuleId() == ruleID {
			return
		}
	}
	t.Fatalf("want a violation of %s, got %v", ruleID, err)
}

func TestCoreV1Api_Rules(t *testing.T) {
	cases := map[string]struct {
		mutate func(p *KubernetesCoreV1ApiPodSpec)
		rule   string
	}{
		"a real pod passes": {func(p *KubernetesCoreV1ApiPodSpec) {}, ""},
		"a pod needs a container": {func(p *KubernetesCoreV1ApiPodSpec) { p.Containers = nil }, "repeated.min_items"},
		"a container needs a name": {func(p *KubernetesCoreV1ApiPodSpec) { p.Containers[0].Name = "" }, "string.min_len"},
		"restart policy is a closed set": {func(p *KubernetesCoreV1ApiPodSpec) { p.RestartPolicy = corePtr("Sometimes") }, "string.in"},
		"a container port is required": {func(p *KubernetesCoreV1ApiPodSpec) { p.Containers[0].Ports[0].ContainerPort = nil }, "required"},
		"a protocol is a closed set": {func(p *KubernetesCoreV1ApiPodSpec) { p.Containers[0].Ports[0].Protocol = corePtr("tcp") }, "string.in"},
		"a resource quantity follows the grammar": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Containers[0].Resources.Limits["memory"] = "2 GB"
		}, "string.pattern"},
		"a size limit follows the grammar": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Volumes[0].EmptyDir.SizeLimit = corePtr("lots")
		}, "string.pattern"},
		"a probe port is a number or a name": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Containers[0].ReadinessProbe.HttpGet.Port = "70000"
		}, "http_get_action.port"},
		"a probe has one handler": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Containers[0].ReadinessProbe.TcpSocket = &KubernetesCoreV1ApiTCPSocketAction{Port: "8888"}
		}, "probe.one_source"},
		"a volume has one source": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Volumes[0].Secret = &KubernetesCoreV1ApiSecretVolumeSource{SecretName: coreLiteral("agent-credentials")}
		}, "volume.one_source"},
		"an env source names exactly one origin": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Containers[0].Env[0].ValueFrom.FieldRef = &KubernetesCoreV1ApiObjectFieldSelector{FieldPath: "metadata.name"}
		}, "env_var_source.one_source"},
		"a secret key selector names its secret": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Containers[0].Env[0].ValueFrom.SecretKeyRef.Name = nil
		}, "required"},
		"a toleration with Exists takes no value": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Tolerations[0].Operator = corePtr("Exists")
		}, "toleration.exists_takes_no_value"},
		"a node selector Gt takes one integer": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Affinity = &KubernetesCoreV1ApiAffinity{NodeAffinity: &KubernetesCoreV1ApiNodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &KubernetesCoreV1ApiNodeSelector{
					NodeSelectorTerms: []*KubernetesCoreV1ApiNodeSelectorTerm{{
						MatchExpressions: []*KubernetesCoreV1ApiNodeSelectorRequirement{{Key: "example.com/gen", Operator: "Gt", Values: []string{"three"}}},
					}},
				},
			}}
		}, "node_selector_requirement.values_by_operator"},
		"a label selector In needs values": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.TopologySpreadConstraints = []*KubernetesCoreV1ApiTopologySpreadConstraint{{
				MaxSkew: corePtr(int32(1)), TopologyKey: "kubernetes.io/hostname", WhenUnsatisfiable: "ScheduleAnyway",
				LabelSelector: &KubernetesCoreV1ApiLabelSelector{
					MatchExpressions: []*KubernetesCoreV1ApiLabelSelectorRequirement{{Key: "app", Operator: "In"}},
				},
			}}
		}, "label_selector_requirement.values_by_operator"},
		"an emptyDir medium is Memory or HugePages": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Volumes[0].EmptyDir.Medium = corePtr("Disk")
		}, "string.pattern"},
		"an iSCSI LUN of zero is valid": {func(p *KubernetesCoreV1ApiPodSpec) {
			p.Volumes[0].EmptyDir = nil
			p.Volumes[0].Iscsi = &KubernetesCoreV1ApiISCSIVolumeSource{TargetPortal: "10.0.0.1:3260", Iqn: "iqn.2001-04.com.example:storage", Lun: corePtr(int32(0))}
		}, ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			pod := corePod()
			c.mutate(pod)
			expectRule(t, pod, c.rule)
		})
	}
}
