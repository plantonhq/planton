package module

// This file maps the Planton DestinationRule traffic-policy proto subtree onto the typed
// crd2pulumi SDK args.
//
// WHY THIS FILE LOOKS REPETITIVE (read before "simplifying"):
// The upstream CRD defines ONE `TrafficPolicy` (and one `PortTrafficPolicy`, one
// `LoadBalancerSettings`, etc.) and reuses each by reference at four reachable paths:
//   1. spec.trafficPolicy
//   2. spec.trafficPolicy.portLevelSettings[]
//   3. spec.subsets[].trafficPolicy
//   4. spec.subsets[].trafficPolicy.portLevelSettings[]
// crd2pulumi does NOT share a Go type across reference paths — it emits a distinct,
// path-named struct at every path (e.g. DestinationRuleSpecTrafficPolicyLoadBalancerArgs vs
// DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerArgs vs the two PortLevelSettings
// variants). These generated structs share no common settable interface, so the builders
// below CANNOT be collapsed into one generic function — doing so will not compile. The proto
// side stays DRY (one shared message per shape); only this adapter layer is duplicated, which
// is the price of compile-time-typed CRD args (the typed crd2pulumi resources are used
// deliberately instead of an untyped CustomResource, so field/structure errors are caught
// at compile time).
//
// To keep the duplication mechanical and low-risk, all leaf scalar mapping goes through the
// shared opt*/strArr/u32IntMap helpers, so each per-path builder is a single declarative
// struct literal. The four path families are prefixed: `tp` (spec.trafficPolicy), `pls`
// (spec...portLevelSettings), `sub` (subset.trafficPolicy), `subPls` (subset...portLevelSettings).

import (
	dr "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesdestinationrule/v1"
	istio "github.com/plantonhq/planton/pkg/kubernetes/kubernetestypes/istio/kubernetes/networking/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// --- shared scalar/collection helpers (path-independent) -----------------------------------
//
// Each returns a nil interface when the proto field is absent (proto3 `optional` pointer nil
// or empty collection), and pulumi omits nil input fields from the rendered CR — so unset
// fields fall through to istiod defaults.

func optStr(p *string) pulumi.StringPtrInput {
	if p == nil {
		return nil
	}
	return pulumi.String(*p)
}

func optBool(p *bool) pulumi.BoolPtrInput {
	if p == nil {
		return nil
	}
	return pulumi.Bool(*p)
}

func optFloat(p *float64) pulumi.Float64PtrInput {
	if p == nil {
		return nil
	}
	return pulumi.Float64(*p)
}

func optI32(p *int32) pulumi.IntPtrInput {
	if p == nil {
		return nil
	}
	return pulumi.Int(int(*p))
}

func optU32(p *uint32) pulumi.IntPtrInput {
	if p == nil {
		return nil
	}
	return pulumi.Int(int(*p))
}

func optU64(p *uint64) pulumi.IntPtrInput {
	if p == nil {
		return nil
	}
	return pulumi.Int(int(*p))
}

func strArr(s []string) pulumi.StringArrayInput {
	if len(s) == 0 {
		return nil
	}
	return pulumi.ToStringArray(s)
}

func u32IntMap(m map[string]uint32) pulumi.IntMapInput {
	if len(m) == 0 {
		return nil
	}
	out := pulumi.IntMap{}
	for k, v := range m {
		out[k] = pulumi.Int(int(v))
	}
	return out
}

// ===========================================================================================
// Family `tp`: spec.trafficPolicy (full TrafficPolicy)
// ===========================================================================================

func buildTrafficPolicy(p *dr.KubernetesDestinationRuleTrafficPolicy) istio.DestinationRuleSpecTrafficPolicyPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyArgs{
		LoadBalancer:      tpLoadBalancer(p.GetLoadBalancer()),
		ConnectionPool:    tpConnectionPool(p.GetConnectionPool()),
		OutlierDetection:  tpOutlierDetection(p.GetOutlierDetection()),
		Tls:               tpTls(p.GetTls()),
		PortLevelSettings: plsArray(p.GetPortLevelSettings()),
		Tunnel:            tpTunnel(p.GetTunnel()),
		ProxyProtocol:     tpProxyProtocol(p.GetProxyProtocol()),
	}
}

