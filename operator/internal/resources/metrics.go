package resources

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// The platform's own metrics surface. The control plane and the runner each
// serve Prometheus text on one port, always: it costs nothing idle, carries no
// customer identifier, and is reachable only inside the cluster (no front door
// routes it). The port is named on the container and on the component's
// Service, so a monitor -- one per component, since the paths differ --
// selects the Service and names the port, never a number:
//
//	selector: app.kubernetes.io/managed-by=planton-operator,
//	          app.kubernetes.io/name=control-plane (or runner)
//	jobLabel: app.kubernetes.io/name   (job="control-plane" / job="runner")
//	port: metrics, path /actuator/prometheus (control plane) or /metrics (runner)
//
// The number is the one the hosted product serves on, so one set of
// dashboards reads both. Traces are the only signal that needs a setting
// (spec.observability.otlpHttpEndpoint), because they need somewhere to go.
const (
	MetricsPort     = 9464
	MetricsPortName = "metrics"
)

// metricsPortEnv is the METRICS_PORT both binaries read to open the surface.
func metricsPortEnv() corev1.EnvVar {
	return corev1.EnvVar{Name: "METRICS_PORT", Value: fmt.Sprintf("%d", MetricsPort)}
}

func metricsContainerPort() corev1.ContainerPort {
	return corev1.ContainerPort{Name: MetricsPortName, ContainerPort: MetricsPort, Protocol: corev1.ProtocolTCP}
}

// metricsServicePort is plain HTTP: appProtocol http, so a mesh or a scraper
// treats it as ordinary web traffic.
func metricsServicePort() corev1.ServicePort {
	return corev1.ServicePort{
		Name:        MetricsPortName,
		Port:        MetricsPort,
		TargetPort:  intstr.FromString(MetricsPortName),
		Protocol:    corev1.ProtocolTCP,
		AppProtocol: new("http"),
	}
}
