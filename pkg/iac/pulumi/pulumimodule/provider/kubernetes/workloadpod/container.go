// Package workloadpod converts the shared workload protos (WorkloadContainer,
// WorkloadPod) into Pulumi Kubernetes core/v1 inputs. Every workload kind's
// Pulumi module (Deployment, StatefulSet, DaemonSet, Job, CronJob) builds its
// pod template through this package, so container semantics — env wiring,
// probes, security hardening, volume derivation — are implemented exactly once
// and behave identically across kinds.
package workloadpod

import (
	kubernetesv1 "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/kubernetes/containerenv"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ImageReference renders a workload image as Kubernetes reads it:
// repo[:tag][@digest]. With a digest, Kubernetes pulls exactly that build and
// the tag stays as its readable name.
func ImageReference(image *kubernetesv1.WorkloadContainerImage) string {
	ref := image.GetRepo()
	if tag := image.GetTag(); tag != "" {
		ref += ":" + tag
	}
	if digest := image.GetDigest(); digest != "" {
		ref += "@" + digest
	}
	return ref
}

// BuildContainer converts one WorkloadContainer into Pulumi container args.
//
// defaultName is used when the spec omits the container name — kinds pass a
// stable role name ("app") for the main container so minimal manifests stay
// minimal; sidecars and init containers are required by validation to carry
// their own names.
//
// envSecretName is the workload-scoped Kubernetes Secret that holds literal
// secret env values (see CollectLiteralEnvSecrets); the env builder wires
// secretKeyRef entries against it.
func BuildContainer(spec *kubernetesv1.WorkloadContainer, defaultName string, envSecretName string) corev1.ContainerArgs {
	name := spec.Name
	if name == "" {
		name = defaultName
	}

	args := corev1.ContainerArgs{
		Name:    pulumi.String(name),
		Image:   pulumi.String(ImageReference(spec.GetImage())),
		Env:     corev1.EnvVarArray(containerenv.BuildEnvVars(spec.Env, envSecretName)),
		EnvFrom: containerenv.BuildEnvFrom(spec.Env),
	}

	if spec.ImagePullPolicy != "" {
		args.ImagePullPolicy = pulumi.String(spec.ImagePullPolicy)
	}
	if len(spec.Command) > 0 {
		args.Command = pulumi.ToStringArray(spec.Command)
	}
	if len(spec.Args) > 0 {
		args.Args = pulumi.ToStringArray(spec.Args)
	}
	if spec.WorkingDir != "" {
		args.WorkingDir = pulumi.String(spec.WorkingDir)
	}

	if len(spec.Ports) > 0 {
		ports := make(corev1.ContainerPortArray, 0, len(spec.Ports))
		for _, p := range spec.Ports {
			portArgs := &corev1.ContainerPortArgs{
				Name:          pulumi.String(p.Name),
				ContainerPort: pulumi.Int(p.ContainerPort),
			}
			// Kubernetes defaults protocol to TCP; only set it when the spec is explicit
			// so the rendered object matches what the API server stores (keeps diffs quiet).
			if p.NetworkProtocol != "" {
				portArgs.Protocol = pulumi.String(p.NetworkProtocol)
			}
			if p.HostPort > 0 {
				portArgs.HostPort = pulumi.Int(p.HostPort)
			}
			ports = append(ports, portArgs)
		}
		args.Ports = ports
	}

	if spec.Resources != nil {
		args.Resources = buildResources(spec.Resources)
	}

	args.LivenessProbe = BuildProbe(spec.LivenessProbe)
	args.ReadinessProbe = BuildProbe(spec.ReadinessProbe)
	args.StartupProbe = BuildProbe(spec.StartupProbe)

	if len(spec.VolumeMounts) > 0 {
		mounts := make(corev1.VolumeMountArray, 0, len(spec.VolumeMounts))
		for _, vm := range spec.VolumeMounts {
			mountArgs := &corev1.VolumeMountArgs{
				Name:      pulumi.String(vm.Name),
				MountPath: pulumi.String(vm.MountPath),
				ReadOnly:  pulumi.Bool(vm.ReadOnly),
			}
			if vm.SubPath != "" {
				mountArgs.SubPath = pulumi.String(vm.SubPath)
			}
			mounts = append(mounts, mountArgs)
		}
		args.VolumeMounts = mounts
	}

	if spec.Lifecycle != nil {
		args.Lifecycle = buildLifecycle(spec.Lifecycle)
	}

	if spec.SecurityContext != nil {
		args.SecurityContext = buildContainerSecurityContext(spec.SecurityContext)
	}

	return args
}

// BuildContainers assembles the pod's container list: the app container first
// (Kubernetes shows containers in declaration order, and tooling conventionally
// treats the first as primary), then sidecars in declared order.
func BuildContainers(app *kubernetesv1.WorkloadContainer, sidecars []*kubernetesv1.WorkloadContainer,
	appDefaultName string, envSecretName string) corev1.ContainerArray {
	containers := make(corev1.ContainerArray, 0, 1+len(sidecars))
	containers = append(containers, BuildContainer(app, appDefaultName, envSecretName))
	for _, sc := range sidecars {
		containers = append(containers, BuildContainer(sc, "", envSecretName))
	}
	return containers
}

// BuildInitContainers converts init containers, preserving declared order —
// Kubernetes runs them sequentially, so order is semantics, not cosmetics.
func BuildInitContainers(initContainers []*kubernetesv1.WorkloadContainer, envSecretName string) corev1.ContainerArray {
	if len(initContainers) == 0 {
		return nil
	}
	containers := make(corev1.ContainerArray, 0, len(initContainers))
	for _, ic := range initContainers {
		containers = append(containers, BuildContainer(ic, "", envSecretName))
	}
	return containers
}

// CollectLiteralEnvSecrets gathers literal secret env values from every
// container handed to it (app, sidecars, and init containers) into one
// name→value map — the data of the single workload-scoped env Secret. One
// Secret per workload (not per container) keeps teardown atomic and matches
// how the env builder references keys: secret env names are unique per
// container by Kubernetes rules, and sharing one Secret across containers is
// safe because keys are only ever read via explicit secretKeyRef.
func CollectLiteralEnvSecrets(containers ...*kubernetesv1.WorkloadContainer) map[string]string {
	data := map[string]string{}
	for _, c := range containers {
		if c == nil || c.Env == nil {
			continue
		}
		for _, s := range c.Env.Secrets {
			// Assert the concrete oneof wrapper: generated wrapper structs carry a
			// plain field (no getter), so an interface assertion on GetValue()
			// would silently match nothing and the Secret would never materialize.
			if src, ok := s.Source.(*kubernetesv1.SecretEnvVar_Value); ok && src.Value != "" {
				data[s.Name] = src.Value
			}
		}
	}
	return data
}

// BuildProbe converts a shared Probe proto into Pulumi probe args.
// Returns nil when the probe is not configured so the field is omitted entirely.
func BuildProbe(protoProbe *kubernetesv1.Probe) *corev1.ProbeArgs {
	if protoProbe == nil {
		return nil
	}

	probe := &corev1.ProbeArgs{}

	// Zero means "not set" for every timing field — Kubernetes applies its own
	// defaults (period 10s, timeout 1s, failure threshold 3), and emitting zeros
	// would be rejected by API validation (minimums are 1).
	if protoProbe.InitialDelaySeconds > 0 {
		probe.InitialDelaySeconds = pulumi.Int(protoProbe.InitialDelaySeconds)
	}
	if protoProbe.PeriodSeconds > 0 {
		probe.PeriodSeconds = pulumi.Int(protoProbe.PeriodSeconds)
	}
	if protoProbe.TimeoutSeconds > 0 {
		probe.TimeoutSeconds = pulumi.Int(protoProbe.TimeoutSeconds)
	}
	if protoProbe.SuccessThreshold > 0 {
		probe.SuccessThreshold = pulumi.Int(protoProbe.SuccessThreshold)
	}
	if protoProbe.FailureThreshold > 0 {
		probe.FailureThreshold = pulumi.Int(protoProbe.FailureThreshold)
	}

	switch handler := protoProbe.Handler.(type) {
	case *kubernetesv1.Probe_Grpc:
		grpcAction := &corev1.GRPCActionArgs{
			Port: pulumi.Int(handler.Grpc.Port),
		}
		if handler.Grpc.Service != "" {
			grpcAction.Service = pulumi.StringPtr(handler.Grpc.Service)
		}
		probe.Grpc = grpcAction

	case *kubernetesv1.Probe_HttpGet:
		httpGet := &corev1.HTTPGetActionArgs{}
		if handler.HttpGet.Path != "" {
			httpGet.Path = pulumi.String(handler.HttpGet.Path)
		}
		switch port := handler.HttpGet.Port.(type) {
		case *kubernetesv1.HTTPGetAction_PortNumber:
			httpGet.Port = pulumi.Int(port.PortNumber)
		case *kubernetesv1.HTTPGetAction_PortName:
			httpGet.Port = pulumi.String(port.PortName)
		}
		if handler.HttpGet.Host != "" {
			httpGet.Host = pulumi.String(handler.HttpGet.Host)
		}
		if handler.HttpGet.Scheme != "" {
			httpGet.Scheme = pulumi.String(handler.HttpGet.Scheme)
		}
		if len(handler.HttpGet.HttpHeaders) > 0 {
			headers := make(corev1.HTTPHeaderArray, 0, len(handler.HttpGet.HttpHeaders))
			for _, h := range handler.HttpGet.HttpHeaders {
				headers = append(headers, &corev1.HTTPHeaderArgs{
					Name:  pulumi.String(h.Name),
					Value: pulumi.String(h.Value),
				})
			}
			httpGet.HttpHeaders = headers
		}
		probe.HttpGet = httpGet

	case *kubernetesv1.Probe_TcpSocket:
		tcpSocket := &corev1.TCPSocketActionArgs{}
		switch port := handler.TcpSocket.Port.(type) {
		case *kubernetesv1.TCPSocketAction_PortNumber:
			tcpSocket.Port = pulumi.Int(port.PortNumber)
		case *kubernetesv1.TCPSocketAction_PortName:
			tcpSocket.Port = pulumi.String(port.PortName)
		}
		if handler.TcpSocket.Host != "" {
			tcpSocket.Host = pulumi.String(handler.TcpSocket.Host)
		}
		probe.TcpSocket = tcpSocket

	case *kubernetesv1.Probe_Exec:
		if len(handler.Exec.Command) > 0 {
			probe.Exec = &corev1.ExecActionArgs{
				Command: pulumi.ToStringArray(handler.Exec.Command),
			}
		}
	}

	return probe
}

func buildResources(res *kubernetesv1.ContainerResources) corev1.ResourceRequirementsArgs {
	args := corev1.ResourceRequirementsArgs{}
	if res.Limits != nil {
		limits := map[string]string{}
		if res.Limits.Cpu != "" {
			limits["cpu"] = res.Limits.Cpu
		}
		if res.Limits.Memory != "" {
			limits["memory"] = res.Limits.Memory
		}
		if len(limits) > 0 {
			args.Limits = pulumi.ToStringMap(limits)
		}
	}
	if res.Requests != nil {
		requests := map[string]string{}
		if res.Requests.Cpu != "" {
			requests["cpu"] = res.Requests.Cpu
		}
		if res.Requests.Memory != "" {
			requests["memory"] = res.Requests.Memory
		}
		if len(requests) > 0 {
			args.Requests = pulumi.ToStringMap(requests)
		}
	}
	return args
}

func buildLifecycle(lc *kubernetesv1.WorkloadContainerLifecycle) *corev1.LifecycleArgs {
	args := &corev1.LifecycleArgs{}
	if h := buildLifecycleHandler(lc.PostStart); h != nil {
		args.PostStart = h
	}
	if h := buildLifecycleHandler(lc.PreStop); h != nil {
		args.PreStop = h
	}
	return args
}

func buildLifecycleHandler(h *kubernetesv1.WorkloadLifecycleHandler) *corev1.LifecycleHandlerArgs {
	if h == nil {
		return nil
	}
	args := &corev1.LifecycleHandlerArgs{}
	switch handler := h.Handler.(type) {
	case *kubernetesv1.WorkloadLifecycleHandler_Exec:
		args.Exec = &corev1.ExecActionArgs{
			Command: pulumi.ToStringArray(handler.Exec.Command),
		}
	case *kubernetesv1.WorkloadLifecycleHandler_HttpGet:
		httpGet := &corev1.HTTPGetActionArgs{}
		if handler.HttpGet.Path != "" {
			httpGet.Path = pulumi.String(handler.HttpGet.Path)
		}
		switch port := handler.HttpGet.Port.(type) {
		case *kubernetesv1.HTTPGetAction_PortNumber:
			httpGet.Port = pulumi.Int(port.PortNumber)
		case *kubernetesv1.HTTPGetAction_PortName:
			httpGet.Port = pulumi.String(port.PortName)
		}
		if handler.HttpGet.Scheme != "" {
			httpGet.Scheme = pulumi.String(handler.HttpGet.Scheme)
		}
		args.HttpGet = httpGet
	case *kubernetesv1.WorkloadLifecycleHandler_TcpSocket:
		tcpSocket := &corev1.TCPSocketActionArgs{}
		switch port := handler.TcpSocket.Port.(type) {
		case *kubernetesv1.TCPSocketAction_PortNumber:
			tcpSocket.Port = pulumi.Int(port.PortNumber)
		case *kubernetesv1.TCPSocketAction_PortName:
			tcpSocket.Port = pulumi.String(port.PortName)
		}
		args.TcpSocket = tcpSocket
	case *kubernetesv1.WorkloadLifecycleHandler_Sleep:
		args.Sleep = &corev1.SleepActionArgs{
			Seconds: pulumi.Int(int(handler.Sleep.Seconds)),
		}
	}
	return args
}

func buildContainerSecurityContext(sc *kubernetesv1.WorkloadContainerSecurityContext) *corev1.SecurityContextArgs {
	args := &corev1.SecurityContextArgs{}
	if sc.Privileged {
		args.Privileged = pulumi.Bool(true)
	}
	if sc.RunAsUser != nil {
		args.RunAsUser = pulumi.Int(int(*sc.RunAsUser))
	}
	if sc.RunAsGroup != nil {
		args.RunAsGroup = pulumi.Int(int(*sc.RunAsGroup))
	}
	if sc.RunAsNonRoot != nil {
		args.RunAsNonRoot = pulumi.Bool(*sc.RunAsNonRoot)
	}
	if sc.ReadOnlyRootFilesystem != nil {
		args.ReadOnlyRootFilesystem = pulumi.Bool(*sc.ReadOnlyRootFilesystem)
	}
	if sc.AllowPrivilegeEscalation != nil {
		args.AllowPrivilegeEscalation = pulumi.Bool(*sc.AllowPrivilegeEscalation)
	}
	if sc.Capabilities != nil {
		capArgs := &corev1.CapabilitiesArgs{}
		if len(sc.Capabilities.Add) > 0 {
			capArgs.Add = pulumi.ToStringArray(sc.Capabilities.Add)
		}
		if len(sc.Capabilities.Drop) > 0 {
			capArgs.Drop = pulumi.ToStringArray(sc.Capabilities.Drop)
		}
		args.Capabilities = capArgs
	}
	if sc.SeccompProfile != nil {
		args.SeccompProfile = buildSeccompProfile(sc.SeccompProfile)
	}
	return args
}

func buildSeccompProfile(sp *kubernetesv1.WorkloadSeccompProfile) *corev1.SeccompProfileArgs {
	args := &corev1.SeccompProfileArgs{
		Type: pulumi.String(sp.Type),
	}
	if sp.LocalhostProfile != "" {
		args.LocalhostProfile = pulumi.StringPtr(sp.LocalhostProfile)
	}
	return args
}