func tpLoadBalancer(p *dr.KubernetesDestinationRuleLoadBalancerSettings) istio.DestinationRuleSpecTrafficPolicyLoadBalancerPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyLoadBalancerArgs{
		Simple:             optStr(p.Simple),
		ConsistentHash:     tpConsistentHash(p.GetConsistentHash()),
		LocalityLbSetting:  tpLocality(p.GetLocalityLbSetting()),
		WarmupDurationSecs: optStr(p.WarmupDurationSecs),
		Warmup:             tpWarmup(p.GetWarmup()),
	}
}

func tpConsistentHash(p *dr.KubernetesDestinationRuleConsistentHashLb) istio.DestinationRuleSpecTrafficPolicyLoadBalancerConsistentHashPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyLoadBalancerConsistentHashArgs{
		HttpHeaderName:         optStr(p.HttpHeaderName),
		HttpCookie:             tpHttpCookie(p.GetHttpCookie()),
		UseSourceIp:            optBool(p.UseSourceIp),
		HttpQueryParameterName: optStr(p.HttpQueryParameterName),
		RingHash:               tpRingHash(p.GetRingHash()),
		Maglev:                 tpMaglev(p.GetMaglev()),
		MinimumRingSize:        optU64(p.MinimumRingSize),
	}
}

func tpHttpCookie(p *dr.KubernetesDestinationRuleHttpCookie) istio.DestinationRuleSpecTrafficPolicyLoadBalancerConsistentHashHttpCookiePtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyLoadBalancerConsistentHashHttpCookieArgs{
		Name: pulumi.String(p.GetName()),
		Path: optStr(p.Path),
		Ttl:  optStr(p.Ttl),
	}
}

func tpRingHash(p *dr.KubernetesDestinationRuleRingHash) istio.DestinationRuleSpecTrafficPolicyLoadBalancerConsistentHashRingHashPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyLoadBalancerConsistentHashRingHashArgs{
		MinimumRingSize: optU64(p.MinimumRingSize),
	}
}

func tpMaglev(p *dr.KubernetesDestinationRuleMagLev) istio.DestinationRuleSpecTrafficPolicyLoadBalancerConsistentHashMaglevPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyLoadBalancerConsistentHashMaglevArgs{
		TableSize: optU64(p.TableSize),
	}
}

func tpWarmup(p *dr.KubernetesDestinationRuleWarmupConfiguration) istio.DestinationRuleSpecTrafficPolicyLoadBalancerWarmupPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyLoadBalancerWarmupArgs{
		Duration:       pulumi.String(p.GetDuration()),
		MinimumPercent: optFloat(p.MinimumPercent),
		Aggression:     optFloat(p.Aggression),
	}
}

func tpLocality(p *dr.KubernetesDestinationRuleLocalityLbSetting) istio.DestinationRuleSpecTrafficPolicyLoadBalancerLocalityLbSettingPtrInput {
	if p == nil {
		return nil
	}
	args := istio.DestinationRuleSpecTrafficPolicyLoadBalancerLocalityLbSettingArgs{
		FailoverPriority: strArr(p.GetFailoverPriority()),
		Enabled:          optBool(p.Enabled),
	}
	if d := p.GetDistribute(); len(d) > 0 {
		arr := istio.DestinationRuleSpecTrafficPolicyLoadBalancerLocalityLbSettingDistributeArray{}
		for _, x := range d {
			arr = append(arr, istio.DestinationRuleSpecTrafficPolicyLoadBalancerLocalityLbSettingDistributeArgs{
				From: optStr(x.From),
				To:   u32IntMap(x.GetTo()),
			})
		}
		args.Distribute = arr
	}
	if f := p.GetFailover(); len(f) > 0 {
		arr := istio.DestinationRuleSpecTrafficPolicyLoadBalancerLocalityLbSettingFailoverArray{}
		for _, x := range f {
			arr = append(arr, istio.DestinationRuleSpecTrafficPolicyLoadBalancerLocalityLbSettingFailoverArgs{
				From: optStr(x.From),
				To:   optStr(x.To),
			})
		}
		args.Failover = arr
	}
	return args
}

func tpConnectionPool(p *dr.KubernetesDestinationRuleConnectionPoolSettings) istio.DestinationRuleSpecTrafficPolicyConnectionPoolPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyConnectionPoolArgs{
		Tcp:  tpTcp(p.GetTcp()),
		Http: tpHttp(p.GetHttp()),
	}
}

