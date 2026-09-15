package module

import (
	"github.com/pkg/errors"
	cloudflarezonesettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezonesettings/v1alpha1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// settingEntry pairs a Cloudflare setting_id with the value the module sends
// to PATCH /zones/{zone_id}/settings/{setting_id}.
type settingEntry struct {
	id    string
	value interface{}
}

// onOff maps the spec's managed booleans to the settings API's on/off strings.
func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// collectSettings walks the spec's typed fan-out and returns one entry per
// MANAGED setting. A nil (unset) field produces no entry -- the module never
// sends defaults, because zone settings have no delete: anything sent once is
// owned until reverted explicitly.
func collectSettings(spec *cloudflarezonesettingsv1alpha1.CloudflareZoneSettingsSpec) []settingEntry {
	entries := make([]settingEntry, 0)

	addBool := func(id string, v *bool) {
		if v != nil {
			entries = append(entries, settingEntry{id: id, value: onOff(*v)})
		}
	}
	addString := func(id string, v *string) {
		if v != nil {
			entries = append(entries, settingEntry{id: id, value: *v})
		}
	}
	addInt := func(id string, v *int64) {
		if v != nil {
			entries = append(entries, settingEntry{id: id, value: *v})
		}
	}

	// On/off toggles (the API represents these as "on"/"off" strings).
	addBool("0rtt", spec.ZeroRtt)
	addBool("advanced_ddos", spec.AdvancedDdos)
	addBool("always_online", spec.AlwaysOnline)
	addBool("always_use_https", spec.AlwaysUseHttps)
	addBool("automatic_https_rewrites", spec.AutomaticHttpsRewrites)
	addBool("brotli", spec.Brotli)
	addBool("browser_check", spec.BrowserCheck)
	addBool("content_converter", spec.ContentConverter)
	addBool("development_mode", spec.DevelopmentMode)
	addBool("early_hints", spec.EarlyHints)
	addBool("email_obfuscation", spec.EmailObfuscation)
	addBool("hotlink_protection", spec.HotlinkProtection)
	addBool("http2", spec.Http2)
	addBool("http3", spec.Http3)
	addBool("ip_geolocation", spec.IpGeolocation)
	addBool("ipv6", spec.Ipv6)
	addBool("long_lived_grpc", spec.LongLivedGrpc)
	addBool("mirage", spec.Mirage)
	addBool("opportunistic_encryption", spec.OpportunisticEncryption)
	addBool("opportunistic_onion", spec.OpportunisticOnion)
	addBool("orange_to_orange", spec.OrangeToOrange)
	addBool("origin_error_page_pass_thru", spec.OriginErrorPagePassThru)
	addBool("prefetch_preload", spec.PrefetchPreload)
	addBool("privacy_pass", spec.PrivacyPass)
	addBool("redirects_for_ai_training", spec.RedirectsForAiTraining)
	addBool("replace_insecure_js", spec.ReplaceInsecureJs)
	addBool("response_buffering", spec.ResponseBuffering)
	addBool("rocket_loader", spec.RocketLoader)
	addBool("search_for_agents", spec.SearchForAgents)
	addBool("server_side_exclude", spec.ServerSideExclude)
	addBool("sha1_support", spec.Sha1Support)
	addBool("sort_query_string_for_cache", spec.SortQueryStringForCache)
	// ssl_recommender: the provider's schema requires the on/off value form on
	// writes (its documented enabled-attribute form fails validation at v5.23.0).
	addBool("ssl_recommender", spec.SslRecommender)
	addBool("tls_1_2_only", spec.Tls_1_2Only)
	addBool("tls_client_auth", spec.TlsClientAuth)
	addBool("true_client_ip_header", spec.TrueClientIpHeader)
	addBool("waf", spec.Waf)
	addBool("webp", spec.Webp)
	addBool("websockets", spec.Websockets)

	// Enum and free-string settings.
	addString("cache_level", spec.CacheLevel)
	addString("cname_flattening", spec.CnameFlattening)
	addString("h2_prioritization", spec.H2Prioritization)
	addString("image_resizing", spec.ImageResizing)
	addString("min_tls_version", spec.MinTlsVersion)
	addString("origin_max_http_version", spec.OriginMaxHttpVersion)
	addString("polish", spec.Polish)
	addString("pseudo_ipv4", spec.PseudoIpv4)
	addString("security_level", spec.SecurityLevel)
	addString("ssl", spec.Ssl)
	addString("tls_1_3", spec.Tls_1_3)
	addString("transformations", spec.Transformations)
	addString("transformations_allowed_origins", spec.TransformationsAllowedOrigins)

	// Numeric settings (value sets validated by the API).
	addInt("browser_cache_ttl", spec.BrowserCacheTtl)
	addInt("challenge_ttl", spec.ChallengeTtl)
	addInt("edge_cache_ttl", spec.EdgeCacheTtl)
	addInt("max_upload", spec.MaxUpload)
	addInt("origin_h2_max_streams", spec.OriginH2MaxStreams)
	addInt("proxy_read_timeout", spec.ProxyReadTimeout)

	// List-valued setting.
	if len(spec.Ciphers) > 0 {
		entries = append(entries, settingEntry{id: "ciphers", value: spec.Ciphers})
	}

	// Object-valued settings. Shapes mirror the settings API's value objects.
	if spec.SecurityHeader != nil {
		entries = append(entries, settingEntry{id: "security_header", value: map[string]interface{}{
			"strict_transport_security": map[string]interface{}{
				"enabled":            spec.SecurityHeader.GetEnabled(),
				"include_subdomains": spec.SecurityHeader.IncludeSubdomains,
				"max_age":            spec.SecurityHeader.MaxAge,
				"nosniff":            spec.SecurityHeader.Nosniff,
				"preload":            spec.SecurityHeader.Preload,
			},
		}})
	}
	if spec.Nel != nil {
		entries = append(entries, settingEntry{id: "nel", value: map[string]interface{}{
			"enabled": spec.Nel.GetEnabled(),
		}})
	}
	if spec.Aegis != nil {
		aegisValue := map[string]interface{}{}
		if spec.Aegis.Enabled != nil {
			aegisValue["enabled"] = *spec.Aegis.Enabled
		}
		if spec.Aegis.PoolId != "" {
			aegisValue["pool_id"] = spec.Aegis.PoolId
		}
		entries = append(entries, settingEntry{id: "aegis", value: aegisValue})
	}
	if spec.AutomaticPlatformOptimization != nil {
		apo := spec.AutomaticPlatformOptimization
		// The APO API requires every member of the value object on writes.
		entries = append(entries, settingEntry{id: "automatic_platform_optimization", value: map[string]interface{}{
			"enabled":              apo.GetEnabled(),
			"cache_by_device_type": apo.CacheByDeviceType,
			"cf":                   apo.Cf,
			"hostnames":            apo.Hostnames,
			"wordpress":            apo.Wordpress,
			"wp_plugin":            apo.WpPlugin,
		}})
	}

	return entries
}

// zoneSettings emits one cloudflare_zone_setting resource per managed setting.
// Resource names embed the setting_id so a spec edit never renames an
// unrelated setting's resource.
func zoneSettings(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareZoneSettings.Spec
	zoneId := spec.ZoneId.GetValue()

	for _, entry := range collectSettings(spec) {
		if _, err := cloudflare.NewZoneSetting(
			ctx,
			"zone_setting_"+entry.id,
			&cloudflare.ZoneSettingArgs{
				ZoneId:    pulumi.String(zoneId),
				SettingId: pulumi.String(entry.id),
				Value:     pulumi.Any(entry.value),
			},
			pulumi.Provider(cloudflareProvider),
		); err != nil {
			return errors.Wrapf(err, "failed to create zone setting %q", entry.id)
		}
	}

	return nil
}
