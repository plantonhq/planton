---
title: "Preset: Rate Limiting for API Endpoints"
description: "Protect APIs from abuse and DDoS with per-IP rate limiting and ban escalation. Traffic under the throttle limit is allowed; exceeding it returns 429. Persistent abusers crossing the ban threshold are..."
type: "preset"
rank: "02"
presetSlug: "02-rate-limiting-api"
componentSlug: "cloud-armor-policy"
componentTitle: "Cloud Armor Policy"
provider: "gcp"
icon: "package"
order: 2
---

# Preset: Rate Limiting for API Endpoints

## Use Case

Protect APIs from abuse and DDoS with per-IP rate limiting and ban escalation. Traffic under the throttle limit is allowed; exceeding it returns 429. Persistent abusers crossing the ban threshold are fully blocked for a period.

## What This Creates

- A CLOUD_ARMOR policy with a single rate-based ban rule (priority 1000) and default allow (2147483647)
- Throttle threshold (100 req/60s): excess requests get `deny(429)`
- Ban threshold (500 req/300s): IPs exceeding this are banned for 3600 seconds (1 hour)

## Rate Limit vs. Ban

A single `rate_based_ban` rule uses two thresholds:

| Threshold | Purpose |
|-----------|---------|
| `rateLimitThreshold` | First line of defense. Requests beyond this get `exceedAction` (e.g., 429). Client can retry after the window. |
| `banThreshold` | Escalation for abuse. IPs exceeding this are fully blocked for `banDurationSec`. No requests processed during the ban. |

The throttle level catches normal over-limit traffic; the ban level catches sustained abuse.

## enforceOnKey Options

`enforceOnKey` defines how requests are grouped for counting:

| Value | Use Case |
|-------|----------|
| `IP` | Per source IP (default for this preset). |
| `ALL` | Single counter for all traffic. |
| `HTTP_HEADER` | Per value of a header; set `enforceOnKeyName` (e.g., `X-API-Key`). |
| `XFF_IP` | Per client IP from `X-Forwarded-For` (useful behind proxies). |
| `HTTP_PATH` | Per URL path. |
| `REGION_CODE` | Per client country/region. |

## Tuning Thresholds

| Parameter | Default | Tuning |
|-----------|---------|--------|
| `rateLimitThreshold` | 100/60s | Throttle level. Lower = stricter; raise for higher-traffic APIs. |
| `banThreshold` | 500/300s | Ban trigger. IPs exceeding this get banned. |
| `banDurationSec` | 3600 | 60–86400. Longer bans deter abuse but can lock out misconfigured clients. |
| `exceedAction` | `deny(429)` | `deny(403)`, `deny(404)`, or `redirect` to a custom page. |