func tpTcp(p *dr.KubernetesDestinationRuleTcpSettings) istio.DestinationRuleSpecTrafficPolicyConnectionPoolTcpPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyConnectionPoolTcpArgs{
		MaxConnections:        optI32(p.MaxConnections),
		ConnectTimeout:        optStr(p.ConnectTimeout),
		TcpKeepalive:          tpKeepalive(p.GetTcpKeepalive()),
		MaxConnectionDuration: optStr(p.MaxConnectionDuration),
		IdleTimeout:           optStr(p.IdleTimeout),
	}
}

func tpKeepalive(p *dr.KubernetesDestinationRuleTcpKeepalive) istio.DestinationRuleSpecTrafficPolicyConnectionPoolTcpTcpKeepalivePtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyConnectionPoolTcpTcpKeepaliveArgs{
		Probes:   optU32(p.Probes),
		Time:     optStr(p.Time),
		Interval: optStr(p.Interval),
	}
}

func tpHttp(p *dr.KubernetesDestinationRuleHttpSettings) istio.DestinationRuleSpecTrafficPolicyConnectionPoolHttpPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyConnectionPoolHttpArgs{
		Http1MaxPendingRequests:  optI32(p.Http1MaxPendingRequests),
		Http2MaxRequests:         optI32(p.Http2MaxRequests),
		MaxRequestsPerConnection: optI32(p.MaxRequestsPerConnection),
		MaxRetries:               optI32(p.MaxRetries),
		IdleTimeout:              optStr(p.IdleTimeout),
		H2UpgradePolicy:          optStr(p.H2UpgradePolicy),
		UseClientProtocol:        optBool(p.UseClientProtocol),
		MaxConcurrentStreams:     optI32(p.MaxConcurrentStreams),
	}
}

func tpOutlierDetection(p *dr.KubernetesDestinationRuleOutlierDetection) istio.DestinationRuleSpecTrafficPolicyOutlierDetectionPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyOutlierDetectionArgs{
		SplitExternalLocalOriginErrors: optBool(p.SplitExternalLocalOriginErrors),
		ConsecutiveLocalOriginFailures: optU32(p.ConsecutiveLocalOriginFailures),
		ConsecutiveGatewayErrors:       optU32(p.ConsecutiveGatewayErrors),
		Consecutive5xxErrors:           optU32(p.Consecutive_5XxErrors),
		Interval:                       optStr(p.Interval),
		BaseEjectionTime:               optStr(p.BaseEjectionTime),
		MaxEjectionPercent:             optI32(p.MaxEjectionPercent),
		MinHealthPercent:               optI32(p.MinHealthPercent),
	}
}

func tpTls(p *dr.KubernetesDestinationRuleClientTlsSettings) istio.DestinationRuleSpecTrafficPolicyTlsPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyTlsArgs{
		Mode:               optStr(p.Mode),
		ClientCertificate:  optStr(p.ClientCertificate),
		PrivateKey:         optStr(p.PrivateKey),
		CaCertificates:     optStr(p.CaCertificates),
		CredentialName:     optStr(p.CredentialName),
		SubjectAltNames:    strArr(p.GetSubjectAltNames()),
		Sni:                optStr(p.Sni),
		InsecureSkipVerify: optBool(p.InsecureSkipVerify),
		CaCrl:              optStr(p.CaCrl),
	}
}

func tpTunnel(p *dr.KubernetesDestinationRuleTunnelSettings) istio.DestinationRuleSpecTrafficPolicyTunnelPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyTunnelArgs{
		Protocol:   optStr(p.Protocol),
		TargetHost: pulumi.String(p.GetTargetHost()),
		TargetPort: pulumi.Int(int(p.GetTargetPort())),
	}
}

func tpProxyProtocol(p *dr.KubernetesDestinationRuleProxyProtocol) istio.DestinationRuleSpecTrafficPolicyProxyProtocolPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyProxyProtocolArgs{
		Version: optStr(p.Version),
	}
}

// ===========================================================================================
// Family `pls`: spec.trafficPolicy.portLevelSettings[] (PortTrafficPolicy)
// ===========================================================================================

func plsArray(items []*dr.KubernetesDestinationRulePortTrafficPolicy) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsArrayInput {
	if len(items) == 0 {
		return nil
	}
	out := istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsArray{}
	for _, p := range items {
		args := istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsArgs{
			LoadBalancer:     plsLoadBalancer(p.GetLoadBalancer()),
			ConnectionPool:   plsConnectionPool(p.GetConnectionPool()),
			OutlierDetection: plsOutlierDetection(p.GetOutlierDetection()),
			Tls:              plsTls(p.GetTls()),
		}
		if sel := p.GetPort(); sel != nil {
			args.Port = istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsPortArgs{
				Number: pulumi.Int(int(sel.GetNumber())),
			}
		}
		out = append(out, args)
	}
	return out
}

