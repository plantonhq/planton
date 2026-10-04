# AwsRoute53ResolverFirewall — Operational Guide

Live-earned judgment lands here as proof runs and adopter operations teach it; the notes below are the forge-time seed.

## Fail-open is a VPC decision, not a rule-group decision

What happens when the firewall cannot evaluate a query (fail closed vs fail open) is configured per VPC, with the VPC's other resolver settings — deliberately not on this kind. A rule group associated to ten VPCs must not fight over one VPC's toggle.

## Priorities are the policy

Evaluation is ascending priority, first match wins — an ALLOW at 100 defeats a BLOCK at 200 for the same domain. Leave gaps (100, 200, 300) so rules can be inserted without renumbering, and remember association priority orders GROUPS the same way when a VPC carries several.

## Mutation protection blocks its own destroy

An association with mutation_protection ENABLED refuses deletion — including the declarative destroy of this very component. Enable it only on associations whose lifecycle is deliberately manual, and disable it before decommissioning.

## ALERT before BLOCK for threat rules

Advanced protection (DGA / DNS tunneling detection) is heuristic. Run new threat rules as ALERT with query logging until the hit pattern is understood, then flip to BLOCK — a HIGH confidence threshold on day one still surprises.

## Domain list pushes are not instant

List contents push through a separate update after create; a partially failed import surfaces as a retry error, never silent success. Wildcards: a bare domain matches itself and all subdomains; a `*.` prefix matches subdomains only.

## The managed lists need an ID lookup

AWS-managed lists (malware, aggregate threat) have account/region-specific IDs and no data source at the pin — resolve the ID once (`aws route53resolver list-firewall-domain-lists`) and pass it as the rule's external list ID.

## Every domain is stored dotted — the modules keep you honest

AWS stores firewall domains (list entries AND the OVERRIDE record's value) as trailing-dot FQDNs and echoes that form on read, and the upstream provider does not suppress the difference — a bare-authored `sinkhole.example.com` would re-plan forever (proven live). Both modules append the dot when it's absent, so write whichever form you like; the deployed form is always canonical.
