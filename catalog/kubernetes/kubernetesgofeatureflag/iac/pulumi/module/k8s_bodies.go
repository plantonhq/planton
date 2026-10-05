package module

import (
	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
)

// The renderers below turn the shared Kubernetes workload messages into the
// plain maps the chart's values expect (the chart passes nodeSelector,
// tolerations, affinity, podSecurityContext and securityContext straight
// through with toYaml). Each has an exact twin in the Terraform module's
// locals.tf.

// resourcesBody renders ContainerResources (nil when nothing is declared).
func resourcesBody(r *kubernetesprovider.ContainerResources) map[string]interface{} {
	if r == nil {
		return nil
	}
	out := map[string]interface{}{}
	if q := cpuMemory(r.GetRequests()); q != nil {
		out["requests"] = q
	}
	if l := cpuMemory(r.GetLimits()); l != nil {
		out["limits"] = l
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cpuMemory(c *kubernetesprovider.CpuMemory) map[string]interface{} {
	if c == nil || (c.GetCpu() == "" && c.GetMemory() == "") {
		return nil
	}
	out := map[string]interface{}{}
	setString(out, "cpu", c.GetCpu())
	setString(out, "memory", c.GetMemory())
	return out
}

// tolerationsBody renders the shared toleration list.
func tolerationsBody(tolerations []*kubernetesprovider.WorkloadToleration) []interface{} {
	out := make([]interface{}, 0, len(tolerations))
	for _, t := range tolerations {
		tol := map[string]interface{}{}
		setString(tol, "key", t.GetKey())
		setString(tol, "operator", t.GetOperator())
		setString(tol, "value", t.GetValue())
		setString(tol, "effect", t.GetEffect())
		if t.TolerationSeconds != nil {
			tol["tolerationSeconds"] = t.GetTolerationSeconds()
		}
		out = append(out, tol)
	}
	return out
}

// affinityBody renders node affinity, pod affinity and pod anti-affinity
// into one Kubernetes affinity object (nil when none is declared).
func affinityBody(node *kubernetesprovider.WorkloadNodeAffinity, pod, antiPod *kubernetesprovider.WorkloadPodAffinity) map[string]interface{} {
	out := map[string]interface{}{}
	if node != nil && (len(node.GetRequired()) > 0 || len(node.GetPreferred()) > 0) {
		na := map[string]interface{}{}
		if len(node.GetRequired()) > 0 {
			terms := []interface{}{}
			for _, t := range node.GetRequired() {
				terms = append(terms, nodeSelectorTerm(t))
			}
			na["requiredDuringSchedulingIgnoredDuringExecution"] = map[string]interface{}{"nodeSelectorTerms": terms}
		}
		if len(node.GetPreferred()) > 0 {
			preferred := []interface{}{}
			for _, p := range node.GetPreferred() {
				preferred = append(preferred, map[string]interface{}{
					"weight":     int(p.GetWeight()),
					"preference": nodeSelectorTerm(p.GetTerm()),
				})
			}
			na["preferredDuringSchedulingIgnoredDuringExecution"] = preferred
		}
		out["nodeAffinity"] = na
	}
	if pa := podAffinityBody(pod); pa != nil {
		out["podAffinity"] = pa
	}
	if pa := podAffinityBody(antiPod); pa != nil {
		out["podAntiAffinity"] = pa
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func nodeSelectorTerm(t *kubernetesprovider.WorkloadNodeSelectorTerm) map[string]interface{} {
	exprs := []interface{}{}
	for _, e := range t.GetMatchExpressions() {
		expr := map[string]interface{}{"key": e.GetKey(), "operator": e.GetOperator()}
		if len(e.GetValues()) > 0 {
			expr["values"] = stringSlice(e.GetValues())
		}
		exprs = append(exprs, expr)
	}
	return map[string]interface{}{"matchExpressions": exprs}
}

func podAffinityBody(a *kubernetesprovider.WorkloadPodAffinity) map[string]interface{} {
	if a == nil || (len(a.GetRequired()) == 0 && len(a.GetPreferred()) == 0) {
		return nil
	}
	out := map[string]interface{}{}
	if len(a.GetRequired()) > 0 {
		terms := []interface{}{}
		for _, t := range a.GetRequired() {
			terms = append(terms, podAffinityTerm(t))
		}
		out["requiredDuringSchedulingIgnoredDuringExecution"] = terms
	}
	if len(a.GetPreferred()) > 0 {
		terms := []interface{}{}
		for _, w := range a.GetPreferred() {
			terms = append(terms, map[string]interface{}{
				"weight":          int(w.GetWeight()),
				"podAffinityTerm": podAffinityTerm(w.GetTerm()),
			})
		}
		out["preferredDuringSchedulingIgnoredDuringExecution"] = terms
	}
	return out
}

func podAffinityTerm(t *kubernetesprovider.WorkloadPodAffinityTerm) map[string]interface{} {
	term := map[string]interface{}{
		"labelSelector": map[string]interface{}{"matchLabels": stringMap(t.GetMatchLabels())},
		"topologyKey":   t.GetTopologyKey(),
	}
	if len(t.GetNamespaces()) > 0 {
		term["namespaces"] = stringSlice(t.GetNamespaces())
	}
	return term
}

// podSecurityContextBody renders the shared pod security context.
func podSecurityContextBody(p *kubernetesprovider.WorkloadPodSecurityContext) map[string]interface{} {
	out := map[string]interface{}{}
	if p == nil {
		return out
	}
	if p.RunAsUser != nil {
		out["runAsUser"] = p.GetRunAsUser()
	}
	if p.RunAsGroup != nil {
		out["runAsGroup"] = p.GetRunAsGroup()
	}
	if p.RunAsNonRoot != nil {
		out["runAsNonRoot"] = p.GetRunAsNonRoot()
	}
	if p.FsGroup != nil {
		out["fsGroup"] = p.GetFsGroup()
	}
	setString(out, "fsGroupChangePolicy", p.GetFsGroupChangePolicy())
	if len(p.GetSupplementalGroups()) > 0 {
		groups := make([]interface{}, 0, len(p.GetSupplementalGroups()))
		for _, g := range p.GetSupplementalGroups() {
			groups = append(groups, g)
		}
		out["supplementalGroups"] = groups
	}
	if len(p.GetSysctls()) > 0 {
		sysctls := make([]interface{}, 0, len(p.GetSysctls()))
		for _, s := range p.GetSysctls() {
			sysctls = append(sysctls, map[string]interface{}{"name": s.GetName(), "value": s.GetValue()})
		}
		out["sysctls"] = sysctls
	}
	if sp := seccompBody(p.GetSeccompProfile()); sp != nil {
		out["seccompProfile"] = sp
	}
	return out
}

// containerSecurityContextBody renders the shared container security
// context.
func containerSecurityContextBody(c *kubernetesprovider.WorkloadContainerSecurityContext) map[string]interface{} {
	out := map[string]interface{}{}
	if c == nil {
		return out
	}
	setTrue(out, "privileged", c.GetPrivileged())
	if c.RunAsUser != nil {
		out["runAsUser"] = c.GetRunAsUser()
	}
	if c.RunAsGroup != nil {
		out["runAsGroup"] = c.GetRunAsGroup()
	}
	if c.RunAsNonRoot != nil {
		out["runAsNonRoot"] = c.GetRunAsNonRoot()
	}
	if c.ReadOnlyRootFilesystem != nil {
		out["readOnlyRootFilesystem"] = c.GetReadOnlyRootFilesystem()
	}
	if c.AllowPrivilegeEscalation != nil {
		out["allowPrivilegeEscalation"] = c.GetAllowPrivilegeEscalation()
	}
	if caps := c.GetCapabilities(); caps != nil && (len(caps.GetAdd()) > 0 || len(caps.GetDrop()) > 0) {
		capabilities := map[string]interface{}{}
		if len(caps.GetAdd()) > 0 {
			capabilities["add"] = stringSlice(caps.GetAdd())
		}
		if len(caps.GetDrop()) > 0 {
			capabilities["drop"] = stringSlice(caps.GetDrop())
		}
		out["capabilities"] = capabilities
	}
	if sp := seccompBody(c.GetSeccompProfile()); sp != nil {
		out["seccompProfile"] = sp
	}
	return out
}

func seccompBody(sp *kubernetesprovider.WorkloadSeccompProfile) map[string]interface{} {
	if sp == nil {
		return nil
	}
	out := map[string]interface{}{}
	setString(out, "type", sp.GetType())
	setString(out, "localhostProfile", sp.GetLocalhostProfile())
	if len(out) == 0 {
		return nil
	}
	return out
}