func plsLoadBalancer(p *dr.KubernetesDestinationRuleLoadBalancerSettings) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerArgs{
		Simple:             optStr(p.Simple),
		ConsistentHash:     plsConsistentHash(p.GetConsistentHash()),
		LocalityLbSetting:  plsLocality(p.GetLocalityLbSetting()),
		WarmupDurationSecs: optStr(p.WarmupDurationSecs),
		Warmup:             plsWarmup(p.GetWarmup()),
	}
}

func plsConsistentHash(p *dr.KubernetesDestinationRuleConsistentHashLb) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashArgs{
		HttpHeaderName:         optStr(p.HttpHeaderName),
		HttpCookie:             plsHttpCookie(p.GetHttpCookie()),
		UseSourceIp:            optBool(p.UseSourceIp),
		HttpQueryParameterName: optStr(p.HttpQueryParameterName),
		RingHash:               plsRingHash(p.GetRingHash()),
		Maglev:                 plsMaglev(p.GetMaglev()),
		MinimumRingSize:        optU64(p.MinimumRingSize),
	}
}

func plsHttpCookie(p *dr.KubernetesDestinationRuleHttpCookie) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashHttpCookiePtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashHttpCookieArgs{
		Name: pulumi.String(p.GetName()),
		Path: optStr(p.Path),
		Ttl:  optStr(p.Ttl),
	}
}

func plsRingHash(p *dr.KubernetesDestinationRuleRingHash) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashRingHashPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashRingHashArgs{
		MinimumRingSize: optU64(p.MinimumRingSize),
	}
}

func plsMaglev(p *dr.KubernetesDestinationRuleMagLev) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashMaglevPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashMaglevArgs{
		TableSize: optU64(p.TableSize),
	}
}

func plsWarmup(p *dr.KubernetesDestinationRuleWarmupConfiguration) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerWarmupPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerWarmupArgs{
		Duration:       pulumi.String(p.GetDuration()),
		MinimumPercent: optFloat(p.MinimumPercent),
		Aggression:     optFloat(p.Aggression),
	}
}

func plsLocality(p *dr.KubernetesDestinationRuleLocalityLbSetting) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingPtrInput {
	if p == nil {
		return nil
	}
	args := istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingArgs{
		FailoverPriority: strArr(p.GetFailoverPriority()),
		Enabled:          optBool(p.Enabled),
	}
	if d := p.GetDistribute(); len(d) > 0 {
		arr := istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingDistributeArray{}
		for _, x := range d {
			arr = append(arr, istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingDistributeArgs{
				From: optStr(x.From),
				To:   u32IntMap(x.GetTo()),
			})
		}
		args.Distribute = arr
	}
	if f := p.GetFailover(); len(f) > 0 {
		arr := istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingFailoverArray{}
		for _, x := range f {
			arr = append(arr, istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingFailoverArgs{
				From: optStr(x.From),
				To:   optStr(x.To),
			})
		}
		args.Failover = arr
	}
	return args
}

func plsConnectionPool(p *dr.KubernetesDestinationRuleConnectionPoolSettings) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsConnectionPoolPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsConnectionPoolArgs{
		Tcp:  plsTcp(p.GetTcp()),
		Http: plsHttp(p.GetHttp()),
	}
}

func plsTcp(p *dr.KubernetesDestinationRuleTcpSettings) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsConnectionPoolTcpPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsConnectionPoolTcpArgs{
		MaxConnections:        optI32(p.MaxConnections),
		ConnectTimeout:        optStr(p.ConnectTimeout),
		TcpKeepalive:          plsKeepalive(p.GetTcpKeepalive()),
		MaxConnectionDuration: optStr(p.MaxConnectionDuration),
		IdleTimeout:           optStr(p.IdleTimeout),
	}
}

func plsKeepalive(p *dr.KubernetesDestinationRuleTcpKeepalive) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsConnectionPoolTcpTcpKeepalivePtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsConnectionPoolTcpTcpKeepaliveArgs{
		Probes:   optU32(p.Probes),
		Time:     optStr(p.Time),
		Interval: optStr(p.Interval),
	}
}

func plsHttp(p *dr.KubernetesDestinationRuleHttpSettings) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsConnectionPoolHttpPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsConnectionPoolHttpArgs{
		Http1MaxPendingRequests:  optI32(p.Http1MaxPendingRequests),
		Http2MaxRequests:         optI32(p.Http2MaxRequests),
		MaxRequestsPerConnection: optI32(p.MaxRequestsPerConnection),
		MaxRetries:               optI32(p.MaxRetries),
		IdleTimeout:              optStr(p.IdleTimeout),
		H2UpgradePolicy:          optStr(p.H2UpgradePolicy),
		UseClientProtocol:        optBool(p.UseClientProtocol),
		MaxConcurrentStreams:     optI32(p.MaxConcurrentStreams),
	}
}

func plsOutlierDetection(p *dr.KubernetesDestinationRuleOutlierDetection) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsOutlierDetectionPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsOutlierDetectionArgs{
		SplitExternalLocalOriginErrors: optBool(p.SplitExternalLocalOriginErrors),
		ConsecutiveLocalOriginFailures: optU32(p.ConsecutiveLocalOriginFailures),
		ConsecutiveGatewayErrors:       optU32(p.ConsecutiveGatewayErrors),
		Consecutive5xxErrors:           optU32(p.Consecutive_5XxErrors),
		Interval:                       optStr(p.Interval),
		BaseEjectionTime:               optStr(p.BaseEjectionTime),
		MaxEjectionPercent:             optI32(p.MaxEjectionPercent),
		MinHealthPercent:               optI32(p.MinHealthPercent),
	}
}

func plsTls(p *dr.KubernetesDestinationRuleClientTlsSettings) istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsTlsPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecTrafficPolicyPortLevelSettingsTlsArgs{
		Mode:               optStr(p.Mode),
		ClientCertificate:  optStr(p.ClientCertificate),
		PrivateKey:         optStr(p.PrivateKey),
		CaCertificates:     optStr(p.CaCertificates),
		CredentialName:     optStr(p.CredentialName),
		SubjectAltNames:    strArr(p.GetSubjectAltNames()),
		Sni:                optStr(p.Sni),
		InsecureSkipVerify: optBool(p.InsecureSkipVerify),
		CaCrl:              optStr(p.CaCrl),
	}
}

// ===========================================================================================
// Family `sub`: spec.subsets[].trafficPolicy (full TrafficPolicy)
// ===========================================================================================

func buildSubsetTrafficPolicy(p *dr.KubernetesDestinationRuleTrafficPolicy) istio.DestinationRuleSpecSubsetsTrafficPolicyPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyArgs{
		LoadBalancer:      subLoadBalancer(p.GetLoadBalancer()),
		ConnectionPool:    subConnectionPool(p.GetConnectionPool()),
		OutlierDetection:  subOutlierDetection(p.GetOutlierDetection()),
		Tls:               subTls(p.GetTls()),
		PortLevelSettings: subPlsArray(p.GetPortLevelSettings()),
		Tunnel:            subTunnel(p.GetTunnel()),
		ProxyProtocol:     subProxyProtocol(p.GetProxyProtocol()),
	}
}

func subLoadBalancer(p *dr.KubernetesDestinationRuleLoadBalancerSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerArgs{
		Simple:             optStr(p.Simple),
		ConsistentHash:     subConsistentHash(p.GetConsistentHash()),
		LocalityLbSetting:  subLocality(p.GetLocalityLbSetting()),
		WarmupDurationSecs: optStr(p.WarmupDurationSecs),
		Warmup:             subWarmup(p.GetWarmup()),
	}
}

func subConsistentHash(p *dr.KubernetesDestinationRuleConsistentHashLb) istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerConsistentHashPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerConsistentHashArgs{
		HttpHeaderName:         optStr(p.HttpHeaderName),
		HttpCookie:             subHttpCookie(p.GetHttpCookie()),
		UseSourceIp:            optBool(p.UseSourceIp),
		HttpQueryParameterName: optStr(p.HttpQueryParameterName),
		RingHash:               subRingHash(p.GetRingHash()),
		Maglev:                 subMaglev(p.GetMaglev()),
		MinimumRingSize:        optU64(p.MinimumRingSize),
	}
}

func subHttpCookie(p *dr.KubernetesDestinationRuleHttpCookie) istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerConsistentHashHttpCookiePtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerConsistentHashHttpCookieArgs{
		Name: pulumi.String(p.GetName()),
		Path: optStr(p.Path),
		Ttl:  optStr(p.Ttl),
	}
}

func subRingHash(p *dr.KubernetesDestinationRuleRingHash) istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerConsistentHashRingHashPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerConsistentHashRingHashArgs{
		MinimumRingSize: optU64(p.MinimumRingSize),
	}
}

func subMaglev(p *dr.KubernetesDestinationRuleMagLev) istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerConsistentHashMaglevPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerConsistentHashMaglevArgs{
		TableSize: optU64(p.TableSize),
	}
}

func subWarmup(p *dr.KubernetesDestinationRuleWarmupConfiguration) istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerWarmupPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerWarmupArgs{
		Duration:       pulumi.String(p.GetDuration()),
		MinimumPercent: optFloat(p.MinimumPercent),
		Aggression:     optFloat(p.Aggression),
	}
}

func subLocality(p *dr.KubernetesDestinationRuleLocalityLbSetting) istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerLocalityLbSettingPtrInput {
	if p == nil {
		return nil
	}
	args := istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerLocalityLbSettingArgs{
		FailoverPriority: strArr(p.GetFailoverPriority()),
		Enabled:          optBool(p.Enabled),
	}
	if d := p.GetDistribute(); len(d) > 0 {
		arr := istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerLocalityLbSettingDistributeArray{}
		for _, x := range d {
			arr = append(arr, istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerLocalityLbSettingDistributeArgs{
				From: optStr(x.From),
				To:   u32IntMap(x.GetTo()),
			})
		}
		args.Distribute = arr
	}
	if f := p.GetFailover(); len(f) > 0 {
		arr := istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerLocalityLbSettingFailoverArray{}
		for _, x := range f {
			arr = append(arr, istio.DestinationRuleSpecSubsetsTrafficPolicyLoadBalancerLocalityLbSettingFailoverArgs{
				From: optStr(x.From),
				To:   optStr(x.To),
			})
		}
		args.Failover = arr
	}
	return args
}

func subConnectionPool(p *dr.KubernetesDestinationRuleConnectionPoolSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyConnectionPoolPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyConnectionPoolArgs{
		Tcp:  subTcp(p.GetTcp()),
		Http: subHttp(p.GetHttp()),
	}
}

func subTcp(p *dr.KubernetesDestinationRuleTcpSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyConnectionPoolTcpPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyConnectionPoolTcpArgs{
		MaxConnections:        optI32(p.MaxConnections),
		ConnectTimeout:        optStr(p.ConnectTimeout),
		TcpKeepalive:          subKeepalive(p.GetTcpKeepalive()),
		MaxConnectionDuration: optStr(p.MaxConnectionDuration),
		IdleTimeout:           optStr(p.IdleTimeout),
	}
}

func subKeepalive(p *dr.KubernetesDestinationRuleTcpKeepalive) istio.DestinationRuleSpecSubsetsTrafficPolicyConnectionPoolTcpTcpKeepalivePtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyConnectionPoolTcpTcpKeepaliveArgs{
		Probes:   optU32(p.Probes),
		Time:     optStr(p.Time),
		Interval: optStr(p.Interval),
	}
}

func subHttp(p *dr.KubernetesDestinationRuleHttpSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyConnectionPoolHttpPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyConnectionPoolHttpArgs{
		Http1MaxPendingRequests:  optI32(p.Http1MaxPendingRequests),
		Http2MaxRequests:         optI32(p.Http2MaxRequests),
		MaxRequestsPerConnection: optI32(p.MaxRequestsPerConnection),
		MaxRetries:               optI32(p.MaxRetries),
		IdleTimeout:              optStr(p.IdleTimeout),
		H2UpgradePolicy:          optStr(p.H2UpgradePolicy),
		UseClientProtocol:        optBool(p.UseClientProtocol),
		MaxConcurrentStreams:     optI32(p.MaxConcurrentStreams),
	}
}

func subOutlierDetection(p *dr.KubernetesDestinationRuleOutlierDetection) istio.DestinationRuleSpecSubsetsTrafficPolicyOutlierDetectionPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyOutlierDetectionArgs{
		SplitExternalLocalOriginErrors: optBool(p.SplitExternalLocalOriginErrors),
		ConsecutiveLocalOriginFailures: optU32(p.ConsecutiveLocalOriginFailures),
		ConsecutiveGatewayErrors:       optU32(p.ConsecutiveGatewayErrors),
		Consecutive5xxErrors:           optU32(p.Consecutive_5XxErrors),
		Interval:                       optStr(p.Interval),
		BaseEjectionTime:               optStr(p.BaseEjectionTime),
		MaxEjectionPercent:             optI32(p.MaxEjectionPercent),
		MinHealthPercent:               optI32(p.MinHealthPercent),
	}
}

func subTls(p *dr.KubernetesDestinationRuleClientTlsSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyTlsPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyTlsArgs{
		Mode:               optStr(p.Mode),
		ClientCertificate:  optStr(p.ClientCertificate),
		PrivateKey:         optStr(p.PrivateKey),
		CaCertificates:     optStr(p.CaCertificates),
		CredentialName:     optStr(p.CredentialName),
		SubjectAltNames:    strArr(p.GetSubjectAltNames()),
		Sni:                optStr(p.Sni),
		InsecureSkipVerify: optBool(p.InsecureSkipVerify),
		CaCrl:              optStr(p.CaCrl),
	}
}

func subTunnel(p *dr.KubernetesDestinationRuleTunnelSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyTunnelPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyTunnelArgs{
		Protocol:   optStr(p.Protocol),
		TargetHost: pulumi.String(p.GetTargetHost()),
		TargetPort: pulumi.Int(int(p.GetTargetPort())),
	}
}

func subProxyProtocol(p *dr.KubernetesDestinationRuleProxyProtocol) istio.DestinationRuleSpecSubsetsTrafficPolicyProxyProtocolPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyProxyProtocolArgs{
		Version: optStr(p.Version),
	}
}

// ===========================================================================================
// Family `subPls`: spec.subsets[].trafficPolicy.portLevelSettings[] (PortTrafficPolicy)
// ===========================================================================================

func subPlsArray(items []*dr.KubernetesDestinationRulePortTrafficPolicy) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsArrayInput {
	if len(items) == 0 {
		return nil
	}
	out := istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsArray{}
	for _, p := range items {
		args := istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsArgs{
			LoadBalancer:     subPlsLoadBalancer(p.GetLoadBalancer()),
			ConnectionPool:   subPlsConnectionPool(p.GetConnectionPool()),
			OutlierDetection: subPlsOutlierDetection(p.GetOutlierDetection()),
			Tls:              subPlsTls(p.GetTls()),
		}
		if sel := p.GetPort(); sel != nil {
			args.Port = istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsPortArgs{
				Number: pulumi.Int(int(sel.GetNumber())),
			}
		}
		out = append(out, args)
	}
	return out
}

func subPlsLoadBalancer(p *dr.KubernetesDestinationRuleLoadBalancerSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerArgs{
		Simple:             optStr(p.Simple),
		ConsistentHash:     subPlsConsistentHash(p.GetConsistentHash()),
		LocalityLbSetting:  subPlsLocality(p.GetLocalityLbSetting()),
		WarmupDurationSecs: optStr(p.WarmupDurationSecs),
		Warmup:             subPlsWarmup(p.GetWarmup()),
	}
}

func subPlsConsistentHash(p *dr.KubernetesDestinationRuleConsistentHashLb) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashArgs{
		HttpHeaderName:         optStr(p.HttpHeaderName),
		HttpCookie:             subPlsHttpCookie(p.GetHttpCookie()),
		UseSourceIp:            optBool(p.UseSourceIp),
		HttpQueryParameterName: optStr(p.HttpQueryParameterName),
		RingHash:               subPlsRingHash(p.GetRingHash()),
		Maglev:                 subPlsMaglev(p.GetMaglev()),
		MinimumRingSize:        optU64(p.MinimumRingSize),
	}
}

func subPlsHttpCookie(p *dr.KubernetesDestinationRuleHttpCookie) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashHttpCookiePtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashHttpCookieArgs{
		Name: pulumi.String(p.GetName()),
		Path: optStr(p.Path),
		Ttl:  optStr(p.Ttl),
	}
}

func subPlsRingHash(p *dr.KubernetesDestinationRuleRingHash) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashRingHashPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashRingHashArgs{
		MinimumRingSize: optU64(p.MinimumRingSize),
	}
}

func subPlsMaglev(p *dr.KubernetesDestinationRuleMagLev) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashMaglevPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerConsistentHashMaglevArgs{
		TableSize: optU64(p.TableSize),
	}
}

func subPlsWarmup(p *dr.KubernetesDestinationRuleWarmupConfiguration) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerWarmupPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerWarmupArgs{
		Duration:       pulumi.String(p.GetDuration()),
		MinimumPercent: optFloat(p.MinimumPercent),
		Aggression:     optFloat(p.Aggression),
	}
}

func subPlsLocality(p *dr.KubernetesDestinationRuleLocalityLbSetting) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingPtrInput {
	if p == nil {
		return nil
	}
	args := istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingArgs{
		FailoverPriority: strArr(p.GetFailoverPriority()),
		Enabled:          optBool(p.Enabled),
	}
	if d := p.GetDistribute(); len(d) > 0 {
		arr := istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingDistributeArray{}
		for _, x := range d {
			arr = append(arr, istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingDistributeArgs{
				From: optStr(x.From),
				To:   u32IntMap(x.GetTo()),
			})
		}
		args.Distribute = arr
	}
	if f := p.GetFailover(); len(f) > 0 {
		arr := istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingFailoverArray{}
		for _, x := range f {
			arr = append(arr, istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsLoadBalancerLocalityLbSettingFailoverArgs{
				From: optStr(x.From),
				To:   optStr(x.To),
			})
		}
		args.Failover = arr
	}
	return args
}

func subPlsConnectionPool(p *dr.KubernetesDestinationRuleConnectionPoolSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsConnectionPoolPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsConnectionPoolArgs{
		Tcp:  subPlsTcp(p.GetTcp()),
		Http: subPlsHttp(p.GetHttp()),
	}
}

func subPlsTcp(p *dr.KubernetesDestinationRuleTcpSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsConnectionPoolTcpPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsConnectionPoolTcpArgs{
		MaxConnections:        optI32(p.MaxConnections),
		ConnectTimeout:        optStr(p.ConnectTimeout),
		TcpKeepalive:          subPlsKeepalive(p.GetTcpKeepalive()),
		MaxConnectionDuration: optStr(p.MaxConnectionDuration),
		IdleTimeout:           optStr(p.IdleTimeout),
	}
}

func subPlsKeepalive(p *dr.KubernetesDestinationRuleTcpKeepalive) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsConnectionPoolTcpTcpKeepalivePtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsConnectionPoolTcpTcpKeepaliveArgs{
		Probes:   optU32(p.Probes),
		Time:     optStr(p.Time),
		Interval: optStr(p.Interval),
	}
}

func subPlsHttp(p *dr.KubernetesDestinationRuleHttpSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsConnectionPoolHttpPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsConnectionPoolHttpArgs{
		Http1MaxPendingRequests:  optI32(p.Http1MaxPendingRequests),
		Http2MaxRequests:         optI32(p.Http2MaxRequests),
		MaxRequestsPerConnection: optI32(p.MaxRequestsPerConnection),
		MaxRetries:               optI32(p.MaxRetries),
		IdleTimeout:              optStr(p.IdleTimeout),
		H2UpgradePolicy:          optStr(p.H2UpgradePolicy),
		UseClientProtocol:        optBool(p.UseClientProtocol),
		MaxConcurrentStreams:     optI32(p.MaxConcurrentStreams),
	}
}

func subPlsOutlierDetection(p *dr.KubernetesDestinationRuleOutlierDetection) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsOutlierDetectionPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsOutlierDetectionArgs{
		SplitExternalLocalOriginErrors: optBool(p.SplitExternalLocalOriginErrors),
		ConsecutiveLocalOriginFailures: optU32(p.ConsecutiveLocalOriginFailures),
		ConsecutiveGatewayErrors:       optU32(p.ConsecutiveGatewayErrors),
		Consecutive5xxErrors:           optU32(p.Consecutive_5XxErrors),
		Interval:                       optStr(p.Interval),
		BaseEjectionTime:               optStr(p.BaseEjectionTime),
		MaxEjectionPercent:             optI32(p.MaxEjectionPercent),
		MinHealthPercent:               optI32(p.MinHealthPercent),
	}
}

func subPlsTls(p *dr.KubernetesDestinationRuleClientTlsSettings) istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsTlsPtrInput {
	if p == nil {
		return nil
	}
	return istio.DestinationRuleSpecSubsetsTrafficPolicyPortLevelSettingsTlsArgs{
		Mode:               optStr(p.Mode),
		ClientCertificate:  optStr(p.ClientCertificate),
		PrivateKey:         optStr(p.PrivateKey),
		CaCertificates:     optStr(p.CaCertificates),
		CredentialName:     optStr(p.CredentialName),
		SubjectAltNames:    strArr(p.GetSubjectAltNames()),
		Sni:                optStr(p.Sni),
		InsecureSkipVerify: optBool(p.InsecureSkipVerify),
		CaCrl:              optStr(p.CaCrl),
	}
}
